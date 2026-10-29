package core

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// ExportPayloadKey discloses only the selected envelope's symmetric key.
// Callers must authenticate that envelope with the viewing key before exporting it.
func ExportPayloadKey(private, envelope []byte, credit bool) ([]byte, error) {
	length, version, domain := 160, byte(2), ordinaryNoteDomain
	if credit {
		length, version, domain = 498, 5, creditNoteDomain
	}
	if (len(envelope) < length || len(envelope) > length+MaxMemoBytes) || envelope[0] != version {
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
	if version == 5 {
		length, domain = 498, creditNoteDomain
	}
	if len(key) != 32 || (len(envelope) < length || len(envelope) > length+MaxMemoBytes) || envelope[0] != version {
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
	plain, e := openPayload(key, envelope, context, 2, 0)
	if e != nil {
		return Note{}, e
	}
	defer clear(plain)
	return readNotePlain(plain, context)
}
func DecryptCreditPayload(key, envelope []byte, context CreditNoteDecryptionContext) (CreditDeliveryState, error) {
	role, e := creditRole(context.Role)
	if e != nil {
		return CreditDeliveryState{}, e
	}
	plain, e := openPayload(key, envelope, context.NoteDecryptionContext, 5, role)
	if e != nil {
		return CreditDeliveryState{}, e
	}
	defer clear(plain)
	return readCreditPlain(plain, context, nil)
}
