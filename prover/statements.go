package prover

import (
	_ "embed"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

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
		if _, e := primitives.HexToElement(input.Hex()); e != nil {
			return ProofStatement{}, errors.New("noncanonical public field")
		}
	}
	schema, ok := layouts()[name]
	if !ok || schema.Keys[target] == "" || len(inputs) != schema.PublicInputCount {
		return ProofStatement{}, errors.New("unsupported circuit, target or public statement width")
	}
	return ProofStatement{name, target, schema.ArtifactSHA256, schema.Keys[target], append([]common.Hash(nil), inputs...)}, nil
}

func newBigString(hash common.Hash) string { return hash.Big().String() }
