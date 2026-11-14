package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
)

// ImportNotes authenticates known openings without claiming discovery coverage.
func (s *WalletSession) ImportNotes(ctx context.Context, inputs []sdk.NoteInput) error {
	inputs, err := frozen(inputs)
	if err != nil {
		return err
	}
	before, err := s.fencedSnapshot(ctx)
	if err != nil {
		return err
	}
	if before.Invalidated {
		return errors.New("wallet requires reset scan")
	}
	if err = s.protocol.ValidateDeployment(ctx, s.options.Provider); err != nil {
		return err
	}
	head, err := s.options.Provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() {
		return errors.New("invalid chain head")
	}
	b := protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	for _, old := range []*protocolrpc.CanonicalBlock{before.ObservedThrough, before.CheckedThrough} {
		if old != nil {
			if err = s.canonical(ctx, *old); err != nil {
				return err
			}
		}
	}
	seen := map[common.Hash]bool{}
	for _, input := range inputs {
		if seen[input.Commitment] {
			return errors.New("duplicate note input")
		}
		seen[input.Commitment] = true
		n := input.Note
		if n.ProtocolContext != s.coreContext() || n.OwnerAddress != primitives.BigIntToElement(s.wallet.OwnerAddress().Big()) {
			return errors.New("note deployment or owner mismatch")
		}
		amount := primitives.ElementToBigInt(n.Amount)
		if amount.Sign() <= 0 || amount.BitLen() > 128 {
			return errors.New("invalid note amount")
		}
		identity, e := primitives.ComputeNoteIdentity(n.OwnerAddress, n.Nonce)
		if e != nil {
			return e
		}
		leaf, e := primitives.ComputeNoteCommitment(n.ChainID, n.ProtocolAddress, identity, n.TokenAddress, n.Amount)
		if e != nil {
			return e
		}
		if common.Hash(leaf.Bytes()) != input.Commitment {
			return errors.New("note commitment mismatch")
		}
		nf, e := s.wallet.ExportNullifier().ForNote(n)
		if e != nil {
			return e
		}
		if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: input.TreeID, LeafIndex: input.LeafIndex, Commitment: input.Commitment}, b); e != nil {
			return e
		}
		if old, ok := before.Notes[input.Commitment]; ok {
			if !old.Note.Equal(n) || old.Nullifier != nf || old.TreeID != input.TreeID || old.LeafIndex != input.LeafIndex {
				return errors.New("import conflicts with stored note")
			}
		} else {
			before.Notes[input.Commitment] = StoredNote{Note: n, Commitment: input.Commitment, Nullifier: nf, TreeID: input.TreeID, LeafIndex: input.LeafIndex}
		}
	}
	spent := map[common.Hash]bool{}
	for k, n := range before.Notes {
		spent[k] = n.Spent
	}
	if err = s.checkKnownState(ctx, &before, b); err != nil {
		return err
	}
	for k, n := range before.Notes {
		n.Spent = n.Spent || spent[k]
		before.Notes[k] = n
	}
	before.ObservedThrough = &b
	before.PreparationSnapshot = nil
	if err = restoreReservations(&before); err != nil {
		return err
	}
	return s.commitPrivateBatch(ctx, before.Scan.Generation, before.Revision, before)
}

// CollectionOptions configures the frozen collection deadline. Zero uses the normal default.
type CollectionOptions struct{ Deadline uint64 }

// PrepareCollection freezes exactly the selected available notes into a full withdrawal.
func (s *WalletSession) PrepareCollection(ctx context.Context, token, recipient common.Address, commitments []common.Hash, options ...CollectionOptions) (*Operation, error) {
	if len(options) > 1 {
		return nil, errors.New("at most one collection option is supported")
	}
	var deadline uint64
	if len(options) == 1 {
		deadline = options[0].Deadline
	}
	if recipient == (common.Address{}) || len(commitments) == 0 {
		return nil, errors.New("collection requires a recipient and selected notes")
	}
	commitments = append([]common.Hash(nil), commitments...)
	state, err := s.fencedSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	sum := new(big.Int)
	seen := map[common.Hash]bool{}
	for _, key := range commitments {
		n, ok := state.Notes[key]
		if seen[key] || !ok || n.Spent || n.ReservedBy != "" || primitives.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) != 0 {
			return nil, errors.New("collection note unavailable or duplicate")
		}
		seen[key] = true
		sum.Add(sum, primitives.ElementToBigInt(n.Note.Amount))
	}
	return s.PrepareTransfer(ctx, sdk.TransferPreparation{Deadline: deadline, TokenAddress: token, Commitments: commitments, Withdrawal: &sdk.Withdrawal{To: recipient, Amount: sum}})
}
