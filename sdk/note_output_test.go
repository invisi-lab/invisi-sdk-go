package sdk

import (
	"context"
	"encoding/json"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"math/big"
	"testing"
)

func TestNoteOutputFactoryIdentityAndImmutability(t *testing.T) {
	amount := big.NewInt(5)
	out, err := NewNoteOutput(NoteOutputOptions{ChainID: big.NewInt(31337), ProtocolAddress: common.HexToAddress("0x01"), TokenAddress: common.HexToAddress("0x02"), Amount: amount, Recipient: IdentityRecipient{Identity: common.HexToHash("0x03")}})
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
	var copy NoteOutput
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

func testOutputs(s *WalletSession, outputs []testPaymentOutput) []NoteOutput {
	result := make([]NoteOutput, 0, len(outputs))
	for _, o := range outputs {
		var r NoteRecipient = InvisiRecipient{Address: o.To}
		if common.IsHexAddress(o.To) {
			r = EVMRecipient{Address: common.HexToAddress(o.To)}
		}
		n, e := NewNoteOutput(NoteOutputOptions{ChainID: core.ElementToBigInt(s.coreContext().ChainID), ProtocolAddress: s.scope.ProtocolAddress, TokenAddress: common.HexToAddress("0x2222222222222222222222222222222222222222"), Amount: o.Amount, Recipient: r, Payload: o.Payload})
		if e != nil {
			panic(e)
		}
		result = append(result, n)
	}
	return result
}

func TestNoteOutputCanonicalNonceAndOpaqueModes(t *testing.T) {
	base := NoteOutputOptions{ChainID: big.NewInt(31337), ProtocolAddress: common.HexToAddress("0x01"), TokenAddress: common.HexToAddress("0x02"), Amount: big.NewInt(5)}
	base.Recipient = EVMRecipient{Address: common.HexToAddress("0x03"), Nonce: big.NewInt(0)}
	out, err := NewNoteOutput(base)
	if err != nil {
		t.Fatal(err)
	}
	if !out.DeliveryPending() || !out.Note().Nonce.IsZero() {
		t.Fatal("offline EVM intent or explicit zero nonce lost")
	}
	for _, nonce := range []*big.Int{big.NewInt(-1), fr.Modulus()} {
		base.Recipient = EVMRecipient{Address: common.HexToAddress("0x03"), Nonce: nonce}
		if _, err = NewNoteOutput(base); err == nil {
			t.Fatal("noncanonical nonce accepted")
		}
	}
	base.Recipient = IdentityRecipient{Identity: common.HexToHash("0x03")}
	base.Payload = []byte{}
	if _, err = NewNoteOutput(base); err == nil {
		t.Fatal("opaque payload accepted")
	}
	base.Payload = nil
	out, err = NewNoteOutput(base)
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
	base := NoteOutputOptions{ChainID: core.ElementToBigInt(session.coreContext().ChainID), ProtocolAddress: session.scope.ProtocolAddress, TokenAddress: token, Amount: big.NewInt(30), Recipient: IdentityRecipient{Identity: common.HexToHash("0x03")}}
	out, err := NewNoteOutput(base)
	if err != nil {
		t.Fatal(err)
	}
	op, err := session.PrepareTransfer(context.Background(), TransferPreparation{TokenAddress: token, Outputs: []NoteOutput{out}})
	if err != nil {
		t.Fatal(err)
	}
	if op.plan.Outputs[0].Note != nil || op.Preview().Outputs[0].Amount.Int64() != 30 {
		t.Fatal("opaque output opening or amount mismatch")
	}
	base.TokenAddress = common.HexToAddress("0x09")
	other, err := NewNoteOutput(base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.output(context.Background(), token, other); err == nil {
		t.Fatal("mismatched token accepted")
	}
}
