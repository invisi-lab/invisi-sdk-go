package core

import (
	"bytes"
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

const ordinaryNoteDomain = "Invisi.NoteDelivery.v1"
const creditNoteDomain = "Invisi.CreditDelivery.v1"

type NoteEncryptionContext struct {
	ProtocolContext
	OutputCommitment fr.Element
}
type NoteDecryptionContext struct {
	NoteEncryptionContext
	TokenAddress, WalletIdentity fr.Element
}
type CreditDeliveryRole string

const (
	CreditOwner   CreditDeliveryRole = "owner"
	CreditSpender CreditDeliveryRole = "spender"
)

type CreditNoteEncryptionContext struct {
	NoteEncryptionContext
	Role CreditDeliveryRole
}
type CreditNoteDecryptionContext struct {
	NoteDecryptionContext
	Role CreditDeliveryRole
}

func creditRole(role CreditDeliveryRole) (byte, error) {
	switch role {
	case CreditOwner:
		return 1, nil
	case CreditSpender:
		return 2, nil
	default:
		return 0, fmt.Errorf("Credit role must be owner or spender")
	}
}

// ViewingPublicKey returns an independent compressed secp256k1 viewing key.
func ViewingPublicKey(privateKey []byte) ([]byte, error) {
	key, err := ethcrypto.ToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid viewing private key: %w", err)
	}
	return ethcrypto.CompressPubkey(&key.PublicKey), nil
}
func validateNote(note Note) error {
	if note.WalletIdentity.IsZero() || note.Amount.IsZero() {
		return fmt.Errorf("wallet identity and amount must be nonzero")
	}
	if err := checkUint(note.Amount, 128, "amount"); err != nil {
		return err
	}
	return checkUint(note.TokenAddress, 160, "tokenAddress")
}
func verifyNote(context NoteEncryptionContext, note Note) error {
	if err := validateProtocolContext(context.ProtocolContext, note.TokenAddress); err != nil {
		return err
	}
	if note.ProtocolContext != context.ProtocolContext {
		return fmt.Errorf("note context mismatch")
	}
	if err := validateNote(note); err != nil {
		return err
	}
	identity, err := ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
	if err != nil {
		return err
	}
	leaf, err := ComputeNoteCommitment(context.ChainID, identity, note.TokenAddress, note.Amount)
	if err != nil {
		return err
	}
	if leaf != context.OutputCommitment {
		return fmt.Errorf("output commitment mismatch")
	}
	return nil
}
func verifyCredit(context NoteEncryptionContext, note CreditNote) error {
	if err := validateProtocolContext(context.ProtocolContext, note.TokenAddress); err != nil {
		return err
	}
	if note.ProtocolContext != context.ProtocolContext {
		return fmt.Errorf("Credit context mismatch")
	}
	if err := validateCredit(note); err != nil {
		return err
	}
	keyHash, err := ComputeCreditSharedKeyHash(note.CreditSharedKey)
	if err != nil {
		return err
	}
	leaf, err := ComputeCreditCommitment(context.ChainID, note.SpenderWalletIdentity, note.OwnerWalletIdentity, note.SpenderPayout, note.OwnerPayout, note.TokenAddress, note.Amount, note.Nonce, note.Expiry, keyHash)
	if err != nil {
		return err
	}
	if leaf != context.OutputCommitment {
		return fmt.Errorf("credit commitment mismatch")
	}
	return nil
}
func appendElement(out []byte, value fr.Element, length int) []byte {
	b := value.Bytes()
	return append(out, b[32-length:]...)
}
func encodeNote(note Note) []byte {
	out := appendElement(nil, note.WalletIdentity, 32)
	out = appendElement(out, note.Amount, 16)
	return appendElement(out, note.Nonce, 32)
}
func encodeCredit(note CreditNote) []byte {
	out := appendElement(nil, note.SpenderWalletIdentity, 32)
	out = appendElement(out, note.OwnerWalletIdentity, 32)
	out = appendElement(out, note.SpenderPayout, 20)
	out = appendElement(out, note.OwnerPayout, 20)
	out = appendElement(out, note.Amount, 16)
	out = appendElement(out, note.Nonce, 32)
	out = appendElement(out, note.Expiry, 8)
	return appendElement(out, note.CreditSharedKey, 32)
}
func decodeElement(data []byte) (fr.Element, error) {
	var value fr.Element
	var padded [32]byte
	if len(data) > 32 {
		return value, fmt.Errorf("field exceeds 32 bytes")
	}
	copy(padded[32-len(data):], data)
	if err := value.SetBytesCanonical(padded[:]); err != nil {
		return value, fmt.Errorf("noncanonical plaintext field: %w", err)
	}
	return value, nil
}
func envelopeAAD(domain string, context NoteEncryptionContext, token fr.Element, header []byte, role byte) ([]byte, error) {
	if err := validateProtocolContext(context.ProtocolContext, token); err != nil {
		return nil, err
	}
	if len(header) != 62 {
		return nil, fmt.Errorf("invalid envelope header")
	}
	out := append([]byte(domain), header...)
	out = append(out, role)
	out = appendElement(out, context.ChainID, 8)
	out = appendElement(out, context.ProtocolAddress, 20)
	out = appendElement(out, token, 20)
	return appendElement(out, context.OutputCommitment, 32), nil
}
func envelopeKey(private, peerKey, salt []byte, domain string, ephemeral, recipient []byte) ([]byte, error) {
	key, err := ethcrypto.ToECDSA(private)
	if err != nil {
		return nil, fmt.Errorf("invalid viewing private key: %w", err)
	}
	if len(peerKey) != 33 || (peerKey[0] != 2 && peerKey[0] != 3) {
		return nil, fmt.Errorf("expected compressed secp256k1 public key")
	}
	peer, err := ethcrypto.DecompressPubkey(peerKey)
	if err != nil {
		return nil, fmt.Errorf("invalid viewing public key: %w", err)
	}
	x, _ := ethcrypto.S256().ScalarMult(peer.X, peer.Y, key.D.Bytes())
	if x == nil {
		return nil, fmt.Errorf("invalid shared point")
	}
	secret := x.FillBytes(make([]byte, 32))
	defer clear(secret)
	info := append([]byte(domain), ephemeral...)
	info = append(info, recipient...)
	derived, err := hkdf.Key(sha256.New, secret, salt, string(info), 32)
	if err != nil {
		return nil, err
	}
	return derived, nil
}
func envelopeCipher(private, peerKey, salt []byte, domain string, ephemeral, recipient []byte) (cipher.AEAD, error) {
	derived, e := envelopeKey(private, peerKey, salt, domain, ephemeral, recipient)
	if e != nil {
		return nil, e
	}
	defer clear(derived)
	block, e := aes.NewCipher(derived)
	if e != nil {
		return nil, e
	}
	return cipher.NewGCM(block)
}
func sealEnvelope(viewingKey, plaintext []byte, context NoteEncryptionContext, token fr.Element, version, role byte, entropy io.Reader) ([]byte, error) {
	domain := ordinaryNoteDomain
	if version == 3 {
		domain = creditNoteDomain
	}
	private := make([]byte, 32)
	defer clear(private)
	var public []byte
	var err error
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
	header := append([]byte{version}, public...)
	header = append(header, salt...)
	header = append(header, iv...)
	aad, err := envelopeAAD(domain, context, token, header, role)
	if err != nil {
		return nil, err
	}
	aead, err := envelopeCipher(private, viewingKey, salt, domain, public, viewingKey)
	if err != nil {
		return nil, err
	}
	return aead.Seal(header, iv, plaintext, aad), nil
}
func openEnvelope(private, envelope []byte, context NoteDecryptionContext, version, role byte) ([]byte, error) {
	length, domain := 158, ordinaryNoteDomain
	if version == 3 {
		length, domain = 400, creditNoteDomain
	}
	if len(envelope) != length {
		return nil, fmt.Errorf("expected %d envelope bytes", length)
	}
	if envelope[0] != version {
		return nil, fmt.Errorf("unsupported envelope version")
	}
	aad, err := envelopeAAD(domain, context.NoteEncryptionContext, context.TokenAddress, envelope[:62], role)
	if err != nil {
		return nil, err
	}
	recipient, err := ViewingPublicKey(private)
	if err != nil {
		return nil, err
	}
	aead, err := envelopeCipher(private, envelope[1:34], envelope[34:50], domain, envelope[1:34], recipient)
	if err != nil {
		return nil, err
	}
	plaintext, err := aead.Open(nil, envelope[50:62], envelope[62:], aad)
	if err != nil {
		return nil, fmt.Errorf("note authentication failed: %w", err)
	}
	return plaintext, nil
}

// EncryptPrivateTransferNote creates a 158-byte envelope with fresh ephemeral key, salt and IV.
func EncryptPrivateTransferNote(viewingKey []byte, note Note, context NoteEncryptionContext) ([]byte, error) {
	if err := verifyNote(context, note); err != nil {
		return nil, err
	}
	plain := encodeNote(note)
	defer clear(plain)
	return sealEnvelope(viewingKey, plain, context, note.TokenAddress, 1, 0, rand.Reader)
}

// DecryptPrivateTransferNote verifies recipient, deployment, canonical fields and commitment.
func DecryptPrivateTransferNote(private, envelope []byte, context NoteDecryptionContext) (Note, error) {
	plain, err := openEnvelope(private, envelope, context, 1, 0)
	if err != nil {
		return Note{}, err
	}
	defer clear(plain)
	return readNotePlain(plain, context)
}

func readNotePlain(plain []byte, context NoteDecryptionContext) (Note, error) {
	var err error
	if len(plain) != 80 {
		return Note{}, fmt.Errorf("invalid note plaintext size")
	}
	note := Note{ProtocolContext: context.ProtocolContext, TokenAddress: context.TokenAddress}
	fields := []*fr.Element{&note.WalletIdentity, &note.Amount, &note.Nonce}
	offsets := []int{0, 32, 48, 80}
	for i, target := range fields {
		value, err := decodeElement(plain[offsets[i]:offsets[i+1]])
		if err != nil {
			return Note{}, err
		}
		*target = value
	}
	if note.WalletIdentity != context.WalletIdentity {
		return Note{}, fmt.Errorf("recipient wallet identity mismatch")
	}
	if err = verifyNote(context.NoteEncryptionContext, note); err != nil {
		return Note{}, err
	}
	return note, nil
}

// NormalizeCreditDeliveryState validates and copies public receiving snapshots.
func NormalizeCreditDeliveryState(state CreditDeliveryState) (CreditDeliveryState, error) {
	owner, err := walletAddressKeys(state.OwnerReceivingKeys)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	spender, err := walletAddressKeys(state.SpenderReceivingKeys)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	return CreditDeliveryState{Note: state.Note, OwnerReceivingKeys: owner, SpenderReceivingKeys: spender}, nil
}
func matchCreditRecipient(state CreditDeliveryState, role byte, public []byte) error {
	keys := state.OwnerReceivingKeys
	if role == 2 {
		keys = state.SpenderReceivingKeys
	}
	if !bytes.Equal(keys.ViewingPublicKey, public) {
		return fmt.Errorf("Credit recipient key differs from its snapshot")
	}
	return nil
}
func encodeCreditState(state CreditDeliveryState) []byte {
	out := encodeCredit(state.Note)
	out = append(out, state.OwnerReceivingKeys.ViewingPublicKey...)
	out = append(out, state.OwnerReceivingKeys.ClueKey...)
	out = append(out, state.SpenderReceivingKeys.ViewingPublicKey...)
	return append(out, state.SpenderReceivingKeys.ClueKey...)
}

// EncryptCreditNote creates a 400-byte role-bound envelope with complete receiving snapshots.
func EncryptCreditNote(viewingKey []byte, state CreditDeliveryState, context CreditNoteEncryptionContext) ([]byte, error) {
	role, err := creditRole(context.Role)
	if err != nil {
		return nil, err
	}
	state, err = NormalizeCreditDeliveryState(state)
	if err != nil {
		return nil, err
	}
	if err = matchCreditRecipient(state, role, viewingKey); err != nil {
		return nil, err
	}
	if err = verifyCredit(context.NoteEncryptionContext, state.Note); err != nil {
		return nil, err
	}
	plain := encodeCreditState(state)
	defer clear(plain)
	return sealEnvelope(viewingKey, plain, context.NoteEncryptionContext, state.Note.TokenAddress, 3, role, rand.Reader)
}
func DecryptCreditNote(private, envelope []byte, context CreditNoteDecryptionContext) (CreditDeliveryState, error) {
	role, err := creditRole(context.Role)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	plain, err := openEnvelope(private, envelope, context.NoteDecryptionContext, 3, role)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	defer clear(plain)
	public, err := ViewingPublicKey(private)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	return readCreditPlain(plain, context, public)
}

func readCreditPlain(plain []byte, context CreditNoteDecryptionContext, public []byte) (CreditDeliveryState, error) {
	if len(plain) != 322 {
		return CreditDeliveryState{}, fmt.Errorf("invalid Credit plaintext size")
	}
	role, err := creditRole(context.Role)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	note := CreditNote{ProtocolContext: context.ProtocolContext, TokenAddress: context.TokenAddress}
	fields := []*fr.Element{&note.SpenderWalletIdentity, &note.OwnerWalletIdentity, &note.SpenderPayout, &note.OwnerPayout, &note.Amount, &note.Nonce, &note.Expiry, &note.CreditSharedKey}
	offsets := []int{0, 32, 64, 84, 104, 120, 152, 160, 192}
	for i, target := range fields {
		value, err := decodeElement(plain[offsets[i]:offsets[i+1]])
		if err != nil {
			return CreditDeliveryState{}, err
		}
		*target = value
	}
	expected := note.OwnerWalletIdentity
	if role == 2 {
		expected = note.SpenderWalletIdentity
	}
	if expected != context.WalletIdentity {
		return CreditDeliveryState{}, fmt.Errorf("recipient wallet identity mismatch")
	}
	if err = verifyCredit(context.NoteEncryptionContext, note); err != nil {
		return CreditDeliveryState{}, err
	}
	state, err := NormalizeCreditDeliveryState(CreditDeliveryState{Note: note,
		OwnerReceivingKeys:   ReceivingKeys{ViewingPublicKey: plain[192:225], ClueKey: plain[225:257]},
		SpenderReceivingKeys: ReceivingKeys{ViewingPublicKey: plain[257:290], ClueKey: plain[290:322]}})
	if err != nil {
		return CreditDeliveryState{}, err
	}
	if public != nil {
		if err = matchCreditRecipient(state, role, public); err != nil {
			return CreditDeliveryState{}, err
		}
	}
	return state, nil
}
