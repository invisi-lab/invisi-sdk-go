package transaction

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func (o *Execution) FundingTypedData() (apitypes.TypedData, error) {
	switch o.Plan.Kind {
	case "shield":
		return protocol.EncodeShieldFundingTypedData(o.ShieldParams())
	case "transact":
		if o.Plan.Target == "noir-recursive" {
			return protocol.EncodeAggregateFundingTypedData(o.Plan.Context, protocol.TransactRequest{Params: o.Plan.Params, Deliveries: o.Deliveries(), Funding: o.Plan.Funding})
		}

		return protocol.EncodeTransactFundingTypedData(o.Plan.Context, o.Bundle.Proof, o.Plan.Params, o.Deliveries(), o.Plan.Funding)

	}
	return apitypes.TypedData{}, errors.New("operation has no public funding")
}

func (o *Execution) AuthorizeFundingInternal(ctx context.Context) (*Execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	typed, err := o.FundingRequest(ctx)
	if err != nil {
		return nil, err
	}
	signer := o.Runtime.Options.FundingSigner
	if signer == nil {
		return nil, &MissingCapabilityError{Capability: "EVM funding signer"}
	}
	if signer.Address() != o.Plan.Funding.Payer {
		return nil, errors.New("funding signer differs from payer")
	}
	copied, err := Frozen(typed)
	if err != nil {
		return nil, err
	}
	signature, err := signer.SignTypedData(ctx, copied)
	if err != nil {
		return nil, err
	}
	if err = o.AttachFunding(ctx, signature); err != nil {
		return nil, err
	}
	return o, nil
}
