package transaction

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func (p *PreparedTransaction) EstimateGas(ctx context.Context, from common.Address) (protocol.GasEstimate, error) {
	provider, ok := p.operation.Runtime.Options.Provider.(protocol.GasEstimator)
	if !ok {
		return protocol.GasEstimate{}, &MissingCapabilityError{Capability: "provider for final eth_estimateGas"}
	}
	call, err := p.ExportTransaction(ctx)
	if err != nil {
		return protocol.GasEstimate{}, err
	}
	return protocol.EstimateTransactionGas(ctx, provider, call, from)
}
