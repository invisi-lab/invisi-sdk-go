package paymentrouter

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"math/big"
)

// CoreCall returns the canonical embedded Core transaction.
func (p DecodedPayment) CoreCall(core common.Address) (protocol.TransactionCall, error) {
	b, e := packCall(coreMethod, p.Proof, p.Request)
	return protocol.TransactionCall{To: core, Data: b, Value: new(big.Int)}, e
}
func methodFor(mode uint8) string {
	switch mode {
	case DirectNative, SponsoredNative:
		return "payNative"
	case DirectToken:
		return "payToken"
	case RelayedToken:
		return "payTokenFor"
	}
	return ""
}
func ValidateIntent(p DecodedPayment, d Deployment, now, expectedNonce *big.Int) error {
	i := p.Intent
	if e := uintCheck(i.AmountInMaximum, 256, true, "budget"); e != nil {
		return e
	}
	for _, v := range []*big.Int{i.Deadline, i.Nonce, now, expectedNonce} {
		if e := uintCheck(v, 256, false, "deadline or nonce"); e != nil {
			return e
		}
	}
	if !nonzero(i.Payer) || i.Payer == d.RouterAddress || !nonzero(i.RefundRecipient) || i.RefundRecipient == d.RouterAddress || i.Deadline.Cmp(now) < 0 || i.Nonce.Cmp(expectedNonce) != 0 {
		return fmt.Errorf("invalid payer, refund recipient, deadline or Router nonce")
	}
	if methodFor(i.Mode) == "" || p.Method != methodFor(i.Mode) || nonzero(i.TokenIn) != (i.Mode == DirectToken || i.Mode == RelayedToken) || nonzero(i.Relayer) != (i.Mode == RelayedToken || i.Mode == SponsoredNative) {
		return fmt.Errorf("invalid payment mode bindings")
	}
	if p.Method != "payTokenFor" && len(p.Signature) != 0 {
		return fmt.Errorf("unexpected intent signature")
	}
	if p.Permit.Deadline != nil {
		if e := uintCheck(p.Permit.Deadline, 256, false, "permit deadline"); e != nil {
			return e
		}
		if p.Method == "payNative" && p.Permit.Deadline.Sign() != 0 {
			return fmt.Errorf("unexpected native permit")
		}
	}
	if (p.Permit.Deadline == nil || p.Permit.Deadline.Sign() == 0) && (p.Permit.V != 0 || p.Permit.R != ([32]byte{}) || p.Permit.S != ([32]byte{})) {
		return fmt.Errorf("inconsistent empty permit")
	}
	call, e := p.CoreCall(d.CoreAddress)
	if e != nil {
		return e
	}
	if _, _, e = inspectCorePayment(call, d, now, true); e != nil {
		return e
	}
	h, e := HashCoreRequest(p.Proof, p.Request)
	if e != nil || h != i.CoreRequestHash {
		return fmt.Errorf("Core request hash mismatch")
	}
	h, e = HashRoute(p.Route)
	if e != nil || h != i.RouteHash {
		return fmt.Errorf("route hash mismatch")
	}
	return nil
}
func ValidateRoute(p DecodedPayment, d Deployment, now *big.Int, settler common.Address) error {
	i, r := p.Intent, p.Route
	output := p.Request.Params.Token
	required := p.Request.Funding.Amount
	if e := uintCheck(required, 128, true, "funding amount"); e != nil {
		return e
	}
	if e := uintCheck(i.AmountInMaximum, 256, true, "input maximum"); e != nil {
		return e
	}
	input := i.TokenIn
	if !nonzero(input) {
		input = d.WETHAddress
	}
	switch r.Channel {
	case ChannelNone:
		if input != output || len(r.Path) != 0 || nonzero(r.Target) || len(r.Data) != 0 || i.AmountInMaximum.Cmp(required) < 0 {
			return fmt.Errorf("invalid no-swap route")
		}
	case ChannelUniswapV3:
		if input == output || nonzero(r.Target) || len(r.Data) != 0 {
			return fmt.Errorf("invalid V3 route")
		}
		return ValidateV3Path(r.Path, input, output)
	case ChannelZeroEx:
		if input == output || len(r.Path) != 0 || d.ZeroEx == nil {
			return fmt.Errorf("invalid 0x route")
		}
		o := ZeroExPayloadOptions{Target: r.Target, Data: r.Data, RouterAddress: d.RouterAddress, TokenOut: output, AmountOut: required, MaxInput: i.AmountInMaximum, SettlerAddress: settler, AllowanceHolderAddress: d.ZeroEx.AllowanceHolderAddress, Now: now}
		if nonzero(i.TokenIn) {
			return ValidateZeroExTokenPayload(i.TokenIn, o)
		}
		v, e := ValidateZeroExPayload(o)
		if e != nil {
			return e
		}
		if !bytes.Equal(r.Data, v.Data) {
			return fmt.Errorf("noncanonical native 0x payload")
		}
	default:
		return fmt.Errorf("unknown route channel")
	}
	return nil
}
func ValidateZeroExTokenPayload(input common.Address, o ZeroExPayloadOptions) error {
	if !nonzero(input) || !nonzero(o.TokenOut) || input == o.TokenOut || !nonzero(o.RouterAddress) || !nonzero(o.Target) || o.Target != o.AllowanceHolderAddress || !nonzero(o.SettlerAddress) || o.SettlerAddress == o.Target || o.SettlerAddress == o.RouterAddress {
		return fmt.Errorf("invalid 0x token deployment")
	}
	if e := uintCheck(o.MaxInput, 256, true, "budget"); e != nil {
		return e
	}
	if e := uintCheck(o.AmountOut, 128, true, "output amount"); e != nil {
		return e
	}
	a, e := strictArgs(zeroABI.Methods["exec"], o.Data)
	if e != nil {
		return e
	}
	if a[0].(common.Address) != o.SettlerAddress || a[1].(common.Address) != input || a[2].(*big.Int).Cmp(o.MaxInput) > 0 || a[3].(common.Address) != o.SettlerAddress {
		return fmt.Errorf("invalid 0x Holder envelope")
	}
	b, e := strictArgs(zeroABI.Methods["execute"], a[4].([]byte))
	if e != nil {
		return e
	}
	s := asTuple[slippage](b[0])
	if s.Recipient != o.RouterAddress || s.BuyToken != o.TokenOut || s.MinAmountOut.Cmp(o.AmountOut) != 0 {
		return fmt.Errorf("invalid 0x output envelope")
	}
	return nil
}
