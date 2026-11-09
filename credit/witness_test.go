package credit

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
	"strings"
	"testing"
)

func TestCreateWitnessAuthenticatesOwnerAndZeroUsage(t *testing.T) {
	s, f := fixture(t)
	private, err := crypto.ToECDSA(common.LeftPadBytes([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	owner := crypto.PubkeyToAddress(private.PublicKey)
	s.OwnerWalletIdentity = core.BigIntToElement(owner.Big())
	c, _ := s.Commitment()
	p := CreateParams{Token: common.BigToAddress(big.NewInt(34)), Owner: owner, RevocationShard: 1, CreditCommitment: common.HexToHash(hex(c)), Deadline: big.NewInt(2000000300)}
	ctx := Context{ProtocolContext: s.ProtocolContext, TokenAddress: p.Token, Timestamp: 2000000000}
	d := f["delivery"].(map[string]any)
	clue := make([]byte, 68)
	clue[64] = 24
	deliveries := []protocol.Delivery{{Ciphertext: bytesHex(t, d["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, d["spenderCiphertext"].(string)), Clue: clue}}
	_, typed, err := p.Statement(ctx, deliveries)
	if err != nil {
		t.Fatal(err)
	}
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := crypto.Sign(digest, private)
	if err != nil {
		t.Fatal(err)
	}
	auth := Authority{PublicKey: crypto.FromECDSAPub(&private.PublicKey), Signature: sig[:64]}
	witness, err := p.Witness(ctx, deliveries, s, auth)
	if err != nil {
		t.Fatal(err)
	}
	if err = witness.WithInputs(func(m map[string]any) error {
		if len(m) != 24 {
			t.Fatalf("witness fields %d", len(m))
		}
		if m["spender_wallet_identity"] != "33" {
			t.Fatal("private spender mismatch")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(witness)
	if !strings.Contains(string(raw), "redacted") {
		t.Fatal("witness leaked")
	}
	auth.Signature[0] ^= 1
	if _, err = p.Witness(ctx, deliveries, s, auth); err == nil {
		t.Fatal("accepted wrong owner signature")
	}
	s.TotalSpent = core.ElementFromUint64(1)
	if _, err = p.Witness(ctx, deliveries, s, auth); err == nil {
		t.Fatal("accepted new grant usage")
	}
}

func TestDrawAndUpdateWitnessBindMembershipAndOrdinaryNote(t *testing.T) {
	s, f := fixture(t)
	key1, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{1}, 32))
	key2, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{2}, 32))
	owner, spender := crypto.PubkeyToAddress(key1.PublicKey), crypto.PubkeyToAddress(key2.PublicKey)
	s.OwnerWalletIdentity = core.BigIntToElement(owner.Big())
	s.SpenderWalletIdentity = core.BigIntToElement(spender.Big())
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	located := LocatedState{State: s, TreeID: 1, Debit: core.ElementFromUint64(7)}
	root, err := located.root()
	if err != nil {
		t.Fatal(err)
	}
	source := protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: root}
	deadline := big.NewInt(2000000300)
	fdelivery := f["delivery"].(map[string]any)
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: bytesHex(t, fdelivery["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, fdelivery["spenderCiphertext"].(string)), Clue: clue}}
	transition, err := s.Draw(located.Debit, c.Timestamp)
	if err != nil {
		t.Fatal(err)
	}
	outputIdentity := core.ElementFromUint64(999)
	out, err := core.ComputeNoteCommitment(c.ChainID, outputIdentity, s.TokenAddress, located.Debit)
	if err != nil {
		t.Fatal(err)
	}
	draw := DrawParams{Token: c.TokenAddress, Sources: []protocol.InputSource{source}, Payers: []Payer{{Payer: owner, Amount: big.NewInt(7)}}, GrantRefs: []GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(transition.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(transition.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(transition.RecoveryAmount.Bytes())}, NoteCommitment: common.Hash(out.Bytes()), TotalAmount: big.NewInt(7), Deadline: deadline}
	dds := append(ds, protocol.Delivery{})
	_, typed, err := draw.Statement(c, dds)
	if err != nil {
		t.Fatal(err)
	}
	digest, _, _ := apitypes.TypedDataAndHash(typed)
	sig, err := crypto.Sign(digest, key2)
	if err != nil {
		t.Fatal(err)
	}
	auth := Authority{PublicKey: crypto.FromECDSAPub(&key2.PublicKey), Signature: sig[:64]}
	witness, err := draw.Witness(c, dds, []LocatedState{located}, outputIdentity, auth)
	if err != nil {
		t.Fatal(err)
	}
	if err = witness.WithInputs(func(m map[string]any) error {
		if m["output_identity"] != "999" {
			t.Fatal("ordinary Note identity missing")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	bad := located
	bad.Path[0] = common.HexToHash("0x03")
	if _, err = draw.Witness(c, dds, []LocatedState{bad}, outputIdentity, auth); err == nil {
		t.Fatal("accepted altered membership")
	}
	if _, err = draw.Witness(c, dds, []LocatedState{located}, core.ElementFromUint64(998), auth); err == nil {
		t.Fatal("accepted wrong ordinary Note")
	}
	policy := Policy{SingleLimit: core.ElementFromUint64(100)}
	updated, err := s.Update(Change{Policy: policy, Expiry: s.Expiry}, deadline.Uint64())
	if err != nil {
		t.Fatal(err)
	}
	p := UpdateParams{Token: c.TokenAddress, Owner: owner, Source: source, Nullifier: common.Hash(updated.Nullifier.Bytes()), CreditCommitment: common.Hash(updated.StateCommitment.Bytes()), RevocationShard: 1, Deadline: deadline}
	_, typed, err = p.Statement(c, ds)
	if err != nil {
		t.Fatal(err)
	}
	digest, _, _ = apitypes.TypedDataAndHash(typed)
	sig, err = crypto.Sign(digest, key1)
	if err != nil {
		t.Fatal(err)
	}
	auth = Authority{PublicKey: crypto.FromECDSAPub(&key1.PublicKey), Signature: sig[:64]}
	if _, err = p.Witness(c, ds, located, Change{Policy: policy, Expiry: s.Expiry}, auth); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Witness(c, ds, bad, Change{Policy: policy, Expiry: s.Expiry}, auth); err == nil {
		t.Fatal("accepted altered update membership")
	}
}
