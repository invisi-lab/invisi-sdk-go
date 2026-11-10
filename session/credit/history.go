package session

import (
	"context"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

type GrantSummary struct {
	Commitment                                             common.Hash
	Owner, Spender, Token                                  common.Address
	Role                                                   credit.Role
	Shard                                                  uint64
	Slot                                                   uint16
	Expiry                                                 uint64
	Unlimited                                              bool
	Consumed, Revoked                                      bool
	ReservedBy                                             string
	SingleLimit, DailyLimit, MonthlyLimit, AvailableCredit *big.Int
	LastDay, LastMonth                                     uint64
	DailySpent, MonthlySpent, TotalSpent                   *big.Int
}
type OperationSummary struct {
	ID, Kind, Status string
	TransactionHash  common.Hash
	Externalized     bool
	Preview          credit.Preview
}
type History struct {
	Grants     []GrantSummary
	Operations []OperationSummary
}

// GetHistory omits grant recovery keys, nonces, envelope keys, authorization signatures and proof witnesses.
func (s *Session) GetHistory(ctx context.Context) (History, error) {
	state, err := s.store.ReadSnapshot(ctx)
	if err != nil {
		return History{}, err
	}
	result := History{Grants: []GrantSummary{}, Operations: []OperationSummary{}}
	for _, record := range state.Grants {
		n := record.Delivery.State
		result.Grants = append(result.Grants, GrantSummary{Commitment: record.Commitment, Owner: common.BigToAddress(primitives.ElementToBigInt(n.OwnerWalletIdentity)), Spender: common.BigToAddress(primitives.ElementToBigInt(n.SpenderWalletIdentity)), Token: common.BigToAddress(primitives.ElementToBigInt(n.TokenAddress)), Role: record.Role, Shard: primitives.ElementToBigInt(n.RevocationShard).Uint64(), Slot: uint16(primitives.ElementToBigInt(n.RevocationSlot).Uint64()), Expiry: primitives.ElementToBigInt(n.Expiry).Uint64(), Consumed: record.Consumed, Revoked: record.Revoked, ReservedBy: record.ReservedBy, SingleLimit: primitives.ElementToBigInt(n.SingleLimit), DailyLimit: primitives.ElementToBigInt(n.DailyLimit), MonthlyLimit: primitives.ElementToBigInt(n.MonthlyLimit), AvailableCredit: primitives.ElementToBigInt(n.AvailableCredit), Unlimited: n.Unlimited, LastDay: primitives.ElementToBigInt(n.LastDay).Uint64(), LastMonth: primitives.ElementToBigInt(n.LastMonth).Uint64(), DailySpent: primitives.ElementToBigInt(n.DailySpent), MonthlySpent: primitives.ElementToBigInt(n.MonthlySpent), TotalSpent: primitives.ElementToBigInt(n.TotalSpent)})
	}
	for _, record := range state.Operations {
		prepared, err := credit.Restore(ctx, record.Intent, nil, nil)
		if err != nil {
			return History{}, err
		}
		preview := prepared.Preview()
		result.Operations = append(result.Operations, OperationSummary{record.ID, preview.Kind, record.Status, record.TransactionHash, record.Externalized, preview})
	}
	sort.Slice(result.Grants, func(i, j int) bool {
		return GrantKey(result.Grants[i].Commitment, result.Grants[i].Role) < GrantKey(result.Grants[j].Commitment, result.Grants[j].Role)
	})
	sort.Slice(result.Operations, func(i, j int) bool { return result.Operations[i].ID < result.Operations[j].ID })
	return result, nil
}
