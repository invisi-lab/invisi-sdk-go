package transaction

import (
	"context"
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func (o *Execution) AuthorizeInternal(ctx context.Context) (*Execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return nil, e
	}
	if e := o.CheckDrawPeriod(ctx); e != nil {
		return nil, e
	}
	if o.Status != "prepared" {
		return nil, errors.New("operation must be prepared before authorization")
	}
	if !o.Plan.NeedsAuthorization {
		return o, nil
	}
	if o.Runtime.Authorization == nil {
		return nil, errors.New("wallet authorization permission required")
	}
	words, e := o.Words()
	if e != nil {
		return nil, e
	}
	base := words[:len(words)-2]
	r, e := protocol.WalletAuthorizationRequest(o.Plan.Context, core.WalletOperation(o.Plan.Kind), o.Plan.Capacity, base)
	if e != nil {
		return nil, e
	}
	requestCopy := r
	requestCopy.Words = append([]fr.Element(nil), r.Words...)
	w, e := o.Runtime.Authorization.Authorize(ctx, requestCopy)
	if e != nil {
		return nil, e
	}
	if e = VerifyAuthorizationWitness(o.Runtime.Wallet.identity, r, w); e != nil {
		return nil, e
	}
	o.Authorization = &w
	o.Status = "authorized"
	if e = o.Persist(ctx, o.Status); e != nil {
		return nil, e
	}
	return o, nil
}

func (o *Execution) FundingTypedData() (apitypes.TypedData, error) {
	switch o.Plan.Kind {
	case "shield":
		return protocol.EncodeShieldFundingTypedData(o.ShieldParams())
	case "transact":
		if o.Plan.Target == "noir-recursive" {
			return protocol.EncodeAggregateFundingTypedData(o.Plan.Context, protocol.TransactRequest{Params: o.Plan.Params, Deliveries: o.Deliveries(), Funding: o.Plan.Funding})
		}

		return protocol.EncodeTransactFundingTypedData(o.Plan.Context, o.Bundle.Proof, o.Plan.Params, o.Deliveries(), o.Plan.Funding)
	case "create-credit":
		return protocol.EncodeCreateCreditFundingTypedData(o.Plan.Context, *o.Plan.CreateCredit, o.Deliveries(), o.Plan.Funding)
	case "top-up-credit":
		return protocol.EncodeTopUpCreditFundingTypedData(o.Plan.Context, *o.Plan.TopUpCredit, o.Deliveries(), o.Plan.Funding)
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
