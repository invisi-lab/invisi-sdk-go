package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func PlanFee(p Plan) (common.Address, *big.Int) {
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

func SetPlanFee(p *Plan, r common.Address, f *big.Int) {
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

func (o *Execution) ApplyBroadcasterFee(ctx context.Context, base Plan, recipient common.Address, fee *big.Int, opt WithBroadcasterOptions) error {
	if base.Kind != "shield" && base.Kind != "cancel-credit" && opt.FeeSource != BroadcasterFeePublicFunding && (base.Funding.Amount == nil || base.Funding.Amount.Sign() == 0 || opt.FeeSource == BroadcasterFeePrivateInputs) {
		var err error
		base, err = o.ExtendBroadcasterInputs(ctx, base, fee)
		if err != nil {
			return err
		}
	}
	p, e := ClonePlan(base)
	if e != nil {
		return e
	}
	o.Plan = p
	SetPlanFee(&o.Plan, recipient, fee)
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
		out := o.Plan.Outputs[change]
		amount := new(big.Int).Sub(out.Amount, fee)
		if amount.Sign() == 0 {
			o.Plan.Outputs = append(o.Plan.Outputs[:change], o.Plan.Outputs[change+1:]...)
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
			recipient, e := o.Runtime.ResolveRecipient(ctx, out.Address)
			if e != nil {
				return e
			}
			delivery, e := EncryptOutput(recipient, note, FieldHash(commit))
			if e != nil {
				return e
			}
			out.Note = &note
			out.Amount = amount
			out.Commitment = FieldHash(commit)
			out.Delivery = delivery
			o.Plan.Outputs[change] = out
		}
		if p.Kind == "transact" {
			o.Plan.Params.OutputCommitments = nil
			for _, out := range o.Plan.Outputs {
				o.Plan.Params.OutputCommitments = append(o.Plan.Params.OutputCommitments, out.Commitment)
			}
			o.Plan.Capacity, e = protocol.SelectTransactCapacity(len(p.Inputs), len(o.Plan.Outputs))
			if e != nil {
				return e
			}
		} else {
			o.Plan.CreateCredit.ChangeCommitment = [32]byte{}
			o.Plan.Deliveries[2] = protocol.Delivery{}
			if len(o.Plan.Outputs) > 0 {
				o.Plan.CreateCredit.ChangeCommitment = o.Plan.Outputs[0].Commitment
				o.Plan.Deliveries[2] = o.Plan.Outputs[0].Delivery
			}
			if o.Plan.TopUpCredit != nil {
				o.Plan.TopUpCredit.Creation = *o.Plan.CreateCredit
			}
		}
		o.Plan.NeedsAuthorization = len(o.Plan.Inputs) > 0 || p.Kind != "transact"
		if p.Funding.Amount.Sign() > 0 && p.FundingMode != "signature" {
			source := FundingSource{Payer: p.Funding.Payer, Amount: new(big.Int).Set(p.Funding.Amount), Authorization: "signature"}
			if p.Funding.Permit.Deadline != nil && p.Funding.Permit.Deadline.Sign() > 0 {
				source.Permit = &p.Funding.Permit
			}
			o.Plan.Funding, o.Plan.FundingMode, e = o.Runtime.Funding(ctx, &source, p.Block, p.Deadline)
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

	o.Plan.Funding, o.Plan.FundingMode, e = o.Runtime.Funding(ctx, &source, p.Block, p.Deadline)
	return e
}
