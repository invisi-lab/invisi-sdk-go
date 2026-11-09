package credit

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Preview struct {
	Kind                                          string
	Token, Owner, FeeRecipient                    common.Address
	GrossAmount, NoteAmount, FeeAmount            *big.Int
	InputCount, StateOutputCount, NoteOutputCount int
	Deadline                                      uint64
}

func (p *Prepared) Preview() Preview {
	p.mu.Lock()
	defer p.mu.Unlock()
	i := p.intent
	result := Preview{Token: i.Context.TokenAddress, GrossAmount: new(big.Int), NoteAmount: new(big.Int), FeeAmount: new(big.Int)}
	recipient, amount := intentFee(i)
	result.FeeRecipient = recipient
	if amount != nil {
		result.FeeAmount.Set(amount)
	}
	switch {
	case i.Create != nil:
		result.Kind = "create-credit"
		result.Owner = i.Create.Owner
		result.StateOutputCount = 1
		result.Deadline = i.Create.Deadline.Uint64()
	case i.Update != nil:
		result.Kind = "update-credit"
		result.Owner = i.Update.Owner
		result.InputCount = 1
		result.StateOutputCount = 1
		result.Deadline = i.Update.Deadline.Uint64()
	case i.Draw != nil:
		result.Kind = "credit-draw"
		result.InputCount = len(i.Inputs)
		result.StateOutputCount = len(i.Inputs)
		result.NoteOutputCount = 1
		result.GrossAmount.Set(i.Draw.TotalAmount)
		result.NoteAmount.Sub(i.Draw.TotalAmount, result.FeeAmount)
		result.Deadline = i.Draw.Deadline.Uint64()
	case i.Revoke != nil:
		result.Kind = "revoke-credit"
		result.Owner = i.Revoke.Owner
		result.Deadline = i.Revoke.Deadline.Uint64()
	}
	return result
}
