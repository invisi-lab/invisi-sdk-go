package primitives

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
	note := Note{ProtocolContext: ProtocolContext{ElementFromUint64(1), ElementFromUint64(2)}, OwnerAddress: ElementFromUint64(3), TokenAddress: ElementFromUint64(4), Amount: ElementFromUint64(5), Nonce: ElementFromUint64(6)}
	identity, _ := ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
	leaf, _ := ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
	return note, private, public, NoteDecryptionContext{NoteEncryptionContext: NoteEncryptionContext{note.ProtocolContext, leaf}, TokenAddress: note.TokenAddress, OwnerAddress: note.OwnerAddress}
}
func TestEmptyMemoUsesCurrentEnvelope(t *testing.T) {
	note, private, public, context := memoNote(t)
	ciphertext, err := EncryptPrivateTransferNote(public, note, context.NoteEncryptionContext)
	if err != nil {
		t.Fatal(err)
	}
	if len(ciphertext) != 160 || ciphertext[0] != 1 {
		t.Fatalf("expected version 1 and 160 bytes, got version %d and %d bytes", ciphertext[0], len(ciphertext))
	}
	if _, err = DecryptPrivateTransferNote(private, ciphertext, context); err != nil {
		t.Fatal(err)
	}
}
func TestMemoRoundTripAndPayloadKey(t *testing.T) {
	note, private, public, context := memoNote(t)
	for _, memo := range [][]byte{nil, []byte("hello\x00"), {0xff, 0xfe, 0, 1}, bytes.Repeat([]byte{0xab}, 256)} {
		note.Payload = memo
		ciphertext, err := EncryptPrivateTransferNote(public, note, context.NoteEncryptionContext)
		if err != nil {
			t.Fatal(err)
		}
		if len(ciphertext) != 160+len(memo) {
			t.Fatal("unexpected envelope size")
		}
		decoded, err := DecryptPrivateTransferNote(private, ciphertext, context)
		if err != nil || !bytes.Equal(decoded.Payload, memo) {
			t.Fatal("memo roundtrip", err)
		}
		key, err := ExportPayloadKey(private, ciphertext)
		if err != nil {
			t.Fatal(err)
		}
		opened, err := DecryptNotePayload(key, ciphertext, context)
		if err != nil || !bytes.Equal(opened.Payload, memo) {
			t.Fatal("payload memo", err)
		}
		encoded, err := json.Marshal(note)
		if err != nil {
			t.Fatal(err)
		}
		var restored Note
		if err = json.Unmarshal(encoded, &restored); err != nil || !bytes.Equal(restored.Payload, memo) {
			t.Fatal("memo storage", err)
		}
		if len(decoded.Payload) > 0 {
			decoded.Payload[0] ^= 1
			if !bytes.Equal(note.Payload, memo) {
				t.Fatal("memo alias")
			}
		}
	}
	note.Payload = make([]byte, MaxCiphertextBytes)
	if _, err := EncryptPrivateTransferNote(public, note, context.NoteEncryptionContext); err == nil {
		t.Fatal("oversize memo accepted")
	}
}
func TestMemoTextValidation(t *testing.T) {
	text := "hello"
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
func TestMalformedMemoLengthsAndUnknownEnvelopeFormat(t *testing.T) {
	note, private, public, context := memoNote(t)
	for _, tail := range [][]byte{{}, {0}, {1, 1}, {0, 0, 1}} {
		plain := append(encodeNote(note)[:80], tail...)
		ciphertext, err := sealEnvelope(public, plain, context.NoteEncryptionContext, note.TokenAddress, 0x01, 0, bytes.NewReader(bytes.Repeat([]byte{1}, 100)))
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
	ciphertext[0] = 2
	if _, err = DecryptPrivateTransferNote(private, ciphertext, context); err == nil {
		t.Fatal("unknown envelope format accepted")
	}
}
