package transaction

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
	"os"
	"strings"
	"testing"
)

func TestPaymentNoteVectors(t *testing.T) {
	raw, _ := os.ReadFile("testdata/payment_notes.json")
	var vectors []struct {
		Name    string `json:"name"`
		Encoded string `json:"encoded"`
	}
	if e := json.Unmarshal(raw, &vectors); e != nil {
		t.Fatal(e)
	}
	for _, v := range vectors {
		t.Run(v.Name, func(t *testing.T) {
			data, e := DecodePaymentNote(v.Encoded)
			if e != nil {
				t.Fatal(e)
			}
			encoded, e := EncodePaymentNote(data)
			if e != nil || encoded != v.Encoded || len(encoded) != 203 {
				t.Fatalf("wire: %v", e)
			}
			note, e := PaymentNoteFromIPN(encoded)
			if e != nil {
				t.Fatal(e)
			}
			if note.Commitment() != data.Commitment {
				t.Fatal("wrong commitment")
			}
			data.Amount = new(big.Int).Sub(data.Amount, big.NewInt(1))
			if data.Amount.Sign() > 0 {
				bad, _ := EncodePaymentNote(data)
				if _, e = PaymentNoteFromIPN(bad); e == nil {
					t.Fatal("wrong amount accepted")
				}
			}
			for _, bad := range []string{encoded + "=", encoded + "A", encoded[:len(encoded)-1], "ipr:" + encoded[4:]} {
				if _, e = DecodePaymentNote(bad); e == nil {
					t.Fatal("invalid wire accepted")
				}
			}
		})
	}
}
func TestPaymentNoteExport(t *testing.T) {
	w, e := WalletFromPrivateKey("0x" + strings.Repeat("01", 32))
	if e != nil {
		t.Fatal(e)
	}
	address, _ := w.GetAddress()
	output, e := NewNoteOutput(PaymentRequest{ChainID: big.NewInt(97), ProtocolAddress: common.HexToAddress("1"), TokenAddress: common.HexToAddress("2"), Amount: big.NewInt(70), Recipient: InvisiRecipient{Address: address}})
	if e != nil {
		t.Fatal(e)
	}
	ipn, e := output.ExportPaymentNote()
	if e != nil {
		t.Fatal(e)
	}
	note, e := PaymentNoteFromIPN(ipn)
	if e != nil {
		t.Fatal(e)
	}
	if note.OwnerAddress() != w.OwnerAddress() || note.Amount().Cmp(big.NewInt(70)) != 0 {
		t.Fatal("wrong opening")
	}
	opaque, e := NewNoteOutput(PaymentRequest{ChainID: big.NewInt(97), ProtocolAddress: common.HexToAddress("1"), TokenAddress: common.HexToAddress("2"), Amount: big.NewInt(70), Recipient: IdentityRecipient{Identity: output.Identity()}})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = opaque.ExportPaymentNote(); e == nil {
		t.Fatal("identity exported")
	}
}

func TestPaymentNoteCanonicalBoundsAndDetachedOutputs(t *testing.T) {
	raw, _ := os.ReadFile("testdata/payment_notes.json")
	var vectors []struct {
		Encoded string `json:"encoded"`
	}
	if e := json.Unmarshal(raw, &vectors); e != nil {
		t.Fatal(e)
	}
	encoded := vectors[0].Encoded
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	i := strings.IndexByte(alphabet, encoded[len(encoded)-1])
	bad := encoded[:len(encoded)-1] + string(alphabet[i|1])
	if _, e := DecodePaymentNote(bad); e == nil {
		t.Fatal("nonzero unused Base64URL bits accepted")
	}
	data, e := DecodePaymentNote(encoded)
	if e != nil {
		t.Fatal(e)
	}
	for _, d := range []PaymentNoteData{{ChainID: big.NewInt(0)}, func() PaymentNoteData { v := data; v.Nonce = new(big.Int).Lsh(big.NewInt(1), 255); return v }(), func() PaymentNoteData { v := data; v.Amount = big.NewInt(0); return v }()} {
		if _, e := EncodePaymentNote(d); e == nil {
			t.Fatal("invalid fields accepted")
		}
	}
	note := paymentNoteOpening(data)
	plan := Plan{Context: protocol.TransactionContext{ChainID: data.ChainID, ProtocolAddress: data.ProtocolAddress, TokenAddress: data.TokenAddress}, Outputs: []PreparedOutput{{Note: &note, Amount: data.Amount, Commitment: data.Commitment, Change: true}}}
	prepared := &PreparedTransaction{operation: &Execution{Plan: plan}}
	outputs := prepared.GetOutputs()
	if len(outputs) != 1 || !outputs[0].Change() {
		t.Fatal("missing final output")
	}
	ipn, e := outputs[0].ExportPaymentNote()
	if e != nil || ipn != encoded {
		t.Fatal("export differs", e)
	}
	prepared.operation.Plan.Outputs[0].Amount.SetInt64(1)
	if outputs[0].Amount().Int64() != 70 {
		t.Fatal("final snapshot was mutable")
	}
}
