package primitives

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// ViewingPublicKey returns an independent compressed secp256k1 viewing key.
func ViewingPublicKey(privateKey []byte) ([]byte, error) {
	key, err := ethcrypto.ToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid viewing private key: %w", err)
	}
	return ethcrypto.CompressPubkey(&key.PublicKey), nil
}

func appendElement(out []byte, value fr.Element, length int) []byte {
	b := value.Bytes()
	return append(out, b[32-length:]...)
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
