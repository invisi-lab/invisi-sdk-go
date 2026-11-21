package broadcaster

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
)

// The transport package cannot import paymentrouter (the payment client imports
// this transport). Derive the Core tuple from the current Protocol ABI and keep
// only the canonical Router/account entry-point argument shapes here.
type relayIntent struct {
	Payer, TokenIn             common.Address
	AmountInMaximum            *big.Int
	RefundRecipient, Relayer   common.Address
	Nonce, Deadline            *big.Int
	CoreRequestHash, RouteHash common.Hash
	Mode                       uint8
}

var relayMethods = func() map[string]abi.Method {
	core, e := protocol.ProtocolABI()
	if e != nil {
		panic(e)
	}
	methods := map[string]abi.Method{}
	for _, m := range core.Methods {
		if m.RawName == "transact" && len(m.Inputs) == 2 && m.Inputs[1].Type.T == abi.TupleTy {
			methods["transact"] = m
			break
		}
	}
	if _, ok := methods["transact"]; !ok {
		panic("missing canonical Core transact")
	}
	tuple := func(fields ...abi.ArgumentMarshaling) abi.Type {
		t, e := abi.NewType("tuple", "", fields)
		if e != nil {
			panic(e)
		}
		return t
	}
	field := func(n, t string) abi.ArgumentMarshaling { return abi.ArgumentMarshaling{Name: n, Type: t} }
	intent := tuple(field("payer", "address"), field("tokenIn", "address"), field("amountInMaximum", "uint256"), field("refundRecipient", "address"), field("relayer", "address"), field("nonce", "uint256"), field("deadline", "uint256"), field("coreRequestHash", "bytes32"), field("routeHash", "bytes32"), field("mode", "uint8"))
	route := tuple(field("channel", "uint8"), field("path", "bytes"), field("target", "address"), field("data", "bytes"))
	permit := tuple(field("deadline", "uint256"), field("v", "uint8"), field("r", "bytes32"), field("s", "bytes32"))
	base := append(append(abi.Arguments{}, methods["transact"].Inputs...), abi.Argument{Type: intent}, abi.Argument{Type: route})
	add := func(name string, args abi.Arguments) {
		methods[name] = abi.NewMethod(name, name, abi.Function, "payable", false, true, args, nil)
	}
	add("payNative", base)
	token := append(append(abi.Arguments{}, base...), abi.Argument{Type: permit}, routerKind("bytes"))
	add("payTokenFor", token)
	add("executeNativePayment", abi.Arguments{routerKind("bytes"), routerKind("uint256"), routerKind("uint256"), routerKind("uint256"), routerKind("address"), routerKind("bytes")})
	return methods
}()

func unpackRelay(name string, data []byte) ([]any, error) {
	m := relayMethods[name]
	if len(data) < 4 || !bytes.Equal(data[:4], m.ID) {
		return nil, fmt.Errorf("unsupported relay entry point: expected %x for %s", m.ID, m.Sig)
	}
	values, e := m.Inputs.Unpack(data[4:])
	if e != nil {
		return nil, e
	}
	canonical, e := m.Inputs.Pack(values...)
	if e != nil || !bytes.Equal(canonical, data[4:]) {
		return nil, fmt.Errorf("noncanonical relay calldata")
	}
	return values, nil
}

// relayDeadline also verifies that data is the exact Core call authorized inside
// the envelope. This is checked again at submission after caller mutation.
func relayDeadline(data []byte, env RouterEnvelope) (*big.Int, error) {
	inner := []byte(env.Data)
	var accountDeadline *big.Int
	var accountValue *big.Int
	var accountRelayer common.Address
	native := len(inner) >= 4 && bytes.Equal(inner[:4], relayMethods["executeNativePayment"].ID)
	if native {
		args, e := unpackRelay("executeNativePayment", inner)
		if e != nil {
			return nil, e
		}
		inner = args[0].([]byte)
		accountValue = args[1].(*big.Int)
		accountDeadline = args[3].(*big.Int)
		accountRelayer = args[4].(common.Address)
	} else if len(env.AuthorizationList) != 0 {
		return nil, fmt.Errorf("unexpected token delegation authorization")
	}
	name := "payTokenFor"
	if native {
		name = "payNative"
	}
	args, e := unpackRelay(name, inner)
	if e != nil {
		return nil, e
	}
	i := abi.ConvertType(args[2], new(relayIntent)).(*relayIntent)
	if native {
		if i.Mode != 3 || i.Payer != env.To || i.TokenIn != (common.Address{}) || accountValue.Cmp(i.AmountInMaximum) != 0 || accountRelayer != i.Relayer {
			return nil, fmt.Errorf("invalid native relay bindings")
		}
	} else if i.Mode != 2 || i.TokenIn == (common.Address{}) || len(args[5].([]byte)) == 0 {
		return nil, fmt.Errorf("invalid token relay bindings")
	}
	if i.Payer == (common.Address{}) || i.Relayer == (common.Address{}) {
		return nil, fmt.Errorf("missing relay payer/relayer")
	}
	core := relayMethods["transact"]
	encoded, e := core.Inputs.Pack(args[0], args[1])
	if e != nil {
		return nil, e
	}
	encoded = append(append([]byte{}, core.ID...), encoded...)
	if !bytes.Equal(data, encoded) {
		return nil, fmt.Errorf("embedded Core calldata mismatch")
	}
	deadline := new(big.Int).Set(i.Deadline)
	if accountDeadline != nil && accountDeadline.Cmp(deadline) < 0 {
		deadline.Set(accountDeadline)
	}
	return deadline, nil
}
