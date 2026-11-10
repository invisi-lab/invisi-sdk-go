package rpc

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func EstimateTransactionGas(ctx context.Context, provider protocol.GasEstimator, call protocol.TransactionCall, from common.Address) (protocol.GasEstimate, error) {
	if err := ctx.Err(); err != nil {
		return protocol.GasEstimate{}, err
	}
	if provider == nil {
		return protocol.GasEstimate{}, errors.New("provider for final eth_estimateGas required")
	}
	if call.To == (common.Address{}) || len(call.Data) == 0 {
		return protocol.GasEstimate{}, errors.New("final EVM transaction required")
	}
	value := new(big.Int)
	if call.Value != nil {
		value.Set(call.Value)
	}
	to := call.To
	units, err := provider.EstimateGas(ctx, ethereum.CallMsg{From: from, To: &to, Data: append([]byte(nil), call.Data...), Value: value})
	if err != nil {
		return protocol.GasEstimate{}, err
	}
	if err := ctx.Err(); err != nil {
		return protocol.GasEstimate{}, err
	}
	return protocol.GasEstimate{Source: "rpc", GasUnits: units}, nil
}
