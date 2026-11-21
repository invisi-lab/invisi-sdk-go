package paymentrouter

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

// GethProvider adapts an existing ethclient without adding signing capabilities.
// eth_estimateGas receives an explicit block tag, unlike ethclient.EstimateGas.
type GethProvider struct{ *ethclient.Client }

func (p GethProvider) EstimateGasAt(ctx context.Context, msg ethereum.CallMsg, block *big.Int) (uint64, error) {
	if p.Client == nil || block == nil || block.Sign() < 0 {
		return 0, fmt.Errorf("client and pinned block required")
	}
	args := map[string]any{"from": msg.From, "data": hexutil.Bytes(msg.Data)}
	if msg.To != nil {
		args["to"] = *msg.To
	}
	if msg.Value != nil {
		args["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		args["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		args["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	if msg.GasFeeCap != nil {
		args["maxFeePerGas"] = (*hexutil.Big)(msg.GasFeeCap)
	}
	if msg.GasTipCap != nil {
		args["maxPriorityFeePerGas"] = (*hexutil.Big)(msg.GasTipCap)
	}
	if len(msg.AuthorizationList) > 0 {
		args["authorizationList"] = msg.AuthorizationList
		args["type"] = hexutil.Uint64(4)
	}
	var out hexutil.Uint64
	e := p.Client.Client().CallContext(ctx, &out, "eth_estimateGas", args, hexutil.EncodeBig(block))
	return uint64(out), e
}

var _ Provider = GethProvider{}
