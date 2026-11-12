package credit

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func TestPreparedCreateFreezesStateAndRequiresProof(t *testing.T) {
	s, f := fixture(t)
	private, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{1}, 32))
	owner := crypto.PubkeyToAddress(private.PublicKey)
	s.OwnerWalletIdentity = primitives.BigIntToElement(owner.Big())
	commitment, _ := s.Commitment()
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	p := CreateParams{Token: c.TokenAddress, Owner: owner, RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}
	d := f["delivery"].(map[string]any)
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: bytesHex(t, d["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, d["spenderCiphertext"].(string)), Clue: clue}}
	prepared, err := PrepareCreate(c, p, ds, s)
	if err != nil {
		t.Fatal(err)
	}
	p.Deadline.SetUint64(1)
	ds[0].Ciphertext[0] = 0
	s.CreditSharedKey = primitives.ElementFromUint64(999)
	typed, err := prepared.SigningRequest()
	if err != nil {
		t.Fatal(err)
	}
	digest, _, _ := apitypes.TypedDataAndHash(typed)
	sig, _ := crypto.Sign(digest, private)
	if err = prepared.WithWalletSignature(sig); err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.Calldata(owner); err == nil {
		t.Fatal("exported unproved create")
	}
}

func TestDrawToNoteFreezesOrdinaryOutputForFeeQuotation(t *testing.T) {
	s, _ := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	transition, _ := s.Draw(primitives.ElementFromUint64(7), c.Timestamp)
	p := DrawParams{Token: c.TokenAddress, Sources: []protocol.InputSource{{TreeId: big.NewInt(1), MerkleRoot: common.HexToHash("0x01")}}, Payers: []Payer{{Payer: common.BigToAddress(big.NewInt(44)), Amount: big.NewInt(7)}}, GrantRefs: []GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(transition.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(transition.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(transition.RecoveryAmount.Bytes())}, TotalAmount: big.NewInt(7), Deadline: big.NewInt(2000000300)}
	receiver, _ := primitives.DeriveWallet(common.LeftPadBytes([]byte{99}, 32))
	keys := &primitives.ReceivingKeys{ViewingPublicKey: receiver.ViewingPublicKey, ClueKey: receiver.ClueKey}
	note := primitives.Note{ProtocolContext: s.ProtocolContext, OwnerAddress: receiver.OwnerAddress, TokenAddress: s.TokenAddress, Amount: primitives.ElementFromUint64(7), Nonce: primitives.ElementFromUint64(100), Payload: []byte("invoice")}
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: []byte{0x82}, Clue: clue}, {Ciphertext: []byte{0x82}, Clue: clue}}
	prepared, err := PrepareDrawToNote(c, p, ds, []LocatedState{{State: s, TreeID: 1, Debit: primitives.ElementFromUint64(7)}}, note, keys)
	if err != nil {
		t.Fatal(err)
	}
	note.Payload[0] = 'X'
	intent, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if intent.OutputNote == nil || string(intent.OutputNote.Payload) != "invoice" || intent.OutputNote.Amount != primitives.ElementFromUint64(7) {
		t.Fatal("ordinary output was not frozen")
	}

	quoteKey, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	client := creditBroadcaster(t, c, quoteKey, common.HexToAddress("0x1234"))
	if err = prepared.WithBroadcaster(context.Background(), client, WithBroadcasterOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	quoted, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if quoted.OutputIdentity != intent.OutputIdentity || quoted.Draw.TotalAmount.Cmp(big.NewInt(7)) != 0 || quoted.Inputs[0].Debit != primitives.ElementFromUint64(7) || quoted.Draw.CreditCommitments[0] != intent.Draw.CreditCommitments[0] {
		t.Fatal("quote changed gross grant accounting or Note identity")
	}
	last := quoted.Deliveries[len(quoted.Deliveries)-1]
	decoded, err := primitives.DecryptPrivateTransferNote(receiver.ViewingPrivateKey, last.Ciphertext, primitives.NoteDecryptionContext{NoteEncryptionContext: primitives.NoteEncryptionContext{ProtocolContext: s.ProtocolContext, OutputCommitment: primitives.BigIntToElement(quoted.Draw.NoteCommitment.Big())}, TokenAddress: s.TokenAddress, OwnerAddress: receiver.OwnerAddress})
	if err != nil || decoded.Amount != primitives.ElementFromUint64(6) || decoded.Nonce != primitives.ElementFromUint64(100) || string(decoded.Payload) != "invoice" {
		t.Fatal("quoted Note did not preserve opening and deliver net amount", err)
	}
}
