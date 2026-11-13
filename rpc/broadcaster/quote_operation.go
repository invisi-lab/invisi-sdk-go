package broadcaster

import (
	"context"
	"errors"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

type FundingPermitRequest struct {
	Payer, TokenAddress, ProtocolAddress common.Address
	Amount                               *big.Int
}

type FeeSource string

const (
	FeePublicFunding FeeSource = "public-funding"
	FeePrivateInputs FeeSource = "private-inputs"
	FeePayout        FeeSource = "payout"
)

type ScheduleOptions struct {
	PublishAfter  string
	ExecuteBefore string
}

type QuoteOptions struct {
	Schedule      *ScheduleOptions
	MaxFeeAmount  *big.Int
	MaxAttempts   int
	FeeSource     FeeSource
	FundingPermit func(context.Context, FundingPermitRequest) (*protocol.TokenPermit, error)
}

// QuoteTarget binds a wallet-owned lifecycle to the RPC client without importing a wallet or Store.
// It must not be copied after binding; the owner-address check rejects copied capabilities.
type QuoteTarget struct {
	self  *QuoteTarget
	quote func(context.Context, *Client, QuoteOptions) error
}

// BindQuoteTarget is for SDK handle constructors. The callback must atomically install the quote
// and preserve the original handle and reservations on failure. It must not export private material.
func BindQuoteTarget(target *QuoteTarget, quote func(context.Context, *Client, QuoteOptions) error) {
	target.self = target
	target.quote = quote
}

func (target *QuoteTarget) quoteCapability() *QuoteTarget { return target }

type quoteTarget interface {
	quoteCapability() *QuoteTarget
}

// Quote freezes the fee before authorization or proving. SDK targets own atomic state updates.
func (c *Client) Quote(ctx context.Context, target quoteTarget, options QuoteOptions) error {
	if c == nil || target == nil {
		return errors.New("broadcaster client and prepared operation required")
	}
	// A promoted embedded method dereferences a typed-nil owner before reaching the capability.
	if value := reflect.ValueOf(target); value.Kind() == reflect.Pointer && value.IsNil() {
		return errors.New("prepared operation is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	capability := target.quoteCapability()
	if capability == nil || capability.self != capability || capability.quote == nil {
		return errors.New("operation does not support broadcaster quoting or was copied")
	}
	if options.MaxFeeAmount != nil {
		options.MaxFeeAmount = new(big.Int).Set(options.MaxFeeAmount)
	}
	if options.Schedule != nil {
		value := *options.Schedule
		options.Schedule = &value
	}
	return capability.quote(ctx, c, options)
}
