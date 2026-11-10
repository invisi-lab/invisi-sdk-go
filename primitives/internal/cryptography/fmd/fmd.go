package fmd

// Frozen S-FMD profile: decaf377-fmd 2.1.1, 24 precision bits and 68-byte clues.
// Detection keys are independent of transaction ownership. This new composition
// is unaudited and is not constant-time: gnark Edwards multiplication, big.Int,
// and early rejection are variable-time. Go does not guarantee secret erasure.
import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/twistededwards"
	"github.com/ethereum/go-ethereum/crypto/blake2b"
)

const fmdPrecision = 24

// fmdHash initializes standard sequential BLAKE2b-512 with a 16-byte padded
// personalization. All frozen FMD inputs fit a single 128-byte block (at most
// 96 bytes). Compression uses geth's existing RFC 7693/EIP-152 implementation.
func fmdHash(label string, parts ...[]byte) [64]byte {
	h := [8]uint64{0x6a09e667f3bcc908, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1, 0x510e527fade682d1, 0x9b05688c2b3e6c1f, 0x1f83d9abfb41bd6b, 0x5be0cd19137e2179}
	h[0] ^= 0x01010040
	var personal [16]byte
	copy(personal[:], "decaf377-fmd."+label)
	h[6] ^= binary.LittleEndian.Uint64(personal[:8])
	h[7] ^= binary.LittleEndian.Uint64(personal[8:])
	var block [128]byte
	n := 0
	for _, part := range parts {
		if n+len(part) > len(block) {
			panic("internal FMD hash input exceeds one block")
		}
		n += copy(block[n:], part)
	}
	var m [16]uint64
	for i := range m {
		m[i] = binary.LittleEndian.Uint64(block[i*8 : (i+1)*8])
	}
	blake2b.F(&h, m, [2]uint64{uint64(n), 0}, true, 12)
	var out [64]byte
	for i, v := range h {
		binary.LittleEndian.PutUint64(out[i*8:], v)
	}
	return out
}
func fmdHashScalar(label string, parts ...[]byte) *big.Int {
	h := fmdHash(label, parts...)
	return new(big.Int).Mod(decafReadLE(h[:]), decafScalarOrder)
}
func fmdChildOffset(clueKey []byte, index int) *big.Int {
	return fmdHashScalar("hkd", clueKey, []byte{byte(index)})
}
func fmdBit(p, shared, q []byte) byte { h := fmdHash("bit", p, shared, q); return h[0] & 1 }
func fmdDetectionScalar(key []byte) (*big.Int, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("detection key must contain exactly 32 bytes")
	}
	n := decafReadLE(key)
	if n.Sign() == 0 || n.Cmp(decafScalarOrder) >= 0 {
		return nil, fmt.Errorf("detection key must be a canonical nonzero Decaf377 scalar")
	}
	return n, nil
}

// DeriveClueKey derives the canonical clue public key from a 32-byte LE scalar.
func DeriveClueKey(key []byte) ([]byte, error) {
	n, e := fmdDetectionScalar(key)
	if e != nil {
		return nil, e
	}
	p := decafMultiply(&decafGenerator, n)
	out, e := decafEncode(&p)
	if e != nil {
		return nil, e
	}
	return out[:], nil
}
func ValidateClueKey(key []byte) error {
	p, e := decafDecode(key)
	if e != nil {
		return e
	}
	if p.X.IsZero() {
		return fmt.Errorf("clue key must not be identity")
	}
	return nil
}

// CreateClue always samples a fresh seed with crypto/rand and encodes 24 bits.
func CreateClue(key []byte) ([]byte, error) {
	if e := ValidateClueKey(key); e != nil {
		return nil, e
	}
	seed := make([]byte, 32)
	if _, e := rand.Read(seed); e != nil {
		return nil, fmt.Errorf("FMD secure randomness: %w", e)
	}
	return CreateClueFromSeed(key, seed)
}

// CreateClueFromSeed is the deterministic interoperability/backend entry point.
// Production seeds must be fresh, independent, CSPRNG-generated 32-byte values.
func CreateClueFromSeed(key, seed []byte) ([]byte, error) {
	if e := ValidateClueKey(key); e != nil {
		return nil, e
	}
	if len(seed) != 32 {
		return nil, fmt.Errorf("clue seed must contain exactly 32 bytes")
	}
	root, _ := decafDecode(key)
	r, z := fmdHashScalar("rdv", key, seed), fmdHashScalar("zdv", key, seed)
	if r.Sign() == 0 {
		return nil, fmt.Errorf("degenerate clue seed")
	}
	p, q := decafMultiply(&decafGenerator, r), decafMultiply(&decafGenerator, z)
	pBytes, e := decafEncode(&p)
	if e != nil {
		return nil, e
	}
	qBytes, e := decafEncode(&q)
	if e != nil {
		return nil, e
	}
	var ciphertext [3]byte
	for i := 0; i < fmdPrecision; i++ {
		offset := decafMultiply(&decafGenerator, fmdChildOffset(key, i))
		var child twistededwards.PointAffine
		child.Add(&root, &offset)
		shared := decafMultiply(&child, r)
		encoded, e := decafEncode(&shared)
		if e != nil {
			return nil, e
		}
		ciphertext[i>>3] |= (fmdBit(pBytes[:], encoded[:], qBytes[:]) ^ 1) << uint(i&7)
	}
	m := fmdHashScalar("sca", pBytes[:], []byte{fmdPrecision}, ciphertext[:])
	y := new(big.Int).Sub(z, m)
	y.Mul(y, new(big.Int).ModInverse(r, decafScalarOrder))
	y.Mod(y, decafScalarOrder)
	if y.Sign() == 0 {
		return nil, fmt.Errorf("degenerate clue seed")
	}
	yb := decafWriteLE(y)
	clue := make([]byte, 68)
	copy(clue, pBytes[:])
	copy(clue[32:64], yb[:])
	clue[64] = fmdPrecision
	copy(clue[65:], ciphertext[:])
	return clue, nil
}

// PreparedClue owns copied immutable data. Invalid cryptographic encodings retain
// a fail-closed state; malformed byte lengths are errors. Fields stay unexported.
type PreparedClue struct {
	valid          bool
	p              twistededwards.PointAffine
	pBytes, qBytes [32]byte
	ciphertext     [3]byte
}

func PrepareClue(clue []byte) (*PreparedClue, error) {
	if len(clue) != 68 {
		return nil, fmt.Errorf("clue must contain exactly 68 bytes")
	}
	out := new(PreparedClue)
	if clue[64] != fmdPrecision {
		return out, nil
	}
	p, e := decafDecode(clue[:32])
	if e != nil || p.X.IsZero() {
		return out, nil
	}
	y := decafReadLE(clue[32:64])
	if y.Sign() == 0 || y.Cmp(decafScalarOrder) >= 0 {
		return out, nil
	}
	copy(out.pBytes[:], clue[:32])
	copy(out.ciphertext[:], clue[65:68])
	m := fmdHashScalar("sca", out.pBytes[:], clue[64:65], out.ciphertext[:])
	yp, mg := decafMultiply(&p, y), decafMultiply(&decafGenerator, m)
	var q twistededwards.PointAffine
	q.Add(&yp, &mg)
	out.qBytes, e = decafEncode(&q)
	if e != nil {
		return out, nil
	}
	out.p = p
	out.valid = true
	return out, nil
}
func DetectClue(key, clue []byte, checkedBits int) (bool, error) {
	if checkedBits < 1 || checkedBits > fmdPrecision {
		return false, fmt.Errorf("checkedBits must be an integer from 1 to 24")
	}
	if _, e := fmdDetectionScalar(key); e != nil {
		return false, e
	}
	prepared, e := PrepareClue(clue)
	if e != nil {
		return false, e
	}
	return DetectPreparedClue(key, prepared, checkedBits)
}
func DetectPreparedClue(key []byte, prepared *PreparedClue, checkedBits int) (bool, error) {
	if checkedBits < 1 || checkedBits > fmdPrecision {
		return false, fmt.Errorf("checkedBits must be an integer from 1 to 24")
	}
	scalar, e := fmdDetectionScalar(key)
	if e != nil {
		return false, e
	}
	if prepared == nil {
		return false, fmt.Errorf("prepared clue is nil")
	}
	if !prepared.valid {
		return false, nil
	}
	clueKey, e := DeriveClueKey(key)
	if e != nil {
		return false, e
	}
	for i := 0; i < checkedBits; i++ {
		child := new(big.Int).Add(scalar, fmdChildOffset(clueKey, i))
		child.Mod(child, decafScalarOrder)
		shared := decafMultiply(&prepared.p, child)
		encoded, e := decafEncode(&shared)
		if e != nil {
			return false, nil
		}
		ciphertextBit := (prepared.ciphertext[i>>3] >> uint(i&7)) & 1
		if ciphertextBit^fmdBit(prepared.pBytes[:], encoded[:], prepared.qBytes[:]) == 0 {
			return false, nil
		}
	}
	return true, nil
}

// IsCanonicalDetectionKey checks the scalar boundary without deriving a public key.
func IsCanonicalDetectionKey(key []byte) bool {
	_, err := fmdDetectionScalar(key)
	return err == nil
}
