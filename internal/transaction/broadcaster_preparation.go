package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
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

type BroadcasterPlan struct {
	Configuration broadcaster.Configuration
	Signed        broadcaster.SignedQuote
	ChainID       *big.Int
	Protocol      common.Address
	Intent        []byte
	RequestID     string
}

func (o *Execution) BroadcasterIntent() ([]byte, error) {
	a, e := protocol.ProtocolABI()
	if e != nil {
		return nil, e
	}
	f, e := protocol.NormalizeFunding(o.Plan.Funding)
	if e != nil {
		return nil, e
	}
	f.Signature = nil
	switch o.Plan.Kind {
	case "shield":
		p := o.ShieldParams()
		if p.Params.FeeAmount == nil {
			p.Params.FeeAmount = new(big.Int)
		}
		return a.Pack("shield", p.Params, p.Outputs, p.Deliveries, f)
	case "transact":
		r, e := protocol.NormalizeTransactRequest(o.Plan.Context, o.Plan.Params, o.Deliveries(), f)
		if e != nil {
			return nil, e
		}
		for _, m := range a.Methods {
			if m.RawName == "transact" && len(m.Inputs) == 2 && m.Inputs[1].Type.T == abi.TupleTy {
				return a.Pack(m.Name, []byte{}, protocol.TransactRequest{Params: r.Params, Deliveries: o.Deliveries(), Funding: r.Funding})
			}
		}

	}
	return nil, errors.New("unsupported broadcaster operation")
}

func (o *Execution) WithBroadcaster(ctx context.Context, c *broadcaster.Client, opt WithBroadcasterOptions) error {
	if c == nil {
		return errors.New("broadcaster client required")
	}
	if o.Status != "prepared" || o.Bundle != nil || o.Authorization != nil || o.Plan.Broadcaster != nil || o.Plan.Target == "noir-recursive" {
		return errors.New("broadcaster must be selected before authorization and proving")
	}
	switch o.Plan.Kind {
	case "shield", "transact":
	default:
		return errors.New("unsupported broadcaster operation")
	}
	switch opt.FeeSource {
	case "", BroadcasterFeePublicFunding, BroadcasterFeePrivateInputs, BroadcasterFeePayout:
	default:
		return errors.New("unsupported fee source")
	}
	if opt.FeeSource == BroadcasterFeePayout || o.Plan.Kind == "shield" && opt.FeeSource == BroadcasterFeePrivateInputs {
		return errors.New("fee source unavailable for operation")
	}

	if Positive(opt.MaxFeeAmount, 128) != nil {
		return errors.New("max fee must be Positive uint128")
	}
	opt.MaxFeeAmount = new(big.Int).Set(opt.MaxFeeAmount)
	if opt.MaxAttempts == 0 {
		opt.MaxAttempts = 4
	}
	if opt.MaxAttempts < 1 || opt.MaxAttempts > 16 {
		return errors.New("max attempts must be 1..16")
	}

	base, e := ClonePlan(o.Plan)
	if e != nil {
		return e
	}
	initial, e := o.BroadcasterIntent()
	if e != nil {
		return e
	}
	fee, e := c.EstimateQuote(ctx, initial)
	if e != nil {
		return e
	}
	amount, _ := new(big.Int).SetString(fee.FeeAmount, 10)
	for attempt := 0; attempt < opt.MaxAttempts; attempt++ {
		if amount.Cmp(opt.MaxFeeAmount) > 0 {
			return errors.New("broadcaster fee exceeds maximum")
		}
		if e = o.ApplyBroadcasterFee(ctx, base, c.Configuration().FeeRecipient, amount, opt); e != nil {
			return e
		}
		data, e := o.BroadcasterIntent()
		if e != nil {
			return e
		}
		required, e := c.EstimateQuote(ctx, data)
		if e != nil {
			return e
		}
		next, _ := new(big.Int).SetString(required.FeeAmount, 10)
		if next.Cmp(amount) > 0 {
			amount = next
			continue
		}
		if opt.FundingPermit != nil {
			if o.Plan.Funding.Amount.Sign() == 0 {
				return errors.New("fundingPermit requires explicit public funding")
			}
			permit, err := opt.FundingPermit(ctx, FundingPermitRequest{Payer: o.Plan.Funding.Payer, Amount: new(big.Int).Set(o.Plan.Funding.Amount), TokenAddress: o.Plan.Context.TokenAddress, ProtocolAddress: o.Plan.Context.ProtocolAddress})
			if err != nil {
				return err
			}
			if permit == nil {
				return errors.New("fundingPermit returned nil")
			}
			o.Plan.Funding.Permit = *permit
			o.Plan.Funding, e = protocol.NormalizeFunding(o.Plan.Funding)
			if e != nil {
				return e
			}
			data, e = o.BroadcasterIntent()
			if e != nil {
				return e
			}
			finalFee, err := c.EstimateQuote(ctx, data)
			if err != nil {
				return err
			}
			finalRequired, _ := new(big.Int).SetString(finalFee.FeeAmount, 10)
			if finalRequired.Cmp(amount) > 0 {
				return errors.New("permit changed required fee; prepare again")
			}
		}
		trust := broadcaster.Trust{ChainID: o.Plan.Context.ChainID, Protocol: o.Plan.Context.ProtocolAddress}
		signed, e := c.GetQuote(ctx, data, trust)
		if e != nil {
			return e
		}

		o.Plan.Broadcaster = &BroadcasterPlan{Configuration: c.Configuration(), Signed: signed, ChainID: new(big.Int).Set(trust.ChainID), Protocol: trust.Protocol, Intent: append([]byte{}, data...)}
		return o.CommitBroadcasterPlan(ctx)
	}
	return errors.New("broadcaster fee did not converge")
}

func (p *PreparedTransaction) WithBroadcaster(ctx context.Context, c *broadcaster.Client, opt WithBroadcasterOptions) (*PreparedTransaction, error) {
	if opt.MaxFeeAmount != nil {
		opt.MaxFeeAmount = new(big.Int).Set(opt.MaxFeeAmount)
	}
	return p.stage(ctx, func(next *Execution) error { return next.WithBroadcaster(ctx, c, opt) })
}
