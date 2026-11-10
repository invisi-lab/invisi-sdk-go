package credit

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	gasrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (p *Prepared) EstimateGas(ctx context.Context, provider protocol.GasEstimator, from common.Address) (protocol.GasEstimate, error) {
	if err := ctx.Err(); err != nil {
		return protocol.GasEstimate{}, err
	}
	if provider == nil {
		return protocol.GasEstimate{}, errors.New("provider for final eth_estimateGas required")
	}
	data, err := p.Calldata(from)
	if err != nil {
		return protocol.GasEstimate{}, err
	}
	p.mu.Lock()
	to := common.BigToAddress(primitives.ElementToBigInt(p.intent.Context.ProtocolAddress))
	p.mu.Unlock()
	return gasrpc.EstimateTransactionGas(ctx, provider, protocol.TransactionCall{To: to, Data: data, Value: new(big.Int)}, from)
}
