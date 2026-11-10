package protocol

import (
	"context"

	"github.com/ethereum/go-ethereum"
)

// GasEstimator uses eth_estimateGas for the final EVM call.
type GasEstimator interface {
	EstimateGas(context.Context, ethereum.CallMsg) (uint64, error)
}

type GasEstimate struct {
	Source   string `json:"source"`
	GasUnits uint64 `json:"gasUnits"`
}
