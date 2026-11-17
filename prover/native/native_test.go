package native

import (
	"context"
	"encoding/json"
	"os"

	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/prover"
)

func TestPinnedCircuitsMatchAllCurrentStatements(t *testing.T) {
	dir := os.Getenv("INVISI_PROVER_RESOURCE_DIR")
	if dir == "" {
		t.Skip("set INVISI_PROVER_RESOURCE_DIR for full authenticated release integration")
	}
	backend, e := NewNativeProver(NativeOptions{Resources: ResourceOptions{Directory: dir, CacheDir: t.TempDir(), Offline: true}})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("../witness-layouts.json")
	if e != nil {
		t.Fatal(e)
	}
	var manifest map[string]struct {
		PublicInputCount int
		Keys             map[string]string
	}
	if e = json.Unmarshal(raw, &manifest); e != nil {
		t.Fatal(e)
	}
	for name, entry := range manifest {
		for target := range entry.Keys {
			statement, e := prover.Statement(name, target, make([]common.Hash, entry.PublicInputCount))
			if e != nil {
				t.Fatal(name, e)
			}
			if _, _, e = backend.assets(context.Background(), statement); e != nil {
				t.Fatal(name, target, e)
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
