package session

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

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
		if n.Spent || primitives.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) != 0 {
			continue
		}
		result.ConfirmedKnown.Add(result.ConfirmedKnown, primitives.ElementToBigInt(n.Note.Amount))
		if n.ReservedBy == "" {
			result.AvailableKnown.Add(result.AvailableKnown, primitives.ElementToBigInt(n.Note.Amount))
			result.NoteCount++
		} else {
			result.Reserved.Add(result.Reserved, primitives.ElementToBigInt(n.Note.Amount))
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
