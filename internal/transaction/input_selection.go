package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (s *Runtime) SelectInputs(ctx context.Context, token common.Address, amount *big.Int, commitments []common.Hash, block protocolrpc.CanonicalBlock) ([]NoteRecord, []protocolrpc.Membership, error) {
	state, e := s.PreparationInventory(ctx)
	if e != nil {
		return nil, nil, e
	}
	if s.ExplicitInventory != nil {
		for _, n := range state.Notes {
			if core.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) != 0 {
				return nil, nil, errors.New("explicit note token mismatch")
			}
		}
	}
	inputs := []NoteRecord{}
	memberships := []protocolrpc.Membership{}
	sum := new(big.Int)
	selected := map[common.Hash]bool{}
	for _, hash := range commitments {
		if selected[hash] {
			return nil, nil, errors.New("duplicate commitment")
		}
		selected[hash] = true
	}
	for _, n := range OrderedNotes(state, token) {
		if n.Note.ProtocolContext != s.CoreContext() || n.Note.OwnerAddress != core.BigIntToElement(s.Wallet.identity.Big()) {
			return nil, nil, errors.New("stored note deployment or owner mismatch")
		}
		expected, e := s.Wallet.ExportNullifier().ForNote(n.Note)
		if e != nil || expected != n.Nullifier {
			return nil, nil, errors.New("stored note nullifier mismatch")
		}
		identity, e := core.ComputeNoteIdentity(n.Note.OwnerAddress, n.Note.Nonce)
		if e != nil {
			return nil, nil, e
		}
		leaf, e := core.ComputeNoteCommitment(n.Note.ChainID, n.Note.ProtocolAddress, identity, n.Note.TokenAddress, n.Note.Amount)
		if e != nil || FieldHash(leaf) != n.Commitment {
			return nil, nil, errors.New("stored note commitment mismatch")
		}

		if len(selected) > 0 && !selected[n.Commitment] {
			continue
		}
		if s.ExplicitInventory == nil && len(selected) == 0 && sum.Cmp(amount) >= 0 {
			break
		}
		contract, e := contracts.NewInvisiProtocolCaller(s.Scope.ProtocolAddress, s.Options.Provider)
		if e != nil {
			return nil, nil, e
		}
		spent, e := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(block.Number)}, n.Nullifier)
		if e != nil {
			return nil, nil, e
		}
		if spent {
			return nil, nil, errors.New("stored note was consumed; refresh wallet")
		}
		m, e := s.RPC.GetMembership(ctx, s.Protocol.Context(), protocolrpc.MembershipInput{TreeID: n.TreeID, LeafIndex: n.LeafIndex, Commitment: n.Commitment}, block)
		if e != nil {
			return nil, nil, e
		}
		inputs = append(inputs, n)
		memberships = append(memberships, m)
		sum.Add(sum, core.ElementToBigInt(n.Note.Amount))
	}
	if sum.Cmp(amount) < 0 || (len(selected) > 0 && len(inputs) != len(selected)) {
		return nil, nil, errors.New("insufficient available known notes")
	}
	return inputs, memberships, nil
}
