package paymentrouter

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func addr(n byte) common.Address {
	var a common.Address
	for i := range a {
		a[i] = n * 17
	}
	return a
}
func bi(n int64) *big.Int { return big.NewInt(n) }
func TestV3ReversedPaths(t *testing.T) {
	p, e := EncodeV3ExactOutputPath([]common.Address{addr(4), addr(5), addr(2)}, []uint32{500, 100})
	if e != nil {
		t.Fatal(e)
	}
	expected := append(addr(2).Bytes(), 0, 0, 100)
	expected = append(expected, addr(5).Bytes()...)
	expected = append(expected, 0, 1, 244)
	expected = append(expected, addr(4).Bytes()...)
	if !bytes.Equal(p, expected) {
		t.Fatalf("wrong reverse path %x", p)
	}
	if e = ValidateV3Path(p, addr(4), addr(2)); e != nil {
		t.Fatal(e)
	}
	for _, v := range []struct {
		tokens []common.Address
		fees   []uint32
	}{{[]common.Address{addr(4)}, nil}, {[]common.Address{addr(4), addr(4)}, []uint32{500}}, {[]common.Address{addr(4), {}}, []uint32{500}}, {[]common.Address{addr(4), addr(2)}, []uint32{0}}, {[]common.Address{addr(4), addr(2)}, []uint32{1000000}}} {
		if _, e := EncodeV3ExactOutputPath(v.tokens, v.fees); e == nil {
			t.Fatal("accepted invalid path")
		}
	}
	if ValidateV3Path(p, addr(2), addr(4)) == nil {
		t.Fatal("accepted wrong endpoints")
	}
}
func nativeData(t *testing.T, recipient, token common.Address, amount, budget, deadline int64, guards int) []byte {
	t.Helper()
	g, e := zeroABI.Pack("NATIVE_CHECK", bi(deadline), bi(budget))
	if e != nil {
		t.Fatal(e)
	}
	actions := [][]byte{}
	for i := 0; i < guards; i++ {
		actions = append(actions, g)
	}
	b, e := zeroABI.Pack("execute", slippage{recipient, token, bi(amount)}, actions, [32]byte{})
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func payloadOptions(data []byte) ZeroExPayloadOptions {
	return ZeroExPayloadOptions{Target: addr(6), Data: data, RouterAddress: addr(3), TokenOut: addr(2), AmountOut: bi(10), MaxInput: bi(100), SettlerAddress: addr(6), AllowanceHolderAddress: addr(7), Now: bi(1000)}
}
func TestNativeGuards(t *testing.T) {
	data := nativeData(t, addr(3), addr(2), 10, 100, 1100, 1)
	v, e := ValidateZeroExPayload(payloadOptions(data))
	if e != nil {
		t.Fatal(e)
	}
	if v.NativeDeadline.Cmp(bi(1100)) != 0 || !bytes.Equal(data, v.Data) {
		t.Fatal("changed direct payload")
	}
	wrapped, e := zeroABI.Pack("exec", addr(6), nativeToken, bi(0), addr(6), data)
	if e != nil {
		t.Fatal(e)
	}
	o := payloadOptions(wrapped)
	o.Target = addr(7)
	if _, e = ValidateZeroExPayload(o); e != nil {
		t.Fatal(e)
	}
	for _, data := range [][]byte{nativeData(t, addr(4), addr(2), 10, 100, 1100, 1), nativeData(t, addr(3), addr(4), 10, 100, 1100, 1), nativeData(t, addr(3), addr(2), 11, 100, 1100, 1), nativeData(t, addr(3), addr(2), 9, 100, 1100, 1), nativeData(t, addr(3), addr(2), 10, 99, 1100, 1), nativeData(t, addr(3), addr(2), 10, 100, 999, 1), nativeData(t, addr(3), addr(2), 10, 100, 1100, 0), nativeData(t, addr(3), addr(2), 10, 100, 1100, 2), {1, 2, 3}} {
		if _, e := ValidateZeroExPayload(payloadOptions(data)); e == nil {
			t.Fatal("accepted hostile payload")
		}
	}
	for _, change := range []string{"operator", "token", "target", "allowance"} {
		operator, token, target, amount := addr(6), nativeToken, addr(6), bi(0)
		switch change {
		case "operator":
			operator = addr(4)
		case "token":
			token = addr(4)
		case "target":
			target = addr(4)
		case "allowance":
			amount = bi(101)
		}
		w, _ := zeroABI.Pack("exec", operator, token, amount, target, data)
		o := payloadOptions(w)
		o.Target = addr(7)
		if _, e := ValidateZeroExPayload(o); e == nil {
			t.Fatal("accepted invalid wrapper", change)
		}
	}
}
