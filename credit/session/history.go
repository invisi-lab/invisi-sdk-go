package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"math/big"
	"sort"
)

type GrantSummary struct {
	Commitment                                           common.Hash
	Owner, Spender, Token                                common.Address
	Role                                                 credit.Role
	Shard                                                uint64
	Slot                                                 uint16
	Expiry                                               uint64
	Consumed, Revoked                                    bool
	ReservedBy                                           string
	SingleLimit, DailyLimit, MonthlyLimit, LifetimeLimit *big.Int
	LastDay, LastMonth                                   uint64
	DailySpent, MonthlySpent, TotalSpent                 *big.Int
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
		result.Grants = append(result.Grants, GrantSummary{Commitment: record.Commitment, Owner: common.BigToAddress(core.ElementToBigInt(n.OwnerWalletIdentity)), Spender: common.BigToAddress(core.ElementToBigInt(n.SpenderWalletIdentity)), Token: common.BigToAddress(core.ElementToBigInt(n.TokenAddress)), Role: record.Role, Shard: core.ElementToBigInt(n.RevocationShard).Uint64(), Slot: uint16(core.ElementToBigInt(n.RevocationSlot).Uint64()), Expiry: core.ElementToBigInt(n.Expiry).Uint64(), Consumed: record.Consumed, Revoked: record.Revoked, ReservedBy: record.ReservedBy, SingleLimit: core.ElementToBigInt(n.SingleLimit), DailyLimit: core.ElementToBigInt(n.DailyLimit), MonthlyLimit: core.ElementToBigInt(n.MonthlyLimit), LifetimeLimit: core.ElementToBigInt(n.LifetimeLimit), LastDay: core.ElementToBigInt(n.LastDay).Uint64(), LastMonth: core.ElementToBigInt(n.LastMonth).Uint64(), DailySpent: core.ElementToBigInt(n.DailySpent), MonthlySpent: core.ElementToBigInt(n.MonthlySpent), TotalSpent: core.ElementToBigInt(n.TotalSpent)})
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
