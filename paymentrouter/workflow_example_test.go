package paymentrouter_test

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-labs/invisi-sdk-go/paymentrouter"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/rpc/broadcaster"
)

// Wallet callbacks return on-chain EOA signatures with v=27/28 (normalize
// geth crypto.Sign parity 0/1); ERC-1271 follows the wallet format. EIP-7702
// authorization yParity stays 0/1.
// These host integration recipes have no Output directive. go test compiles
// them without calling a provider, wallet or Broadcaster. core must be a real
// independently verified proof exported after PreparePayment(...).Prove(ctx).
func ExampleClient_BuildPayment() {
	build := func(ctx context.Context, client *paymentrouter.Client,
		core protocol.TransactionCall, payer, input common.Address,
		budget, deadline *big.Int, route paymentrouter.Route,
		ensureApproval func(context.Context, protocol.TransactionCall) error,
	) (protocol.TransactionCall, error) {
		approval, err := client.ApprovalTransaction(input, budget)
		if err != nil {
			return protocol.TransactionCall{}, err
		}
		// Host submits approval and waits for its canonical receipt. Alternatively,
		// sign an ERC-2612 permit and attach PaymentOptions.Permit, if supported.
		if err = ensureApproval(ctx, approval); err != nil {
			return protocol.TransactionCall{}, err
		}
		call, err := client.BuildPayment(ctx, paymentrouter.PaymentOptions{
			Core: core, Mode: paymentrouter.DirectToken, Payer: payer,
			TokenIn: input, AmountInMaximum: budget, RefundRecipient: payer,
			Deadline: deadline, Route: route,
		})
		if err != nil {
			return call, err
		}
		_, err = client.SimulateCanonicalPayment(ctx, call, payer)
		return call, err // Host signs/sends only after successful simulation.
	}
	_ = build
}

func ExampleClient_QuoteRoute() {
	quote := func(ctx context.Context, client *paymentrouter.Client,
		adapter paymentrouter.QuoteAdapter, input, output common.Address,
		fundingIncludingFee, budget, deadline *big.Int,
	) (paymentrouter.SwapQuote, paymentrouter.Route, error) {
		// Host configures a V3 or 0x adapter. Zero input means native. Identical
		// input/output or native -> WETH uses ChannelNone without a swap quote.
		return client.QuoteRoute(ctx, adapter, paymentrouter.QuoteParameters{
			TokenIn: input, TokenOut: output, AmountOut: fundingIncludingFee,
			MaxInput: budget, Deadline: deadline,
		}) // BuildPayment uses the returned quote's MaxInput, Deadline and Route.
	}
	_ = quote
}

func ExampleClient_BuildPayment_relayedToken() {
	relay := func(ctx context.Context, client *paymentrouter.Client,
		service *broadcaster.Client, trust broadcaster.Trust,
		opts paymentrouter.PaymentOptions, maxFee *big.Int,
		prove func(context.Context) (protocol.TransactionCall, error),
		sign func(context.Context, apitypes.TypedData) ([]byte, error), submit bool,
	) (broadcaster.SignedQuote, broadcaster.Status, error) {
		var quote broadcaster.SignedQuote
		var status broadcaster.Status
		// opts.Core is prepared.ExportRouterDraft(ctx), with finalized bounded fee
		// and route. prove must retain its exact Core business fields.
		// Relayer is the authenticated Broadcaster EVM sender.
		opts.Mode, opts.Signature = paymentrouter.RelayedToken, nil
		unsigned, err := client.BuildPaymentDraft(ctx, opts)
		if err != nil {
			return quote, status, err
		}
		draftEnv := paymentrouter.RouterEnvelope{To: unsigned.To, Data: unsigned.Data, Value: "0"}
		fee, err := service.EstimateRouterQuote(ctx, opts.Core.Data, draftEnv)
		if err != nil {
			return quote, status, err
		}
		if err = checkRouterFee(opts.Core, fee, maxFee); err != nil {
			return quote, status, err
		}
		quote, err = service.GetRouterQuote(ctx, opts.Core.Data, draftEnv, trust)
		if err != nil {
			return quote, status, err
		}
		// Only after the formal quote: produce the proof and sign its actual hash.
		opts.Core, err = prove(ctx)
		if err != nil {
			return quote, status, err
		}
		unsigned, err = client.BuildPayment(ctx, opts)
		if err != nil {
			return quote, status, err
		}
		decoded, err := paymentrouter.DecodeRouterCall(unsigned.Data)
		if err != nil {
			return quote, status, err
		}
		d := client.Deployment()
		typed, err := paymentrouter.IntentTypedData(d.ChainID, d.RouterAddress, decoded.Intent)
		if err != nil {
			return quote, status, err
		}
		opts.Signature, err = sign(ctx, typed) // Explicit wallet approval.
		if err != nil {
			return quote, status, err
		}
		opts.Nonce = decoded.Intent.Nonce // Reject a concurrent nonce change.
		call, err := client.BuildPayment(ctx, opts)
		if err != nil {
			return quote, status, err
		}
		env := paymentrouter.RouterEnvelope{To: call.To, Data: call.Data, Value: "0"}
		if _, err = client.SimulateEnvelope(ctx, env, opts.Relayer); err != nil {
			return quote, status, err
		}
		if !submit {
			return quote, status, nil
		}
		// Host persists core, env and quote before this explicit submission.
		status, err = service.SubmitRouterTransaction(ctx, opts.Core.Data, env, quote)
		return quote, status, err // Persist the returned opaque status.ID.
	}
	_ = relay
}

func ExampleClient_BuildSponsoredPayment() {
	build := func(ctx context.Context, client *paymentrouter.Client,
		inner paymentrouter.PaymentOptions, policy paymentrouter.SponsoredPaymentOptions,
		routerCodeHash common.Hash,
		sign func(context.Context, apitypes.TypedData) ([]byte, error),
	) (paymentrouter.RouterEnvelope, error) {
		// Host reads account action nonce into policy.Payment.Nonce, pins runtime
		// hashes and sets policy.Player to inner.Payer. AuthorizationList contains
		// an explicitly wallet-signed SignDelegation authorization if installation
		// is needed; omit it for an existing approved delegation. EOA authorization
		// nonce and account action nonce are independent of the Router nonce.
		inner.Mode = paymentrouter.SponsoredNative
		call, err := client.BuildPayment(ctx, inner)
		if err != nil {
			return paymentrouter.RouterEnvelope{}, err
		}
		policy.Payment.RouterCalldata = call.Data
		policy.Payment.Value = call.Value // Player supplies this inner native cap.
		policy.Payment.Relayer = inner.Relayer
		policy.Payment.Deadline = inner.Deadline
		d := client.Deployment()
		typed, err := paymentrouter.NativePaymentTypedData(d.ChainID, policy.Player,
			policy.Implementation, d.RouterAddress, routerCodeHash, policy.Payment)
		if err != nil {
			return paymentrouter.RouterEnvelope{}, err
		}
		policy.Payment.Signature, err = sign(ctx, typed)
		if err != nil {
			return paymentrouter.RouterEnvelope{}, err
		}
		env, err := client.BuildSponsoredPayment(ctx, policy)
		if err != nil {
			return env, err
		}
		_, err = client.SimulateEnvelope(ctx, env, inner.Relayer, policy)
		// env.Value is "0". Use core.Data and this exact envelope with the Router
		// estimate/quote/submit methods as above. Fee or route changes require a new
		// proof/account signature/quote; never broadcast fixture proofs.
		return env, err
	}
	_ = build
}

// Check the host-selected fee before requesting the final signed quote.
// The service still authenticates current fee/gas at quotation and submission.
func checkRouterFee(core protocol.TransactionCall, fee broadcaster.Fee, maxFee *big.Int) error {
	intent, err := broadcaster.DecodeIntent(core.Data)
	if err != nil {
		return err
	}
	amount, ok := new(big.Int).SetString(fee.FeeAmount, 10)
	if !ok || maxFee == nil || maxFee.Sign() <= 0 || amount.Sign() <= 0 {
		return fmt.Errorf("positive required fee and explicit fee cap required")
	}
	if intent.FeeAmount.Cmp(maxFee) > 0 || amount.Cmp(maxFee) > 0 {
		return fmt.Errorf("Core fee or required fee exceeds host cap")
	}
	if intent.FeeRecipient != fee.FeeRecipient || intent.FeeAmount.Cmp(amount) < 0 {
		return fmt.Errorf("rebuild Core fee, prove and sign again before quoting")
	}
	return nil
}
