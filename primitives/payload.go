package primitives

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// ExportPayloadKey discloses the selected output's full plaintext, including its amount and payload.
// Funds disclosure also enables spend tracking and a full sweep back to the original owner.
// Callers must authenticate that envelope with the viewing key before exporting it.
func ExportPayloadKey(private, envelope []byte) ([]byte, error) {
	length, version, domain := 160, byte(0x01), ordinaryNoteDomain
	if (len(envelope) < length || len(envelope) > MaxCiphertextBytes) || envelope[0] != version {
		return nil, fmt.Errorf("unsupported delivery envelope")
	}
	public, e := ViewingPublicKey(private)
	if e != nil {
		return nil, e
	}
	return envelopeKey(private, envelope[1:34], envelope[34:50], domain, envelope[1:34], public)
}
func openPayload(key, envelope []byte, context NoteDecryptionContext, version, role byte) ([]byte, error) {
	length, domain := 160, ordinaryNoteDomain
	if len(key) != 32 || (len(envelope) < length || len(envelope) > MaxCiphertextBytes) || envelope[0] != version {
		return nil, fmt.Errorf("invalid payload key or envelope")
	}
	aad, e := envelopeAAD(domain, context.NoteEncryptionContext, context.TokenAddress, envelope[:62], role)
	if e != nil {
		return nil, e
	}
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	aead, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	return aead.Open(nil, envelope[50:62], envelope[62:], aad)
}
func DecryptNotePayload(key, envelope []byte, context NoteDecryptionContext) (Note, error) {
	plain, e := openPayload(key, envelope, context, 0x01, 0)
	if e != nil {
		return Note{}, e
	}
	defer clear(plain)
	return readNotePlain(plain, context)
}
