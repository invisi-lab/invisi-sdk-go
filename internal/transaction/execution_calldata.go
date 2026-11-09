package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func (o *Execution) Deliveries() []protocol.Delivery {
	if o.Plan.Kind != "transact" && o.Plan.Kind != "shield" {
		return o.Plan.Deliveries
	}
	result := []protocol.Delivery{}
	for _, out := range o.Plan.Outputs {
		result = append(result, out.Delivery)
	}
	return result
}

func (o *Execution) Words() ([][32]byte, error) {
	switch o.Plan.Kind {
	case "transact":
		return protocol.TransactUserPublicInputs(o.Plan.Context, o.Plan.Params, o.Deliveries(), o.Plan.Funding)

	}
	return nil, errors.New("operation has no private proof")
}

func (o *Execution) ReadyInternal(ctx context.Context) (*Execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return nil, e
	}
	if o.Plan.Kind != "shield" || o.Status != "prepared" {
		return nil, errors.New("Ready is for prepared shield operations")
	}
	if o.Plan.FundingMode == "signature" {
		return nil, errors.New("authorize shield funding before readiness")
	}
	if e := o.Encode(); e != nil {
		return nil, e
	}
	o.Status = "ready"
	return o, o.Persist(ctx, o.Status)
}

func (o *Execution) ShieldParams() protocol.BuildShieldCalldataParams {
	outputs := []protocol.ShieldOutput{}
	for _, out := range o.Plan.Outputs {
		outputs = append(outputs, protocol.ShieldOutput{Identity: out.Identity, Amount: new(big.Int).Set(out.Amount)})
	}
	var caller *common.Address
	if o.Runtime.Options.Signer != nil {
		address := o.Runtime.Options.Signer.From
		caller = &address
	}
	return protocol.BuildShieldCalldataParams{Context: o.Plan.Context, Params: o.Plan.Shield, Outputs: outputs, Deliveries: o.Deliveries(), Funding: o.Plan.Funding, Caller: caller}
}

func (o *Execution) Encode() error {
	var call *protocol.Calldata
	var e error
	var caller *common.Address
	if o.Runtime.Options.Signer != nil {
		address := o.Runtime.Options.Signer.From
		caller = &address
	}
	proof := []byte(nil)
	if o.Bundle != nil {
		proof = o.Bundle.Proof
	}
	switch o.Plan.Kind {
	case "shield":
		call, e = protocol.BuildShieldCalldata(o.ShieldParams())
	case "transact":
		call, e = protocol.BuildTransactCalldata(protocol.BuildTransactCalldataParams{Context: o.Plan.Context, Proof: proof, Params: o.Plan.Params, Deliveries: o.Deliveries(), Funding: o.Plan.Funding, Caller: caller})
	default:
		return errors.New("unsupported Note operation")
	}
	if e != nil {
		return e
	}
	o.Data = append([]byte(nil), call.Data...)
	return nil
}
