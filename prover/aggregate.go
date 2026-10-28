package prover

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

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
	if len(members) != 4 && len(members) != 8 && len(members) != 12 {
		return result, errors.New("aggregation requires 4, 8 or 12 members")
	}
	if len(options) > 1 {
		return result, errors.New("expected at most one proving option")
	}
	opt := Options{}
	if len(options) == 1 {
		opt = options[0]
	}
	var key struct {
		VerificationKeyFields []string    `json:"verificationKeyFields"`
		KeyHash               string      `json:"keyHash"`
		InnerVKCommitment     common.Hash `json:"innerVkCommitment"`
	}
	if e := json.Unmarshal(recursiveKeyJSON, &key); e != nil {
		return result, e
	}
	proofs, statements := [][]string{}, [][]string{}
	words := []common.Hash{}
	seenN, seenO := map[common.Hash]bool{}, map[common.Hash]bool{}
	for i, member := range members {
		bundle := CloneBundle(member.Bundle)
		statement := CloneStatement(member.Statement)
		if bundle.CircuitName != "transact_user_slots14" || bundle.Target != "noir-recursive" {
			return result, errors.New("aggregation requires recursive slots14 members")
		}
		if e := AssertProofBundle(bundle, statement); e != nil {
			return result, e
		}
		if i > 0 {
			for j := 0; j < 3; j++ {
				if bundle.PublicInputs[j] != words[j] {
					return result, errors.New("aggregate context mismatch")
				}
			}
		}
		n, m := bundle.PublicInputs[3].Big().Uint64(), bundle.PublicInputs[4].Big().Uint64()
		if n+m > 14 {
			return result, errors.New("invalid aggregate member shape")
		}
		for j := uint64(0); j < n+m; j++ {
			v := bundle.PublicInputs[34+j]
			seen := seenO
			if j < n {
				seen = seenN
			}
			if seen[v] {
				return result, errors.New("duplicate aggregate nullifier or commitment")
			}
			seen[v] = true
		}
		valid, e := a.backend.Verify(ctx, CloneBundle(bundle), CloneStatement(statement), opt)
		if e != nil {
			return result, e
		}
		if !valid {
			return result, errors.New("aggregate member verification failed")
		}
		p := []string{}
		for j := 0; j < 458; j++ {
			p = append(p, common.BytesToHash(bundle.Proof[j*32:(j+1)*32]).Hex())
		}
		proofs = append(proofs, p)
		row := []string{}
		for _, v := range bundle.PublicInputs {
			row = append(row, v.Hex())
		}
		statements = append(statements, row)
		words = append(words, bundle.PublicInputs...)
	}
	words = append(words, key.InnerVKCommitment)
	expected, e := Statement(fmt.Sprintf("transact_aggregate_14_%d", len(members)), "evm", words)
	if e != nil {
		return result, e
	}
	witness, e := NewLocalWitness(map[string]any{"verification_key": key.VerificationKeyFields, "key_hash": key.KeyHash, "proofs": proofs, "statements": statements, "inner_vk_commitment": key.InnerVKCommitment.Hex()})
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
