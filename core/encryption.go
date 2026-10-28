package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

const ordinaryNoteDomain = "InvisibleProtocol.PrivateTransferNote.v1"
const creditNoteDomain = "InvisibleProtocol.CreditNote.v1"

// NoteEncryptionContext authenticates the deployment and output commitment.
type NoteEncryptionContext struct {
	ProtocolContext
	OutputCommitment fr.Element
}

// NoteDecryptionContext supplies the token address omitted from the encrypted note.
type NoteDecryptionContext struct {
	NoteEncryptionContext
	TokenAddress fr.Element
}

// ViewingPublicKey returns a compressed 33-byte secp256k1 viewing key.
// Viewing keys are independent of note secrets and Credit shared keys.
func ViewingPublicKey(privateKey []byte) ([]byte, error) {
	key, err := ethcrypto.ToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid viewing private key: %w", err)
	}
	return ethcrypto.CompressPubkey(&key.PublicKey), nil
}

func validateNote(note Note) error {
	if err := checkUint(note.Amount, 128, "amount"); err != nil {
		return err
	}
	if note.Amount.IsZero() {
		return fmt.Errorf("amount must be positive")
	}
	if err := checkUint(note.ReceivingAddress, 160, "receivingAddress"); err != nil {
		return err
	}
	return checkUint(note.TokenAddress, 160, "tokenAddress")
}

func verifyNote(context NoteEncryptionContext, note Note) error {
	if err := validateProtocolContext(context.ProtocolContext, note.TokenAddress); err != nil {
		return err
	}
	if err := validateNote(note); err != nil {
		return err
	}
	identity, err := ComputeIdentityCommitment(note.NoteSecret, note.ReceivingAddress)
	if err != nil {
		return err
	}
	commitment, err := ComputeNoteCommitment(context.ChainID, identity, note.TokenAddress, note.Amount, note.Nonce)
	if err != nil {
		return err
	}
	if commitment != context.OutputCommitment {
		return fmt.Errorf("output commitment mismatch")
	}
	return nil
}
func verifyCredit(context NoteEncryptionContext, note CreditNote) error {
	if err := validateProtocolContext(context.ProtocolContext, note.TokenAddress); err != nil {
		return err
	}
	if err := validateCredit(note); err != nil {
		return err
	}
	if note.Expiry.IsZero() {
		return fmt.Errorf("credit expiry must be positive")
	}
	hash, err := ComputeCreditSharedKeyHash(note.CreditSharedKey)
	if err != nil {
		return err
	}
	commitment, err := ComputeCreditCommitment(context.ChainID, note.SpenderIdentityCommitment, note.OwnerIdentityCommitment, note.TokenAddress, note.Amount, note.Nonce, note.Expiry, hash)
	if err != nil {
		return err
	}
	if commitment != context.OutputCommitment {
		return fmt.Errorf("credit commitment mismatch")
	}
	return nil
}
func appendElement(out []byte, value fr.Element, length int) []byte {
	b := value.Bytes()
	return append(out, b[32-length:]...)
}
func encodeNote(note Note) []byte {
	out := make([]byte, 0, 100)
	out = appendElement(out, note.Amount, 16)
	out = appendElement(out, note.Nonce, 32)
	out = appendElement(out, note.NoteSecret, 32)
	return appendElement(out, note.ReceivingAddress, 20)
}
func encodeCredit(note CreditNote) []byte {
	out := make([]byte, 0, 176)
	out = appendElement(out, note.SpenderIdentityCommitment, 32)
	out = appendElement(out, note.OwnerIdentityCommitment, 32)
	out = appendElement(out, note.Amount, 16)
	out = appendElement(out, note.Nonce, 32)
	out = appendElement(out, note.Expiry, 32)
	return appendElement(out, note.CreditSharedKey, 32)
}
func decodeElement(data []byte) (fr.Element, error) {
	var value fr.Element
	var padded [32]byte
	if len(data) > len(padded) {
		return value, fmt.Errorf("field exceeds 32 bytes")
	}
	copy(padded[32-len(data):], data)
	if err := value.SetBytesCanonical(padded[:]); err != nil {
		return value, fmt.Errorf("noncanonical plaintext field: %w", err)
	}
	return value, nil
}
func envelopeAAD(domain string, context NoteEncryptionContext, token fr.Element) ([]byte, error) {
	if err := validateProtocolContext(context.ProtocolContext, token); err != nil {
		return nil, err
	}
	out := append([]byte(nil), domain...)
	out = appendElement(out, context.ChainID, 32)
	out = appendElement(out, context.ProtocolAddress, 20)
	out = appendElement(out, token, 20)
	return appendElement(out, context.OutputCommitment, 32), nil
}
func envelopeCipher(privateKey, publicKey, salt []byte, domain string) (cipher.AEAD, error) {
	key, err := ethcrypto.ToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid viewing private key: %w", err)
	}
	if len(publicKey) != 33 || (publicKey[0] != 2 && publicKey[0] != 3) {
		return nil, fmt.Errorf("expected compressed secp256k1 public key")
	}
	peer, err := ethcrypto.DecompressPubkey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid viewing public key: %w", err)
	}
	x, _ := ethcrypto.S256().ScalarMult(peer.X, peer.Y, key.D.Bytes())
	if x == nil {
		return nil, fmt.Errorf("invalid shared point")
	}
	secret := x.FillBytes(make([]byte, 32))
	derived, err := hkdf.Key(sha256.New, secret, salt, domain, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(derived)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
func sealEnvelope(viewingKey, plaintext []byte, context NoteEncryptionContext, token fr.Element, version byte, entropy io.Reader) ([]byte, error) {
	domain := ordinaryNoteDomain
	if version == 2 {
		domain = creditNoteDomain
	}
	aad, err := envelopeAAD(domain, context, token)
	if err != nil {
		return nil, err
	}
	private := make([]byte, 32)
	var public []byte
	for {
		if _, err = io.ReadFull(entropy, private); err != nil {
			return nil, err
		}
		public, err = ViewingPublicKey(private)
		if err == nil {
			break
		}
	}
	salt := make([]byte, 16)
	if _, err = io.ReadFull(entropy, salt); err != nil {
		return nil, err
	}
	iv := make([]byte, 12)
	if _, err = io.ReadFull(entropy, iv); err != nil {
		return nil, err
	}
	aead, err := envelopeCipher(private, viewingKey, salt, domain)
	if err != nil {
		return nil, err
	}
	header := append([]byte{version}, public...)
	header = append(header, salt...)
	header = append(header, iv...)
	if version == 2 {
		aad = append(aad, header...)
	}
	return aead.Seal(header, iv, plaintext, aad), nil
}
func openEnvelope(private, envelope []byte, context NoteDecryptionContext, version byte) ([]byte, error) {
	length, domain := 178, ordinaryNoteDomain
	if version == 2 {
		length, domain = 254, creditNoteDomain
	}
	if len(envelope) != length {
		return nil, fmt.Errorf("expected %d envelope bytes", length)
	}
	if envelope[0] != version {
		return nil, fmt.Errorf("unsupported envelope version")
	}
	aad, err := envelopeAAD(domain, context.NoteEncryptionContext, context.TokenAddress)
	if err != nil {
		return nil, err
	}
	if version == 2 {
		aad = append(aad, envelope[:62]...)
	}
	aead, err := envelopeCipher(private, envelope[1:34], envelope[34:50], domain)
	if err != nil {
		return nil, err
	}
	plaintext, err := aead.Open(nil, envelope[50:62], envelope[62:], aad)
	if err != nil {
		return nil, fmt.Errorf("note authentication failed: %w", err)
	}
	return plaintext, nil
}

// EncryptPrivateTransferNote seals the current 178-byte v1 ordinary envelope,
// using a fresh ephemeral key, salt and IV.
func EncryptPrivateTransferNote(viewingKey []byte, note Note, context NoteEncryptionContext) ([]byte, error) {
	if err := verifyNote(context, note); err != nil {
		return nil, err
	}
	return sealEnvelope(viewingKey, encodeNote(note), context, note.TokenAddress, 1, rand.Reader)
}

// DecryptPrivateTransferNote authenticates context, canonical fields and commitment.
func DecryptPrivateTransferNote(privateKey, envelope []byte, context NoteDecryptionContext) (Note, error) {
	var note Note
	plaintext, err := openEnvelope(privateKey, envelope, context, 1)
	if err != nil {
		return note, err
	}
	fields := []*fr.Element{&note.Amount, &note.Nonce, &note.NoteSecret, &note.ReceivingAddress}
	offsets := []int{0, 16, 48, 80, 100}
	for i, target := range fields {
		value, err := decodeElement(plaintext[offsets[i]:offsets[i+1]])
		if err != nil {
			return Note{}, err
		}
		*target = value
	}
	note.TokenAddress = context.TokenAddress
	if err = verifyNote(context.NoteEncryptionContext, note); err != nil {
		return Note{}, err
	}
	return note, nil
}

// EncryptCreditNote seals the current 254-byte v2 Credit envelope. The full
// header is authenticated and no viewing or ordinary note secret is serialized.
func EncryptCreditNote(viewingKey []byte, note CreditNote, context NoteEncryptionContext) ([]byte, error) {
	if err := verifyCredit(context, note); err != nil {
		return nil, err
	}
	return sealEnvelope(viewingKey, encodeCredit(note), context, note.TokenAddress, 2, rand.Reader)
}

// DecryptCreditNote authenticates the complete header, context and commitment.
func DecryptCreditNote(privateKey, envelope []byte, context NoteDecryptionContext) (CreditNote, error) {
	var note CreditNote
	plaintext, err := openEnvelope(privateKey, envelope, context, 2)
	if err != nil {
		return note, err
	}
	fields := []*fr.Element{&note.SpenderIdentityCommitment, &note.OwnerIdentityCommitment, &note.Amount, &note.Nonce, &note.Expiry, &note.CreditSharedKey}
	offsets := []int{0, 32, 64, 80, 112, 144, 176}
	for i, target := range fields {
		value, err := decodeElement(plaintext[offsets[i]:offsets[i+1]])
		if err != nil {
			return CreditNote{}, err
		}
		*target = value
	}
	note.TokenAddress = context.TokenAddress
	if err = verifyCredit(context.NoteEncryptionContext, note); err != nil {
		return CreditNote{}, err
	}
	return note, nil
}
