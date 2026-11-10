package primitives

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

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
	length, domain := 160, ordinaryNoteDomain
	if len(envelope) < length || len(envelope) > MaxCiphertextBytes {
		return nil, fmt.Errorf("expected %d..%d envelope bytes", length, MaxCiphertextBytes)
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
