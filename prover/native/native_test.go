package native

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-labs/invisi-sdk-go/prover"
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

func TestPrefetchCombinedCircuitOffline(t *testing.T) {
	directory := os.Getenv("INVISI_PROVER_RESOURCE_DIR")
	if directory == "" {
		t.Skip("set INVISI_PROVER_RESOURCE_DIR to a local directory containing both pinned sets")
	}
	if e := PrefetchResources(context.Background(), ResourceOptions{Directory: directory, CacheDir: t.TempDir(), Offline: true}, []string{"circuits/create_credit_initial_payment.json", "circuits/create_credit_initial_payment.evm.vk"}); e != nil {
		t.Fatal(e)
	}
}

func TestHistoricalCatalogRejectsCurrentCreditStatements(t *testing.T) {
	backend, err := NewNativeProver(NativeOptions{Resources: ResourceOptions{LegacyCombined: true, CacheDir: t.TempDir(), Offline: true}})
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, fixture := range currentFixtures(t) {
		if !strings.Contains(fixture.Name, "credit") {
			continue
		}
		request := fixtureRequest(t, fixture)
		if _, _, err = backend.assets(context.Background(), request.ProofStatement); err == nil || !strings.Contains(err.Error(), "historical catalog") {
			t.Fatalf("%s accepted historical Credit assets: %v", fixture.Name, err)
		}
		checked++
	}
	if checked != 4 {
		t.Fatalf("expected four Credit statement families, got %d", checked)
	}
}
