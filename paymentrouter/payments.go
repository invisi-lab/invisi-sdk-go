package paymentrouter

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/holiman/uint256"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"math/big"
)

// PaymentOptions binds the original Core call to a canonical Router intent.
// Nonce is read from the Router when nil. An empty relay signature builds an
// unsigned intent for wallet signing; ValidatePayment requires authorization.
type PaymentOptions struct {
	Core                                     protocol.TransactionCall
	Mode                                     uint8
	Payer, TokenIn, RefundRecipient, Relayer common.Address
	AmountInMaximum, Nonce, Deadline         *big.Int
	Route                                    Route
	Permit                                   Permit
	Signature                                []byte
}

func (c *Client) routerNonce(ctx context.Context, payer common.Address, f deploymentFrame) (*big.Int, error) {
	b, e := canonical.Pack("nonces", payer)
	if e != nil {
		return nil, e
	}
	out, e := c.provider.CallContract(ctx, ethereum.CallMsg{To: &c.deployment.RouterAddress, Data: b}, f.BlockNumber)
	if e != nil {
		return nil, e
	}
	a, e := canonical.Methods["nonces"].Outputs.Unpack(out)
	if e != nil {
		return nil, e
	}
	return a[0].(*big.Int), nil
}

// QuoteRoute obtains and validates a quote for native or ERC20 input. The returned
// Route and input budget must be bound into the signed intent by BuildPayment.
func (c *Client) QuoteRoute(ctx context.Context, a QuoteAdapter, p QuoteParameters) (SwapQuote, Route, error) {
	var q SwapQuote
	var r Route
	if a == nil {
		return q, r, fmt.Errorf("adapter required")
	}
	p = cloneParameters(p)
	f, e := frame(ctx, c.provider, c.deployment, a.Channel() == ZeroEx)
	if e != nil {
		return q, r, e
	}
	if e = validateParameters(p, f.Now); e != nil {
		return q, r, e
	}
	if quoteInput(SwapQuoteRequest{QuoteParameters: p, Deployment: c.deployment}) == p.TokenOut {
		return q, r, fmt.Errorf("same-token payment uses ChannelNone")
	}
	q, e = a.Quote(ctx, cloneRequest(SwapQuoteRequest{p, cloneDeployment(c.deployment), copyInt(f.Now), copyInt(f.BlockNumber), f.SettlerAddress}))
	if e != nil {
		return q, r, e
	}
	q = cloneQuote(q)
	if q.Channel != a.Channel() || q.ChainID == nil || q.ChainID.Cmp(c.deployment.ChainID) != 0 || q.RouterAddress != c.deployment.RouterAddress || q.TokenOut != p.TokenOut {
		return q, r, fmt.Errorf("quote deployment or token mismatch")
	}
	if e = validateParameters(QuoteParameters{AmountOut: q.AmountOut, MaxInput: q.MaxInput, Deadline: q.Deadline}, f.Now); e != nil {
		return q, r, e
	}
	if e = uintCheck(q.EstimatedInput, 256, true, "estimated input"); e != nil {
		return q, r, e
	}
	if q.AmountOut.Cmp(p.AmountOut) != 0 || q.MaxInput.Cmp(p.MaxInput) > 0 || q.EstimatedInput.Cmp(q.MaxInput) > 0 || q.Deadline.Cmp(p.Deadline) > 0 {
		return q, r, fmt.Errorf("quote exceeds authorized amounts or deadline")
	}
	switch q.Channel {
	case UniswapV3:
		r = Route{Channel: ChannelUniswapV3, Path: q.ReversePath}
	case ZeroEx:
		if q.SettlerAddress != f.SettlerAddress {
			return q, r, fmt.Errorf("stale Settler")
		}
		r = Route{Channel: ChannelZeroEx, Target: q.Target, Data: q.Data}
	default:
		return q, r, fmt.Errorf("unsupported channel")
	}
	payment := DecodedPayment{Intent: Intent{TokenIn: p.TokenIn, AmountInMaximum: q.MaxInput}, Route: r, Request: protocol.TransactRequest{Params: protocol.CrossTreeTransactParams{Token: p.TokenOut}, Funding: protocol.FundingParams{Amount: p.AmountOut}}}
	if e = ValidateRoute(payment, c.deployment, f.Now, f.SettlerAddress); e != nil {
		return q, r, e
	}
	e = assertFrame(ctx, c.provider, f)
	return q, r, e
}

// BuildPayment returns canonical calldata for all four modes. SponsoredNative
// returns the inner Router call; wrap it with BuildSponsoredPayment before relay.
func (c *Client) BuildPayment(ctx context.Context, o PaymentOptions) (protocol.TransactionCall, error) {
	return c.buildPaymentPhase(ctx, o, false)
}

// BuildPaymentDraft builds the quote envelope before proving or wallet signing.
func (c *Client) BuildPaymentDraft(ctx context.Context, o PaymentOptions) (protocol.TransactionCall, error) {
	return c.buildPaymentPhase(ctx, o, true)
}
func (c *Client) buildPaymentPhase(ctx context.Context, o PaymentOptions, draft bool) (protocol.TransactionCall, error) {
	var empty protocol.TransactionCall
	f, e := frame(ctx, c.provider, c.deployment, o.Route.Channel == ChannelZeroEx)
	if e != nil {
		return empty, e
	}
	proof, req, e := inspectCorePaymentPhase(o.Core, c.deployment, f.Now, true, draft)
	if e != nil {
		return empty, e
	}
	nonce, e := c.routerNonce(ctx, o.Payer, f)
	if e != nil {
		return empty, e
	}
	if o.Nonce != nil && o.Nonce.Cmp(nonce) != 0 {
		return empty, fmt.Errorf("stale Router nonce")
	}
	ch, e := HashCoreRequest(proof, req)
	if e != nil {
		return empty, e
	}
	rh, e := HashRoute(o.Route)
	if e != nil {
		return empty, e
	}
	p := DecodedPayment{Method: methodFor(o.Mode), Proof: proof, Request: req, Intent: Intent{Payer: o.Payer, TokenIn: o.TokenIn, AmountInMaximum: o.AmountInMaximum, RefundRecipient: o.RefundRecipient, Relayer: o.Relayer, Nonce: nonce, Deadline: o.Deadline, CoreRequestHash: ch, RouteHash: rh, Mode: o.Mode}, Route: o.Route, Permit: o.Permit, Signature: o.Signature}
	if e = validateIntentPhase(p, c.deployment, f.Now, nonce, draft); e != nil {
		return empty, e
	}
	if e = ValidateRoute(p, c.deployment, f.Now, f.SettlerAddress); e != nil {
		return empty, e
	}
	if !draft && o.Mode == RelayedToken && len(o.Signature) != 0 {
		h, e := HashIntent(c.deployment.ChainID, c.deployment.RouterAddress, p.Intent)
		if e != nil {
			return empty, e
		}
		if e = c.verifyPayer(ctx, h, o.Signature, o.Payer, f); e != nil {
			return empty, e
		}
	}
	data, e := EncodeRouterCall(p)
	if e != nil {
		return empty, e
	}
	if e = assertFrame(ctx, c.provider, f); e != nil {
		return empty, e
	}
	value := new(big.Int)
	if o.Mode == DirectNative || o.Mode == SponsoredNative {
		value.Set(o.AmountInMaximum)
	}
	return protocol.TransactionCall{To: c.deployment.RouterAddress, Data: data, Value: value}, nil
}

// ValidatePayment refreshes Router deployment, nonce and registry state and checks
// sender/value bindings before any transaction is submitted.
func (c *Client) ValidatePayment(ctx context.Context, call protocol.TransactionCall, from common.Address) error {
	p, e := DecodeRouterCall(call.Data)
	if e != nil {
		return e
	}
	f, e := frame(ctx, c.provider, c.deployment, p.Route.Channel == ChannelZeroEx)
	if e != nil {
		return e
	}
	if call.To != c.deployment.RouterAddress || !nonzero(from) {
		return fmt.Errorf("invalid transaction target or sender")
	}
	nonce, e := c.routerNonce(ctx, p.Intent.Payer, f)
	if e != nil {
		return e
	}
	if e = ValidateIntent(p, c.deployment, f.Now, nonce); e != nil {
		return e
	}
	if e = ValidateRoute(p, c.deployment, f.Now, f.SettlerAddress); e != nil {
		return e
	}
	expected := new(big.Int)
	sender := p.Intent.Payer
	if p.Intent.Mode == DirectNative || p.Intent.Mode == SponsoredNative {
		expected.Set(p.Intent.AmountInMaximum)
	}
	if p.Intent.Mode == RelayedToken {
		sender = p.Intent.Relayer
		h, e := HashIntent(c.deployment.ChainID, c.deployment.RouterAddress, p.Intent)
		if e != nil {
			return e
		}
		if e = c.verifyPayer(ctx, h, p.Signature, p.Intent.Payer, f); e != nil {
			return e
		}
	}
	if call.Value == nil || call.Value.Cmp(expected) != 0 || from != sender {
		return fmt.Errorf("sender or value mismatch")
	}
	return assertFrame(ctx, c.provider, f)
}

type PaymentSimulation struct {
	GasEstimate      uint64
	AmountIn, Refund *big.Int
}

func (c *Client) SimulateCanonicalPayment(ctx context.Context, call protocol.TransactionCall, from common.Address) (PaymentSimulation, error) {
	var result PaymentSimulation
	if e := c.ValidatePayment(ctx, call, from); e != nil {
		return result, e
	}
	p, _ := DecodeRouterCall(call.Data)
	f, e := frame(ctx, c.provider, c.deployment, p.Route.Channel == ChannelZeroEx)
	if e != nil {
		return result, e
	}
	msg := ethereum.CallMsg{From: from, To: &call.To, Data: call.Data, Value: call.Value}
	raw, e := c.provider.CallContract(ctx, msg, f.BlockNumber)
	if e != nil {
		return result, e
	}
	gas, e := c.provider.EstimateGasAt(ctx, msg, f.BlockNumber)
	if e != nil {
		return result, e
	}
	a, e := canonical.Methods[p.Method].Outputs.Unpack(raw)
	if e != nil {
		return result, e
	}
	spent := a[0].(*big.Int)
	if spent.Cmp(p.Intent.AmountInMaximum) > 0 {
		return result, fmt.Errorf("simulated input exceeds budget")
	}
	if e = assertFrame(ctx, c.provider, f); e != nil {
		return result, e
	}
	return PaymentSimulation{gas, spent, new(big.Int).Sub(p.Intent.AmountInMaximum, spent)}, nil
}

// ApprovalTransaction authorizes the Router, never the DEX or Settler, to pull input.
func (c *Client) ApprovalTransaction(token common.Address, amount *big.Int) (protocol.TransactionCall, error) {
	if !nonzero(token) || token == c.deployment.RouterAddress {
		return protocol.TransactionCall{}, fmt.Errorf("invalid input token")
	}
	if e := uintCheck(amount, 256, false, "approval amount"); e != nil {
		return protocol.TransactionCall{}, e
	}
	a := mustABI(`[{"type":"function","name":"approve","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"type":"bool"}]}]`)
	b, e := a.Pack("approve", c.deployment.RouterAddress, amount)
	return protocol.TransactionCall{To: token, Data: b, Value: new(big.Int)}, e
}
func SignTypedData(key *ecdsa.PrivateKey, t apitypes.TypedData) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("signing key required")
	}
	h, e := typedDigest(t, nil)
	if e != nil {
		return nil, e
	}
	s, e := crypto.Sign(h[:], key)
	if e == nil {
		s[64] += 27
	}
	return s, e
}
func VerifySignature(h common.Hash, signature []byte, signer common.Address) error {
	if len(signature) != 65 {
		return fmt.Errorf("expected EOA 65-byte signature; contract signatures require on-chain simulation")
	}
	s := append([]byte(nil), signature...)
	if s[64] >= 27 {
		s[64] -= 27
	}
	if s[64] > 1 || !crypto.ValidateSignatureValues(s[64], new(big.Int).SetBytes(s[:32]), new(big.Int).SetBytes(s[32:64]), true) {
		return fmt.Errorf("invalid signature")
	}
	p, e := crypto.SigToPub(h[:], s)
	if e != nil {
		return e
	}
	if crypto.PubkeyToAddress(*p) != signer {
		return fmt.Errorf("signature signer mismatch")
	}
	return nil
}

// SignDelegation requires an explicit wallet decision; nonce is the EIP-7702
// authorization nonce, independent of both payment action nonces.
func SignDelegation(key *ecdsa.PrivateKey, chain *big.Int, implementation common.Address, nonce uint64) (types.SetCodeAuthorization, error) {
	if key == nil {
		return types.SetCodeAuthorization{}, fmt.Errorf("signing key required")
	}
	if e := uintCheck(chain, 256, true, "chain ID"); e != nil {
		return types.SetCodeAuthorization{}, e
	}
	if !nonzero(implementation) {
		return types.SetCodeAuthorization{}, fmt.Errorf("implementation required")
	}
	return types.SignSetCode(key, types.SetCodeAuthorization{ChainID: *uint256.MustFromBig(chain), Address: implementation, Nonce: nonce})
}

func (c *Client) verifyPayer(ctx context.Context, h common.Hash, sig []byte, payer common.Address, f deploymentFrame) error {
	code, e := c.provider.CodeAt(ctx, payer, f.BlockNumber)
	if e != nil {
		return e
	}
	if len(code) == 0 {
		return VerifySignature(h, sig, payer)
	}
	a := mustABI(`[{"name":"isValidSignature","type":"function","inputs":[{"type":"bytes32"},{"type":"bytes"}],"outputs":[{"type":"bytes4"}]}]`)
	b, e := a.Pack("isValidSignature", h, sig)
	if e != nil {
		return e
	}
	raw, e := c.provider.CallContract(ctx, ethereum.CallMsg{To: &payer, Data: b}, f.BlockNumber)
	if e != nil {
		return e
	}
	v, e := a.Methods["isValidSignature"].Outputs.Unpack(raw)
	if e != nil {
		return e
	}
	if v[0].([4]byte) != ([4]byte{0x16, 0x26, 0xba, 0x7e}) {
		return fmt.Errorf("invalid ERC1271 signature")
	}
	return nil
}

type SponsoredPaymentOptions struct {
	Player, Implementation common.Address
	ImplementationCodeHash common.Hash
	Payment                AccountPayment
	AuthorizationList      []types.SetCodeAuthorization
}

// BuildSponsoredPayment validates pinned implementation, installed or explicitly
// supplied delegation, both action nonces and the account authorization.
func (c *Client) BuildSponsoredPayment(ctx context.Context, o SponsoredPaymentOptions) (RouterEnvelope, error) {
	return c.buildSponsoredPhase(ctx, o, false)
}

// BuildSponsoredDraft validates pinned dependencies and live nonces without recovering unsigned wallet authorizations.
func (c *Client) BuildSponsoredDraft(ctx context.Context, o SponsoredPaymentOptions) (RouterEnvelope, error) {
	return c.buildSponsoredPhase(ctx, o, true)
}
func (c *Client) buildSponsoredPhase(ctx context.Context, o SponsoredPaymentOptions, draft bool) (RouterEnvelope, error) {
	var env RouterEnvelope
	p, e := DecodeRouterCall(o.Payment.RouterCalldata)
	if e != nil {
		return env, e
	}
	if p.Intent.Mode != SponsoredNative || p.Intent.Payer != o.Player || o.Payment.Value == nil || o.Payment.Value.Cmp(p.Intent.AmountInMaximum) != 0 || o.Payment.Relayer != p.Intent.Relayer {
		return env, fmt.Errorf("account and Router payment bindings mismatch")
	}
	if !draft {
		if e = c.ValidatePayment(ctx, protocol.TransactionCall{To: c.deployment.RouterAddress, Data: o.Payment.RouterCalldata, Value: o.Payment.Value}, o.Player); e != nil {
			return env, e
		}
	}
	f, e := frame(ctx, c.provider, c.deployment, p.Route.Channel == ChannelZeroEx)
	if e != nil {
		return env, e
	}
	if draft {
		n, err := c.routerNonce(ctx, o.Player, f)
		if err != nil {
			return env, err
		}
		if err = ValidateDraftIntent(p, c.deployment, f.Now, n); err != nil {
			return env, err
		}
		if err = ValidateRoute(p, c.deployment, f.Now, f.SettlerAddress); err != nil {
			return env, err
		}
	}
	if !nonzero(o.Implementation) || o.ImplementationCodeHash == (common.Hash{}) {
		return env, fmt.Errorf("pinned account implementation and code hash required")
	}
	code, e := c.provider.CodeAt(ctx, o.Implementation, f.BlockNumber)
	if e != nil {
		return env, e
	}
	if crypto.Keccak256Hash(code) != o.ImplementationCodeHash {
		return env, fmt.Errorf("account implementation code hash mismatch")
	}
	read := func(to common.Address, name string) (any, error) {
		b, e := account.Pack(name)
		if e != nil {
			return nil, e
		}
		raw, e := c.provider.CallContract(ctx, ethereum.CallMsg{To: &to, Data: b}, f.BlockNumber)
		if e != nil {
			return nil, e
		}
		a, e := account.Methods[name].Outputs.Unpack(raw)
		if e != nil {
			return nil, e
		}
		return a[0], nil
	}
	router, e := read(o.Implementation, "router")
	if e != nil {
		return env, e
	}
	impl, e := read(o.Implementation, "implementation")
	if e != nil {
		return env, e
	}
	pinned, e := read(o.Implementation, "routerCodeHash")
	if e != nil {
		return env, e
	}
	runtime, e := c.provider.CodeAt(ctx, c.deployment.RouterAddress, f.BlockNumber)
	if e != nil {
		return env, e
	}
	routerHash := crypto.Keccak256Hash(runtime)
	if router.(common.Address) != c.deployment.RouterAddress || impl.(common.Address) != o.Implementation || common.Hash(pinned.([32]byte)) != routerHash {
		return env, fmt.Errorf("account dependency mismatch")
	}
	playerCode, e := c.provider.CodeAt(ctx, o.Player, f.BlockNumber)
	if e != nil {
		return env, e
	}
	installed, delegated := types.ParseDelegation(playerCode)
	if len(playerCode) > 0 && (!delegated || installed != o.Implementation) {
		return env, fmt.Errorf("existing account implementation requires explicit wallet management")
	}
	if len(o.AuthorizationList) > 1 {
		return env, fmt.Errorf("only the player's delegation authorization is allowed")
	}
	if len(o.AuthorizationList) == 1 {
		a := o.AuthorizationList[0]
		authority := o.Player
		var e error
		if !draft {
			authority, e = a.Authority()
		}
		if e != nil {
			return env, e
		}
		if authority != o.Player || a.Address != o.Implementation || a.ChainID.ToBig().Cmp(c.deployment.ChainID) != 0 {
			return env, fmt.Errorf("invalid delegation authorization")
		}
	} else if !delegated {
		return env, fmt.Errorf("delegation authorization required")
	}
	nonce := new(big.Int)
	if !delegated {
		reader, ok := c.provider.(AccountStateProvider)
		if !ok {
			return env, fmt.Errorf("provider must expose account storage and EOA nonce for delegation")
		}
		slot := new(big.Int).Sub(crypto.Keccak256Hash([]byte("invisi.payment-router.account.storage.v1")).Big(), big.NewInt(1))
		raw, e := reader.StorageAt(ctx, o.Player, common.BigToHash(slot), f.BlockNumber)
		if e != nil {
			return env, e
		}
		nonce.SetBytes(raw)
	}
	if len(o.AuthorizationList) == 1 {
		reader, ok := c.provider.(AccountStateProvider)
		if !ok {
			return env, fmt.Errorf("provider must expose account storage and EOA nonce for delegation")
		}
		n, e := reader.NonceAt(ctx, o.Player, f.BlockNumber)
		if e != nil {
			return env, e
		}
		if o.AuthorizationList[0].Nonce != n {
			return env, fmt.Errorf("stale EIP-7702 authorization nonce")
		}
	}
	if delegated {
		v, e := read(o.Player, "nonce")
		if e != nil {
			return env, e
		}
		nonce = v.(*big.Int)
	}
	if uintCheck(o.Payment.Nonce, 256, false, "account nonce") != nil || o.Payment.Nonce.Cmp(nonce) != 0 || uintCheck(o.Payment.Deadline, 256, false, "account deadline") != nil || o.Payment.Deadline.Cmp(f.Now) < 0 {
		return env, fmt.Errorf("stale account nonce or deadline")
	}
	h, e := HashNativePayment(c.deployment.ChainID, o.Player, o.Implementation, c.deployment.RouterAddress, routerHash, o.Payment)
	if e != nil {
		return env, e
	}
	if !draft {
		if e = VerifySignature(h, o.Payment.Signature, o.Player); e != nil {
			return env, e
		}
	}
	b, e := EncodeAccountCall(o.Payment)
	if e != nil {
		return env, e
	}
	if e = assertFrame(ctx, c.provider, f); e != nil {
		return env, e
	}
	return RouterEnvelope{To: o.Player, Data: b, Value: "0", AuthorizationList: append([]types.SetCodeAuthorization(nil), o.AuthorizationList...)}, nil
}

// SimulateEnvelope estimates the actual outer transaction, including type-4
// delegation installation, account execution and the complete Router call.
func (c *Client) SimulateEnvelope(ctx context.Context, env RouterEnvelope, from common.Address, policies ...SponsoredPaymentOptions) (uint64, error) {
	if env.Value != "0" || !nonzero(from) {
		return 0, fmt.Errorf("relay envelope must carry zero principal")
	}
	f, e := frame(ctx, c.provider, c.deployment, false)
	if e != nil {
		return 0, e
	}
	if env.To == c.deployment.RouterAddress {
		if len(env.AuthorizationList) != 0 {
			return 0, fmt.Errorf("unexpected Router delegation")
		}
		if e = c.ValidatePayment(ctx, protocol.TransactionCall{To: env.To, Data: env.Data, Value: new(big.Int)}, from); e != nil {
			return 0, e
		}
	} else {
		if len(policies) != 1 {
			return 0, fmt.Errorf("pinned sponsored payment policy required")
		}
		checked, e := c.BuildSponsoredPayment(ctx, policies[0])
		if e != nil {
			return 0, e
		}
		a, e := EnvelopeHash(c.deployment.ChainID, c.deployment.CoreAddress, checked)
		if e != nil {
			return 0, e
		}
		b, e := EnvelopeHash(c.deployment.ChainID, c.deployment.CoreAddress, env)
		if e != nil {
			return 0, e
		}
		if a != b {
			return 0, fmt.Errorf("sponsored envelope differs from validated payment")
		}
		p, e := DecodeAccountCall(env.Data)
		if e != nil {
			return 0, e
		}
		if p.Relayer != from {
			return 0, fmt.Errorf("account relayer mismatch")
		}
		inner, e := DecodeRouterCall(p.RouterCalldata)
		if e != nil {
			return 0, e
		}
		if inner.Intent.Mode != SponsoredNative || inner.Intent.Payer != env.To || inner.Intent.Relayer != from {
			return 0, fmt.Errorf("invalid account envelope")
		}
	}
	msg := ethereum.CallMsg{From: from, To: &env.To, Data: env.Data, Value: new(big.Int), AuthorizationList: env.AuthorizationList}
	if _, e = c.provider.CallContract(ctx, msg, f.BlockNumber); e != nil {
		return 0, e
	}
	gas, e := c.provider.EstimateGasAt(ctx, msg, f.BlockNumber)
	if e != nil {
		return 0, e
	}
	return gas, assertFrame(ctx, c.provider, f)
}

// AccountStateProvider is required when validating a new delegation authorization.
type AccountStateProvider interface {
	StorageAt(context.Context, common.Address, common.Hash, *big.Int) ([]byte, error)
	NonceAt(context.Context, common.Address, *big.Int) (uint64, error)
}
