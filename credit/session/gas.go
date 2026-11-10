package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
)

func (o *Operation) EstimateGas(ctx context.Context, provider protocol.GasEstimator, from common.Address) (protocol.GasEstimate, error) {
	if provider == nil {
		return protocol.GasEstimate{}, errors.New("provider for final eth_estimateGas required")
	}
	data, err := o.Calldata(ctx, from)
	if err != nil {
		return protocol.GasEstimate{}, err
	}
	intent, err := o.prepared.ExportIntent()
	if err != nil {
		return protocol.GasEstimate{}, err
	}
	to := common.BigToAddress(core.ElementToBigInt(intent.Context.ProtocolAddress))
	return protocol.EstimateTransactionGas(ctx, provider, protocol.TransactionCall{To: to, Data: data, Value: new(big.Int)}, from)
}
