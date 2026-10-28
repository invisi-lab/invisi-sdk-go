// Package prover defines the local proof boundary shared by protocol builders and client operations.
package prover

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

type ProofStatement struct {
	CircuitName, Target, ArtifactSHA256, VKSHA256 string
	ExpectedPublicInputs                          []common.Hash
}
type ProofBundle struct {
	FormatVersion                                            int
	CircuitName, Target, BBVersion, ArtifactSHA256, VKSHA256 string
	Proof                                                    []byte
	ProofSHA256                                              string
	PublicInputs                                             []common.Hash
}
type Progress struct{ Phase string }
type Options struct{ OnProgress func(Progress) }
type LocalWitness struct{ encoded []byte }

func NewLocalWitness(inputs map[string]any) (*LocalWitness, error) {
	raw, e := json.Marshal(inputs)
	if e != nil {
		return nil, e
	}
	return &LocalWitness{raw}, nil
}
func (w *LocalWitness) WithInputs(consumer func(map[string]any) error) error {
	if w == nil {
		return errors.New("local witness required")
	}
	var value map[string]any
	decoder := json.NewDecoder(bytes.NewReader(w.encoded))
	decoder.UseNumber()
	if e := decoder.Decode(&value); e != nil {
		return e
	}
	return consumer(value)
}
func (w *LocalWitness) String() string   { return "LocalWitness([redacted])" }
func (w *LocalWitness) GoString() string { return w.String() }
func (w *LocalWitness) MarshalJSON() ([]byte, error) {
	return []byte(`{"localWitness":"[redacted]"}`), nil
}

type ProvingRequest struct {
	ProofStatement
	Witness *LocalWitness
}
type Prover interface {
	Prove(context.Context, ProvingRequest, Options) (ProofBundle, error)
	Verify(context.Context, ProofBundle, ProofStatement, Options) (bool, error)
}

//go:embed witness-layouts.json
var layoutJSON []byte

type layout struct {
	ArtifactSHA256   string            `json:"artifactSha256"`
	Keys             map[string]string `json:"keys"`
	PublicInputCount int               `json:"publicInputCount"`
	Public           []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
		Array bool   `json:"array"`
	} `json:"public"`
	Private []string `json:"private"`
}

func layouts() map[string]layout {
	var result map[string]layout
	if e := json.Unmarshal(layoutJSON, &result); e != nil {
		panic("invalid embedded witness schemas")
	}
	return result
}
func Statement(name, target string, inputs []common.Hash) (ProofStatement, error) {
	for _, input := range inputs {
		if _, e := core.HexToElement(input.Hex()); e != nil {
			return ProofStatement{}, errors.New("noncanonical public field")
		}
	}
	schema, ok := layouts()[name]
	if !ok || schema.Keys[target] == "" || len(inputs) != schema.PublicInputCount {
		return ProofStatement{}, errors.New("unsupported circuit, target or public statement width")
	}
	return ProofStatement{name, target, schema.ArtifactSHA256, schema.Keys[target], append([]common.Hash(nil), inputs...)}, nil
}
func Request(name, target string, private map[string]any, inputs []common.Hash) (ProvingRequest, error) {
	if strings.HasPrefix(name, "transact_aggregate_") {
		return ProvingRequest{}, errors.New("use Aggregator to construct recursive aggregate witnesses")
	}
	statement, e := Statement(name, target, inputs)
	if e != nil {
		return ProvingRequest{}, e
	}
	schema := layouts()[name]
	if len(private) != len(schema.Private) {
		return ProvingRequest{}, errors.New("private witness layout mismatch")
	}
	all := map[string]any{}
	for _, key := range schema.Private {
		value, ok := private[key]
		if !ok {
			return ProvingRequest{}, fmt.Errorf("missing private witness field %s", key)
		}
		all[key] = value
	}
	at := 0
	for _, field := range schema.Public {
		if field.Array {
			values := []string{}
			for i := 0; i < field.Count; i++ {
				values = append(values, newBigString(inputs[at+i]))
			}
			all[field.Name] = values
		} else {
			all[field.Name] = newBigString(inputs[at])
		}
		at += field.Count
	}
	w, e := NewLocalWitness(all)
	return ProvingRequest{statement, w}, e
}
func AssertProofBundle(bundle ProofBundle, statement ProofStatement) error {
	expected, e := Statement(statement.CircuitName, statement.Target, statement.ExpectedPublicInputs)
	if e != nil {
		return e
	}
	if statement.ArtifactSHA256 != expected.ArtifactSHA256 || statement.VKSHA256 != expected.VKSHA256 || bundle.FormatVersion != 1 || bundle.BBVersion != "6.0.0" || bundle.CircuitName != statement.CircuitName || bundle.Target != statement.Target || bundle.ArtifactSHA256 != statement.ArtifactSHA256 || bundle.VKSHA256 != statement.VKSHA256 || len(bundle.Proof) == 0 || len(bundle.PublicInputs) != len(statement.ExpectedPublicInputs) {
		return errors.New("proof bundle identity mismatch")
	}
	if len(bundle.Proof)%32 != 0 || (bundle.Target == "noir-recursive" && len(bundle.Proof) != 458*32) {
		return errors.New("invalid proof field encoding or recursive width")
	}
	h := sha256.Sum256(bundle.Proof)
	if hex.EncodeToString(h[:]) != bundle.ProofSHA256 {
		return errors.New("proof digest mismatch")
	}
	for i, v := range bundle.PublicInputs {
		if v != statement.ExpectedPublicInputs[i] {
			return errors.New("proof public inputs mismatch")
		}
	}
	return nil
}
func newBigString(hash common.Hash) string { return hash.Big().String() }
