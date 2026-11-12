package transaction

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
)

// CreditInput supplies an application-selected grant and its authenticated tree position.
// Applications own persistence and reservations for detached operations.
type CreditInput struct {
	Delivery   credit.DeliveryState
	Commitment common.Hash
	TreeID     uint64
	LeafIndex  uint16
}
type ExplicitCreateCreditPreparation struct {
	TransactionIdentity
	credit.CreatePreparation
}
type ExplicitUpdateCreditPreparation struct {
	TransactionIdentity
	credit.UpdatePreparation
	Inputs []CreditInput
}
type ExplicitDrawCreditPreparation struct {
	TransactionIdentity
	credit.DrawPreparation
	Inputs []CreditInput
}
type ExplicitRevokeCreditPreparation struct {
	TransactionIdentity
	credit.RevokePreparation
}

func (i *Invisi) creditBuilder(ctx context.Context, identity TransactionIdentity, inputs []CreditInput) (credit.PreparationBuilder, error) {
	inputs, err := Frozen(inputs)
	if err != nil {
		return credit.PreparationBuilder{}, err
	}
	r, err := i.explicit(ctx, identity, ExplicitFunds{})
	if err != nil {
		return credit.PreparationBuilder{}, err
	}
	r = r.preparationRuntime()
	grants := map[common.Hash]CreditInput{}
	for _, g := range inputs {
		if _, ok := grants[g.Commitment]; ok {
			return credit.PreparationBuilder{}, errors.New("duplicate Credit input")
		}
		grants[g.Commitment] = g
	}
	var block protocolrpc.CanonicalBlock
	roots := map[uint64]common.Hash{}
	return credit.PreparationBuilder{
		Owner:   identity.Wallet.GetAddress,
		Resolve: r.ResolveRecipient,
		Frame: func(ctx context.Context, token common.Address, deadline uint64) (credit.Context, uint64, error) {
			c, b, d, err := r.Frame(ctx, token, deadline)
			if err != nil {
				return credit.Context{}, 0, err
			}
			block = b
			return credit.Context{ProtocolContext: r.CoreContext(), TokenAddress: token, Timestamp: c.Now.Uint64()}, d, nil
		},
		Allocation: func(ctx context.Context, c credit.Context) (uint64, uint16, error) {
			caller, err := contracts.NewInvisiProtocolCaller(r.Scope.ProtocolAddress, r.Options.Provider)
			if err != nil {
				return 0, 0, err
			}
			a, err := caller.CreditAllocation(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(block.Number)}, identity.Wallet.ownerAddress, c.TokenAddress)
			return a.Shard, a.Slot, err
		},
		Input: func(ctx context.Context, c credit.Context, h common.Hash, role credit.Role) (credit.PreparationInput, error) {
			g, ok := grants[h]
			if !ok {
				return credit.PreparationInput{}, errors.New("selected Credit input unavailable")
			}
			s := g.Delivery.State
			who := s.OwnerWalletIdentity
			if role == credit.Spender {
				who = s.SpenderWalletIdentity
			}
			leaf, err := s.Commitment()
			if err != nil {
				return credit.PreparationInput{}, err
			}
			if s.ProtocolContext != c.ProtocolContext || common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)) != c.TokenAddress || common.BigToAddress(primitives.ElementToBigInt(who)) != identity.Wallet.ownerAddress || common.Hash(leaf.Bytes()) != h {
				return credit.PreparationInput{}, errors.New("Credit role, commitment or deployment mismatch")
			}
			caller, err := contracts.NewInvisiProtocolCaller(r.Scope.ProtocolAddress, r.Options.Provider)
			if err != nil {
				return credit.PreparationInput{}, err
			}
			opts := &bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(block.Number)}
			bitmap, err := caller.CreditRevocationBitmap(opts, common.BigToAddress(primitives.ElementToBigInt(s.OwnerWalletIdentity)), c.TokenAddress, primitives.ElementToBigInt(s.RevocationShard).Uint64(), uint8(primitives.ElementToBigInt(s.RevocationSlot).Uint64()/256))
			if err != nil {
				return credit.PreparationInput{}, err
			}
			if bitmap.Bit(int(primitives.ElementToBigInt(s.RevocationSlot).Uint64()%256)) != 0 {
				return credit.PreparationInput{}, errors.New("Credit grant is revoked")
			}
			nf, err := s.Nullifier()
			if err != nil {
				return credit.PreparationInput{}, err
			}
			spent, err := caller.IsNullifierSpent(opts, common.Hash(nf.Bytes()))
			if err != nil {
				return credit.PreparationInput{}, err
			}
			if spent {
				return credit.PreparationInput{}, errors.New("Credit grant is consumed")
			}
			request := protocolrpc.MembershipInput{Commitment: h, TreeID: g.TreeID, LeafIndex: uint32(g.LeafIndex)}
			if root, ok := roots[g.TreeID]; ok {
				request.Root = &root
			}
			m, err := r.membership(ctx, request, block)
			if err != nil {
				return credit.PreparationInput{}, err
			}
			known, err := caller.RootHistory(opts, new(big.Int).SetUint64(g.TreeID), m.Root)
			if err != nil {
				return credit.PreparationInput{}, err
			}
			if !known {
				return credit.PreparationInput{}, errors.New("Credit root is unknown")
			}
			roots[g.TreeID] = m.Root
			located := credit.LocatedState{State: s, TreeID: g.TreeID, LeafIndex: g.LeafIndex}
			if len(m.Path) != len(located.Path) {
				return credit.PreparationInput{}, fmt.Errorf("Credit membership depth mismatch")
			}
			copy(located.Path[:], m.Path)
			return credit.PreparationInput{Located: located, Delivery: g.Delivery, Source: protocol.InputSource{TreeId: new(big.Int).SetUint64(g.TreeID), MerkleRoot: m.Root}}, nil
		},
	}, nil
}
func (i *Invisi) PrepareCreateCredit(ctx context.Context, p ExplicitCreateCreditPreparation) (result *credit.Prepared, failure error) {
	defer func() { failure = WrapPreparation("prepareCreateCredit", failure) }()
	p.CreatePreparation = p.CreatePreparation.Clone()
	b, err := i.creditBuilder(ctx, p.TransactionIdentity, nil)
	if err != nil {
		return nil, err
	}
	return b.Create(ctx, p.CreatePreparation)
}
func (i *Invisi) PrepareUpdateCredit(ctx context.Context, p ExplicitUpdateCreditPreparation) (result *credit.Prepared, failure error) {
	defer func() { failure = WrapPreparation("prepareUpdateCredit", failure) }()
	p.UpdatePreparation = p.UpdatePreparation.Clone()
	b, err := i.creditBuilder(ctx, p.TransactionIdentity, p.Inputs)
	if err != nil {
		return nil, err
	}
	return b.Update(ctx, p.UpdatePreparation)
}
func (i *Invisi) PrepareDrawCredit(ctx context.Context, p ExplicitDrawCreditPreparation) (result *credit.Prepared, failure error) {
	defer func() { failure = WrapPreparation("prepareDrawCredit", failure) }()
	p.DrawPreparation = p.DrawPreparation.Clone()
	b, err := i.creditBuilder(ctx, p.TransactionIdentity, p.Inputs)
	if err != nil {
		return nil, err
	}
	return b.Draw(ctx, p.DrawPreparation)
}
func (i *Invisi) PrepareRevokeCredit(ctx context.Context, p ExplicitRevokeCreditPreparation) (result *credit.Prepared, failure error) {
	defer func() { failure = WrapPreparation("prepareRevokeCredit", failure) }()
	p.RevokePreparation = p.RevokePreparation.Clone()
	b, err := i.creditBuilder(ctx, p.TransactionIdentity, nil)
	if err != nil {
		return nil, err
	}
	return b.Revoke(ctx, p.RevokePreparation)
}
