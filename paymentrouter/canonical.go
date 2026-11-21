package paymentrouter

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"math/big"
)

//go:embed PaymentRouter.abi.json
var canonicalJSON string

//go:embed PaymentRouterAccount.abi.json
var accountJSON string
var canonical = mustABI(canonicalJSON)
var account = mustABI(accountJSON)

func CanonicalRouterABI() abi.ABI      { return mustABI(canonicalJSON) }
func PaymentRouterAccountABI() abi.ABI { return mustABI(accountJSON) }

const (
	DirectNative uint8 = iota
	DirectToken
	RelayedToken
	SponsoredNative
)
const (
	ChannelNone uint8 = iota
	ChannelUniswapV3
	ChannelZeroEx
)

type Intent struct {
	Payer, TokenIn             common.Address
	AmountInMaximum            *big.Int
	RefundRecipient, Relayer   common.Address
	Nonce, Deadline            *big.Int
	CoreRequestHash, RouteHash common.Hash
	Mode                       uint8
}
type Route struct {
	Channel uint8
	Path    []byte
	Target  common.Address
	Data    []byte
}
type Permit struct {
	Deadline *big.Int
	V        uint8
	R, S     [32]byte
}
type DecodedPayment struct {
	Method    string
	Proof     []byte
	Request   protocol.TransactRequest
	Intent    Intent
	Route     Route
	Permit    Permit
	Signature []byte
}
type AccountPayment struct {
	RouterCalldata         []byte
	Value, Nonce, Deadline *big.Int
	Relayer                common.Address
	Signature              []byte
}

// pack handles unset tuple integers before go-ethereum reflects on their values.
func pack(args abi.Arguments, values ...any) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			data = nil
			err = fmt.Errorf("invalid ABI value: %v", r)
		}
	}()
	return args.Pack(values...)
}
func packCall(m abi.Method, values ...any) ([]byte, error) {
	b, e := pack(m.Inputs, values...)
	if e != nil {
		return nil, e
	}
	return append(bytes.Clone(m.ID), b...), nil
}
func strictArgs(m abi.Method, b []byte) ([]any, error) {
	if len(b) < 4 || !bytes.Equal(b[:4], m.ID) {
		return nil, fmt.Errorf("unexpected selector")
	}
	a, e := m.Inputs.Unpack(b[4:])
	if e != nil {
		return nil, e
	}
	canon, e := pack(m.Inputs, a...)
	if e != nil || !bytes.Equal(canon, b[4:]) {
		return nil, fmt.Errorf("noncanonical %s encoding", m.Name)
	}
	return a, nil
}
func EncodeRouterCall(p DecodedPayment) ([]byte, error) {
	a := []any{p.Proof, p.Request, p.Intent, p.Route}
	switch p.Method {
	case "payNative":
	case "payToken", "payTokenFor":
		if p.Permit.Deadline == nil {
			p.Permit.Deadline = new(big.Int)
		}
		a = append(a, p.Permit)
		if p.Method == "payTokenFor" {
			a = append(a, p.Signature)
		}
	default:
		return nil, fmt.Errorf("unknown payment method %q", p.Method)
	}
	return packCall(canonical.Methods[p.Method], a...)
}
func asTuple[T any](v any) T { return *abi.ConvertType(v, new(T)).(*T) }
func DecodeRouterCall(data []byte) (DecodedPayment, error) {
	var p DecodedPayment
	m, e := canonical.MethodById(data)
	if e != nil {
		return p, e
	}
	switch m.Name {
	case "payNative", "payToken", "payTokenFor":
	default:
		return p, fmt.Errorf("unexpected payment method")
	}
	a, e := strictArgs(*m, data)
	if e != nil {
		return p, e
	}
	p = DecodedPayment{Method: m.Name, Proof: a[0].([]byte), Request: asTuple[protocol.TransactRequest](a[1]), Intent: asTuple[Intent](a[2]), Route: asTuple[Route](a[3])}
	if p.Intent.Mode > SponsoredNative || p.Route.Channel > ChannelZeroEx {
		return DecodedPayment{}, fmt.Errorf("invalid enum")
	}
	if len(a) > 4 {
		p.Permit = asTuple[Permit](a[4])
	}
	if len(a) > 5 {
		p.Signature = a[5].([]byte)
	}
	return p, nil
}
func EncodeAccountCall(p AccountPayment) ([]byte, error) {
	return packCall(account.Methods["executeNativePayment"], p.RouterCalldata, p.Value, p.Nonce, p.Deadline, p.Relayer, p.Signature)
}
func DecodeAccountCall(b []byte) (AccountPayment, error) {
	a, e := strictArgs(account.Methods["executeNativePayment"], b)
	if e != nil {
		return AccountPayment{}, e
	}
	return AccountPayment{a[0].([]byte), a[1].(*big.Int), a[2].(*big.Int), a[3].(*big.Int), a[4].(common.Address), a[5].([]byte)}, nil
}
func hashArgs(args abi.Arguments, values ...any) (common.Hash, error) {
	b, e := pack(args, values...)
	if e != nil {
		return common.Hash{}, e
	}
	return crypto.Keccak256Hash(b), nil
}
func kind(name string) abi.Argument {
	t, e := abi.NewType(name, "", nil)
	if e != nil {
		panic(e)
	}
	return abi.Argument{Type: t}
}
func HashCoreRequest(proof []byte, r protocol.TransactRequest) (common.Hash, error) {
	return hashArgs(abi.Arguments{kind("bytes32"), coreMethod.Inputs[1]}, crypto.Keccak256Hash(proof), r)
}
func HashRoute(r Route) (common.Hash, error) {
	return hashArgs(canonical.Methods["payNative"].Inputs[3:4], r)
}
