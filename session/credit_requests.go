package session

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

func (s *WalletSession) buildCreditRequest(ctx context.Context, request any) (result *credit.Prepared, failure error) {
	defer func() { failure = transaction.WrapPreparation("prepareCredit", failure) }()
	r, err := s.preparationRuntime(ctx)
	if err != nil {
		return nil, err
	}
	prepared, err := s.buildCreditRequestWithRuntime(ctx, r, request)
	if errors.Is(err, transaction.ErrPreparationSnapshotMissing) {
		return s.buildCreditRequestWithRuntime(ctx, s.runtimeSnapshot(), request)
	}
	return prepared, err
}
func (s *WalletSession) buildCreditRequestWithRuntime(ctx context.Context, r *transaction.Runtime, request any) (*credit.Prepared, error) {
	var block protocolrpc.CanonicalBlock
	if r.Snapshot != nil {
		r.Options.Provider = r.Snapshot.Provider(r.Options.Provider)
	}
	roots := map[uint64]common.Hash{}
	builder := credit.PreparationBuilder{
		Owner:   s.wallet.GetAddress,
		Resolve: r.ResolveRecipient,
		Frame: func(ctx context.Context, token common.Address, deadline uint64) (credit.Context, uint64, error) {
			c, b, d, err := r.Frame(ctx, token, deadline)
			if err != nil {
				return credit.Context{}, 0, err
			}
			block = b
			return credit.Context{ProtocolContext: s.coreContext(), TokenAddress: token, Timestamp: c.Now.Uint64()}, d, err
		},
		Allocation: func(ctx context.Context, c credit.Context) (uint64, uint16, error) {
			caller, err := contracts.NewInvisiProtocolCaller(s.scope.ProtocolAddress, r.Options.Provider)
			if err != nil {
				return 0, 0, err
			}
			allocation, err := caller.CreditAllocation(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(block.Number)}, s.scope.OwnerAddress, c.TokenAddress)
			return allocation.Shard, allocation.Slot, err
		},
		Input: func(ctx context.Context, c credit.Context, commitment common.Hash, role credit.Role) (credit.PreparationInput, error) {
			state, err := s.fencedSnapshot(ctx)
			if err != nil {
				return credit.PreparationInput{}, err
			}
			g, ok := state.PostpaidCredit.Grants[creditsession.GrantKey(commitment, role)]
			if !ok || g.Consumed || g.Revoked || g.ReservedBy != "" {
				return credit.PreparationInput{}, fmt.Errorf("confirmed unspent Credit role unavailable")
			}
			n := g.Delivery.State
			identity := n.OwnerWalletIdentity
			if role == credit.Spender {
				identity = n.SpenderWalletIdentity
			}
			leaf, err := n.Commitment()
			if err != nil {
				return credit.PreparationInput{}, err
			}
			if n.ProtocolContext != c.ProtocolContext || common.BigToAddress(primitives.ElementToBigInt(n.TokenAddress)) != c.TokenAddress || common.BigToAddress(primitives.ElementToBigInt(identity)) != s.scope.OwnerAddress || common.Hash(leaf.Bytes()) != commitment {
				return credit.PreparationInput{}, fmt.Errorf("Credit role, commitment or deployment mismatch")
			}
			input := protocolrpc.MembershipInput{TreeID: g.TreeID, LeafIndex: uint32(g.LeafIndex), Commitment: commitment}
			if root, ok := roots[g.TreeID]; ok {
				input.Root = &root
			}
			var m protocolrpc.Membership
			if r.Snapshot != nil {
				m, err = r.Snapshot.Membership(input, block)
			} else {
				m, err = s.rpc.GetMembership(ctx, s.protocol.Context(), input, block)
			}
			if err != nil {
				return credit.PreparationInput{}, err
			}
			roots[g.TreeID] = m.Root
			located := credit.LocatedState{State: n, TreeID: g.TreeID, LeafIndex: g.LeafIndex}
			if len(m.Path) != len(located.Path) {
				return credit.PreparationInput{}, fmt.Errorf("Credit membership depth mismatch")
			}
			copy(located.Path[:], m.Path)
			return credit.PreparationInput{Located: located, Delivery: g.Delivery, Source: protocol.InputSource{TreeId: new(big.Int).SetUint64(g.TreeID), MerkleRoot: m.Root}}, nil
		},
	}
	switch p := request.(type) {
	case credit.CreatePreparation:
		return builder.Create(ctx, p)
	case credit.UpdatePreparation:
		return builder.Update(ctx, p)
	case credit.DrawPreparation:
		return builder.Draw(ctx, p)
	case credit.RevokePreparation:
		return builder.Revoke(ctx, p)
	default:
		return nil, fmt.Errorf("unsupported Credit preparation request")
	}
}
