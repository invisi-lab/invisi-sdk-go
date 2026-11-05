package sdk

import (
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

func (o *Operation) ExportAggregateMember(ctx context.Context) (OperationAggregateMember, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return OperationAggregateMember{}, e
	}
	if o.plan.Broadcaster != nil {
		return OperationAggregateMember{}, errors.New("broadcaster operations cannot be aggregate members")
	}
	if (o.status != "proved" && o.status != "ready") || o.plan.Kind != "transact" || o.plan.Target != "noir-recursive" || o.plan.Capacity != 14 || o.bundle == nil || len(o.bundle.Proof) != 458*32 {
		return OperationAggregateMember{}, errors.New("ready recursive slots14 member required")
	}
	if o.plan.FundingMode == "signature" && len(o.plan.Funding.Signature) == 0 {
		return OperationAggregateMember{}, errors.New("aggregate funding authorization required")
	}
	words, e := o.words()
	if e != nil {
		return OperationAggregateMember{}, e
	}
	inputs := []common.Hash{}
	for _, word := range words {
		inputs = append(inputs, word)
	}
	statement, e := prover.Statement(o.circuitName(), "noir-recursive", inputs)
	if e != nil {
		return OperationAggregateMember{}, e
	}
	if e = prover.AssertProofBundle(*o.bundle, statement); e != nil {
		return OperationAggregateMember{}, e
	}
	member := OperationAggregateMember{*o.bundle, statement, o.plan.Context, protocol.TransactRequest{Params: o.plan.Params, Deliveries: o.deliveries(), Funding: o.plan.Funding}}
	raw, e := json.Marshal(member)
	if e != nil {
		return OperationAggregateMember{}, e
	}
	var result OperationAggregateMember
	e = json.Unmarshal(raw, &result)
	return result, e
}
func (o *Operation) matchesTransactionData(data []byte) bool {
	if o.plan.Target != "noir-recursive" {
		return common.Bytes2Hex(data) == common.Bytes2Hex(o.data)
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
	want, _ := json.Marshal(protocol.TransactRequest{Params: o.plan.Params, Deliveries: o.deliveries(), Funding: o.plan.Funding})
	matches := 0
	for _, request := range requests {
		got, _ := json.Marshal(request)
		if string(got) == string(want) {
			matches++
		}
	}
	return matches == 1
}
