package credit

import (
	"encoding/json"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"math/big"
	"os"
	"reflect"
	"testing"
)

func TestInitialPaymentStatementGolden(t *testing.T) {
	raw, err := os.ReadFile("testdata/initial-payment-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Deliveries []struct {
			Ciphertext hexutil.Bytes
			Clue       hexutil.Bytes
		}
		PublicInputs []common.Hash
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	ds := make([]protocol.Delivery, len(f.Deliveries))
	for i, d := range f.Deliveries {
		ds[i] = protocol.Delivery{Ciphertext: d.Ciphertext, Clue: d.Clue}
	}
	yes := true
	c := Context{ProtocolContext: primitives.ProtocolContext{ChainID: primitives.ElementFromUint64(1), ProtocolAddress: primitives.BigIntToElement(common.HexToAddress("0x1111111111111111111111111111111111111111").Big())}, TokenAddress: common.HexToAddress("0x2222222222222222222222222222222222222222"), Timestamp: 2000000000}
	p := CreateParams{Token: c.TokenAddress, Owner: common.HexToAddress("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf"), RevocationShard: 1, CreditCommitment: common.HexToHash("0x1d8f4fe1b37a592a185db11788b2e9b68f77106c6a09a54b079e39e119bd48e9"), Deadline: big.NewInt(2000000100), FeeRecipient: common.HexToAddress("0x4444444444444444444444444444444444444444"), FeeAmount: big.NewInt(1), HasInitialPayment: &yes, TotalAmount: big.NewInt(7), NoteCommitment: common.HexToHash("0x1e8da2ed5930db94d005d57f5e832ee14da4a5dedf06ec72a3805840460098d6")}
	inputs, _, err := p.Statement(c, ds)
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 18 {
		t.Fatalf("width %d", len(inputs))
	}
	for i := range inputs {
		if inputs[i] != f.PublicInputs[i] {
			t.Fatalf("public input %d mismatch", i)
		}
	}
	if _, err = p.Calldata(c, []byte{1}, ds); err != nil {
		t.Fatal(err)
	}
	p.TotalAmount = big.NewInt(1)
	if _, _, err = p.Statement(c, ds); err == nil {
		t.Fatal("accepted fee consuming payment")
	}
}

func TestInitialPaymentWitnessGolden(t *testing.T) {
	raw, err := os.ReadFile("testdata/initial-payment-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Deliveries []struct {
			Ciphertext hexutil.Bytes
			Clue       hexutil.Bytes
		}
		Witness      map[string]json.RawMessage
		PublicInputs []common.Hash
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	field := func(name string) fr.Element {
		var text string
		if e := json.Unmarshal(f.Witness[name], &text); e != nil {
			t.Fatal(e)
		}
		n, ok := new(big.Int).SetString(text, 10)
		if !ok {
			t.Fatal(name)
		}
		return primitives.BigIntToElement(n)
	}
	c := Context{ProtocolContext: primitives.ProtocolContext{ChainID: field("chain_id"), ProtocolAddress: field("protocol_address")}, TokenAddress: common.BigToAddress(primitives.ElementToBigInt(field("token_address"))), Timestamp: 2000000000}
	s := State{ProtocolContext: c.ProtocolContext, TokenAddress: field("token_address"), OwnerWalletIdentity: field("owner_address"), SpenderWalletIdentity: field("spender_wallet_identity"), RevocationShard: field("revocation_shard"), RevocationSlot: field("revocation_slot"), Nonce: field("credit_nonce"), Expiry: field("expiry"), CreditSharedKey: field("credit_shared_key"), AvailableCredit: field("available_credit")}
	var limits []string
	if e := json.Unmarshal(f.Witness["limits"], &limits); e != nil {
		t.Fatal(e)
	}
	for i, p := range []*fr.Element{&s.SingleLimit, &s.DailyLimit, &s.MonthlyLimit} {
		n, _ := new(big.Int).SetString(limits[i], 10)
		*p = primitives.BigIntToElement(n)
	}
	initial := primitives.ElementToBigInt(s.AvailableCredit)
	debit := field("total_amount")
	s.Usage, err = s.drawUsage(debit, c.Timestamp)
	if err != nil {
		t.Fatal(err)
	}
	s.AvailableCredit.Sub(&s.AvailableCredit, &debit)
	yes := true
	p := CreateParams{Token: c.TokenAddress, Owner: common.BigToAddress(primitives.ElementToBigInt(s.OwnerWalletIdentity)), RevocationShard: 1, CreditCommitment: f.PublicInputs[6], Deadline: big.NewInt(2000000100), FeeRecipient: common.BigToAddress(f.PublicInputs[10].Big()), FeeAmount: big.NewInt(1), HasInitialPayment: &yes, TotalAmount: big.NewInt(7), NoteCommitment: f.PublicInputs[12]}
	output := primitives.Note{ProtocolContext: c.ProtocolContext, TokenAddress: s.TokenAddress, OwnerAddress: s.SpenderWalletIdentity, Nonce: field("output_nonce"), Amount: primitives.ElementFromUint64(6)}
	auth := Authority{PublicKey: []byte{4}}
	for _, name := range []string{"public_key_x", "public_key_y", "signature"} {
		var b []int
		if e := json.Unmarshal(f.Witness[name], &b); e != nil {
			t.Fatal(e)
		}
		for _, v := range b {
			if name == "signature" {
				auth.Signature = append(auth.Signature, byte(v))
			} else {
				auth.PublicKey = append(auth.PublicKey, byte(v))
			}
		}
	}
	ds := make([]protocol.Delivery, len(f.Deliveries))
	for i, d := range f.Deliveries {
		ds[i] = protocol.Delivery{Ciphertext: d.Ciphertext, Clue: d.Clue}
	}
	w, err := p.createWitness(c, ds, s, auth, initial, &output)
	if err != nil {
		t.Fatal(err)
	}
	err = w.WithInputs(func(values map[string]any) error {
		for k, want := range f.Witness {
			got, e := json.Marshal(values[k])
			if e != nil {
				t.Fatal(e)
			}
			var a, b any
			json.Unmarshal(want, &a)
			json.Unmarshal(got, &b)
			if !reflect.DeepEqual(a, b) {
				t.Errorf("witness %s mismatch: %s != %s", k, got, want)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	output.OwnerAddress = s.OwnerWalletIdentity
	if _, err = p.createWitness(c, ds, s, auth, initial, &output); err == nil {
		t.Fatal("accepted nonspender Note")
	}
}
