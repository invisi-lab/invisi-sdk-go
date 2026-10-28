package core

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Poseidon2Hash computes the Poseidon2 sponge hash over the given field
// elements. The implementation is bit-for-bit compatible with Barretenberg
// (bb.js) used in Noir/Aztec circuits.
//
// Sponge construction:
//   - IV:      state = [0, 0, 0, len(inputs) * 2^64]
//   - Absorb:  add inputs in chunks of rate=3 to state[0..2], then permute
//   - Squeeze:  return state[0] after final permutation
func Poseidon2Hash(inputs ...fr.Element) (fr.Element, error) {
	if len(inputs) == 0 {
		return fr.Element{}, errors.New("poseidon2: input must not be empty")
	}
	initPoseidon2Constants()

	// Build IV: len(inputs) << 64, placed in state[rate] (the capacity slot).
	var iv fr.Element
	iv.SetUint64(uint64(len(inputs)))
	iv.Mul(&iv, &twoTo64)

	// Initialize sponge state.
	var state [poseidon2T]fr.Element
	state[poseidon2Rate] = iv

	// Absorb phase: buffer inputs and permute when the buffer is full.
	var cache [poseidon2Rate]fr.Element
	cacheSize := 0

	for i := range inputs {
		if cacheSize == poseidon2Rate {
			// Duplex: add cache to state, permute, reset cache.
			performDuplex(&state, &cache)
			cache = [poseidon2Rate]fr.Element{}
			cache[0] = inputs[i]
			cacheSize = 1
		} else {
			cache[cacheSize] = inputs[i]
			cacheSize++
		}
	}

	// Squeeze: final duplex with whatever is in the cache (padded with zeros).
	performDuplex(&state, &cache)

	return state[0], nil
}

// performDuplex adds the cache into the sponge state and applies the
// permutation. This matches Barretenberg's perform_duplex: all rate
// elements are added (unused positions are zero).
func performDuplex(state *[poseidon2T]fr.Element, cache *[poseidon2Rate]fr.Element) {
	for i := 0; i < poseidon2Rate; i++ {
		state[i].Add(&state[i], &cache[i])
	}
	poseidon2Permutation(state)
}

// poseidon2Permutation applies the Poseidon2 permutation in-place on a
// 4-element state. Follows the Barretenberg reference implementation:
//
//  1. matMulExternal(state)                              — initial linear layer
//  2. For r = 0..3 (first half full rounds):
//     addRC → sBoxFull → matMulExternal
//  3. For r = 4..59 (partial rounds):
//     state[0] += rc[r][0] → sBox(state[0]) → matMulInternal
//  4. For r = 60..63 (second half full rounds):
//     addRC → sBoxFull → matMulExternal
func poseidon2Permutation(state *[poseidon2T]fr.Element) {
	// Initial linear layer.
	matMulExternal(state)

	// First half full rounds (rounds 0-3).
	for r := 0; r < 4; r++ {
		addRoundConstants(state, r)
		sBoxFull(state)
		matMulExternal(state)
	}

	// Partial rounds (rounds 4-59).
	for r := 4; r < 60; r++ {
		state[0].Add(&state[0], &poseidon2RC[r][0])
		sBox(&state[0])
		matMulInternal(state)
	}

	// Second half full rounds (rounds 60-63).
	for r := 60; r < 64; r++ {
		addRoundConstants(state, r)
		sBoxFull(state)
		matMulExternal(state)
	}
}

// addRoundConstants adds the round constants for round r to the state.
func addRoundConstants(state *[poseidon2T]fr.Element, r int) {
	for i := 0; i < poseidon2T; i++ {
		state[i].Add(&state[i], &poseidon2RC[r][i])
	}
}

// sBox computes x^5 in-place (the Poseidon2 S-box for BN254).
func sBox(x *fr.Element) {
	var x2, x4 fr.Element
	x2.Mul(x, x)     // x^2
	x4.Mul(&x2, &x2) // x^4
	x.Mul(x, &x4)    // x^5
}

// sBoxFull applies the S-box to all 4 state elements.
func sBoxFull(state *[poseidon2T]fr.Element) {
	for i := 0; i < poseidon2T; i++ {
		sBox(&state[i])
	}
}

// matMulExternal applies the external 4x4 MDS matrix multiplication.
// Implements the Barretenberg matrix_multiplication_4x4 which realizes
// the matrix [5,7,1,3; 4,6,1,1; 1,3,5,7; 1,1,4,6].
func matMulExternal(state *[poseidon2T]fr.Element) {
	var t0, t1, t2, t3, t4, t5, t6, t7 fr.Element

	t0.Add(&state[0], &state[1]) // t0 = s0 + s1
	t1.Add(&state[2], &state[3]) // t1 = s2 + s3

	t2.Double(&state[1]) // t2 = 2*s1
	t2.Add(&t2, &t1)     // t2 = 2*s1 + s2 + s3

	t3.Double(&state[3]) // t3 = 2*s3
	t3.Add(&t3, &t0)     // t3 = s0 + s1 + 2*s3

	t4.Double(&t1)   // t4 = 2*(s2+s3)
	t4.Double(&t4)   // t4 = 4*(s2+s3)
	t4.Add(&t4, &t3) // t4 = s0 + s1 + 4*s2 + 6*s3

	t5.Double(&t0)   // t5 = 2*(s0+s1)
	t5.Double(&t5)   // t5 = 4*(s0+s1)
	t5.Add(&t5, &t2) // t5 = 4*s0 + 6*s1 + s2 + s3

	t6.Add(&t3, &t5) // t6 = 5*s0 + 7*s1 + s2 + 3*s3
	t7.Add(&t2, &t4) // t7 = s0 + 3*s1 + 5*s2 + 7*s3

	state[0].Set(&t6)
	state[1].Set(&t5)
	state[2].Set(&t7)
	state[3].Set(&t4)
}

// matMulInternal applies the internal matrix multiplication:
//
//	result[i] = (D_i - 1) * x[i] + sum(x)
//
// where D_i - 1 are the diagonal constants from poseidon2Diag.
func matMulInternal(state *[poseidon2T]fr.Element) {
	var sum fr.Element
	sum.Set(&state[0])
	for i := 1; i < poseidon2T; i++ {
		sum.Add(&sum, &state[i])
	}
	for i := 0; i < poseidon2T; i++ {
		state[i].Mul(&state[i], &poseidon2Diag[i])
		state[i].Add(&state[i], &sum)
	}
}

// twoTo64 is the constant 2^64, used for the sponge IV encoding.
var twoTo64 fr.Element

func init() {
	// 2^64 = 18446744073709551616
	twoTo64.SetString("18446744073709551616")
}
