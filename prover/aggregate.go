package prover

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

//go:embed recursive-key.json
var recursiveKeyJSON []byte

type AggregateMember struct {
	Bundle    ProofBundle
	Statement ProofStatement
}
type Aggregator struct{ backend Prover }

func NewAggregator(backend Prover) (*Aggregator, error) {
	if backend == nil {
		return nil, errors.New("local prover required")
	}
	return &Aggregator{backend}, nil
}
func CloneStatement(statement ProofStatement) ProofStatement {
	statement.ExpectedPublicInputs = append([]common.Hash(nil), statement.ExpectedPublicInputs...)
	return statement
}
func CloneBundle(bundle ProofBundle) ProofBundle {
	bundle.Proof = append([]byte(nil), bundle.Proof...)
	bundle.PublicInputs = append([]common.Hash(nil), bundle.PublicInputs...)
	return bundle
}
func (a *Aggregator) Aggregate(ctx context.Context, members []AggregateMember, options ...Options) (ProofBundle, error) {
	var result ProofBundle
	if len(members) < 1 || len(members) > 5 {
		return result, errors.New("aggregation requires one through five members")
	}
	if len(options) > 1 {
		return result, errors.New("expected at most one proving option")
	}
	opt := Options{}
	if len(options) == 1 {
		opt = options[0]
	}
	frozen := make([]AggregateMember, len(members))
	for i, member := range members {
		frozen[i] = AggregateMember{Bundle: CloneBundle(member.Bundle), Statement: CloneStatement(member.Statement)}
	}
	members = frozen
	var keys map[string]struct {
		VerificationKeyFields []string    `json:"verificationKeyFields"`
		KeyHash               string      `json:"keyHash"`
		InnerVKCommitment     common.Hash `json:"innerVkCommitment"`
	}
	if e := json.Unmarshal(recursiveKeyJSON, &keys); e != nil {
		return result, e
	}
	proofs, statements := [][]string{}, [][]string{}
	capacities := []int{}
	seenN, seenO := map[common.Hash]bool{}, map[common.Hash]bool{}
	for i, member := range members {
		bundle, statement := member.Bundle, member.Statement
		if e := AssertProofBundle(bundle, statement); e != nil {
			return result, e
		}
		capacity := 0
		for c := 1; c <= 10; c++ {
			if bundle.CircuitName == fmt.Sprintf("transact_user_slots%d", c) {
				capacity = c
				break
			}
		}
		if capacity == 0 || bundle.Target != "noir-recursive" {
			return result, errors.New("aggregation requires native C1 through C10 recursive members")
		}
		if e := validateAggregateUser(bundle.PublicInputs, capacity); e != nil {
			return result, e
		}
		if i > 0 {
			for j := 0; j < 2; j++ {
				if bundle.PublicInputs[j] != members[0].Bundle.PublicInputs[j] {
					return result, errors.New("aggregate chain or protocol context mismatch")
				}
			}
		}
		n, m := int(bundle.PublicInputs[3].Big().Uint64()), int(bundle.PublicInputs[4].Big().Uint64())
		for j := 0; j < n+m; j++ {
			value := bundle.PublicInputs[6+2*capacity+j]
			seen := seenO
			if j < n {
				seen = seenN
			}
			if value == (common.Hash{}) || seen[value] {
				return result, errors.New("zero or duplicate aggregate nullifier or commitment")
			}
			seen[value] = true
		}
		proof := []string{}
		for j := 0; j < 458; j++ {
			proof = append(proof, common.BytesToHash(bundle.Proof[j*32:(j+1)*32]).Hex())
		}
		proofs = append(proofs, proof)
		row := []string{}
		for _, word := range bundle.PublicInputs {
			row = append(row, word.Hex())
		}
		statements = append(statements, row)
		capacities = append(capacities, capacity)
	}
	homogeneous, small := true, true
	for _, c := range capacities {
		homogeneous = homogeneous && c == capacities[0]
		small = small && c <= 5
	}
	dedicated := homogeneous && (len(members) == 1 || len(members) == 2 || len(members) == 5)
	if !dedicated && (!small || len(members) < 2) {
		return result, errors.New("unsupported aggregate capacity/count combination")
	}
	for _, member := range members {
		bundle, statement := member.Bundle, member.Statement
		valid, e := a.backend.Verify(ctx, CloneBundle(bundle), CloneStatement(statement), opt)
		if e != nil {
			return result, e
		}
		if !valid {
			return result, errors.New("aggregate member verification failed")
		}
	}
	words := []common.Hash{}
	values := map[string]any{"proofs": proofs, "statements": statements}
	name := fmt.Sprintf("transact_aggregate_%d_%d", capacities[0], len(members))
	if dedicated {
		key := keys[fmt.Sprint(capacities[0])]
		if len(key.VerificationKeyFields) != 115 {
			return result, errors.New("missing trusted recursive key")
		}
		for _, member := range members {
			words = append(words, member.Bundle.PublicInputs...)
		}
		words = append(words, key.InnerVKCommitment)
		values["verification_key"] = key.VerificationKeyFields
		values["key_hash"] = key.KeyHash
		values["inner_vk_commitment"] = key.InnerVKCommitment.Hex()
	} else {
		name = fmt.Sprintf("transact_aggregate_mixed_1_5_%d", len(members))
		tags := []string{}
		for _, c := range capacities {
			tag := common.BigToHash(big.NewInt(int64(c)))
			words = append(words, tag)
			tags = append(tags, fmt.Sprint(c))
		}
		for i, row := range statements {
			for len(row) < 32 {
				row = append(row, common.Hash{}.Hex())
			}
			statements[i] = row
			for _, value := range row {
				words = append(words, common.HexToHash(value))
			}
		}
		allKeys, hashes, pins := [][]string{}, []string{}, []string{}
		for c := 1; c <= 5; c++ {
			key := keys[fmt.Sprint(c)]
			if len(key.VerificationKeyFields) != 115 {
				return result, errors.New("missing trusted mixed recursive key")
			}
			allKeys = append(allKeys, key.VerificationKeyFields)
			hashes = append(hashes, key.KeyHash)
			pins = append(pins, key.InnerVKCommitment.Hex())
			words = append(words, key.InnerVKCommitment)
		}
		values["statements"] = statements
		values["capacities"] = tags
		values["verification_keys"] = allKeys
		values["key_hashes"] = hashes
		values["inner_vk_commitments"] = pins
	}
	expected, e := Statement(name, "evm", words)
	if e != nil {
		return result, e
	}
	witness, e := NewLocalWitness(values)
	if e != nil {
		return result, e
	}
	bundle, e := a.backend.Prove(ctx, ProvingRequest{CloneStatement(expected), witness}, opt)
	if e != nil {
		return result, e
	}
	bundle = CloneBundle(bundle)
	if e = AssertProofBundle(bundle, expected); e != nil {
		return result, e
	}
	valid, e := a.backend.Verify(ctx, CloneBundle(bundle), CloneStatement(expected), opt)
	if e != nil {
		return result, e
	}
	if !valid {
		return result, errors.New("aggregate proof verification failed")
	}
	return bundle, nil
}

// Canonical native capacity and complete EIP712 digest must match the same public statement.
func validateAggregateUser(inputs []common.Hash, capacity int) error {
	if len(inputs) != 17+3*capacity {
		return errors.New("invalid native member width")
	}
	n, m := inputs[3].Big(), inputs[4].Big()
	if !n.IsUint64() || !m.IsUint64() || n.Uint64() > uint64(capacity) || m.Uint64() > uint64(capacity) || n.Uint64()+m.Uint64() != uint64(capacity) {
		return errors.New("noncanonical aggregate member capacity")
	}
	request := core.WalletAuthorizationRequest{Context: core.ProtocolContext{ChainID: core.BigIntToElement(inputs[0].Big()), ProtocolAddress: core.BigIntToElement(inputs[1].Big())}, Operation: core.WalletTransact, Capacity: capacity}
	for _, word := range inputs[:len(inputs)-2] {
		request.Words = append(request.Words, core.BigIntToElement(word.Big()))
	}
	digest, err := core.ComputeWalletAuthorizationDigest(request)
	if err != nil {
		return err
	}
	high, low := new(big.Int).SetBytes(digest[:16]), new(big.Int).SetBytes(digest[16:])
	if inputs[len(inputs)-2].Big().Cmp(high) != 0 || inputs[len(inputs)-1].Big().Cmp(low) != 0 {
		return errors.New("aggregate member authorization digest mismatch")
	}
	return nil
}
