package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
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

	}
}

func (o *Execution) ApplyBroadcasterFee(ctx context.Context, base Plan, recipient common.Address, fee *big.Int, opt BroadcasterQuoteOptions) error {
	_, previousFee := PlanFee(base)
	requestedFee := new(big.Int).Set(fee)
	fee = new(big.Int).Sub(fee, previousFee)
	if fee.Sign() > 0 && base.Kind != "shield" && opt.FeeSource != BroadcasterFeePublicFunding && (base.Funding.Amount == nil || base.Funding.Amount.Sign() == 0 || opt.FeeSource == BroadcasterFeePrivateInputs) {
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
	SetPlanFee(&o.Plan, recipient, requestedFee)

	if fee.Sign() < 0 && p.Kind == "transact" && opt.FeeSource != BroadcasterFeePublicFunding {
		hasChange := false
		for _, out := range p.Outputs {
			if out.Change {
				hasChange = true
				break
			}
		}
		if !hasChange {
			address, e := o.Runtime.Wallet.GetAddress()
			if e != nil {
				return e
			}
			if o.Runtime.ChangeAddress != "" {
				address = o.Runtime.ChangeAddress
			}
			out, e := o.Runtime.ChangeOutput(ctx, p.Context.TokenAddress, address, new(big.Int).Neg(fee))
			if e != nil {
				return e
			}
			out.Change = true
			p.Outputs = append(p.Outputs, out)
			o.Plan.Outputs = append(o.Plan.Outputs, out)
			fee = new(big.Int)
		}
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
	if change < 0 && fee.Sign() == 0 {
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
			note.Amount = primitives.BigIntToElement(amount)
			commit, e := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, primitives.BigIntToElement(out.Identity.Big()), note.TokenAddress, note.Amount)
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
			o.Plan.Params.NoteCommitments = nil
			for _, out := range o.Plan.Outputs {
				o.Plan.Params.NoteCommitments = append(o.Plan.Params.NoteCommitments, out.Commitment)
			}
			o.Plan.Capacity, e = protocol.SelectTransactCapacity(len(p.Inputs), len(o.Plan.Outputs))
			if e != nil {
				return e
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
	if gross.Sign() <= 0 {
		return errors.New("public funding fee replacement must leave positive funding")
	}
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
