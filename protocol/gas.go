package protocol

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// GasEstimator uses eth_estimateGas for the final EVM call.
type GasEstimator interface {
	EstimateGas(context.Context, ethereum.CallMsg) (uint64, error)
}

type GasEstimate struct {
	Source   string `json:"source"`
	GasUnits uint64 `json:"gasUnits"`
}

func EstimateTransactionGas(ctx context.Context, provider GasEstimator, call TransactionCall, from common.Address) (GasEstimate, error) {
	if err := ctx.Err(); err != nil {
		return GasEstimate{}, err
	}
	if provider == nil {
		return GasEstimate{}, errors.New("provider for final eth_estimateGas required")
	}
	if call.To == (common.Address{}) || len(call.Data) == 0 {
		return GasEstimate{}, errors.New("final EVM transaction required")
	}
	value := new(big.Int)
	if call.Value != nil {
		value.Set(call.Value)
	}
	to := call.To
	units, err := provider.EstimateGas(ctx, ethereum.CallMsg{From: from, To: &to, Data: append([]byte(nil), call.Data...), Value: value})
	if err != nil {
		return GasEstimate{}, err
	}
	if err := ctx.Err(); err != nil {
		return GasEstimate{}, err
	}
	return GasEstimate{Source: "rpc", GasUnits: units}, nil
}
