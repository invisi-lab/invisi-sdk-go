package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	transaction "github.com/invisi-labs/invisi-sdk-go/internal/transaction"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/rpc/contracts"
	creditsession "github.com/invisi-labs/invisi-sdk-go/session/credit"
	"math/big"
)

// Known state is captured at one numeric block/hash. Workers only return values;
// the caller merges them after every read and the canonical fence succeed.
func (s *WalletSession) checkKnownState(ctx context.Context, v *WalletState, b protocolrpc.CanonicalBlock) error {
	c, e := contracts.NewInvisiCoreCaller(s.scope.ProtocolAddress, s.options.Provider)
	if e != nil {
		return e
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(b.Number)}
	cache := scanCache(ctx)
	type statusInput struct {
		owner, token common.Address
		grantID      *big.Int
	}
	statuses := map[string]statusInput{}
	grantKeys := map[string]string{}
	nullifiers := map[common.Hash]bool{}
	for _, n := range v.Notes {
		nullifiers[n.Nullifier] = true
	}
	for key, g := range v.PostpaidCredit.Grants {
		nullifiers[g.Nullifier] = true
		state := g.Delivery.State
		owner := common.BigToAddress(primitives.ElementToBigInt(state.OwnerWalletIdentity))
		token := common.BigToAddress(primitives.ElementToBigInt(state.TokenAddress))
		id := primitives.ElementToBigInt(state.GrantID)
		ns := fmt.Sprintf("%s:%s:%s", owner, token, id)
		statuses[ns] = statusInput{owner, token, id}
		grantKeys[key] = ns
	}
	type read struct {
		nf common.Hash
		ns string
	}
	type value struct {
		spent  bool
		status uint8
	}
	jobs := []read{}
	spent := map[common.Hash]bool{}
	statusValues := map[string]uint8{}
	for nf := range nullifiers {
		if cache != nil {
			if v, ok := cache.spent[spentCheckKey{Block: b, Nullifier: nf}]; ok {
				spent[nf] = v
				continue
			}
		}
		jobs = append(jobs, read{nf: nf})
	}
	for ns := range statuses {
		if cache != nil {
			if v, ok := cache.statuses[statusCheckKey{Block: b, Namespace: ns}]; ok {
				statusValues[ns] = v
				continue
			}
		}
		jobs = append(jobs, read{ns: ns})
	}
	values, e := transaction.MapPreparation(ctx, len(jobs), func(ctx context.Context, i int) (value, error) {
		job := jobs[i]
		local := *opts
		local.Context = ctx
		if job.ns == "" {
			v, e := c.IsNullifierSpent(&local, job.nf)
			return value{spent: v}, e
		}
		q := statuses[job.ns]
		v, e := c.CreditStatusValue(&local, q.owner, q.token, q.grantID)
		if e != nil {
			return value{}, e
		}
		if v > 2 {
			return value{}, errors.New("invalid Credit status")
		}
		return value{status: v}, nil
	})
	if e != nil {
		return e
	}
	if e = s.canonical(ctx, b); e != nil {
		return e
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	for i, job := range jobs {
		if job.ns == "" {
			spent[job.nf] = values[i].spent
			if cache != nil {
				cache.trim()
				cache.spent[spentCheckKey{Block: b, Nullifier: job.nf}] = values[i].spent
			}
		} else {
			statusValues[job.ns] = values[i].status
			if cache != nil {
				cache.trim()
				cache.statuses[statusCheckKey{Block: b, Namespace: job.ns}] = values[i].status
			}
		}
	}
	for key, n := range v.Notes {
		n.Spent = spent[n.Nullifier]
		checked := b
		n.CheckedThrough = &checked
		v.Notes[key] = n
	}
	for key, g := range v.PostpaidCredit.Grants {
		g.Consumed = spent[g.Nullifier]
		previousStatus := g
		if e = g.SetGrantStatus(statusValues[grantKeys[key]], &b); e != nil {
			return e
		}
		g = g.MergeGrantStatus(previousStatus)
		v.PostpaidCredit.Grants[key] = g
	}
	v.CheckedThrough = &b
	v.Scan.ChainHead = &b
	return nil
}

// StateRefreshOptions keeps inventory refresh separate from optional transaction
// material preloading. The zero value preserves offline-preparation support.
type StateRefreshOptions struct{ InventoryOnly bool }

func (s *WalletSession) RefreshState(ctx context.Context) error {
	return s.RefreshStateWithOptions(ctx, StateRefreshOptions{})
}

// RefreshStateWithOptions optionally omits transaction-material preloading while
// preserving the same live state, deployment and recovery checks.
func (s *WalletSession) RefreshStateWithOptions(ctx context.Context, options StateRefreshOptions) error {
	inventoryOnly := options.InventoryOnly
	// A final live refresh must recapture all consumption/revocation reads in its
	// preparation snapshot and must not reuse the discovery scan's state cache.
	ctx = context.WithValue(ctx, scanStateCacheKey{}, (*scanStateCache)(nil))
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
	snap := newPreparationSnapshot(s, before, b, head)
	recorder := transaction.NewSnapshotRecorder(s.options.Provider, snap)
	captured := *s
	captured.options.Provider = recorder
	if e = s.protocol.ValidateDeployment(ctx, recorder); e != nil {
		return e
	}
	if e = captured.checkKnownState(ctx, &before, b); e != nil {
		return e
	}
	if inventoryOnly {
		before.PreparationSnapshot = nil
	} else if e = s.finishPreparationSnapshot(ctx, &before, recorder); e != nil {
		return e
	}
	if _, e = recorder.ChainID(ctx); e != nil {
		return e
	}
	current, e := s.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(b.Number))
	if e != nil {
		return e
	}
	if current == nil || current.Hash() != b.Hash {
		return errors.New("state refresh canonical anchor changed")
	}
	if e = s.commitPrivateBatch(ctx, before.Scan.Generation, before.Revision, before); e != nil {
		return e
	}
	snapshot, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return e
	}
	for _, r := range snapshot.Operations {
		if !r.Externalized || r.TransactionHash == (common.Hash{}) || r.Status == "confirmed" || r.Status == "reverted" {
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
	return credit.ReconcilePending(ctx, s.CreditReconciler("latest"))
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
	if state.Scan.CreditReplayFrom > from {
		from = state.Scan.CreditReplayFrom
	}
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
		if o.SkipStateRefresh {
			checkpointed, err := s.checkpointEmptyCreditReplay(ctx, page, required, revision, end)
			if err != nil {
				return err
			}
			if checkpointed {
				if o.OnProgress != nil {
					o.OnProgress(HistoryProgress{"committed", "credit-recovery", q, target})
				}
				from = end + 1
				continue
			}
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
		filter, e := contracts.NewInvisiCoreFilterer(s.scope.ProtocolAddress, nil)
		if e != nil {
			return e
		}
		// Index both roles and extend it while recovering successors. A later
		// transition in the same transaction/page can spend a recovered child.
		parents := make(map[common.Hash][]creditsession.StoredState, len(next.PostpaidCredit.Grants))
		for _, previous := range next.PostpaidCredit.Grants {
			parents[previous.Nullifier] = append(parents[previous.Nullifier], previous)
		}
		for _, tx := range page.Transactions {
			for _, log := range tx.Receipt.Logs {
				if log == nil || log.Address != s.scope.ProtocolAddress {
					continue
				}
				draw, e := filter.ParseCreditStateTransition(*log)
				if e != nil {
					continue
				}
				for j, nf := range draw.State.Nullifiers {
					for _, previous := range parents[common.Hash(nf)] {
						commitment := common.Hash(draw.State.CreditCommitments[j])
						key := creditsession.GrantKey(commitment, previous.Role)
						if _, ok := next.PostpaidCredit.Grants[key]; ok {
							continue
						}
						recovered, e := previous.Delivery.State.Recover(primitives.BigIntToElement(new(big.Int).SetBytes(nf[:])), primitives.BigIntToElement(commitment.Big()), primitives.BigIntToElement(new(big.Int).SetBytes(draw.State.RecoveryAmounts[j][:])), tx.Header.Time)
						if e != nil {
							return e
						}
						if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: draw.OutputTreeId.Uint64(), LeafIndex: draw.CreditLeafIndices[j], Commitment: commitment}, target); e != nil {
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
						child.LeafIndex = uint16(draw.CreditLeafIndices[j])
						child.Consumed = false
						statusBlock := protocolrpc.CanonicalBlock{Number: tx.Header.Number.Uint64(), Hash: tx.Header.Hash(), ParentHash: tx.Header.ParentHash, Timestamp: tx.Header.Time}
						_ = child.SetGrantStatus(1, &statusBlock)
						child = child.MergeGrantStatusFrom(next.PostpaidCredit.Grants)
						child.ReservedBy = ""
						next.PostpaidCredit.Grants[key] = child
						parents[child.Nullifier] = append(parents[child.Nullifier], child)
						result.AddedCredits++
					}
				}
			}
		}
		next.Scan.CreditReplayCursor = end
		if end == required && next.Scan.CreditReplayRevision == revision {
			next.Scan.CreditReplayTarget = 0
			next.Scan.CreditReplayCursor = 0
			next.Scan.CreditReplayFrom = 0
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
