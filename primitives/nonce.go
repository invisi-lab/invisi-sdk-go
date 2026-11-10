package primitives

import (
	"crypto/rand"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// FieldModulus returns a fresh copy of the BN254 scalar field modulus.
func FieldModulus() *big.Int { return fr.Modulus() }

// GenerateNonce samples uniformly from the canonical BN254 field.
func GenerateNonce() (fr.Element, error) {
	n, err := rand.Int(rand.Reader, fr.Modulus())
	if err != nil {
		return fr.Element{}, err
	}
	return BigIntToElement(n), nil
}
