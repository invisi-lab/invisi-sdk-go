package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
)

type FundingPermitRequest struct {
	Payer, TokenAddress, ProtocolAddress common.Address
	Amount                               *big.Int
}
type BroadcasterFeeSource string

const (
	BroadcasterFeePublicFunding BroadcasterFeeSource = "public-funding"
	BroadcasterFeePrivateInputs BroadcasterFeeSource = "private-inputs"
	BroadcasterFeePayout        BroadcasterFeeSource = "payout"
)

type WithBroadcasterOptions struct {
	MaxFeeAmount  *big.Int
	MaxAttempts   int
	FeeSource     BroadcasterFeeSource
	FundingPermit func(context.Context, FundingPermitRequest) (*protocol.TokenPermit, error)
}
type broadcasterPlan struct {
	Configuration broadcaster.Configuration
	Signed        broadcaster.SignedQuote
	ChainID       *big.Int
	Protocol      common.Address
	Intent        []byte
}

func planFee(p operationPlan) (common.Address, *big.Int) {
	var r common.Address
	var f *big.Int
	switch p.Kind {
	case "shield":
		r, f = p.Shield.FeeRecipient, p.Shield.FeeAmount
	case "transact":
		r, f = p.Params.FeeRecipient, p.Params.FeeAmount
	case "create-credit":
		r, f = p.CreateCredit.FeeRecipient, p.CreateCredit.FeeAmount
	case "top-up-credit":
		r, f = p.TopUpCredit.Creation.FeeRecipient, p.TopUpCredit.Creation.FeeAmount
	case "cancel-credit":
		r, f = p.CancelCredit.FeeRecipient, p.CancelCredit.FeeAmount
	}
	if f == nil {
		return r, new(big.Int)
	}
	return r, new(big.Int).Set(f)
}
func setPlanFee(p *operationPlan, r common.Address, f *big.Int) {
	switch p.Kind {
	case "shield":
		p.Shield.FeeRecipient, p.Shield.FeeAmount = r, new(big.Int).Set(f)
	case "transact":
		p.Params.FeeRecipient, p.Params.FeeAmount = r, new(big.Int).Set(f)
	case "create-credit", "top-up-credit":
		p.CreateCredit.FeeRecipient, p.CreateCredit.FeeAmount = r, new(big.Int).Set(f)
		if p.TopUpCredit != nil {
			p.TopUpCredit.Creation = *p.CreateCredit
		}
	case "cancel-credit":
		p.CancelCredit.FeeRecipient, p.CancelCredit.FeeAmount = r, new(big.Int).Set(f)
	}
}
func (o *Operation) broadcasterIntent() ([]byte, error) {
	a, e := protocol.ProtocolABI()
	if e != nil {
		return nil, e
	}
	f, e := protocol.NormalizeFunding(o.plan.Funding)
	if e != nil {
		return nil, e
	}
	f.Signature = nil
	switch o.plan.Kind {
	case "shield":
		p := o.shieldParams()
		if p.Params.FeeAmount == nil {
			p.Params.FeeAmount = new(big.Int)
		}
		return a.Pack("shield", p.Params, p.Outputs, p.Deliveries, f)
	case "transact":
		r, e := protocol.NormalizeTransactRequest(o.plan.Context, o.plan.Params, o.deliveries(), f)
		if e != nil {
			return nil, e
		}
		for _, m := range a.Methods {
			if m.RawName == "transact" && len(m.Inputs) == 2 && m.Inputs[1].Type.T == abi.TupleTy {
				return a.Pack(m.Name, []byte{}, protocol.TransactRequest{Params: r.Params, Deliveries: o.deliveries(), Funding: r.Funding})
			}
		}
	case "create-credit":
		r, e := protocol.ValidateCreateCreditRequest(o.plan.Context, *o.plan.CreateCredit, o.deliveries(), f)
		if e != nil {
			return nil, e
		}
		return a.Pack("createCredit", []byte{}, r.Params, o.deliveries(), r.Funding)
	case "top-up-credit":
		r, e := protocol.ValidateTopUpCreditRequest(o.plan.Context, *o.plan.TopUpCredit, o.deliveries(), f)
		if e != nil {
			return nil, e
		}
		return a.Pack("topUpCredit", []byte{}, r.Params, o.deliveries(), r.Funding)
	case "cancel-credit":
		r, e := protocol.ValidateCancelCreditRequest(o.plan.Context, *o.plan.CancelCredit)
		if e != nil {
			return nil, e
		}
		return a.Pack("cancelCredit", []byte{}, r.Params)
	}
	return nil, errors.New("unsupported broadcaster operation")
}
func (o *Operation) applyBroadcasterFee(ctx context.Context, base operationPlan, recipient common.Address, fee *big.Int, opt WithBroadcasterOptions) error {
	if base.Kind != "shield" && base.Kind != "cancel-credit" && opt.FeeSource != BroadcasterFeePublicFunding && (base.Funding.Amount == nil || base.Funding.Amount.Sign() == 0 || opt.FeeSource == BroadcasterFeePrivateInputs) {
		var err error
		base, err = o.extendBroadcasterInputs(ctx, base, fee)
		if err != nil {
			return err
		}
	}
	p, e := clonePlan(base)
	if e != nil {
		return e
	}
	o.plan = p
	setPlanFee(&o.plan, recipient, fee)
	if p.Kind == "cancel-credit" {
		if fee.Cmp(p.CancelCredit.Amount) >= 0 {
			return errors.New("broadcaster fee consumes cancellation payout")
		}
		return nil
	}
	change := -1
	for j, out := range p.Outputs {
		if out.Change && out.Amount.Cmp(fee) >= 0 {
			change = j
			break
		}
	}
	if opt.FeeSource == BroadcasterFeePublicFunding {
		change = -1
	}
	if change < 0 && opt.FeeSource == BroadcasterFeePrivateInputs {
		return errors.New("insufficient existing private change for fee")
	}
	if change >= 0 {
		out := o.plan.Outputs[change]
		amount := new(big.Int).Sub(out.Amount, fee)
		if amount.Sign() == 0 {
			o.plan.Outputs = append(o.plan.Outputs[:change], o.plan.Outputs[change+1:]...)
		} else {
			if out.Note == nil {
				return errors.New("missing private change note")
			}
			note := *out.Note
			note.Amount = core.BigIntToElement(amount)
			commit, e := core.ComputeNoteCommitment(note.ChainID, core.BigIntToElement(out.Identity.Big()), note.TokenAddress, note.Amount)
			if e != nil {
				return e
			}
			recipient, e := o.session.resolve(ctx, out.Address)
			if e != nil {
				return e
			}
			delivery, e := encryptOutput(recipient, note, fieldHash(commit))
			if e != nil {
				return e
			}
			out.Note = &note
			out.Amount = amount
			out.Commitment = fieldHash(commit)
			out.Delivery = delivery
			o.plan.Outputs[change] = out
		}
		if p.Kind == "transact" {
			o.plan.Params.OutputCommitments = nil
			for _, out := range o.plan.Outputs {
				o.plan.Params.OutputCommitments = append(o.plan.Params.OutputCommitments, out.Commitment)
			}
			o.plan.Capacity, e = protocol.SelectTransactCapacity(len(p.Inputs), len(o.plan.Outputs))
			if e != nil {
				return e
			}
		} else {
			o.plan.CreateCredit.ChangeCommitment = [32]byte{}
			o.plan.Deliveries[2] = protocol.Delivery{}
			if len(o.plan.Outputs) > 0 {
				o.plan.CreateCredit.ChangeCommitment = o.plan.Outputs[0].Commitment
				o.plan.Deliveries[2] = o.plan.Outputs[0].Delivery
			}
			if o.plan.TopUpCredit != nil {
				o.plan.TopUpCredit.Creation = *o.plan.CreateCredit
			}
		}
		o.plan.NeedsAuthorization = len(o.plan.Inputs) > 0 || p.Kind != "transact"
		if p.Funding.Amount.Sign() > 0 && p.FundingMode != "signature" {
			source := FundingSource{Payer: p.Funding.Payer, Amount: new(big.Int).Set(p.Funding.Amount), Authorization: "signature"}
			if p.Funding.Permit.Deadline != nil && p.Funding.Permit.Deadline.Sign() > 0 {
				source.Permit = &p.Funding.Permit
			}
			o.plan.Funding, o.plan.FundingMode, e = o.session.funding(ctx, &source, p.Block, p.Deadline)
			if e != nil {
				return e
			}
		}
		return nil
	}
	gross := new(big.Int).Add(p.Funding.Amount, fee)
	payer := p.Funding.Payer
	if payer == (common.Address{}) {
		return errors.New("public funding fee requires an explicit payer")
	}
	source := FundingSource{Payer: payer, Amount: gross, Authorization: "signature", Deadline: p.Deadline}
	if p.FundingMode == "signature" {
		source.SigNonce = p.Funding.SigNonce
		if p.Funding.Deadline != nil {
			source.Deadline = p.Funding.Deadline.Uint64()
		}
	}
	if p.Funding.Permit.Deadline != nil && p.Funding.Permit.Deadline.Sign() > 0 && opt.FundingPermit == nil {
		return errors.New("gross funding requires a fresh token permit callback")
	}

	o.plan.Funding, o.plan.FundingMode, e = o.session.funding(ctx, &source, p.Block, p.Deadline)
	return e
}
func (o *Operation) withBroadcaster(ctx context.Context, c *broadcaster.Client, opt WithBroadcasterOptions) error {
	if c == nil {
		return errors.New("broadcaster client required")
	}
	if o.status != "prepared" || o.bundle != nil || o.authorization != nil || o.plan.Broadcaster != nil || o.plan.Target == "noir-recursive" {
		return errors.New("broadcaster must be selected before authorization and proving")
	}
	switch o.plan.Kind {
	case "shield", "transact", "create-credit", "top-up-credit", "cancel-credit":
	default:
		return errors.New("unsupported broadcaster operation")
	}
	switch opt.FeeSource {
	case "", BroadcasterFeePublicFunding, BroadcasterFeePrivateInputs, BroadcasterFeePayout:
	default:
		return errors.New("unsupported fee source")
	}
	if o.plan.Kind == "cancel-credit" {
		if opt.FeeSource != "" && opt.FeeSource != BroadcasterFeePayout {
			return errors.New("cancellation fee requires payout source")
		}
	} else if opt.FeeSource == BroadcasterFeePayout || o.plan.Kind == "shield" && opt.FeeSource == BroadcasterFeePrivateInputs {
		return errors.New("fee source unavailable for operation")
	}
	if positive(opt.MaxFeeAmount, 128) != nil {
		return errors.New("max fee must be positive uint128")
	}
	opt.MaxFeeAmount = new(big.Int).Set(opt.MaxFeeAmount)
	if opt.MaxAttempts == 0 {
		opt.MaxAttempts = 4
	}
	if opt.MaxAttempts < 1 || opt.MaxAttempts > 16 {
		return errors.New("max attempts must be 1..16")
	}
	base, e := clonePlan(o.plan)
	if e != nil {
		return e
	}
	initial, e := o.broadcasterIntent()
	if e != nil {
		return e
	}
	fee, e := c.Fee(ctx, o.plan.Context.TokenAddress, initial)
	if e != nil {
		return e
	}
	amount, _ := new(big.Int).SetString(fee.FeeAmount, 10)
	for attempt := 0; attempt < opt.MaxAttempts; attempt++ {
		if amount.Cmp(opt.MaxFeeAmount) > 0 {
			return errors.New("broadcaster fee exceeds maximum")
		}
		if e = o.applyBroadcasterFee(ctx, base, c.Configuration().FeeRecipient, amount, opt); e != nil {
			return e
		}
		data, e := o.broadcasterIntent()
		if e != nil {
			return e
		}
		required, e := c.Fee(ctx, o.plan.Context.TokenAddress, data)
		if e != nil {
			return e
		}
		next, _ := new(big.Int).SetString(required.FeeAmount, 10)
		if next.Cmp(amount) > 0 {
			amount = next
			continue
		}
		if opt.FundingPermit != nil {
			if o.plan.Funding.Amount.Sign() == 0 {
				return errors.New("fundingPermit requires explicit public funding")
			}
			permit, err := opt.FundingPermit(ctx, FundingPermitRequest{Payer: o.plan.Funding.Payer, Amount: new(big.Int).Set(o.plan.Funding.Amount), TokenAddress: o.plan.Context.TokenAddress, ProtocolAddress: o.plan.Context.ProtocolAddress})
			if err != nil {
				return err
			}
			if permit == nil {
				return errors.New("fundingPermit returned nil")
			}
			o.plan.Funding.Permit = *permit
			o.plan.Funding, e = protocol.NormalizeFunding(o.plan.Funding)
			if e != nil {
				return e
			}
			data, e = o.broadcasterIntent()
			if e != nil {
				return e
			}
			finalFee, err := c.Fee(ctx, o.plan.Context.TokenAddress, data)
			if err != nil {
				return err
			}
			finalRequired, _ := new(big.Int).SetString(finalFee.FeeAmount, 10)
			if finalRequired.Cmp(amount) > 0 {
				return errors.New("permit changed required fee; prepare again")
			}
		}
		trust := broadcaster.Trust{ChainID: o.plan.Context.ChainID, Protocol: o.plan.Context.ProtocolAddress}
		signed, e := c.Quote(ctx, data, trust)
		if e != nil {
			return e
		}
		if e = broadcaster.VerifyQuote(signed, broadcaster.Trust{ChainID: trust.ChainID, Protocol: trust.Protocol, Signer: c.Configuration().QuoteSigner}, data, o.plan.Block.Timestamp); e != nil {
			return e
		}
		o.plan.Broadcaster = &broadcasterPlan{Configuration: c.Configuration(), Signed: signed, ChainID: new(big.Int).Set(trust.ChainID), Protocol: trust.Protocol, Intent: append([]byte{}, data...)}
		return o.commitBroadcasterPlan(ctx)
	}
	return errors.New("broadcaster fee did not converge")
}
func (o *Operation) WithBroadcaster(ctx context.Context, c *broadcaster.Client, opt WithBroadcasterOptions) (*Operation, error) {
	if opt.MaxFeeAmount != nil {
		opt.MaxFeeAmount = new(big.Int).Set(opt.MaxFeeAmount)
	}
	return o.stage(ctx, func(next *Operation) error { return next.withBroadcaster(ctx, c, opt) }, true)
}
func (p *PreparedTransaction) WithBroadcaster(ctx context.Context, c *broadcaster.Client, opt WithBroadcasterOptions) (*PreparedTransaction, error) {
	if opt.MaxFeeAmount != nil {
		opt.MaxFeeAmount = new(big.Int).Set(opt.MaxFeeAmount)
	}
	return p.stage(ctx, func(next *Operation) error { return next.withBroadcaster(ctx, c, opt) })
}

func (o *Operation) extendBroadcasterInputs(ctx context.Context, base operationPlan, fee *big.Int) (operationPlan, error) {
	change := new(big.Int)
	index := -1
	for j, out := range base.Outputs {
		if out.Change {
			change.Set(out.Amount)
			index = j
			break
		}
	}
	if change.Cmp(fee) >= 0 {
		return base, nil
	}
	if o.session.explicitInventory != nil {
		return base, errors.New("explicit private inputs cannot cover broadcaster fee")
	}
	need := new(big.Int).Sub(fee, change)
	inputs, memberships, e := o.session.selectInputs(ctx, base.Context.TokenAddress, need, nil, base.Block)
	if e != nil {
		return base, e
	}
	p, e := clonePlan(base)
	if e != nil {
		return base, e
	}
	p.Inputs = append(p.Inputs, inputs...)
	p.Memberships = append(p.Memberships, memberships...)
	for _, n := range inputs {
		change.Add(change, core.ElementToBigInt(n.Note.Amount))
	}
	if index < 0 {
		address, e := o.session.wallet.GetAddress()
		if e != nil {
			return base, e
		}
		if o.session.changeAddress != "" {
			address = o.session.changeAddress
		}
		out, e := o.session.changeOutput(ctx, base.Context.TokenAddress, address, change)
		if e != nil {
			return base, e
		}
		out.Change = true
		p.Outputs = append(p.Outputs, out)
	} else {
		out := p.Outputs[index]
		out.Amount = change
		if out.Note == nil {
			return base, errors.New("missing change note")
		}
		note := *out.Note
		note.Amount = core.BigIntToElement(change)
		out.Note = &note
		p.Outputs[index] = out
	}
	sources, e := sourcesFor(p.Memberships)
	if e != nil {
		return base, e
	}
	nullifiers := make([][32]byte, len(p.Inputs))
	for j, n := range p.Inputs {
		nullifiers[j] = n.Nullifier
	}
	if p.Kind == "transact" {
		p.Params.Sources = sources
		p.Params.Nullifiers = nullifiers
	} else {
		if len(p.Inputs) > 10 {
			return base, errors.New("Credit broadcaster fee requires more than ten funds inputs")
		}
		p.CreateCredit.Sources = sources
		p.CreateCredit.Nullifiers = nullifiers
		if p.TopUpCredit != nil {
			p.TopUpCredit.Creation = *p.CreateCredit
		}
	}
	return p, nil
}
func (o *Operation) commitBroadcasterPlan(ctx context.Context) error {
	if o.session.explicitInventory != nil {
		return ctx.Err()
	}
	encoded, e := json.Marshal(o.plan)
	if e != nil {
		return e
	}
	return o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		record, ok := state.Operations[o.plan.ID]
		if !ok || record.WorkerID != o.workerID || state.Invalidated {
			return errors.New("stale operation worker")
		}
		for _, n := range o.plan.Inputs {
			current, ok := state.Notes[n.Commitment]
			if !ok || current.Spent || (current.ReservedBy != "" && current.ReservedBy != o.plan.ID) || !current.Note.Equal(n.Note) || current.TreeID != n.TreeID || current.LeafIndex != n.LeafIndex {
				return errors.New("fee input changed during preparation")
			}
		}
		record.Inputs = nil
		for _, n := range o.plan.Inputs {
			current := state.Notes[n.Commitment]
			current.ReservedBy = o.plan.ID
			state.Notes[n.Commitment] = current
			record.Inputs = append(record.Inputs, n.Commitment)
		}
		record.Plan = encoded
		state.Operations[o.plan.ID] = record
		return nil
	})
}
