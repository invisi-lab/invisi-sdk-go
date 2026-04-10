package core

import (
	"crypto/rand"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// FieldModulus is the BN254 scalar field modulus.
var FieldModulus, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10,
)

// GenerateNonce generates a cryptographically secure random nonce
// within the BN254 scalar field.
func GenerateNonce() (fr.Element, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return fr.Element{}, err
	}
	raw := new(big.Int).SetBytes(buf)
	raw.Mod(raw, FieldModulus)
	var e fr.Element
	e.SetBigInt(raw)
	return e, nil
}
