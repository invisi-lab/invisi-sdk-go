package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
	"math/big"
)

func (s *WalletSession) checkKnownState(ctx context.Context, v *WalletState, b protocolrpc.CanonicalBlock) error {
	c, e := contracts.NewInvisiProtocolCaller(s.scope.ProtocolAddress, s.options.Provider)
	if e != nil {
		return e
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(b.Number)}
	for key, n := range v.Notes {
		n.Spent, e = c.IsNullifierSpent(opts, n.Nullifier)
		if e != nil {
			return e
		}
		checked := b
		n.CheckedThrough = &checked
		v.Notes[key] = n
	}
	bitmaps := map[string]*big.Int{}
	for key, g := range v.PostpaidCredit.Grants {
		g.Consumed, e = c.IsNullifierSpent(opts, g.Nullifier)
		if e != nil {
			return e
		}
		state := g.Delivery.State
		owner := common.BigToAddress(primitives.ElementToBigInt(state.OwnerWalletIdentity))
		token := common.BigToAddress(primitives.ElementToBigInt(state.TokenAddress))
		shard := primitives.ElementToBigInt(state.RevocationShard).Uint64()
		slot := primitives.ElementToBigInt(state.RevocationSlot).Uint64()
		k := fmt.Sprintf("%s:%s:%d:%d", owner, token, shard, slot/256)
		bits := bitmaps[k]
		if bits == nil {
			bits, e = c.CreditRevocationBitmap(opts, owner, token, shard, uint8(slot/256))
			if e != nil {
				return e
			}
			if bits == nil || bits.Sign() < 0 || bits.BitLen() > 256 {
				return errors.New("invalid revocation bitmap")
			}
			bitmaps[k] = bits
		}
		g.Revoked = bits.Bit(int(slot%256)) == 1
		v.PostpaidCredit.Grants[key] = g
	}
	if e = s.canonical(ctx, b); e != nil {
		return e
	}
	v.CheckedThrough = &b
	v.Scan.ChainHead = &b
	return nil
}
func (s *WalletSession) RefreshState(ctx context.Context) error {
	head, e := s.options.Provider.HeaderByNumber(ctx, nil)
	if e != nil {
		return e
	}
	if head == nil || head.Number == nil {
		return errors.New("invalid chain head")
	}
	b := protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	public, e := s.publicHistory.Snapshot(ctx)
	if e != nil {
		return e
	}
	if e = s.publicHistory.checkCanonical(ctx, public); e != nil {
		_ = s.invalidate(ctx)
		return e
	}
	before, e := s.fencedSnapshot(ctx)
	if e != nil {
		return e
	}
	if before.Invalidated {
		return errors.New("wallet requires reset scan")
	}
	if e = s.checkKnownState(ctx, &before, b); e != nil {
		return e
	}
	if e = s.commitPrivateBatch(ctx, before.Scan.Generation, before.Revision, before); e != nil {
		return e
	}
	snapshot, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return e
	}
	for _, r := range snapshot.Operations {
		if !r.Externalized || r.TransactionHash == (common.Hash{}) || r.Status == "confirmed" || r.Status == "cancelled" || r.Status == "reverted" {
			continue
		}
		var plan transaction.Plan
		if e = json.Unmarshal(r.Plan, &plan); e != nil {
			return e
		}
		o := &Operation{session: s, plan: plan, status: r.Status, data: append([]byte(nil), r.CallData...), transactionHash: r.TransactionHash, workerID: r.WorkerID}
		if _, e = o.reconcileInternal(ctx); e != nil {
			return e
		}
	}
	credit, e := s.Credit()
	if e != nil {
		return e
	}
	return credit.ReconcilePending(ctx, s.CreditReconciler(1))
}

func (s *WalletSession) replayCredits(ctx context.Context, target protocolrpc.CanonicalBlock, size uint64, o ScanOptions, result *RefreshResult) error {
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return e
	}
	if state.Scan.CreditReplayTarget == 0 {
		return nil
	}
	if target.Number < state.Scan.CreditReplayTarget {
		return nil
	}
	from := s.protocol.Deployment().Protocol.FromBlock
	if state.Scan.CreditReplayCursor >= from {
		from = state.Scan.CreditReplayCursor + 1
	}
	required, revision := state.Scan.CreditReplayTarget, state.Scan.CreditReplayRevision
	for from <= required {
		end := from + size - 1
		if end < from || end > required {
			end = required
		}
		q := BlockRange{from, end}
		page, e := s.publicHistory.Page(ctx, q, false)
		if e != nil {
			return e
		}
		before, e := s.options.Store.ReadSnapshot(ctx, s.scope)
		if e != nil {
			return e
		}
		if before.Scan.CreditReplayRevision != revision || before.Scan.CreditReplayTarget != required {
			return nil
		}
		next, e := CloneState(before)
		if e != nil {
			return e
		}
		for _, tx := range page.Transactions {
			filter, e := contracts.NewInvisiProtocolFilterer(s.scope.ProtocolAddress, nil)
			if e != nil {
				return e
			}
			for _, log := range tx.Receipt.Logs {
				if log == nil || log.Address != s.scope.ProtocolAddress {
					continue
				}
				draw, e := filter.ParseCreditDraw(*log)
				if e != nil {
					continue
				}
				for _, previous := range next.PostpaidCredit.Grants {
					for j, nf := range draw.Params.Nullifiers {
						if previous.Nullifier != common.Hash(nf) {
							continue
						}
						commitment := common.Hash(draw.Params.CreditCommitments[j])
						key := creditsession.GrantKey(commitment, previous.Role)
						if _, ok := next.PostpaidCredit.Grants[key]; ok {
							continue
						}
						recovered, e := previous.Delivery.State.Recover(primitives.BigIntToElement(new(big.Int).SetBytes(nf[:])), primitives.BigIntToElement(commitment.Big()), primitives.BigIntToElement(new(big.Int).SetBytes(draw.Params.RecoveryAmounts[j][:])), tx.Header.Time)
						if e != nil {
							return e
						}
						if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: draw.OutputTreeId.Uint64(), LeafIndex: draw.LeafIndices[j], Commitment: commitment}, target); e != nil {
							return e
						}
						child := previous
						child.Delivery.State = recovered
						child.Commitment = commitment
						nullifier, e := recovered.Nullifier()
						if e != nil {
							return e
						}
						child.Nullifier = common.Hash(nullifier.Bytes())
						child.TreeID = draw.OutputTreeId.Uint64()
						child.LeafIndex = uint16(draw.LeafIndices[j])
						child.Consumed = false
						child.Revoked = false
						child.ReservedBy = ""
						next.PostpaidCredit.Grants[key] = child
						result.AddedCredits++
					}
				}
			}
		}
		next.Scan.CreditReplayCursor = end
		if end == required && next.Scan.CreditReplayRevision == revision {
			next.Scan.CreditReplayTarget = 0
			next.Scan.CreditReplayCursor = 0
		}
		if !o.SkipStateRefresh {
			head := next.Scan.ChainHead
			if head != nil {
				if e = s.checkKnownState(ctx, &next, *head); e != nil {
					return e
				}
			}
		}
		if e = restoreCreditReservations(&next.PostpaidCredit); e != nil {
			return e
		}
		e = s.commitPrivateBatch(ctx, next.Scan.Generation, before.Revision, next)
		if e != nil {
			return e
		}
		if o.OnProgress != nil {
			o.OnProgress(HistoryProgress{"committed", "credit-recovery", q, target})
		}
		if end == required {
			break
		}
		from = end + 1
	}
	return nil
}

func (s *WalletSession) fencedSnapshot(ctx context.Context) (WalletState, error) {
	v, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return v, e
	}
	if s.publicHistory != nil {
		p, e := s.publicHistory.Snapshot(ctx)
		if e != nil {
			return v, e
		}
		if p.Generation != v.Scan.Generation {
			v.Invalidated = true
		}
	}
	return v, nil
}

func (s *WalletSession) commitPrivateBatch(ctx context.Context, generation, revision uint64, next WalletState) error {
	public, e := s.publicHistory.Snapshot(ctx)
	if e != nil {
		return e
	}
	if public.Generation != generation {
		return errors.New("public history generation changed")
	}
	if e = s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		if v.Revision != revision {
			return errors.New("wallet changed during private batch")
		}
		*v = next
		return nil
	}); e != nil {
		return e
	}
	public, e = s.publicHistory.Snapshot(ctx)
	if e != nil {
		return e
	}
	if public.Generation != generation {
		_ = s.invalidate(context.WithoutCancel(ctx))
		return errors.New("public history generation changed during private batch")
	}
	return nil
}
