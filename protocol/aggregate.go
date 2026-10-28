package protocol

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type AggregateRequest struct {
	Proof    []byte
	Requests []TransactRequest
	Caller   *common.Address
}
type TransactionCall struct {
	To    common.Address
	Data  []byte
	Value *big.Int
}

func (p *Protocol) Aggregate(request AggregateRequest) (TransactionCall, error) {
	chain, _ := p.deployment.ChainNumber()
	encoded, e := BuildTransactAggregateCalldata(BuildTransactAggregateCalldataParams{Context: TransactionContext{ChainID: chain, ProtocolAddress: p.deployment.Protocol.Address}, Proof: request.Proof, Requests: request.Requests, Caller: request.Caller})
	if e != nil {
		return TransactionCall{}, e
	}
	return TransactionCall{To: p.deployment.Protocol.Address, Data: encoded.Data, Value: new(big.Int)}, nil
}
