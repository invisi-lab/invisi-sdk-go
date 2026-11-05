package transaction

import (
	"context"
	"errors"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/invisi-lab/invisi-sdk-go/core"
	gasmodel "github.com/invisi-lab/invisi-sdk-go/internal/gas"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type Provider interface {
	protocolrpc.ChainProvider
	TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error)
}

type InvisiOptions struct {
	GasBenchmarks   []gasmodel.GasBenchmarkProfile
	ProtocolRPC     *protocolrpc.ProtocolRpcClient
	Provider        Provider
	Prover          Prover
	FundingVerifier FundingAuthorizationVerifier
}

type Invisi struct{ options InvisiOptions }

func NewInvisi(options InvisiOptions) *Invisi {
	options.GasBenchmarks = gasmodel.CopyGasProfiles(options.GasBenchmarks)
	return &Invisi{options}
}

func (s *Runtime) CoreContext() core.ProtocolContext {
	return core.ProtocolContext{ChainID: core.ElementFromUint64(s.Scope.ChainID), ProtocolAddress: core.BigIntToElement(s.Scope.ProtocolAddress.Big())}
}

func (s *Runtime) Canonical(ctx context.Context, block protocolrpc.CanonicalBlock) error {
	h, e := s.Options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return e
	}
	if h == nil || h.Number == nil || h.Number.Uint64() != block.Number || h.Hash() != block.Hash || h.ParentHash != block.ParentHash {
		return errors.New("canonical wallet checkpoint changed")
	}
	return nil
}

func CreditKey(hash common.Hash, role string) string { return hash.Hex() + ":" + role }

func OrderedNotes(state Inventory, token common.Address) []NoteRecord {
	result := []NoteRecord{}
	for _, n := range state.Notes {
		if !n.Spent && n.ReservedBy == "" && core.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) == 0 {
			result = append(result, n)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := core.ElementToBigInt(result[i].Note.Amount), core.ElementToBigInt(result[j].Note.Amount)
		if a.Cmp(b) == 0 {
			return result[i].Commitment.Hex() < result[j].Commitment.Hex()
		}
		return a.Cmp(b) < 0
	})
	return result
}

// preparationInventory reads application inputs directly in explicit mode. It never implements or calls a WalletStore.
func (s *Runtime) PreparationInventory(ctx context.Context) (Inventory, error) {
	if err := ctx.Err(); err != nil {
		return Inventory{}, err
	}
	if s.ExplicitInventory != nil {
		return *s.ExplicitInventory, nil
	}
	return s.InventoryReader(ctx)
}
