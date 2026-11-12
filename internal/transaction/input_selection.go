package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
)

func (s *Runtime) SelectInputs(ctx context.Context, token common.Address, amount *big.Int, commitments []common.Hash, block protocolrpc.CanonicalBlock) ([]NoteRecord, []protocolrpc.Membership, error) {
	state, e := s.PreparationInventory(ctx)
	if e != nil {
		return nil, nil, e
	}
	if s.ExplicitInventory != nil {
		for _, n := range state.Notes {
			if primitives.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) != 0 {
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
		if n.Note.ProtocolContext != s.CoreContext() || n.Note.OwnerAddress != primitives.BigIntToElement(s.Wallet.ownerAddress.Big()) {
			return nil, nil, errors.New("stored note deployment or owner mismatch")
		}
		expected, e := s.Wallet.ExportNullifier().ForNote(n.Note)
		if e != nil || expected != n.Nullifier {
			return nil, nil, errors.New("stored note nullifier mismatch")
		}
		identity, e := primitives.ComputeNoteIdentity(n.Note.OwnerAddress, n.Note.Nonce)
		if e != nil {
			return nil, nil, e
		}
		leaf, e := primitives.ComputeNoteCommitment(n.Note.ChainID, n.Note.ProtocolAddress, identity, n.Note.TokenAddress, n.Note.Amount)
		if e != nil || FieldHash(leaf) != n.Commitment {
			return nil, nil, errors.New("stored note commitment mismatch")
		}

		if len(selected) > 0 && !selected[n.Commitment] {
			continue
		}
		if s.ExplicitInventory == nil && len(selected) == 0 && sum.Cmp(amount) >= 0 {
			break
		}
		inputs = append(inputs, n)
		sum.Add(sum, primitives.ElementToBigInt(n.Note.Amount))
	}
	if sum.Cmp(amount) < 0 || (len(selected) > 0 && len(inputs) != len(selected)) {
		return nil, nil, errors.New("insufficient available known notes")
	}
	// Pin one indexed root per tree before concurrent paths are requested.
	first := map[uint64]int{}
	leaders := []int{}
	for i, n := range inputs {
		if _, ok := first[n.TreeID]; !ok {
			first[n.TreeID] = i
			leaders = append(leaders, i)
		}
	}
	memberships = make([]protocolrpc.Membership, len(inputs))
	read := func(ctx context.Context, i int, root *common.Hash) (protocolrpc.Membership, error) {
		n := inputs[i]
		contract, err := contracts.NewInvisiProtocolCaller(s.Scope.ProtocolAddress, s.Options.Provider)
		if err != nil {
			return protocolrpc.Membership{}, err
		}
		spent, err := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(block.Number)}, n.Nullifier)
		if err != nil {
			return protocolrpc.Membership{}, err
		}
		if spent {
			return protocolrpc.Membership{}, errors.New("stored note was consumed; refresh wallet")
		}
		return s.membership(ctx, protocolrpc.MembershipInput{TreeID: n.TreeID, LeafIndex: n.LeafIndex, Commitment: n.Commitment, Root: root}, block)
	}
	_, e = MapPreparation(ctx, len(leaders), func(ctx context.Context, j int) (struct{}, error) {
		i := leaders[j]
		m, err := read(ctx, i, nil)
		memberships[i] = m
		return struct{}{}, err
	})
	if e != nil {
		return nil, nil, e
	}
	rest := []int{}
	for i, n := range inputs {
		if first[n.TreeID] != i {
			rest = append(rest, i)
		}
	}
	_, e = MapPreparation(ctx, len(rest), func(ctx context.Context, j int) (struct{}, error) {
		i := rest[j]
		root := memberships[first[inputs[i].TreeID]].Root
		m, err := read(ctx, i, &root)
		memberships[i] = m
		return struct{}{}, err
	})
	if e != nil {
		return nil, nil, e
	}
	return inputs, memberships, nil
}

func (s *Runtime) membership(ctx context.Context, input protocolrpc.MembershipInput, block protocolrpc.CanonicalBlock) (protocolrpc.Membership, error) {
	if s.Snapshot != nil {
		return s.Snapshot.Membership(input, block)
	}
	load := func() ([]byte, error) {
		m, err := s.RPC.GetMembership(ctx, s.Protocol.Context(), input, block)
		if err != nil {
			return nil, err
		}
		return json.Marshal(m)
	}
	var raw []byte
	var err error
	if provider, ok := s.Options.Provider.(*preparationProvider); ok {
		key := fmt.Sprintf("membership:%s:%d:%d:%s:%v", block.Hash, input.TreeID, input.LeafIndex, input.Commitment, input.Root)
		raw, err = provider.reads.read(ctx, key, load)
	} else {
		raw, err = load()
	}
	if err != nil {
		return protocolrpc.Membership{}, err
	}
	var membership protocolrpc.Membership
	err = json.Unmarshal(raw, &membership)
	return membership, err
}
