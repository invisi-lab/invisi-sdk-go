package core

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/invisi-lab/invisi-sdk-go/internal/cryptography/poseidon"
)

// Poseidon2Hash computes the canonical BN254 Poseidon2 sponge hash.
func Poseidon2Hash(inputs ...fr.Element) (fr.Element, error) {
	return poseidon.Poseidon2Hash(inputs...)
}
