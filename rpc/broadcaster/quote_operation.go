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

// OperationTarget binds a wallet-owned lifecycle to the RPC client without importing a wallet or Store.
// It must not be copied after binding; the owner-address check rejects copied capabilities.
type OperationTarget struct {
	self   *OperationTarget
	quote  func(context.Context, *Client, QuoteOptions) error
	submit func(context.Context, *Client) (Status, error)
	cancel func(context.Context, *Client) (Status, error)
}

// BindOperationTarget 仅供 SDK 句柄构造器绑定状态机；回调负责原子持久化与保留失败后的预留。
func BindOperationTarget(target *OperationTarget, quote func(context.Context, *Client, QuoteOptions) error, submit, cancel func(context.Context, *Client) (Status, error)) {
	target.self = target
	target.quote = quote
	target.submit, target.cancel = submit, cancel
}

func (target *OperationTarget) operationCapability() *OperationTarget { return target }

type operationTarget interface {
	operationCapability() *OperationTarget
}

// Quote freezes the fee before authorization or proving. SDK targets own atomic state updates.
func (c *Client) Quote(ctx context.Context, target operationTarget, options QuoteOptions) error {
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
	capability := target.operationCapability()
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

func (c *Client) operation(ctx context.Context, target operationTarget, cancel bool) (Status, error) {
	if c == nil || target == nil {
		return Status{}, errors.New("broadcaster client and operation required")
	}
	if v := reflect.ValueOf(target); v.Kind() == reflect.Pointer && v.IsNil() {
		return Status{}, errors.New("operation is nil")
	}
	if err := ctx.Err(); err != nil {
		return Status{}, err
	}
	capability := target.operationCapability()
	if capability == nil || capability.self != capability {
		return Status{}, errors.New("operation capability is unbound or copied")
	}
	action := capability.submit
	if cancel {
		action = capability.cancel
	}
	if action == nil {
		return Status{}, errors.New("operation capability unavailable")
	}
	return action(ctx, c)
}

func (c *Client) Submit(ctx context.Context, target operationTarget) (Status, error) {
	return c.operation(ctx, target, false)
}
func (c *Client) Cancel(ctx context.Context, target operationTarget) (Status, error) {
	return c.operation(ctx, target, true)
}
