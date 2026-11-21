package prover

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"math/big"
)

// VerificationKeyLoader is an optional Prover capability bound to its resource policy.
type VerificationKeyLoader interface {
	LoadVerificationKey(context.Context, string, string, Options) ([]byte, error)
}
type RecursiveVerificationKey struct {
	Fields     []string
	Hash       string
	Commitment common.Hash
}
type recursiveKeyPin struct {
	KeyHash           string `json:"keyHash"`
	InnerVKCommitment string `json:"innerVkCommitment"`
}

// LoadRecursiveVerificationKey authenticates the binary, canonical field words and commitment.
// Decoded keys are not retained; each caller receives independent fields.
func LoadRecursiveVerificationKey(ctx context.Context, name string, loader VerificationKeyLoader, options Options) (RecursiveVerificationKey, error) {
	var result RecursiveVerificationKey
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if loader == nil {
		return result, errors.New("recursive aggregation requires a VerificationKeyLoader on the backend or AggregatorOptions")
	}
	var pins map[string]recursiveKeyPin
	if err := json.Unmarshal(recursiveKeyJSON, &pins); err != nil {
		return result, err
	}
	var capacity int
	if _, err := fmt.Sscanf(name, "transact_user_slots%d", &capacity); err != nil || name != fmt.Sprintf("transact_user_slots%d", capacity) {
		return result, errors.New("unsupported recursive circuit")
	}
	pin, ok := pins[fmt.Sprint(capacity)]
	if !ok {
		return result, errors.New("missing trusted recursive key metadata")
	}
	digest, err := VerificationKeySHA256(name, "noir-recursive")
	if err != nil {
		return result, err
	}
	data, err := loader.LoadVerificationKey(ctx, name, "noir-recursive", options)
	if err != nil {
		return result, err
	}
	return authenticateRecursiveKey(ctx, append([]byte(nil), data...), digest, pin)
}
func authenticateRecursiveKey(ctx context.Context, data []byte, digest string, pin recursiveKeyPin) (RecursiveVerificationKey, error) {
	var result RecursiveVerificationKey
	if err := ctx.Err(); err != nil {
		return result, err
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != digest {
		return result, errors.New("recursive verification key checksum mismatch")
	}
	if len(data) != 115*32 {
		return result, errors.New("invalid recursive verification key binary width")
	}
	canonical := func(value string) (fr.Element, error) {
		var field fr.Element
		if len(value) != 66 || value[:2] != "0x" {
			return field, errors.New("noncanonical recursive verification key field")
		}
		raw, err := hex.DecodeString(value[2:])
		if err != nil {
			return field, err
		}
		if err = field.SetBytesCanonical(raw); err != nil {
			return field, errors.New("noncanonical recursive verification key field")
		}
		return field, nil
	}
	hash, err := canonical(pin.KeyHash)
	if err != nil {
		return result, err
	}
	commitment, err := canonical(pin.InnerVKCommitment)
	if err != nil {
		return result, err
	}
	domain, _ := new(big.Int).SetString("494e564953495f5245435552534956455f564b31", 16)
	fields := []fr.Element{primitives.BigIntToElement(domain)}
	for i := 0; i < 115; i++ {
		raw := data[i*32 : (i+1)*32]
		var field fr.Element
		if err = field.SetBytesCanonical(raw); err != nil {
			return RecursiveVerificationKey{}, errors.New("noncanonical recursive verification key field")
		}
		fields = append(fields, field)
		result.Fields = append(result.Fields, "0x"+hex.EncodeToString(raw))
	}
	fields = append(fields, hash)
	actual, err := primitives.Poseidon2Hash(fields...)
	if err != nil {
		return RecursiveVerificationKey{}, err
	}
	if err = ctx.Err(); err != nil {
		return RecursiveVerificationKey{}, err
	}
	if !actual.Equal(&commitment) {
		return RecursiveVerificationKey{}, errors.New("recursive verification key commitment mismatch")
	}
	result.Hash = pin.KeyHash
	result.Commitment = common.HexToHash(pin.InnerVKCommitment)
	return result, nil
}
