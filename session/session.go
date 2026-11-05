package session

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/internal/receipts"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"

	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (s *WalletSession) Scope() WalletScope { return s.scope }

func (s *WalletSession) Protocol() *protocol.Protocol { return s.protocol }

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
		return nil
	})
}

type RefreshOptions struct {
	FromBlock    uint64
	ToBlock      *uint64
	DetectionKey []byte
	Reset        bool
	PageSize     uint64
}

type RefreshResult struct {
	Block                    *protocolrpc.CanonicalBlock
	AddedNotes, AddedCredits int
}

func (s *WalletSession) Refresh(ctx context.Context, options RefreshOptions) (RefreshResult, error) {
	options.DetectionKey = append([]byte(nil), options.DetectionKey...)
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
	if s.viewing == nil {
		return result, errors.New("viewing permission required")
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return result, e
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
			state.Credits = map[string]StoredCredit{}
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
		page, e := s.rpc.GetUpdates(ctx, s.protocol.Context(), protocolrpc.ProtocolUpdateRange{FromBlock: from, ToBlock: to, DetectionKey: options.DetectionKey})
		if e != nil {
			return result, e
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
		credits := []StoredCredit{}
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
			ns, cs, ds, e := s.readTransaction(ctx, hash, finalBlock)
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
			before.Credits[transaction.CreditKey(c.Commitment, c.Role)] = c
		}
		for _, draw := range draws {
			for _, previous := range before.Credits {
				for j, nf := range draw.Params.Nullifiers {
					if previous.Nullifier != common.Hash(nf) {
						continue
					}
					recovered, e := core.RecoverCreditRemainder(s.coreContext(), previous.State.Note, core.CreditRemainderRecovery{CreditNullifier: core.BigIntToElement(new(big.Int).SetBytes(nf[:])), RemainderCommitment: core.BigIntToElement(new(big.Int).SetBytes(draw.Params.RemainderCommitments[j][:])), RecoveryAmount: core.BigIntToElement(new(big.Int).SetBytes(draw.Params.RecoveryAmounts[j][:]))}, draw.Timestamp)
					if e != nil {
						return result, e
					}
					if recovered == nil {
						continue
					}
					commitment := common.Hash(draw.Params.RemainderCommitments[j])
					// An authenticated child envelope carries metadata that proof-only recovery cannot reconstruct.
					if _, ok := before.Credits[transaction.CreditKey(commitment, previous.Role)]; ok {
						continue
					}
					if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: draw.OutputTreeId.Uint64(), LeafIndex: draw.RemainderLeafIndices[j], Commitment: commitment}, finalBlock); e != nil {
						return result, e
					}
					child := previous
					child.State.Note = *recovered
					child.Commitment = commitment
					keyHash, _ := core.ComputeCreditSharedKeyHash(recovered.CreditSharedKey)
					leaf, e := core.ComputeCreditCommitment(recovered.ChainID, recovered.SpenderWalletIdentity, recovered.OwnerWalletIdentity, recovered.SpenderPayout, recovered.OwnerPayout, recovered.TokenAddress, recovered.Amount, recovered.Nonce, recovered.Expiry, keyHash, recovered.CreditSpendingState)
					if e != nil {
						return result, e
					}
					cn, e := core.ComputeCreditNullifier(recovered.ChainID, recovered.CreditSharedKey, leaf)
					if e != nil {
						return result, e
					}
					child.Nullifier = transaction.FieldHash(cn)
					child.TreeID = draw.OutputTreeId.Uint64()
					child.LeafIndex = draw.RemainderLeafIndices[j]
					child.Spent = false
					child.ReservedBy = ""
					credits = append(credits, child)
					before.Credits[transaction.CreditKey(child.Commitment, child.Role)] = child
				}
			}
		}
		for _, n := range notes {
			before.Notes[n.Commitment] = n
		}
		for _, c := range credits {
			before.Credits[transaction.CreditKey(c.Commitment, c.Role)] = c
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
		for key, c := range before.Credits {
			spent, e := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(end)}, c.Nullifier)
			if e != nil {
				return result, e
			}
			spentCredits[key] = spent
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
				k := transaction.CreditKey(c.Commitment, c.Role)
				old, ok := state.Credits[k]
				if ok {
					c.ReservedBy = old.ReservedBy
				} else {
					result.AddedCredits++
				}
				state.Credits[k] = c
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
				c, ok := state.Credits[key]
				if !ok {
					return errors.New("refresh Credit disappeared before spent-state commit")
				}
				c.Spent = spent
				state.Credits[key] = c
			}
			if state.ObservedThrough == nil || state.ObservedThrough.Number <= page.Block.Number {
				state.ObservedThrough = &page.Block
			}
			state.CheckedThrough = &finalBlock
			if err := restoreReservations(state); err != nil {
				return err
			}
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

type authenticatedCreditDraw struct {
	*contracts.InvisiProtocolCreditDraw
	Timestamp uint64
}

func (s *WalletSession) readTransaction(ctx context.Context, hash common.Hash, anchor protocolrpc.CanonicalBlock) ([]StoredNote, []StoredCredit, []*authenticatedCreditDraw, error) {
	receipt, e := s.options.Provider.TransactionReceipt(ctx, hash)
	if e != nil {
		return nil, nil, nil, e
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful || receipt.TxHash != hash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() || receipt.BlockNumber.Uint64() > anchor.Number {
		return nil, nil, nil, errors.New("invalid candidate receipt")
	}
	header, e := s.options.Provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if e != nil {
		return nil, nil, nil, e
	}
	if header == nil || header.Number == nil || header.Number.Cmp(receipt.BlockNumber) != 0 || header.Hash() != receipt.BlockHash {
		return nil, nil, nil, errors.New("candidate receipt is not canonical")
	}
	logs := []types.Log{}
	for _, l := range receipt.Logs {
		if l == nil {
			return nil, nil, nil, errors.New("nil receipt log")
		}
		if l.Address != s.scope.ProtocolAddress {
			continue
		}
		if l.Removed || l.TxHash != hash || l.BlockHash != receipt.BlockHash || l.BlockNumber != receipt.BlockNumber.Uint64() {
			return nil, nil, nil, errors.New("invalid receipt log binding")
		}
		logs = append(logs, *l)
	}
	decoder, e := receipts.NewDecoder()
	if e != nil {
		return nil, nil, nil, e
	}
	batch, e := decoder.Decode(receipts.Block{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash, ParentHash: header.ParentHash, Timestamp: header.Time}, logs)
	if e != nil {
		return nil, nil, nil, e
	}
	notes := []StoredNote{}
	credits := []StoredCredit{}
	draws := []*authenticatedCreditDraw{}
	filter, e := contracts.NewInvisiProtocolFilterer(s.scope.ProtocolAddress, nil)
	if e != nil {
		return nil, nil, nil, e
	}
	for _, l := range logs {
		if draw, e := filter.ParseCreditDraw(l); e == nil {
			draws = append(draws, &authenticatedCreditDraw{draw, header.Time})
		}
	}
	for _, output := range batch.Outputs {
		// Omitted Funds delivery keeps its tree position but has no envelope to decrypt.
		if len(output.Ciphertext) == 0 && len(output.Clue) == 0 {
			continue
		}
		context := core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: s.coreContext(), OutputCommitment: core.BigIntToElement(output.Commitment.Big())}, TokenAddress: core.BigIntToElement(output.Token.Big()), OwnerAddress: core.BigIntToElement(s.wallet.OwnerAddress().Big())}
		if note, e := s.viewing.DecryptNote(output.Ciphertext, context); e == nil {
			if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: output.TreeID, LeafIndex: output.LeafIndex, Commitment: output.Commitment}, anchor); e != nil {
				return nil, nil, nil, e
			}
			nf, e := s.wallet.ExportNullifier().ForNote(note)
			if e != nil {
				return nil, nil, nil, e
			}
			notes = append(notes, StoredNote{Note: note, Commitment: output.Commitment, Nullifier: nf, TreeID: output.TreeID, LeafIndex: output.LeafIndex, TransactionHash: hash, Block: &protocolrpc.CanonicalBlock{Number: header.Number.Uint64(), Hash: header.Hash(), ParentHash: header.ParentHash, Timestamp: header.Time}})
			continue
		}
		for _, role := range []core.CreditDeliveryRole{core.CreditOwner, core.CreditSpender} {
			state, e := s.viewing.DecryptCredit(output.Ciphertext, core.CreditNoteDecryptionContext{NoteDecryptionContext: context, Role: role})
			if e != nil {
				continue
			}
			if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: output.TreeID, LeafIndex: output.LeafIndex, Commitment: output.Commitment}, anchor); e != nil {
				return nil, nil, nil, e
			}
			nf, e := core.ComputeCreditNullifier(state.Note.ChainID, state.Note.CreditSharedKey, context.OutputCommitment)
			if e != nil {
				return nil, nil, nil, e
			}
			credits = append(credits, StoredCredit{State: state, Commitment: output.Commitment, Nullifier: transaction.FieldHash(nf), TreeID: output.TreeID, LeafIndex: output.LeafIndex, Role: string(role)})
		}
	}
	return notes, credits, draws, nil
}

type BalanceSnapshot struct {
	AvailableKnown, ConfirmedKnown, Reserved *big.Int
	TokenAddress                             common.Address
	Scope                                    WalletScope
	Invalidated                              bool
	CheckedThrough                           *protocolrpc.CanonicalBlock
	NoteCount                                int
}

func (s *WalletSession) Balance(ctx context.Context, token common.Address) (BalanceSnapshot, error) {
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		return BalanceSnapshot{}, e
	}
	checked := state.CheckedThrough
	if checked == nil {
		checked = state.ObservedThrough
	}
	result := BalanceSnapshot{AvailableKnown: new(big.Int), ConfirmedKnown: new(big.Int), Reserved: new(big.Int), CheckedThrough: checked, TokenAddress: token, Scope: s.scope, Invalidated: state.Invalidated}
	for _, n := range state.Notes {
		if n.Spent || core.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) != 0 {
			continue
		}
		result.ConfirmedKnown.Add(result.ConfirmedKnown, core.ElementToBigInt(n.Note.Amount))
		if n.ReservedBy == "" {
			result.AvailableKnown.Add(result.AvailableKnown, core.ElementToBigInt(n.Note.Amount))
			result.NoteCount++
		} else {
			result.Reserved.Add(result.Reserved, core.ElementToBigInt(n.Note.Amount))
		}
	}
	if state.Invalidated || checked == nil {
		result.AvailableKnown = nil
	}
	if state.Invalidated {
		result.CheckedThrough = nil
	}
	return result, nil
}

func (s *WalletSession) String() string {
	return fmt.Sprintf("WalletSession(%s,%s)", strconv.FormatUint(s.scope.ChainID, 10), s.wallet.OwnerAddress().Hex())
}
