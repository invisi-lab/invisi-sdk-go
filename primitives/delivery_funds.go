package primitives

import (
	"crypto/rand"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func encodeNote(note Note) []byte {
	out := appendElement(nil, note.OwnerAddress, 32)
	out = appendElement(out, note.Amount, 16)
	return appendPayload(appendElement(out, note.Nonce, 32), note.Payload)
}

// EncryptPrivateTransferNote creates a v1 funds envelope with fresh ephemeral key, salt and IV.
func EncryptPrivateTransferNote(viewingKey []byte, note Note, context NoteEncryptionContext) ([]byte, error) {
	if err := verifyNote(context, note); err != nil {
		return nil, err
	}
	plain := encodeNote(note)
	if len(plain)+78 > MaxCiphertextBytes {
		return nil, fmt.Errorf("envelope exceeds resource bound")
	}
	defer clear(plain)
	return sealEnvelope(viewingKey, plain, context, note.TokenAddress, 0x01, 0, rand.Reader)
}

// DecryptPrivateTransferNote verifies recipient, deployment, canonical fields and commitment.
func DecryptPrivateTransferNote(private, envelope []byte, context NoteDecryptionContext) (Note, error) {
	plain, err := openEnvelope(private, envelope, context, 0x01, 0)
	if err != nil {
		return Note{}, err
	}
	defer clear(plain)
	return readNotePlain(plain, context)
}

func readNotePlain(plain []byte, context NoteDecryptionContext) (Note, error) {
	memo, err := readPayload(plain, 80)
	if err != nil {
		return Note{}, fmt.Errorf("invalid note plaintext: %w", err)
	}
	note := Note{ProtocolContext: context.ProtocolContext, TokenAddress: context.TokenAddress, Payload: memo}
	fields := []*fr.Element{&note.OwnerAddress, &note.Amount, &note.Nonce}
	offsets := []int{0, 32, 48, 80}
	for i, target := range fields {
		value, err := decodeElement(plain[offsets[i]:offsets[i+1]])
		if err != nil {
			return Note{}, err
		}
		*target = value
	}
	if note.OwnerAddress != context.OwnerAddress {
		return Note{}, fmt.Errorf("recipient owner address mismatch")
	}
	if err = verifyNote(context.NoteEncryptionContext, note); err != nil {
		return Note{}, err
	}
	return note, nil
}

// DecodeNotePlaintext checks identity, context and commitment for native wallet responses.
func DecodeNotePlaintext(plain []byte, context NoteDecryptionContext) (Note, error) {
	return readNotePlain(plain, context)
}

// EncodeNotePlaintext returns the canonical wallet plaintext bytes.
func EncodeNotePlaintext(note Note) []byte { return encodeNote(note) }
