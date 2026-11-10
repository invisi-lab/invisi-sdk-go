package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func (o *Operation) EstimateGas(ctx context.Context, from common.Address) (protocol.GasEstimate, error) {
	provider, ok := o.session.options.Provider.(protocol.GasEstimator)
	if !ok {
		return protocol.GasEstimate{}, &sdk.MissingCapabilityError{Capability: "provider for final eth_estimateGas"}
	}
	call, err := o.ExportTransaction(ctx)
	if err != nil {
		return protocol.GasEstimate{}, err
	}
	return protocol.EstimateTransactionGas(ctx, provider, call, from)
}
