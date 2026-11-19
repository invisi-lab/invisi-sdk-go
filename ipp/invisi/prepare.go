// Package invisi adapts validated IPP requests to existing SDK preparation inputs.
package invisi

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/ipp"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
)

// Target binds a trusted wallet deployment and preparation capabilities. Callers can bind
// WalletSession.PrepareTransfer and Credit().PrepareCreate, or explicit stateless SDK calls.
// The callbacks only prepare; authorization, proofs and submission remain with the caller.
type Target[T any] struct {
	Context             protocolrpc.DeploymentContext
	PrepareTransfer     func(context.Context, sdk.TransferPreparation) (T, error)
	PrepareCreateCredit func(context.Context, sdk.CreateCreditPreparation) (T, error)
}
type Options struct {
	Now func() time.Time
	// CreationABI selects the deployed selector for pure Credit creation.
	CreationABI string
	// InitialPaymentFee is a local actual fee; IPP terms contain only the merchant net amount.
	InitialPaymentFee *credit.Fee
}

func PrepareRequest[T any](ctx context.Context, value *ipp.Request, target Target[T], options Options) (T, error) {
	var zero T
	b, e := ipp.EncodeJSON(value)
	if e != nil {
		return zero, &ipp.Error{Code: "invalid_message"}
	}
	r, e := ipp.ParsePaymentRequest(b, "")
	if e != nil {
		return zero, e
	}
	now := time.Now()
	if options.Now != nil {
		now = options.Now()
	}
	expires, _ := time.Parse("2006-01-02T15:04:05Z", r.ExpiresAt)
	if r.Status != "requires_action" || !expires.After(now) {
		return zero, &ipp.Error{Code: "execution_not_allowed"}
	}
	var network, protocolAddress, tokenAddress string
	if r.Mode == ipp.Payment {
		network, protocolAddress, tokenAddress = r.Payment.Network, r.Payment.ProtocolAddress, r.Payment.TokenAddress
	} else {
		network, protocolAddress, tokenAddress = r.Credit.Network, r.Credit.ProtocolAddress, r.Credit.TokenAddress
	}
	chainID, _ := new(big.Int).SetString(network[7:], 10)
	if chainID.Uint64() != target.Context.ChainID || common.HexToAddress(protocolAddress) != target.Context.ProtocolAddress {
		return zero, &ipp.Error{Code: "deployment_mismatch"}
	}
	if e = ctx.Err(); e != nil {
		return zero, e
	}
	deadline := uint64(expires.Unix())
	token := common.HexToAddress(tokenAddress)
	if options.InitialPaymentFee != nil && (r.Mode != ipp.Credit || r.Credit == nil || r.Credit.InitialPayment == nil) {
		return zero, &ipp.Error{Code: "execution_not_allowed"}
	}
	if r.Mode == ipp.Payment {
		if target.PrepareTransfer == nil {
			return zero, &ipp.Error{Code: "missing_capability"}
		}
		amount, _ := new(big.Int).SetString(r.Payment.Amount, 10)
		out, e := sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: chainID, ProtocolAddress: common.HexToAddress(protocolAddress), TokenAddress: token, Amount: amount, Recipient: sdk.IdentityRecipient{Identity: common.HexToHash(r.Payment.Recipient.Identity)}})
		if e != nil {
			return zero, e
		}
		return target.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{out}, Deadline: deadline})
	}
	if target.PrepareCreateCredit == nil {
		return zero, &ipp.Error{Code: "missing_capability"}
	}
	c := r.Credit
	amount, _ := new(big.Int).SetString(c.AvailableCredit, 10)
	key, _ := new(big.Int).SetString(c.CreditSharedKey[2:], 16)
	single, _ := new(big.Int).SetString(c.SingleLimit, 10)
	daily, _ := new(big.Int).SetString(c.DailyLimit, 10)
	monthly, _ := new(big.Int).SetString(c.MonthlyLimit, 10)
	expiry, _ := time.Parse("2006-01-02T15:04:05Z", c.ExpiresAt)
	preparation := sdk.CreateCreditPreparation{TokenAddress: token, Spender: c.Spender, CreditSharedKey: key, AvailableCredit: amount, Unlimited: c.Unlimited, SingleLimit: primitives.BigIntToElement(single), DailyLimit: primitives.BigIntToElement(daily), MonthlyLimit: primitives.BigIntToElement(monthly), Expiry: uint64(expiry.Unix()), Deadline: deadline, CreationABI: options.CreationABI}
	if c.InitialPayment != nil {
		gross, _ := new(big.Int).SetString(c.InitialPayment.Amount, 10)
		if fee := options.InitialPaymentFee; fee != nil {
			if fee.Amount == nil || fee.Amount.Sign() < 0 || (fee.Amount.Sign() == 0) != (fee.Recipient == (common.Address{})) {
				return zero, &ipp.Error{Code: "execution_not_allowed"}
			}
			gross.Add(gross, fee.Amount)
			preparation.Fee = &credit.Fee{Recipient: fee.Recipient, Amount: new(big.Int).Set(fee.Amount)}
		}
		if gross.Sign() <= 0 || gross.BitLen() > 126 {
			return zero, &ipp.Error{Code: "execution_not_allowed"}
		}
		preparation.InitialPayment = &credit.InitialPayment{Amount: gross}
	}
	return target.PrepareCreateCredit(ctx, preparation)
}
