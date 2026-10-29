package core

import (
	"bytes"
	"encoding/json"
	"testing"
)

func memoNote(t *testing.T) (Note, []byte, []byte, NoteDecryptionContext) {
	t.Helper()
	private := make([]byte, 32)
	private[31] = 1
	public, err := ViewingPublicKey(private)
	if err != nil {
		t.Fatal(err)
	}
	note := Note{ProtocolContext: ProtocolContext{ElementFromUint64(1), ElementFromUint64(2)}, WalletIdentity: ElementFromUint64(3), TokenAddress: ElementFromUint64(4), Amount: ElementFromUint64(5), Nonce: ElementFromUint64(6)}
	identity, _ := ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
	leaf, _ := ComputeNoteCommitment(note.ChainID, identity, note.TokenAddress, note.Amount)
	return note, private, public, NoteDecryptionContext{NoteEncryptionContext{note.ProtocolContext, leaf}, note.TokenAddress, note.WalletIdentity}
}
func TestEmptyMemoUsesCurrentEnvelope(t *testing.T) {
	note, private, public, context := memoNote(t)
	ciphertext, err := EncryptPrivateTransferNote(public, note, context.NoteEncryptionContext)
	if err != nil {
		t.Fatal(err)
	}
	if len(ciphertext) != 160 || ciphertext[0] != 2 {
		t.Fatalf("expected version 2 and 160 bytes, got version %d and %d bytes", ciphertext[0], len(ciphertext))
	}
	if _, err = DecryptPrivateTransferNote(private, ciphertext, context); err != nil {
		t.Fatal(err)
	}
}
func TestMemoRoundTripAndPayloadKey(t *testing.T) {
	note, private, public, context := memoNote(t)
	for _, memo := range [][]byte{nil, []byte("你好\x00"), {0xff, 0xfe, 0, 1}, bytes.Repeat([]byte{0xab}, 256)} {
		note.Memo = memo
		ciphertext, err := EncryptPrivateTransferNote(public, note, context.NoteEncryptionContext)
		if err != nil {
			t.Fatal(err)
		}
		if len(ciphertext) != 160+len(memo) {
			t.Fatal("unexpected envelope size")
		}
		decoded, err := DecryptPrivateTransferNote(private, ciphertext, context)
		if err != nil || !bytes.Equal(decoded.Memo, memo) {
			t.Fatal("memo roundtrip", err)
		}
		key, err := ExportPayloadKey(private, ciphertext, false)
		if err != nil {
			t.Fatal(err)
		}
		opened, err := DecryptNotePayload(key, ciphertext, context)
		if err != nil || !bytes.Equal(opened.Memo, memo) {
			t.Fatal("payload memo", err)
		}
		encoded, err := json.Marshal(note)
		if err != nil {
			t.Fatal(err)
		}
		var restored Note
		if err = json.Unmarshal(encoded, &restored); err != nil || !bytes.Equal(restored.Memo, memo) {
			t.Fatal("memo storage", err)
		}
		if len(decoded.Memo) > 0 {
			decoded.Memo[0] ^= 1
			if !bytes.Equal(note.Memo, memo) {
				t.Fatal("memo alias")
			}
		}
	}
	note.Memo = make([]byte, 257)
	if _, err := EncryptPrivateTransferNote(public, note, context.NoteEncryptionContext); err == nil {
		t.Fatal("oversize memo accepted")
	}
}
func TestMemoTextValidation(t *testing.T) {
	text := "你好"
	memo, err := MemoFromText(text)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := MemoText(memo); err != nil || got != text {
		t.Fatal(got, err)
	}
	if _, err := MemoFromText(string([]byte{0xff})); err == nil {
		t.Fatal("invalid UTF8 text accepted")
	}
	if _, err := MemoText([]byte{0xff}); err == nil {
		t.Fatal("binary memo decoded as text")
	}
	if _, err := MemoFromText(string(bytes.Repeat([]byte{0x61}, 257))); err == nil {
		t.Fatal("257 byte text accepted")
	}
}
func TestMalformedMemoLengthsAndRetiredVersions(t *testing.T) {
	note, private, public, context := memoNote(t)
	for _, tail := range [][]byte{{}, {0}, {1, 1}, {0, 0, 1}, append([]byte{1, 1}, make([]byte, 257)...)} {
		plain := append(encodeNote(note)[:80], tail...)
		ciphertext, err := sealEnvelope(public, plain, context.NoteEncryptionContext, note.TokenAddress, 2, 0, bytes.NewReader(bytes.Repeat([]byte{1}, 100)))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = DecryptPrivateTransferNote(private, ciphertext, context); err == nil {
			t.Fatal("malformed memo accepted")
		}
	}
	ciphertext, err := EncryptPrivateTransferNote(public, note, context.NoteEncryptionContext)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext[0] = 1
	if _, err = DecryptPrivateTransferNote(private, ciphertext, context); err == nil {
		t.Fatal("retired version accepted")
	}
}

func TestCreditMemoRoundTripAndRecoveryOmitsUnknownMemo(t *testing.T) {
	note, private, public, noteContext := memoNote(t)
	wallet, err := DeriveWallet(private)
	if err != nil {
		t.Fatal(err)
	}
	keys := ReceivingKeys{ViewingPublicKey: public, ClueKey: wallet.ClueKey}
	credit := CreditNote{ProtocolContext: note.ProtocolContext, SpenderWalletIdentity: note.WalletIdentity, OwnerWalletIdentity: note.WalletIdentity, SpenderPayout: note.WalletIdentity, OwnerPayout: note.WalletIdentity, TokenAddress: note.TokenAddress, Amount: ElementFromUint64(100), Nonce: note.Nonce, Expiry: ElementFromUint64(1000), CreditSharedKey: ElementFromUint64(7), Memo: bytes.Repeat([]byte{0xff}, 256)}
	keyHash, _ := ComputeCreditSharedKeyHash(credit.CreditSharedKey)
	leaf, _ := ComputeCreditCommitment(credit.ChainID, credit.SpenderWalletIdentity, credit.OwnerWalletIdentity, credit.SpenderPayout, credit.OwnerPayout, credit.TokenAddress, credit.Amount, credit.Nonce, credit.Expiry, keyHash)
	state := CreditDeliveryState{Note: credit, OwnerReceivingKeys: keys, SpenderReceivingKeys: keys}
	for _, role := range []CreditDeliveryRole{CreditOwner, CreditSpender} {
		context := CreditNoteDecryptionContext{NoteDecryptionContext: noteContext, Role: role}
		context.OutputCommitment = leaf
		ciphertext, err := EncryptCreditNote(public, state, CreditNoteEncryptionContext{NoteEncryptionContext: context.NoteEncryptionContext, Role: role})
		if err != nil {
			t.Fatal(err)
		}
		if len(ciphertext) != 658 || ciphertext[0] != 4 {
			t.Fatal("Credit envelope width/version")
		}
		decoded, err := DecryptCreditNote(private, ciphertext, context)
		if err != nil || !bytes.Equal(decoded.Note.Memo, credit.Memo) {
			t.Fatal("Credit memo", err)
		}
		decoded.Note.Memo[0] = 0
		if credit.Memo[0] != 0xff {
			t.Fatal("Credit memo aliased")
		}
		key, err := ExportPayloadKey(private, ciphertext, true)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err = DecryptCreditPayload(key, ciphertext, context)
		if err != nil || !bytes.Equal(decoded.Note.Memo, credit.Memo) {
			t.Fatal("Credit payload memo", err)
		}
		if role == CreditOwner {
			context.Role = CreditSpender
		} else {
			context.Role = CreditOwner
		}
		if _, err = DecryptCreditNote(private, ciphertext, context); err == nil {
			t.Fatal("role switch accepted")
		}
	}
	nf, _ := ComputeCreditNullifier(credit.ChainID, credit.CreditSharedKey, leaf)
	child := credit
	child.Amount = ElementFromUint64(70)
	child.Nonce, _ = DeriveRemainderNonce(credit.ChainID, credit.ProtocolAddress, credit.TokenAddress, nf, credit.CreditSharedKey)
	childLeaf, _ := ComputeCreditCommitment(child.ChainID, child.SpenderWalletIdentity, child.OwnerWalletIdentity, child.SpenderPayout, child.OwnerPayout, child.TokenAddress, child.Amount, child.Nonce, child.Expiry, keyHash)
	pad, _ := ComputeCreditRecoveryPad(credit.ChainID, credit.ProtocolAddress, credit.TokenAddress, nf, credit.CreditSharedKey)
	amount := child.Amount
	amount.Add(&amount, &pad)
	recovered, err := RecoverCreditRemainder(credit.ProtocolContext, credit, CreditRemainderRecovery{nf, childLeaf, amount})
	if err != nil || recovered == nil || recovered.Amount != child.Amount {
		t.Fatal("financial recovery failed", err)
	}
	if len(recovered.Memo) != 0 {
		t.Fatal("financial recovery fabricated child memo")
	}
	normalized, err := NormalizeCreditDeliveryState(state)
	if err != nil {
		t.Fatal(err)
	}
	normalized.Note.Memo[0] = 0
	if credit.Memo[0] != 0xff {
		t.Fatal("normalized memo alias")
	}
	creditContext := CreditNoteDecryptionContext{NoteDecryptionContext: noteContext, Role: CreditOwner}
	creditContext.OutputCommitment = leaf
	for _, tail := range [][]byte{{}, {0}, {1, 1}, {0, 0, 1}, append([]byte{1, 1}, make([]byte, 257)...)} {
		plain := append(encodeCreditState(state)[:322], tail...)
		ciphertext, err := sealEnvelope(public, plain, creditContext.NoteEncryptionContext, credit.TokenAddress, 4, 1, bytes.NewReader(bytes.Repeat([]byte{1}, 100)))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = DecryptCreditNote(private, ciphertext, creditContext); err == nil {
			t.Fatal("malformed Credit memo accepted")
		}
	}
	state.Note.Memo = make([]byte, 257)
	if _, err := EncryptCreditNote(public, state, CreditNoteEncryptionContext{NoteEncryptionContext: NoteEncryptionContext{credit.ProtocolContext, leaf}, Role: CreditOwner}); err == nil {
		t.Fatal("257 byte Credit memo accepted")
	}
}
