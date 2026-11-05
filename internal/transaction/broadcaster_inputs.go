package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

func (o *Execution) ExtendBroadcasterInputs(ctx context.Context, base Plan, fee *big.Int) (Plan, error) {
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
	if o.Runtime.ExplicitInventory != nil {
		return base, errors.New("explicit private inputs cannot cover broadcaster fee")
	}
	need := new(big.Int).Sub(fee, change)
	inputs, memberships, e := o.Runtime.SelectInputs(ctx, base.Context.TokenAddress, need, nil, base.Block)
	if e != nil {
		return base, e
	}
	p, e := ClonePlan(base)
	if e != nil {
		return base, e
	}
	p.Inputs = append(p.Inputs, inputs...)
	p.Memberships = append(p.Memberships, memberships...)
	for _, n := range inputs {
		change.Add(change, core.ElementToBigInt(n.Note.Amount))
	}
	if index < 0 {
		address, e := o.Runtime.Wallet.GetAddress()
		if e != nil {
			return base, e
		}
		if o.Runtime.ChangeAddress != "" {
			address = o.Runtime.ChangeAddress
		}
		out, e := o.Runtime.ChangeOutput(ctx, base.Context.TokenAddress, address, change)
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
	sources, e := SourcesFor(p.Memberships)
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
