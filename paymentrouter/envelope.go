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
