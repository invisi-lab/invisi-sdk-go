// Package paymentrouter provides detached native ETH payment preparation, quoting,
// encoding and simulation. It never signs or broadcasts a transaction.
package paymentrouter

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
)

// NativeToken is the immutable 0x native ETH sentinel.
const NativeToken = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"

var nativeToken = common.HexToAddress(NativeToken)
var zeroABI = mustABI(`[
 {"type":"function","name":"execute","inputs":[{"name":"slippage","type":"tuple","components":[{"name":"recipient","type":"address"},{"name":"buyToken","type":"address"},{"name":"minAmountOut","type":"uint256"}]},{"name":"actions","type":"bytes[]"},{"name":"zid","type":"bytes32"}],"outputs":[{"type":"bool"}]},
 {"type":"function","name":"exec","inputs":[{"name":"operator","type":"address"},{"name":"token","type":"address"},{"name":"amount","type":"uint256"},{"name":"target","type":"address"},{"name":"data","type":"bytes"}],"outputs":[{"type":"bytes"}]},
 {"type":"function","name":"NATIVE_CHECK","inputs":[{"type":"uint256"},{"type":"uint256"}],"outputs":[]}
]`)

func mustABI(s string) abi.ABI {
	a, e := abi.JSON(strings.NewReader(s))
	if e != nil {
		panic(e)
	}
	return a
}
func uintCheck(v *big.Int, bits int, positive bool, label string) error {
	if v == nil || v.Sign() < 0 || v.BitLen() > bits || (positive && v.Sign() == 0) {
		return fmt.Errorf("invalid %s", label)
	}
	return nil
}
func copyInt(v *big.Int) *big.Int {
	if v == nil {
		return nil
	}
	return new(big.Int).Set(v)
}
func nonzero(a common.Address) bool { return a != (common.Address{}) }
func unpack(a abi.ABI, name string, data []byte) ([]any, error) {
	m := a.Methods[name]
	if len(data) < 4 || !bytes.Equal(data[:4], m.ID) {
		return nil, fmt.Errorf("invalid %s selector", name)
	}
	return m.Inputs.Unpack(data[4:])
}

// EncodeV3ExactOutputPath converts forward token/fee inputs to a reversed exact-output path.
func EncodeV3ExactOutputPath(tokens []common.Address, fees []uint32) ([]byte, error) {
	if len(tokens) < 2 || len(fees) != len(tokens)-1 {
		return nil, fmt.Errorf("invalid V3 path length")
	}
	for _, a := range tokens {
		if !nonzero(a) {
			return nil, fmt.Errorf("path token must be nonzero")
		}
	}
	out := append([]byte(nil), tokens[len(tokens)-1].Bytes()...)
	for i := len(fees) - 1; i >= 0; i-- {
		f := fees[i]
		if f == 0 || f >= 1000000 || tokens[i] == tokens[i+1] {
			return nil, fmt.Errorf("invalid V3 fee or adjacent tokens")
		}
		out = append(out, byte(f>>16), byte(f>>8), byte(f))
		out = append(out, tokens[i].Bytes()...)
	}
	return out, nil
}
func ValidateV3Path(path []byte, weth, output common.Address) error {
	if len(path) < 43 || (len(path)-20)%23 != 0 {
		return fmt.Errorf("invalid V3 path length")
	}
	if common.BytesToAddress(path[:20]) != output || common.BytesToAddress(path[len(path)-20:]) != weth {
		return fmt.Errorf("V3 path endpoints mismatch")
	}
	tokens := []common.Address{}
	fees := []uint32{}
	for i := 0; i < len(path); i += 23 {
		tokens = append(tokens, common.BytesToAddress(path[i:i+20]))
		if i+20 < len(path) {
			fees = append(fees, uint32(path[i+20])<<16|uint32(path[i+21])<<8|uint32(path[i+22]))
		}
	}
	_, e := EncodeV3ExactOutputPath(tokens, fees)
	return e
}

type ZeroExPayloadOptions struct {
	Target                                 common.Address
	Data                                   []byte
	RouterAddress, TokenOut                common.Address
	AmountOut, MaxInput                    *big.Int
	SettlerAddress, AllowanceHolderAddress common.Address
	Now                                    *big.Int
}
type CheckedZeroExPayload struct {
	Target         common.Address
	Data           []byte
	SettlerAddress common.Address
	NativeDeadline *big.Int
}
type slippage struct {
	Recipient, BuyToken common.Address
	MinAmountOut        *big.Int
}

// ValidateZeroExPayload authenticates the native envelope, leaving route actions opaque.
// Returned calldata is canonicalized, including a possible AllowanceHolder wrapper.
func ValidateZeroExPayload(o ZeroExPayloadOptions) (CheckedZeroExPayload, error) {
	fail := func(s string) (CheckedZeroExPayload, error) { return CheckedZeroExPayload{}, fmt.Errorf("%s", s) }
	for _, a := range []common.Address{o.Target, o.SettlerAddress, o.AllowanceHolderAddress, o.RouterAddress} {
		if !nonzero(a) {
			return fail("0x address must be nonzero")
		}
	}
	if o.SettlerAddress == o.AllowanceHolderAddress || o.SettlerAddress == o.RouterAddress {
		return fail("invalid Settler deployment")
	}
	for _, v := range []struct {
		n *big.Int
		b int
		p bool
		l string
	}{{o.MaxInput, 256, true, "native budget"}, {o.AmountOut, 128, true, "funding amount"}, {o.Now, 64, false, "chain timestamp"}} {
		if e := uintCheck(v.n, v.b, v.p, v.l); e != nil {
			return CheckedZeroExPayload{}, e
		}
	}
	data := o.Data
	allowance := new(big.Int)
	wrapped := o.Target == o.AllowanceHolderAddress
	if wrapped {
		a, e := unpack(zeroABI, "exec", data)
		if e != nil {
			return CheckedZeroExPayload{}, e
		}
		if a[0].(common.Address) != o.SettlerAddress || a[3].(common.Address) != o.SettlerAddress {
			return fail("0x nested target mismatch")
		}
		if a[1].(common.Address) != nativeToken {
			return fail("0x input must be native ETH")
		}
		allowance = a[2].(*big.Int)
		if allowance.Cmp(o.MaxInput) > 0 {
			return fail("0x allowance exceeds budget")
		}
		data = a[4].([]byte)
	} else if o.Target != o.SettlerAddress {
		return fail("0x target is not the current Settler")
	}
	a, e := unpack(zeroABI, "execute", data)
	if e != nil {
		return CheckedZeroExPayload{}, e
	}
	s := *abi.ConvertType(a[0], new(slippage)).(*slippage)
	if s.Recipient != o.RouterAddress {
		return fail("0x recipient mismatch")
	}
	if s.BuyToken != o.TokenOut {
		return fail("0x output token mismatch")
	}
	if s.MinAmountOut.Cmp(o.AmountOut) != 0 {
		return fail("0x exact output amount mismatch")
	}
	actions := a[1].([][]byte)
	var deadline *big.Int
	for _, act := range actions {
		if len(act) < 4 {
			return fail("malformed 0x action")
		}
		if !bytes.Equal(act[:4], zeroABI.Methods["NATIVE_CHECK"].ID) {
			continue
		}
		if len(act) != 68 || deadline != nil {
			return fail("invalid or duplicate native guard")
		}
		g, e := unpack(zeroABI, "NATIVE_CHECK", act)
		if e != nil {
			return CheckedZeroExPayload{}, e
		}
		deadline = g[0].(*big.Int)
		if deadline.Cmp(o.Now) < 0 || g[1].(*big.Int).Cmp(o.MaxInput) != 0 {
			return fail("0x native deadline or budget mismatch")
		}
	}
	if deadline == nil {
		return fail("missing 0x native guard")
	}
	data, e = zeroABI.Pack("execute", slippage{o.RouterAddress, o.TokenOut, o.AmountOut}, actions, a[2])
	if e != nil {
		return CheckedZeroExPayload{}, e
	}
	if wrapped {
		data, e = zeroABI.Pack("exec", o.SettlerAddress, nativeToken, allowance, o.SettlerAddress, data)
	}
	return CheckedZeroExPayload{o.Target, data, o.SettlerAddress, deadline}, e
}
