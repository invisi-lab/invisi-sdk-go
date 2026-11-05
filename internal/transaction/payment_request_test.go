package transaction

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"
)

func TestPaymentRequestVectors(t *testing.T) {
	raw, e := os.ReadFile("testdata/payment_requests.json")
	if e != nil {
		t.Fatal(e)
	}
	var file struct {
		Vectors []struct {
			Name      string `json:"name"`
			Encoded   string `json:"encoded"`
			BinaryHex string `json:"binaryHex"`
		}
	}
	if e = json.Unmarshal(raw, &file); e != nil {
		t.Fatal(e)
	}
	for _, v := range file.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			request, e := DecodePaymentRequest(v.Encoded)
			if e != nil {
				t.Fatal(e)
			}
			encoded, e := EncodePaymentRequest(request)
			if e != nil || encoded != v.Encoded {
				t.Fatalf("roundtrip: %s %v", encoded, e)
			}
			output, e := NewNoteOutput(v.Encoded)
			if e != nil {
				t.Fatal(e)
			}
			if output.IPP() != v.Encoded {
				t.Fatal("request was rewritten")
			}
			if _, opaque := request.Recipient.(IdentityRecipient); !opaque {
				nonce := big.NewInt(9)
				var fixed *big.Int
				switch r := request.Recipient.(type) {
				case EVMRecipient:
					fixed = r.Nonce
				case InvisiRecipient:
					fixed = r.Nonce
				}
				if fixed != nil {
					nonce = fixed
				}
				selected, e := NewNoteOutput(v.Encoded, NoteOutputExecutionOptions{Nonce: nonce})
				if e != nil {
					t.Fatal(e)
				}
				if selected.Nonce().Cmp(nonce) != 0 || selected.IPP() != v.Encoded {
					t.Fatal("nonce execution changed request")
				}
			}
		})
	}
}

func TestPaymentRequestExecutionAndBounds(t *testing.T) {
	var file struct {
		Vectors []struct {
			Encoded string `json:"encoded"`
		}
	}
	raw, _ := os.ReadFile("testdata/payment_requests.json")
	if e := json.Unmarshal(raw, &file); e != nil {
		t.Fatal(e)
	}
	for _, v := range file.Vectors {
		for _, bad := range []string{v.Encoded + "=", v.Encoded + "A", v.Encoded[:len(v.Encoded)-1], "ipn:" + v.Encoded[4:]} {
			if _, e := DecodePaymentRequest(bad); e == nil {
				t.Fatal("accepted malformed request")
			}
		}
	}
	request, e := DecodePaymentRequest(file.Vectors[0].Encoded)
	if e != nil {
		t.Fatal(e)
	}
	first, e := NewNoteOutput(request)
	if e != nil {
		t.Fatal(e)
	}
	second, e := NewNoteOutput(request)
	if e != nil {
		t.Fatal(e)
	}
	if first.Nonce().Cmp(second.Nonce()) == 0 {
		t.Fatal("reused nonce")
	}
	if _, e = NewNoteOutput(first.IPP(), NoteOutputExecutionOptions{AddressLookup: "chain"}); e == nil {
		t.Fatal("lookup accepted without envelope")
	}
	r := request.Recipient.(EVMRecipient)
	r.Nonce = big.NewInt(0)
	request.Recipient = r
	fixed, e := EncodePaymentRequest(request)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = NewNoteOutput(fixed, NoteOutputExecutionOptions{Nonce: big.NewInt(1)}); e == nil {
		t.Fatal("fixed nonce overwritten")
	}
	r.Envelope = nil
	r.Nonce = nil
	r.AddressLookup = "chain"
	request.Recipient = r
	output, e := NewNoteOutput(request)
	if e != nil {
		t.Fatal(e)
	}
	if output.saved.Lookup != "chain" {
		t.Fatal("local lookup lost")
	}
	decoded, e := DecodePaymentRequest(output.IPP())
	if e != nil {
		t.Fatal(e)
	}
	if decoded.Recipient.(EVMRecipient).AddressLookup != "" {
		t.Fatal("lookup encoded")
	}
	request.Payload = make([]byte, 3936)
	encoded, e := EncodePaymentRequest(request)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = DecodePaymentRequest(encoded); e != nil {
		t.Fatal(e)
	}
	request.Payload = make([]byte, 3937)
	if _, e = EncodePaymentRequest(request); e == nil {
		t.Fatal("oversize payload accepted")
	}
	output.saved.IPP = file.Vectors[0].Encoded
	if e = matchPaymentRequest(output.saved); e == nil {
		t.Fatal("mutated request accepted")
	}
}

func TestPaymentRequestMaximumAndSnapshot(t *testing.T) {
	raw, _ := os.ReadFile("testdata/payment_requests.json")
	var file struct {
		Vectors []struct {
			Encoded string `json:"encoded"`
		}
	}
	if e := json.Unmarshal(raw, &file); e != nil {
		t.Fatal(e)
	}
	for _, v := range file.Vectors {
		request, e := DecodePaymentRequest(v.Encoded)
		if e != nil {
			t.Fatal(e)
		}
		if r, ok := request.Recipient.(InvisiRecipient); ok {
			r.Nonce = big.NewInt(0)
			request.Recipient = r
			request.Payload = make([]byte, 3936)
			encoded, e := EncodePaymentRequest(request)
			if e != nil {
				t.Fatal(e)
			}
			if len(encoded) != 5500 {
				t.Fatal("maximum length mismatch", len(encoded))
			}
			if _, e = DecodePaymentRequest(encoded); e != nil {
				t.Fatal(e)
			}
			output, e := NewNoteOutput(v.Encoded)
			if e != nil {
				t.Fatal(e)
			}
			saved, e := json.Marshal(output)
			if e != nil {
				t.Fatal(e)
			}
			var restored NoteOutput
			if e = json.Unmarshal(saved, &restored); e != nil {
				t.Fatal(e)
			}
			if restored.IPP() != v.Encoded || restored.Commitment() != output.Commitment() {
				t.Fatal("snapshot changed output or request")
			}
			return
		}
	}
	t.Fatal("no Invisi fixture")
}
