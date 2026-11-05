package transaction

import (
	"context"
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

func (s *Runtime) PrepareTopUpCredit(ctx context.Context, p TopUpCreditPreparation) (result *Execution, failure error) {
	defer func() { failure = WrapPreparation("prepareTopUpCredit", failure) }()
	var err error
	p, err = Frozen(p)
	if err != nil {
		return nil, err
	}
	_, block, _, e := s.Frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	old, membership, e := s.CreditInput(ctx, p.TokenAddress, p.Credit, "owner", block)
	if e != nil {
		return nil, e
	}
	if e = Positive(p.AdditionalAmount, 126); e != nil {
		return nil, e
	}
	amount := new(big.Int).Add(core.ElementToBigInt(old.State.Note.Amount), p.AdditionalAmount)
	expiry := p.Expiry
	if expiry == nil {
		expiry = core.ElementToBigInt(old.State.Note.Expiry)
	}
	return s.PrepareCreditCreation(ctx, CreateCreditPreparation{TokenAddress: p.TokenAddress, Amount: amount, Expiry: expiry, Funding: p.Funding, Deadline: p.Deadline, Payload: p.Payload, SingleLimit: p.SingleLimit, DailyLimit: p.DailyLimit, MonthlyLimit: p.MonthlyLimit}, &old, &membership)
}
