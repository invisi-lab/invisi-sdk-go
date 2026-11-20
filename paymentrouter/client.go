package paymentrouter

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
	"math/big"
)

// Client owns a detached trusted deployment and an unsigned read-only provider.
type Client struct {
	deployment Deployment
	provider   Provider
}

func NewClient(d Deployment, p Provider) (*Client, error) {
	d, e := normalizeDeployment(d)
	if e != nil {
		return nil, e
	}
	if p == nil {
		return nil, fmt.Errorf("provider required")
	}
	return &Client{d, p}, nil
}
func (c *Client) Deployment() Deployment { return cloneDeployment(c.deployment) }
func (c *Client) PreparePayment(ctx context.Context, o PreparePaymentOptions) (*sdk.PreparedTransaction, error) {
	o.Outputs = append([]sdk.NoteOutput(nil), o.Outputs...)
	if o.Fee != nil {
		v := *o.Fee
		v.Amount = copyInt(v.Amount)
		o.Fee = &v
	}
	if _, e := frame(ctx, c.provider, c.deployment, false); e != nil {
		return nil, e
	}
	if o.Invisi == nil || o.Protocol == nil || o.Wallet == nil {
		return nil, fmt.Errorf("SDK, wallet and protocol required")
	}
	pc := o.Protocol.Context()
	if pc.ChainID != c.deployment.ChainID.Uint64() || pc.ProtocolAddress != c.deployment.CoreAddress {
		return nil, fmt.Errorf("SDK protocol deployment mismatch")
	}
	if len(o.Outputs) < 2 {
		return nil, fmt.Errorf("pure funding requires at least two Note outputs")
	}
	if o.TokenAddress == c.deployment.WETHAddress {
		return nil, fmt.Errorf("WETH output is not supported")
	}
	amount := new(big.Int)
	if o.Fee != nil && o.Fee.Amount != nil {
		amount.Set(o.Fee.Amount)
	}
	for _, out := range o.Outputs {
		a := out.Amount()
		if e := uintCheck(a, 128, true, "output amount"); e != nil {
			return nil, e
		}
		amount.Add(amount, a)
	}
	if e := uintCheck(amount, 128, true, "total funding amount"); e != nil {
		return nil, e
	}
	return o.Invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{TransactionIdentity: sdk.TransactionIdentity{Wallet: o.Wallet, Protocol: o.Protocol}, TransferPreparation: sdk.TransferPreparation{TokenAddress: o.TokenAddress, Outputs: o.Outputs, Fee: o.Fee, Deadline: o.Deadline, Funding: &sdk.FundingSource{Payer: c.deployment.RouterAddress, Amount: amount, Authorization: "caller"}}})
}
func validateParameters(p QuoteParameters, now *big.Int) error {
	for _, v := range []struct {
		n *big.Int
		b int
		p bool
		l string
	}{{p.AmountOut, 128, true, "funding amount"}, {p.MaxInput, 256, true, "ETH budget"}, {p.Deadline, 64, false, "quote deadline"}} {
		if e := uintCheck(v.n, v.b, v.p, v.l); e != nil {
			return e
		}
	}
	if slippageBps(p) < 0 || slippageBps(p) > 10000 {
		return fmt.Errorf("invalid slippageBps")
	}
	if now != nil && p.Deadline.Cmp(now) < 0 {
		return fmt.Errorf("quote expired")
	}
	return nil
}
func (c *Client) checkQuote(q SwapQuote, f deploymentFrame) error {
	if q.ChainID == nil || q.ChainID.Cmp(c.deployment.ChainID) != 0 || q.RouterAddress != c.deployment.RouterAddress {
		return fmt.Errorf("quote deployment mismatch")
	}
	if e := validateParameters(QuoteParameters{TokenOut: q.TokenOut, AmountOut: q.AmountOut, MaxInput: q.MaxInput, Deadline: q.Deadline}, f.Now); e != nil {
		return e
	}
	if e := uintCheck(q.EstimatedInput, 256, false, "estimated input"); e != nil {
		return e
	}
	if q.EstimatedInput.Cmp(q.MaxInput) > 0 {
		return fmt.Errorf("estimated input exceeds budget")
	}
	switch q.Channel {
	case UniswapV3:
		return ValidateV3Path(q.ReversePath, c.deployment.WETHAddress, q.TokenOut)
	case ZeroEx:
		if c.deployment.ZeroEx == nil || !nonzero(f.SettlerAddress) {
			return fmt.Errorf("0x deployment required")
		}
		if q.SettlerAddress != f.SettlerAddress {
			return fmt.Errorf("0x Settler changed; refresh the quote")
		}
		v, e := c.checkZero(q, f)
		if e != nil {
			return e
		}
		if q.Deadline.Cmp(v.NativeDeadline) > 0 {
			return fmt.Errorf("quote outlives native guard")
		}
		return nil
	default:
		return fmt.Errorf("unsupported swap channel")
	}
}
func (c *Client) checkZero(q SwapQuote, f deploymentFrame) (CheckedZeroExPayload, error) {
	return ValidateZeroExPayload(ZeroExPayloadOptions{Target: q.Target, Data: q.Data, RouterAddress: c.deployment.RouterAddress, TokenOut: q.TokenOut, AmountOut: q.AmountOut, MaxInput: q.MaxInput, SettlerAddress: f.SettlerAddress, AllowanceHolderAddress: c.deployment.ZeroEx.AllowanceHolderAddress, Now: f.Now})
}
func (c *Client) Quote(ctx context.Context, a QuoteAdapter, p QuoteParameters) (SwapQuote, error) {
	p = cloneParameters(p)
	fail := func(e error) (SwapQuote, error) { return SwapQuote{}, e }
	if e := ctx.Err(); e != nil {
		return fail(e)
	}
	if a == nil {
		return fail(fmt.Errorf("adapter required"))
	}
	if e := validateParameters(p, nil); e != nil {
		return fail(e)
	}
	if p.TokenOut == c.deployment.WETHAddress {
		return fail(fmt.Errorf("WETH output is not supported"))
	}
	f, e := frame(ctx, c.provider, c.deployment, a.Channel() == ZeroEx)
	if e != nil {
		return fail(e)
	}
	if e = validateParameters(p, f.Now); e != nil {
		return fail(e)
	}
	r := SwapQuoteRequest{p, cloneDeployment(c.deployment), copyInt(f.Now), copyInt(f.BlockNumber), f.SettlerAddress}
	q, e := a.Quote(ctx, cloneRequest(r))
	if e != nil {
		return fail(e)
	}
	q = cloneQuote(q)
	if q.Channel != a.Channel() {
		return fail(fmt.Errorf("quote channel mismatch"))
	}
	if e = c.checkQuote(q, f); e != nil {
		return fail(e)
	}
	if q.TokenOut != p.TokenOut || q.AmountOut.Cmp(p.AmountOut) != 0 {
		return fail(fmt.Errorf("quoted payment amount or token changed"))
	}
	if q.MaxInput.Cmp(p.MaxInput) > 0 || q.Deadline.Cmp(p.Deadline) > 0 {
		return fail(fmt.Errorf("quote exceeds authorized budget or deadline"))
	}
	if e = assertFrame(ctx, c.provider, f); e != nil {
		return fail(e)
	}
	current, e := frame(ctx, c.provider, c.deployment, q.Channel == ZeroEx)
	if e != nil {
		return fail(e)
	}
	if e = c.checkQuote(q, current); e != nil {
		return fail(e)
	}
	return q, ctx.Err()
}

// InspectCorePayment decodes and validates the existing proof statement without modifying it.
func InspectCorePayment(call protocol.TransactionCall, d Deployment, now *big.Int) ([]byte, protocol.TransactRequest, error) {
	var r protocol.TransactRequest
	fail := func(s string) ([]byte, protocol.TransactRequest, error) { return nil, r, fmt.Errorf("%s", s) }
	if call.To != d.CoreAddress || call.Value == nil || call.Value.Sign() != 0 {
		return fail("core payment deployment/value mismatch")
	}
	if len(call.Data) < 4 || string(call.Data[:4]) != string(coreMethod.ID) {
		return fail("invalid core transact selector")
	}
	a, e := coreMethod.Inputs.Unpack(call.Data[4:])
	if e != nil {
		return nil, r, e
	}
	proof := a[0].([]byte)
	if len(proof) == 0 {
		return fail("core proof is required")
	}
	r = *abi.ConvertType(a[1], new(protocol.TransactRequest)).(*protocol.TransactRequest)
	f, e := protocol.NormalizeFunding(r.Funding)
	if e != nil {
		return nil, r, e
	}
	if f.Payer != d.RouterAddress {
		return fail("core proof funding payer must be the Router")
	}
	if f.Amount.Sign() <= 0 || f.SigNonce.Sign() != 0 || f.Deadline.Sign() != 0 || len(f.Signature) != 0 || f.Permit.Deadline.Sign() != 0 || f.Permit.V != 0 || f.Permit.R != ([32]byte{}) || f.Permit.S != ([32]byte{}) {
		return fail("canonical direct Router funding is required")
	}
	p := r.Params
	if len(p.Sources) != 0 || len(p.Nullifiers) != 0 || p.WithdrawalAmount.Sign() != 0 || nonzero(p.WithdrawalAddress) {
		return fail("Router payment requires zero private inputs and no withdrawal")
	}
	if p.Token == d.WETHAddress {
		return fail("WETH output is not supported")
	}
	_, e = protocol.NormalizeTransactRequest(protocol.TransactionContext{ChainID: d.ChainID, ProtocolAddress: d.CoreAddress, TokenAddress: p.Token, Now: now}, p, r.Deliveries, f)
	return proof, r, e
}

// BuildPaymentTransaction accepts a core call or *sdk.PreparedTransaction with a ready proof.
func (c *Client) BuildPaymentTransaction(ctx context.Context, payment any, quote SwapQuote, recipient common.Address) (protocol.TransactionCall, error) {
	q := cloneQuote(quote)
	var empty protocol.TransactionCall
	if !nonzero(recipient) || recipient == c.deployment.RouterAddress {
		return empty, fmt.Errorf("invalid refund recipient")
	}
	var call protocol.TransactionCall
	switch p := payment.(type) {
	case protocol.TransactionCall:
		call = p
	case *sdk.PreparedTransaction:
		if p == nil {
			return empty, fmt.Errorf("payment required")
		}
		v, e := p.ExportTransaction(ctx)
		if e != nil {
			return empty, e
		}
		call = v
	default:
		return empty, fmt.Errorf("core call or prepared transaction required")
	}
	call.Data = append([]byte(nil), call.Data...)
	call.Value = copyInt(call.Value)
	f, e := frame(ctx, c.provider, c.deployment, q.Channel == ZeroEx)
	if e != nil {
		return empty, e
	}
	if e = c.checkQuote(q, f); e != nil {
		return empty, e
	}
	proof, r, e := InspectCorePayment(call, c.deployment, f.Now)
	if e != nil {
		return empty, e
	}
	if r.Params.Token != q.TokenOut {
		return empty, fmt.Errorf("quote output token mismatch")
	}
	if r.Funding.Amount.Cmp(q.AmountOut) != 0 {
		return empty, fmt.Errorf("quote funding amount mismatch")
	}
	if q.Deadline.Cmp(r.Params.Deadline) > 0 {
		return empty, fmt.Errorf("quote outlives core proof deadline")
	}
	var data []byte
	if q.Channel == UniswapV3 {
		data, e = routerABI.Pack("pay", proof, r, q.ReversePath, q.Deadline, recipient)
	} else {
		v, err := c.checkZero(q, f)
		if err != nil {
			return empty, err
		}
		data, e = routerABI.Pack("payWithZeroEx", proof, r, v.Target, v.Data, q.Deadline, recipient)
	}
	if e != nil {
		return empty, e
	}
	if e = assertFrame(ctx, c.provider, f); e != nil {
		return empty, e
	}
	return protocol.TransactionCall{To: c.deployment.RouterAddress, Data: data, Value: copyInt(q.MaxInput)}, nil
}

// SimulatePayment calls and estimates the entire payable Router transaction at one block.
func (c *Client) SimulatePayment(ctx context.Context, input protocol.TransactionCall, from common.Address) (RouterSimulation, error) {
	var result RouterSimulation
	call := input
	call.Data = append([]byte(nil), input.Data...)
	call.Value = copyInt(input.Value)
	if !nonzero(from) {
		return result, fmt.Errorf("sender must be nonzero")
	}
	if call.To != c.deployment.RouterAddress {
		return result, fmt.Errorf("simulation target must be the payment Router")
	}
	if e := uintCheck(call.Value, 256, true, "ETH budget"); e != nil {
		return result, e
	}
	if len(call.Data) < 4 {
		return result, fmt.Errorf("unsupported Router call")
	}
	m, e := routerABI.MethodById(call.Data[:4])
	if e != nil {
		return result, fmt.Errorf("unsupported Router call")
	}
	a, e := m.Inputs.Unpack(call.Data[4:])
	if e != nil {
		return result, e
	}
	f, e := frame(ctx, c.provider, c.deployment, m.Name == "payWithZeroEx")
	if e != nil {
		return result, e
	}
	core, e := coreMethod.Inputs.Pack(a[0], a[1])
	if e != nil {
		return result, e
	}
	_, r, e := InspectCorePayment(protocol.TransactionCall{To: c.deployment.CoreAddress, Data: append(append([]byte(nil), coreMethod.ID...), core...), Value: new(big.Int)}, c.deployment, f.Now)
	if e != nil {
		return result, e
	}
	i := 3
	if m.Name == "payWithZeroEx" {
		i = 4
	}
	deadline := a[i].(*big.Int)
	if e = uintCheck(deadline, 64, false, "swap deadline"); e != nil {
		return result, e
	}
	if deadline.Cmp(f.Now) < 0 {
		return result, fmt.Errorf("Router swap expired")
	}
	recipient := a[i+1].(common.Address)
	if !nonzero(recipient) || recipient == c.deployment.RouterAddress {
		return result, fmt.Errorf("invalid refund recipient")
	}
	if m.Name == "pay" {
		e = ValidateV3Path(a[2].([]byte), c.deployment.WETHAddress, r.Params.Token)
	} else {
		_, e = c.checkZero(SwapQuote{Target: a[2].(common.Address), Data: a[3].([]byte), TokenOut: r.Params.Token, AmountOut: r.Funding.Amount, MaxInput: call.Value}, f)
	}
	if e != nil {
		return result, e
	}
	msg := ethereum.CallMsg{From: from, To: &call.To, Data: call.Data, Value: call.Value}
	raw, e := c.provider.CallContract(ctx, msg, copyInt(f.BlockNumber))
	if e != nil {
		return result, e
	}
	gas, e := c.provider.EstimateGasAt(ctx, msg, copyInt(f.BlockNumber))
	if e != nil {
		return result, e
	}
	v, e := m.Outputs.Unpack(raw)
	if e != nil {
		return result, e
	}
	spent := v[0].(*big.Int)
	if spent.Cmp(call.Value) > 0 {
		return result, fmt.Errorf("simulated input exceeds ETH budget")
	}
	if e = assertFrame(ctx, c.provider, f); e != nil {
		return result, e
	}
	return RouterSimulation{gas, spent, new(big.Int).Sub(call.Value, spent)}, nil
}
