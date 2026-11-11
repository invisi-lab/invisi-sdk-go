package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type OperationAggregateMember struct {
	Bundle    prover.ProofBundle
	Statement prover.ProofStatement
	Context   protocol.TransactionContext
	Request   protocol.TransactRequest
}

func (o *Execution) ExportAggregateMember(ctx context.Context) (OperationAggregateMember, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return OperationAggregateMember{}, e
	}
	if o.Plan.Broadcaster != nil {
		return OperationAggregateMember{}, errors.New("broadcaster operations cannot be aggregate members")
	}
	if (o.Status != "proved" && o.Status != "ready") || o.Plan.Kind != "transact" || o.Plan.Target != "noir-recursive" || (o.Plan.Capacity < 1 || o.Plan.Capacity > 10) || o.Bundle == nil || len(o.Bundle.Proof) != 458*32 {
		return OperationAggregateMember{}, errors.New("ready recursive C1..C10 member required")
	}
	if o.Plan.FundingMode == "signature" && len(o.Plan.Funding.Signature) == 0 {
		return OperationAggregateMember{}, errors.New("aggregate funding authorization required")
	}
	words, e := o.Words()
	if e != nil {
		return OperationAggregateMember{}, e
	}
	inputs := []common.Hash{}
	for _, word := range words {
		inputs = append(inputs, word)
	}
	statement, e := prover.Statement(o.CircuitName(), "noir-recursive", inputs)
	if e != nil {
		return OperationAggregateMember{}, e
	}
	if e = prover.AssertProofBundle(*o.Bundle, statement); e != nil {
		return OperationAggregateMember{}, e
	}
	member := OperationAggregateMember{*o.Bundle, statement, o.Plan.Context, protocol.TransactRequest{Params: o.Plan.Params, Deliveries: o.Deliveries(), Funding: o.Plan.Funding}}
	raw, e := json.Marshal(member)
	if e != nil {
		return OperationAggregateMember{}, e
	}
	var result OperationAggregateMember
	e = json.Unmarshal(raw, &result)
	return result, e
}

func (o *Execution) MatchesTransactionData(data []byte) bool {
	if o.Plan.Target != "noir-recursive" {
		return common.Bytes2Hex(data) == common.Bytes2Hex(o.Data)
	}
	a, e := protocol.ProtocolABI()
	if e != nil || len(data) < 4 {
		return false
	}
	method, e := a.MethodById(data[:4])
	if e != nil || method.RawName != "transact" || len(method.Inputs) != 2 || method.Inputs[1].Type.T != abi.SliceTy {
		return false
	}
	values, e := method.Inputs.Unpack(data[4:])
	if e != nil {
		return false
	}
	requests := *abi.ConvertType(values[1], new([]protocol.TransactRequest)).(*[]protocol.TransactRequest)
	encoded, err := method.Inputs.Pack(values...)
	if err != nil || !bytes.Equal(encoded, data[4:]) {
		return false
	}
	argument := abi.Arguments{{Type: *method.Inputs[1].Type.Elem}}
	want, err := argument.Pack(protocol.TransactRequest{Params: o.Plan.Params, Deliveries: o.Deliveries(), Funding: o.Plan.Funding})
	if err != nil {
		return false
	}
	matches := 0
	for _, request := range requests {
		got, err := argument.Pack(request)
		if err == nil && bytes.Equal(got, want) {
			matches++
		}
	}
	return matches == 1
}
