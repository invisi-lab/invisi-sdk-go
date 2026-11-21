package paymentrouter

import (
	"bytes"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"testing"
)

func TestCanonicalWire(t *testing.T) {
	args, err := coreMethod.Inputs.Unpack(coreCall(t).Data[4:])
	if err != nil {
		t.Fatal(err)
	}
	p := DecodedPayment{Method: "payNative", Proof: args[0].([]byte), Request: *abi.ConvertType(args[1], new(protocol.TransactRequest)).(*protocol.TransactRequest), Intent: Intent{AmountInMaximum: bi(2), Nonce: bi(0), Deadline: bi(3)}}
	for _, method := range []string{"payNative", "payToken", "payTokenFor"} {
		p.Method = method
		p.Permit.Deadline = bi(0)
		wire, e := EncodeRouterCall(p)
		if e != nil {
			t.Fatal(e)
		}
		q, e := DecodeRouterCall(wire)
		if e != nil || q.Method != method {
			t.Fatal(q, e)
		}
		again, e := EncodeRouterCall(q)
		if e != nil || !bytes.Equal(again, wire) {
			t.Fatal(e)
		}
		if _, e = DecodeRouterCall(append(wire, 0)); e == nil {
			t.Fatal("accepted suffix")
		}
	}
}
