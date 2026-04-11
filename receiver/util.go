package receiver

import "math/big"

// bn254FieldModulus is the BN254 scalar field modulus.
var bn254FieldModulus, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10,
)

// StringIdToBigInt converts a string identifier (such as an invoice ID or
// refund ID) to a *big.Int suitable for use as a field element.
//
// Algorithm (matching TS SDK):
//  1. Encode string as UTF-8 bytes
//  2. Interpret as big-endian integer: for each byte, result = (result << 8) | byte
//  3. Reduce modulo BN254 scalar field modulus
func StringIdToBigInt(id string) *big.Int {
	result := new(big.Int)
	for i := 0; i < len(id); i++ {
		result.Lsh(result, 8)
		result.Or(result, big.NewInt(int64(id[i])))
	}
	result.Mod(result, bn254FieldModulus)
	return result
}
