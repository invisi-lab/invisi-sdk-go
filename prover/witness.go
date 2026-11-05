package prover

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

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

func (w *LocalWitness) String() string { return "LocalWitness([redacted])" }

func (w *LocalWitness) GoString() string { return w.String() }

func (w *LocalWitness) MarshalJSON() ([]byte, error) {
	return []byte(`{"localWitness":"[redacted]"}`), nil
}

type ProvingRequest struct {
	ProofStatement
	Witness *LocalWitness
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
