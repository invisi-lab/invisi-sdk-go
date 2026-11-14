package transaction

import (
	"context"
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/settlement"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

type FundingPermitRequest = broadcaster.FundingPermitRequest

type BroadcasterFeeSource = broadcaster.FeeSource

const (
	BroadcasterFeePublicFunding = broadcaster.FeePublicFunding
	BroadcasterFeePrivateInputs = broadcaster.FeePrivateInputs
	BroadcasterFeePayout        = broadcaster.FeePayout
)

type ScheduleOptions = broadcaster.ScheduleOptions
type BroadcasterQuoteOptions = broadcaster.QuoteOptions

type BroadcasterPlan struct {
	CancelToken           string `json:"cancelToken,omitempty"`
	CancellationConfirmed bool   `json:"cancellationConfirmed,omitempty"`
	Configuration         broadcaster.Configuration
	Signed                broadcaster.SignedQuote
	ChainID               *big.Int
	Protocol              common.Address
	Intent                []byte
	Request               *broadcaster.Submission
	RequestID             string
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

func (o *Execution) QuoteBroadcaster(ctx context.Context, c *broadcaster.Client, opt BroadcasterQuoteOptions) error {
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
	if o.Plan.Kind == "shield" && (opt.FeeSource == BroadcasterFeePrivateInputs || opt.FeeSource == BroadcasterFeePayout) {
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

	var schedule *broadcaster.Schedule
	var cancelToken string
	var scheduleMinimum *big.Int
	if opt.Schedule != nil {
		caps, err := c.ServiceStatus(ctx, broadcaster.Trust{ChainID: o.Plan.Context.ChainID, Protocol: o.Plan.Context.ProtocolAddress})
		if err != nil {
			return err
		}
		if !caps.Scheduling.Enabled || caps.SubmissionHalted {
			return errors.New("broadcaster scheduling unavailable")
		}
		request := *opt.Schedule
		after, err := settlement.ScheduleSeconds(request.PublishAfter)
		if err != nil {
			return err
		}
		delay, _ := settlement.ScheduleSeconds(caps.Scheduling.MaxDelaySeconds)
		now := uint64(time.Now().Unix())
		if after <= now || after-now > delay {
			return errors.New("schedule outside broadcaster delay bounds")
		}
		if request.ExecuteBefore == "" {
			window, _ := settlement.ScheduleSeconds(caps.Scheduling.DefaultExecutionWindowSeconds)
			request.ExecuteBefore = strconv.FormatUint(after+window, 10)
		}
		before, err := settlement.ScheduleSeconds(request.ExecuteBefore)
		if err != nil {
			return err
		}
		maxWindow, _ := settlement.ScheduleSeconds(caps.Scheduling.MaxExecutionWindowSeconds)
		if before <= after || before-after > maxWindow {
			return errors.New("schedule outside broadcaster execution window")
		}
		slack, _ := new(big.Int).SetString(caps.Scheduling.MinDeadlineSlackSeconds, 10)
		scheduleMinimum = new(big.Int).Add(new(big.Int).SetUint64(before), slack)
		opt.Schedule = &request
		token, hash, err := broadcaster.NewCancelToken()
		if err != nil {
			return err
		}
		cancelToken = token
		schedule = &broadcaster.Schedule{PublishAfter: opt.Schedule.PublishAfter, ExecuteBefore: opt.Schedule.ExecuteBefore, CancelTokenHash: hash}
		if err = schedule.Validate(); err != nil {
			return err
		}
	}
	base, e := ClonePlan(o.Plan)
	if e != nil {
		return e
	}
	initial, e := o.BroadcasterIntent()
	if e != nil {
		return e
	}
	if e = validateScheduledDeadlines(initial, scheduleMinimum); e != nil {
		return e
	}
	fee, e := c.EstimateQuote(ctx, initial, schedule)
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
		required, e := c.EstimateQuote(ctx, data, schedule)
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
			finalFee, err := c.EstimateQuote(ctx, data, schedule)
			if err != nil {
				return err
			}
			finalRequired, _ := new(big.Int).SetString(finalFee.FeeAmount, 10)
			if finalRequired.Cmp(amount) > 0 {
				return errors.New("permit changed required fee; prepare again")
			}
		}
		trust := broadcaster.Trust{ChainID: o.Plan.Context.ChainID, Protocol: o.Plan.Context.ProtocolAddress}
		if e = validateScheduledDeadlines(data, scheduleMinimum); e != nil {
			return e
		}
		signed, e := c.GetQuote(ctx, data, trust, schedule)
		if e != nil {
			return e
		}

		o.Plan.Broadcaster = &BroadcasterPlan{CancelToken: cancelToken, Configuration: c.Configuration(), Signed: signed, ChainID: new(big.Int).Set(trust.ChainID), Protocol: trust.Protocol, Intent: append([]byte{}, data...)}
		return o.CommitBroadcasterPlan(ctx)
	}
	return errors.New("broadcaster fee did not converge")
}

func (p *PreparedTransaction) bindBroadcasterQuote() *PreparedTransaction {
	broadcaster.BindOperationTarget(&p.OperationTarget, func(ctx context.Context, c *broadcaster.Client, opt broadcaster.QuoteOptions) error {
		_, err := p.stage(ctx, func(next *Execution) error { return next.QuoteBroadcaster(ctx, c, opt) }, true)
		return err
	}, p.submitViaBroadcaster, p.cancelViaBroadcaster)
	return p
}

func validateScheduledDeadlines(data []byte, minimum *big.Int) error {
	if minimum == nil {
		return nil
	}
	request, err := settlement.DecodeSingle(data)
	if err != nil {
		return err
	}
	if request.Statement().Deadline.Cmp(minimum) < 0 {
		return errors.New("intent deadline does not cover scheduled window and slack")
	}
	if request.HasFunding() && request.Funding.Amount.Sign() > 0 {
		if request.Funding.Deadline.Cmp(minimum) < 0 {
			return errors.New("funding deadline does not cover scheduled window and slack")
		}
		if request.Funding.Permit.Deadline.Sign() != 0 && request.Funding.Permit.Deadline.Cmp(minimum) < 0 {
			return errors.New("permit deadline does not cover scheduled window and slack")
		}
	}
	return nil
}

func (BroadcasterPlan) String() string     { return "BroadcasterPlan([sensitive])" }
func (b BroadcasterPlan) GoString() string { return b.String() }
