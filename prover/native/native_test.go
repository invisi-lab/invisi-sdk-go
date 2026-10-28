package native

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

func TestBundledCircuitsMatchCurrentTSManifest(t *testing.T) {
	raw, e := resources.ReadFile("circuits/manifest.json")
	if e != nil {
		t.Fatal(e)
	}
	var manifest struct {
		Entries []struct {
			Name             string
			PublicInputCount int
			Keys             map[string]json.RawMessage
		}
	}
	if e = json.Unmarshal(raw, &manifest); e != nil {
		t.Fatal(e)
	}
	for _, entry := range manifest.Entries {
		for target := range entry.Keys {
			statement, e := prover.Statement(entry.Name, target, make([]common.Hash, entry.PublicInputCount))
			if e != nil {
				t.Fatal(entry.Name, e)
			}
			if _, _, e = assets(statement); e != nil {
				t.Fatal(entry.Name, target, e)
			}
		}
	}
}
func TestNativeProverConfiguresOfficialWitnessExecutorByDefault(t *testing.T) {
	backend, e := NewNativeProver(NativeOptions{})
	if e != nil {
		t.Fatal(e)
	}
	executor, ok := backend.options.WitnessExecutor.(noirWitnessExecutor)
	if !ok || executor.binary != "noir-execute" {
		t.Fatal("default local executor missing")
	}
}
