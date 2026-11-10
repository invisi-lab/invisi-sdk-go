package chain

import (
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func (p *Protocol) Aggregate(request protocol.AggregateRequest) (protocol.TransactionCall, error) {
	chain, _ := p.deployment.ChainNumber()
	encoded, e := protocol.BuildTransactAggregateCalldata(protocol.BuildTransactAggregateCalldataParams{Context: protocol.TransactionContext{ChainID: chain, ProtocolAddress: p.deployment.Protocol.Address}, Proof: request.Proof, Requests: request.Requests, Caller: request.Caller})
	if e != nil {
		return protocol.TransactionCall{}, e
	}
	return protocol.TransactionCall{To: p.deployment.Protocol.Address, Data: encoded.Data, Value: new(big.Int)}, nil
}
