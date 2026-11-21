package routerexample

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/paymentrouter"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
)

// Prepare100PlusFee creates two 50-token Notes and a 0.2 Core-token fee.
// decimals, token, recipients and fee recipient must describe the real deployment.
// No private Note inputs or Provenance input proof are involved in public funding.
func Prepare100PlusFee(ctx context.Context, client *paymentrouter.Client,
	invisi *sdk.Invisi, wallet *sdk.Wallet, pool *sdk.Protocol,
	token, feeRecipient common.Address, recipients [2]string, decimals uint8, deadline uint64,
) (*sdk.PreparedTransaction, *big.Int, error) {
	if client == nil || invisi == nil || wallet == nil || pool == nil {
		return nil, nil, fmt.Errorf("Router, SDK, wallet and Protocol required")
	}
	if decimals == 0 || decimals > 30 {
		return nil, nil, fmt.Errorf("use authenticated token decimals in [1,30]")
	}
	unit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	fee := new(big.Int).Div(new(big.Int).Set(unit), big.NewInt(5))
	outputs := make([]sdk.NoteOutput, 0, 2)
	for _, recipient := range recipients {
		output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
			ChainID: new(big.Int).SetUint64(pool.Context().ChainID), ProtocolAddress: pool.Context().ProtocolAddress,
			TokenAddress: token, Amount: new(big.Int).Mul(big.NewInt(50), unit),
			Recipient: sdk.InvisiRecipient{Address: recipient},
		})
		if err != nil {
			return nil, nil, err
		}
		outputs = append(outputs, output)
	}
	prepared, err := client.PreparePayment(ctx, paymentrouter.PreparePaymentOptions{
		Invisi: invisi, Wallet: wallet, Protocol: pool, TokenAddress: token, Outputs: outputs,
		Fee: &sdk.Fee{Recipient: feeRecipient, Amount: fee}, Deadline: deadline,
	})
	return prepared, new(big.Int).Add(new(big.Int).Mul(big.NewInt(100), unit), fee), err
}

// RouteV3 uses actual on-chain exact-output quotes. Native input is the zero address;
// configure the route with WETH as its input token. QuoteRoute encodes reverse paths.
func RouteV3(ctx context.Context, client *paymentrouter.Client, provider paymentrouter.Provider,
	quoter common.Address, routes []paymentrouter.V3Route, params paymentrouter.QuoteParameters,
) (paymentrouter.SwapQuote, paymentrouter.Route, error) {
	adapter, err := paymentrouter.NewUniswapV3Adapter(provider, quoter, routes)
	if err != nil {
		return paymentrouter.SwapQuote{}, paymentrouter.Route{}, err
	}
	return client.QuoteRoute(ctx, adapter, params)
}

// RouteZeroEx accepts an authenticated backend transport. It does not fabricate a
// quote, read an API key, or assume that opaque liquidity action bytes are executable.
func RouteZeroEx(ctx context.Context, client *paymentrouter.Client,
	transport paymentrouter.ZeroExQuoteTransport, params paymentrouter.QuoteParameters,
) (paymentrouter.SwapQuote, paymentrouter.Route, error) {
	adapter, err := paymentrouter.NewZeroExAdapter(transport)
	if err != nil {
		return paymentrouter.SwapQuote{}, paymentrouter.Route{}, err
	}
	return client.QuoteRoute(ctx, adapter, params)
}

// SelfPayment covers self-paid ERC20 and native input with V3, 0x or ChannelNone.
// opts.Core must come from a real prepared.Prove/ExportTransaction. The caller
// explicitly opts into sending and provides persistence plus an external signer.
func SelfPayment(ctx context.Context, client *paymentrouter.Client, opts paymentrouter.PaymentOptions,
	ensureApproval func(context.Context, protocol.TransactionCall) error,
	persist func(protocol.TransactionCall) error,
	send func(context.Context, protocol.TransactionCall) (common.Hash, error), submit bool,
) (protocol.TransactionCall, common.Hash, error) {
	var hash common.Hash
	if client == nil || persist == nil || (submit && send == nil) || (opts.Mode == paymentrouter.DirectToken && ensureApproval == nil) {
		return protocol.TransactionCall{}, hash, fmt.Errorf("Router, approval, persistence and send capabilities required")
	}
	if opts.Mode != paymentrouter.DirectToken && opts.Mode != paymentrouter.DirectNative {
		return protocol.TransactionCall{}, hash, fmt.Errorf("select DirectToken or DirectNative")
	}
	if opts.Mode == paymentrouter.DirectToken {
		approval, err := client.ApprovalTransaction(opts.TokenIn, opts.AmountInMaximum)
		if err != nil {
			return protocol.TransactionCall{}, hash, err
		}
		// Host approves the Router, then waits for a canonical successful receipt.
		if err = ensureApproval(ctx, approval); err != nil {
			return protocol.TransactionCall{}, hash, err
		}
	}
	call, err := client.BuildPayment(ctx, opts)
	if err != nil {
		return call, hash, err
	}
	if _, err = client.SimulateCanonicalPayment(ctx, call, opts.Payer); err != nil {
		return call, hash, err
	}
	if err = persist(call); err != nil {
		return call, hash, err
	}
	if submit {
		hash, err = send(ctx, call)
	}
	return call, hash, err
}

// CoreProofCallback closes over the exact prepared statement used for quotation.
// Pass this callback to RelayToken or SponsorNative; it only runs after the quote.
func CoreProofCallback(prepared *sdk.PreparedTransaction) func(context.Context) (protocol.TransactionCall, error) {
	return func(ctx context.Context) (protocol.TransactionCall, error) {
		if prepared == nil {
			return protocol.TransactionCall{}, fmt.Errorf("prepared Core payment required")
		}
		ready, err := prepared.Prove(ctx)
		if err != nil {
			return protocol.TransactionCall{}, err
		}
		return ready.ExportTransaction(ctx)
	}
}
