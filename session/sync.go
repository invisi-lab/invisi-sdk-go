package session

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	creditsession "github.com/invisi-lab/invisi-sdk-go/credit/session"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (s *WalletSession) invalidate(ctx context.Context) error {
	return s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		state.Invalidated = true
		for id, record := range state.Operations {
			if (record.Status == "failed" || record.Status == "expired") && !record.Externalized {
				continue
			}
			record.Status = "reorged"
			record.ResolvedAt = nil
			record.WorkerID = ""
			state.Operations[id] = record
		}
		for id, record := range state.PostpaidCredit.Operations {
			if record.Externalized {
				record.Status = "reorged"
				record.WorkerID = ""
				state.PostpaidCredit.Operations[id] = record
			}
		}
		return nil
	})
}

type RefreshOptions struct {
	FromBlock uint64
	ToBlock   *uint64
	Reset     bool
	PageSize  uint64
}

type RefreshResult struct {
	Block                    *protocolrpc.CanonicalBlock
	AddedNotes, AddedCredits int
}

func (s *WalletSession) refreshInternal(ctx context.Context, options RefreshOptions, resolve walletResolver, pages *refreshPages) (RefreshResult, error) {
	if options.ToBlock != nil {
		height := *options.ToBlock
		options.ToBlock = &height
	}
	select {
	case s.refreshGate <- struct{}{}:
		defer func() { <-s.refreshGate }()
	case <-ctx.Done():
		return RefreshResult{}, ctx.Err()
	}

	var result RefreshResult
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return result, e
	}
	if state.PostpaidCredit.Version != 2 {
		return result, errors.New("legacy Credit store is unsupported")
	}
	// A direct reset can first observe a reorg; fence exposed journals before rebuilding.
	if options.Reset && !state.Invalidated {
		for _, saved := range []*protocolrpc.CanonicalBlock{state.ObservedThrough, state.CheckedThrough, state.DiscoveryThrough} {
			if saved == nil {
				continue
			}
			if err := s.canonical(ctx, *saved); err != nil {
				if e = s.invalidate(ctx); e != nil {
					return result, e
				}
				state, e = s.options.Store.ReadSnapshot(ctx, s.scope)
				if e != nil {
					return result, e
				}
				break
			}
		}
	}
	expectedRevision := state.Revision
	if state.ObservedThrough != nil && !options.Reset {
		if e = s.canonical(ctx, *state.ObservedThrough); e != nil {
			_ = s.invalidate(ctx)
			return result, e
		}
	}
	if state.CheckedThrough != nil && !options.Reset {
		if e = s.canonical(ctx, *state.CheckedThrough); e != nil {
			_ = s.invalidate(ctx)
			return result, e
		}
	}
	if state.DiscoveryThrough != nil && !options.Reset {
		if err := s.canonical(ctx, *state.DiscoveryThrough); err != nil {
			_ = s.invalidate(ctx)
			return result, err
		}
	}
	if state.Invalidated && !options.Reset {
		return result, errors.New("wallet requires reset refresh")
	}
	status, e := s.rpc.GetStatus(ctx, s.protocol.Context())
	if e != nil {
		return result, e
	}
	if status.IndexedHead == nil {
		return result, errors.New("Protocol RPC has not indexed deployment")
	}
	end := status.IndexedHead.Number
	if options.ToBlock != nil {
		end = *options.ToBlock
	}
	if end > status.IndexedHead.Number {
		return result, errors.New("Protocol RPC is behind the requested range")
	}
	if options.FromBlock < s.protocol.Deployment().Protocol.FromBlock || options.FromBlock > end {
		return result, errors.New("invalid refresh range")
	}
	if !options.Reset && state.ObservedThrough != nil && end < state.ObservedThrough.Number {
		return result, errors.New("refresh anchor predates the saved checkpoint")
	}
	finalHeader, e := s.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(end))
	if e != nil {
		return result, e
	}
	if finalHeader == nil || finalHeader.Number == nil || finalHeader.Number.Uint64() != end {
		return result, errors.New("invalid refresh anchor")
	}
	finalBlock := protocolrpc.CanonicalBlock{Number: end, Hash: finalHeader.Hash(), ParentHash: finalHeader.ParentHash, Timestamp: finalHeader.Time}
	if !options.Reset && state.CheckedThrough != nil && end < state.CheckedThrough.Number {
		return result, errors.New("Protocol RPC is behind the latest checked spend state")
	}
	size := options.PageSize
	if size == 0 {
		size = protocolrpc.MaxUpdateBlocks
	}
	if size > protocolrpc.MaxUpdateBlocks {
		return result, errors.New("pageSize must be 1..10000")
	}
	if options.Reset {
		if e = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
			if state.Revision != expectedRevision {
				return errors.New("wallet inventory changed before reset; retry refresh")
			}
			state.Notes = map[common.Hash]StoredNote{}
			state.PostpaidCredit.Grants = map[string]creditsession.StoredState{}
			state.ObservedThrough = nil
			state.CheckedThrough = nil
			state.DiscoveryThrough = nil
			state.Invalidated = true
			return nil
		}); e != nil {
			return result, e
		}
		expectedRevision++
	}
	for from := options.FromBlock; from <= end; {
		to := from + size - 1
		if to < from || to > end {
			to = end
		}
		query := protocolrpc.ProtocolUpdateRange{FromBlock: from, ToBlock: to}
		notePage, e := pages.notes(ctx, s, query)
		if e != nil {
			return result, e
		}
		creditPage, e := pages.credits(ctx, s, query)
		if e != nil {
			return result, e
		}
		if creditPage.Block != notePage.Block {
			return result, errors.New("discovery pages have different anchors")
		}
		page := struct {
			Block             protocolrpc.CanonicalBlock
			TransactionHashes []common.Hash
		}{Block: notePage.Block}
		seen := map[common.Hash]bool{}
		add := func(hash common.Hash) {
			if !seen[hash] {
				seen[hash] = true
				page.TransactionHashes = append(page.TransactionHashes, hash)
			}
		}
		for _, note := range notePage.Notes {
			add(note.TransactionHash)
		}
		for _, event := range creditPage.Transitions {
			add(event.TransactionHash)
		}

		// Empty intermediate pages need no local reconciliation; the final page still checks every spent flag.
		if len(page.TransactionHashes) == 0 && to < end {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			from = to + 1
			continue
		}
		notes := []StoredNote{}
		credits := []creditsession.StoredState{}
		draws := []*authenticatedCreditDraw{}
		// Process complete receipts in canonical transaction order, independently of service ordering.
		type receiptOrder struct {
			hash  common.Hash
			block uint64
			index uint
		}
		order := []receiptOrder{}
		for _, hash := range page.TransactionHashes {
			receipt, e := s.options.Provider.TransactionReceipt(ctx, hash)
			if e != nil {
				return result, e
			}
			if receipt == nil || receipt.BlockNumber == nil || receipt.BlockNumber.Uint64() < from || receipt.BlockNumber.Uint64() > to {
				return result, errors.New("candidate transaction is outside the requested range")
			}
			order = append(order, receiptOrder{hash, receipt.BlockNumber.Uint64(), receipt.TransactionIndex})
		}
		sort.Slice(order, func(i, j int) bool {
			if order[i].block == order[j].block {
				return order[i].index < order[j].index
			}
			return order[i].block < order[j].block
		})
		page.TransactionHashes = nil
		for _, entry := range order {
			page.TransactionHashes = append(page.TransactionHashes, entry.hash)
		}
		for _, hash := range page.TransactionHashes {
			ns, cs, ds, e := s.readTransaction(ctx, hash, finalBlock, resolve)
			if e != nil {
				return result, e
			}
			notes = append(notes, ns...)
			credits = append(credits, cs...)
			draws = append(draws, ds...)
		}
		if e = s.canonical(ctx, finalBlock); e != nil {
			_ = s.invalidate(ctx)
			return result, e
		}
		// Perform all network checks before the atomic local state update.
		before, e := s.options.Store.ReadSnapshot(ctx, s.scope)
		if e != nil {
			return result, e
		}
		if before.Revision != expectedRevision {
			return result, errors.New("wallet inventory changed during refresh; retry")
		}
		for _, c := range credits {
			before.PostpaidCredit.Grants[creditsession.GrantKey(c.Commitment, c.Role)] = c
		}
		for _, draw := range draws {
			for _, previous := range before.PostpaidCredit.Grants {
				for j, nf := range draw.Params.Nullifiers {
					if previous.Nullifier != common.Hash(nf) {
						continue
					}
					recovered, e := previous.Delivery.State.Recover(core.BigIntToElement(new(big.Int).SetBytes(nf[:])), core.BigIntToElement(new(big.Int).SetBytes(draw.Params.CreditCommitments[j][:])), core.BigIntToElement(new(big.Int).SetBytes(draw.Params.RecoveryAmounts[j][:])), draw.Timestamp)
					if e != nil {
						return result, e
					}
					commitment := common.Hash(draw.Params.CreditCommitments[j])
					if _, ok := before.PostpaidCredit.Grants[creditsession.GrantKey(commitment, previous.Role)]; ok {
						continue
					}
					if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: draw.OutputTreeId.Uint64(), LeafIndex: draw.LeafIndices[j], Commitment: commitment}, finalBlock); e != nil {
						return result, e
					}
					child := previous
					child.Delivery.State = recovered
					child.Commitment = commitment
					nullifier, e := recovered.Nullifier()
					if e != nil {
						return result, e
					}
					child.Nullifier = common.Hash(nullifier.Bytes())
					child.TreeID = draw.OutputTreeId.Uint64()
					child.LeafIndex = uint16(draw.LeafIndices[j])
					child.Consumed = false
					child.Revoked = false
					child.ReservedBy = ""
					credits = append(credits, child)
					before.PostpaidCredit.Grants[creditsession.GrantKey(child.Commitment, child.Role)] = child
				}
			}
		}

		for _, n := range notes {
			before.Notes[n.Commitment] = n
		}
		for _, c := range credits {
			before.PostpaidCredit.Grants[creditsession.GrantKey(c.Commitment, c.Role)] = c
		}
		contract, e := contracts.NewInvisiProtocolCaller(s.scope.ProtocolAddress, s.options.Provider)
		if e != nil {
			return result, e
		}
		spentNotes := map[common.Hash]bool{}
		spentCredits := map[string]bool{}
		for hash, n := range before.Notes {
			spent, e := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(end)}, n.Nullifier)
			if e != nil {
				return result, e
			}
			spentNotes[hash] = spent
		}
		revokedCredits := map[string]bool{}
		bitmaps := map[string]*big.Int{}
		for key, c := range before.PostpaidCredit.Grants {
			spent, e := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(end)}, c.Nullifier)
			if e != nil {
				return result, e
			}
			spentCredits[key] = spent
			grant := c.Delivery.State
			owner := common.BigToAddress(core.ElementToBigInt(grant.OwnerWalletIdentity))
			token := common.BigToAddress(core.ElementToBigInt(grant.TokenAddress))
			shard := core.ElementToBigInt(grant.RevocationShard).Uint64()
			slot := core.ElementToBigInt(grant.RevocationSlot).Uint64()
			word := uint8(slot / 256)
			bitmapKey := fmt.Sprintf("%s:%s:%d:%d", owner.Hex(), token.Hex(), shard, word)
			bits, ok := bitmaps[bitmapKey]
			if !ok {
				bits, e = contract.CreditRevocationBitmap(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(end)}, owner, token, shard, word)
				if e != nil {
					return result, e
				}
				if bits == nil || bits.Sign() < 0 || bits.BitLen() > 256 {
					return result, errors.New("invalid revocation bitmap")
				}
				bitmaps[bitmapKey] = bits
			}
			revokedCredits[key] = bits.Bit(int(slot%256)) == 1

		}
		if e = s.canonical(ctx, finalBlock); e != nil {
			_ = s.invalidate(ctx)
			return result, e
		}
		e = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
			if state.Revision != expectedRevision {
				return errors.New("wallet inventory changed before refresh commit; retry")
			}
			if (state.ObservedThrough != nil && state.ObservedThrough.Number > end) || (state.CheckedThrough != nil && state.CheckedThrough.Number > end) {
				return errors.New("concurrent refresh advanced beyond this anchor")
			}

			for _, n := range notes {
				old, ok := state.Notes[n.Commitment]
				if ok {
					n.ReservedBy = old.ReservedBy
				} else {
					result.AddedNotes++
				}
				state.Notes[n.Commitment] = n
			}
			for _, c := range credits {
				k := creditsession.GrantKey(c.Commitment, c.Role)
				old, ok := state.PostpaidCredit.Grants[k]
				if ok {
					c.ReservedBy = old.ReservedBy
				} else {
					result.AddedCredits++
				}
				state.PostpaidCredit.Grants[k] = c
			}

			for hash, spent := range spentNotes {
				n, ok := state.Notes[hash]
				if !ok {
					return errors.New("refresh note disappeared before spent-state commit")
				}
				n.Spent = spent
				state.Notes[hash] = n
			}
			for key, spent := range spentCredits {
				c, ok := state.PostpaidCredit.Grants[key]
				if !ok {
					return errors.New("refresh Credit disappeared before spent-state commit")
				}
				c.Consumed = spent
				c.Revoked = revokedCredits[key]
				state.PostpaidCredit.Grants[key] = c
			}
			if state.ObservedThrough == nil || state.ObservedThrough.Number <= page.Block.Number {
				state.ObservedThrough = &page.Block
			}
			state.CheckedThrough = &finalBlock
			if err := restoreReservations(state); err != nil {
				return err
			}
			if err := restoreCreditReservations(&state.PostpaidCredit); err != nil {
				return err
			}
			if state.PostpaidCredit.Revision == ^uint64(0) {
				return errors.New("Credit revision exhausted")
			}
			state.PostpaidCredit.Revision++
			state.Invalidated = to != end && options.Reset
			return nil
		})
		if e != nil {
			return result, e
		}
		expectedRevision++
		result.Block = &page.Block
		if to == end {
			break
		}
		from = to + 1
	}
	if result.Block != nil {
		err := s.options.Store.Update(ctx, s.scope, func(current *WalletState) error {
			if current.Revision != expectedRevision {
				return errors.New("wallet inventory changed before discovery checkpoint commit; retry")
			}
			start := s.protocol.Deployment().Protocol.FromBlock
			if current.DiscoveryThrough != nil {
				start = current.DiscoveryThrough.Number + 1
			}
			if options.FromBlock <= start && (current.DiscoveryThrough == nil || result.Block.Number >= current.DiscoveryThrough.Number) {
				copy := *result.Block
				current.DiscoveryThrough = &copy
			}
			return nil
		})
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

// Replay unresolved Credit reservations after a reset without releasing externalized authority.
func restoreCreditReservations(state *creditsession.State) error {
	for id, record := range state.Operations {
		if record.Status == "confirmed" || record.Status == "reverted" || record.Status == "cancelled" {
			continue
		}
		for _, input := range record.Intent.Inputs {
			c, e := input.State.Commitment()
			if e != nil {
				return e
			}
			commitment := common.Hash(c.Bytes())
			for key, grant := range state.Grants {
				if grant.Commitment != commitment || grant.Consumed {
					continue
				}
				if grant.ReservedBy != "" && grant.ReservedBy != id {
					return errors.New("conflicting persisted Credit reservations")
				}
				grant.ReservedBy = id
				state.Grants[key] = grant
			}
		}
	}
	return nil
}
