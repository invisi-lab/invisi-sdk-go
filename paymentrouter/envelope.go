package paymentrouter

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/rpc/broadcaster"
	"math/big"
)

type RouterEnvelope = broadcaster.RouterEnvelope

func EnvelopeHash(chainID *big.Int, core common.Address, env RouterEnvelope) (common.Hash, error) {
	return broadcaster.EnvelopeHash(chainID, core, env)
}

func RouterIntentHash(chainID *big.Int, core common.Address, env RouterEnvelope) (common.Hash, error) {
	return broadcaster.RouterIntentHash(chainID, core, env)
}
func NormalizeRouterEnvelope(env RouterEnvelope) (RouterEnvelope, error) {
	return broadcaster.NormalizeRouterEnvelope(env)
}

type RouterGasInput = broadcaster.RouterGasInput
type RouterGasResult = broadcaster.RouterGasResult
type RouterGasBreakdown = broadcaster.RouterGasBreakdown

const RouterGasModelID = broadcaster.RouterGasModelID

func EstimateRouterGas(input RouterGasInput) RouterGasResult {
	return broadcaster.EstimateRouterGas(input)
}
