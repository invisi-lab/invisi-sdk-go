package session

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestNoteOutputFactoryIdentityAndImmutability(t *testing.T) {
	amount := big.NewInt(5)
	out, err := sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: big.NewInt(31337), ProtocolAddress: common.HexToAddress("0x01"), TokenAddress: common.HexToAddress("0x02"), Amount: amount, Recipient: sdk.IdentityRecipient{Identity: common.HexToHash("0x03")}})
	if err != nil {
		t.Fatal(err)
	}
	amount.SetInt64(8)
	if out.Amount().Int64() != 5 {
		t.Fatal("factory retained mutable amount")
	}
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var copy sdk.NoteOutput
	if err = json.Unmarshal(raw, &copy); err != nil {
		t.Fatal(err)
	}
	if copy.Commitment() != out.Commitment() {
		t.Fatal("snapshot changed")
	}
	if err = json.Unmarshal([]byte(`{"Amount":5}`), &copy); err == nil {
		t.Fatal("fabricated output accepted")
	}
}

type testPaymentOutput struct {
	To      string
	Amount  *big.Int
	Payload []byte
}

func testOutputs(s *WalletSession, outputs []testPaymentOutput) []sdk.NoteOutput {
	result := make([]sdk.NoteOutput, 0, len(outputs))
	for _, o := range outputs {
		var r sdk.NoteRecipient = sdk.InvisiRecipient{Address: o.To}
		if common.IsHexAddress(o.To) {
			r = sdk.EVMRecipient{Address: common.HexToAddress(o.To)}
		}
		n, e := sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: core.ElementToBigInt(s.coreContext().ChainID), ProtocolAddress: s.scope.ProtocolAddress, TokenAddress: common.HexToAddress("0x2222222222222222222222222222222222222222"), Amount: o.Amount, Recipient: r, Payload: o.Payload})
		if e != nil {
			panic(e)
		}
		result = append(result, n)
	}
	return result
}

func TestNoteOutputCanonicalNonceAndOpaqueModes(t *testing.T) {
	base := sdk.PaymentRequest{ChainID: big.NewInt(31337), ProtocolAddress: common.HexToAddress("0x01"), TokenAddress: common.HexToAddress("0x02"), Amount: big.NewInt(5)}
	base.Recipient = sdk.EVMRecipient{Address: common.HexToAddress("0x03"), Nonce: big.NewInt(0)}
	out, err := sdk.NewNoteOutput(base)
	if err != nil {
		t.Fatal(err)
	}
	if !out.DeliveryPending() || !out.Note().Nonce.IsZero() {
		t.Fatal("offline EVM intent or explicit zero nonce lost")
	}
	for _, nonce := range []*big.Int{big.NewInt(-1), fr.Modulus()} {
		base.Recipient = sdk.EVMRecipient{Address: common.HexToAddress("0x03"), Nonce: nonce}
		if _, err = sdk.NewNoteOutput(base); err == nil {
			t.Fatal("noncanonical nonce accepted")
		}
	}
	base.Recipient = sdk.IdentityRecipient{Identity: common.HexToHash("0x03")}
	base.Payload = []byte{}
	if _, err = sdk.NewNoteOutput(base); err == nil {
		t.Fatal("opaque payload accepted")
	}
	base.Payload = nil
	out, err = sdk.NewNoteOutput(base)
	if err != nil {
		t.Fatal(err)
	}
	if out.Note() != nil || out.DeliveryPending() || len(out.Delivery().Ciphertext) > 0 {
		t.Fatal("opaque identity fabricated an opening or envelope")
	}
	if out.AssertCommitment(common.Hash{}) == nil {
		t.Fatal("commitment assertion accepted mismatch")
	}
}
func TestNoteOutputPreparationUsesDirectIdentityAndChecksContext(t *testing.T) {
	session, token, _ := testSession(t)
	base := sdk.PaymentRequest{ChainID: core.ElementToBigInt(session.coreContext().ChainID), ProtocolAddress: session.scope.ProtocolAddress, TokenAddress: token, Amount: big.NewInt(30), Recipient: sdk.IdentityRecipient{Identity: common.HexToHash("0x03")}}
	out, err := sdk.NewNoteOutput(base)
	if err != nil {
		t.Fatal(err)
	}
	op, err := session.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{out}})
	if err != nil {
		t.Fatal(err)
	}
	if op.plan.Outputs[0].Note != nil || op.Preview().Outputs[0].Amount.Int64() != 30 {
		t.Fatal("opaque output opening or amount mismatch")
	}
	base.TokenAddress = common.HexToAddress("0x09")
	other, err := sdk.NewNoteOutput(base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.runtime.Output(context.Background(), token, other); err == nil {
		t.Fatal("mismatched token accepted")
	}
}

func TestFinalOutputsExportPaymentNotesIncludingChange(t *testing.T) {
	s, token, _ := testSession(t)
	address, _ := s.wallet.GetAddress()
	op, e := s.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	outputs := op.GetOutputs()
	if len(outputs) != 2 {
		t.Fatalf("outputs: %d", len(outputs))
	}
	for _, output := range outputs {
		ipn, e := output.ExportPaymentNote()
		if e != nil {
			t.Fatal(e)
		}
		note, e := sdk.PaymentNoteFromIPN(ipn)
		if e != nil {
			t.Fatal(e)
		}
		if note.Commitment() != output.Commitment() || note.Amount().Cmp(output.Amount()) != 0 {
			t.Fatal("final output export differs")
		}
	}
	if outputs[1].Amount().Cmp(big.NewInt(70)) != 0 || !outputs[1].Change() {
		t.Fatal("missing change")
	}
}
