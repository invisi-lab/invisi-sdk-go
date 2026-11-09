package native

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"os"
	"strings"
	"testing"
	"time"
)

type publicFixture struct {
	Label, Name, Target string
	Witness             map[string]any
	PublicInputs        []common.Hash
}

func currentFixtures(t *testing.T) []publicFixture {
	t.Helper()
	raw, err := os.ReadFile("testdata/current-proving.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct{ Fixtures []publicFixture }
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	return data.Fixtures
}
func fixtureRequest(t *testing.T, f publicFixture) prover.ProvingRequest {
	t.Helper()
	s, err := prover.Statement(f.Name, f.Target, f.PublicInputs)
	if err != nil {
		t.Fatal(err)
	}
	w, err := prover.NewLocalWitness(f.Witness)
	if err != nil {
		t.Fatal(err)
	}
	return prover.ProvingRequest{ProofStatement: s, Witness: w}
}
func TestRealOfficialWitnessExecutor(t *testing.T) {
	binary := os.Getenv("NOIR_EXECUTE_PATH")
	if binary == "" {
		t.Skip("set NOIR_EXECUTE_PATH to Noir1.0.0 noir-execute")
	}
	scratch := t.TempDir()
	t.Setenv("TMPDIR", scratch)
	request := fixtureRequest(t, currentFixtures(t)[0])
	artifact, _, err := assets(request.ProofStatement)
	if err != nil {
		t.Fatal(err)
	}
	executor := noirWitnessExecutor{binary: binary, timeout: time.Minute}
	witness, err := executor.Execute(context.Background(), artifact, request.Witness)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(witness)
	if !bytes.HasPrefix(witness, []byte{0x1f, 0x8b}) {
		t.Fatal("not a compressed ACVM witness")
	}
	var invalid *prover.LocalWitness
	err = request.Witness.WithInputs(func(inputs map[string]any) error {
		inputs["signature"] = [64]byte{}
		invalid, err = prover.NewLocalWitness(inputs)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = executor.Execute(context.Background(), artifact, invalid); err == nil {
		t.Fatal("invalid signature accepted")
	}
	if strings.Contains(err.Error(), "public_key") || strings.Contains(err.Error(), "signature =") {
		t.Fatal("private diagnostics leaked")
	}
	leftovers, err := os.ReadDir(scratch)
	if err != nil || len(leftovers) != 0 {
		t.Fatal("temporary private witness files retained", err)
	}
}
func TestRealNativeCurrentCreditNoteAndMixedAggregation(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine proving")
	}
	backend, err := NewNativeProver(NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
	if err != nil {
		t.Fatal(err)
	}
	members := []prover.AggregateMember{}
	for _, fixture := range currentFixtures(t) {
		t.Run(fixture.Label, func(t *testing.T) {
			request := fixtureRequest(t, fixture)
			phases := []string{}
			bundle, err := backend.Prove(context.Background(), request, prover.Options{OnProgress: func(p prover.Progress) { phases = append(phases, p.Phase) }})
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(phases, ",") != "witness,proving,verifying,complete" {
				t.Fatal(phases)
			}
			valid, err := backend.Verify(context.Background(), bundle, request.ProofStatement, prover.Options{})
			if err != nil || !valid {
				t.Fatal("independent verification failed", err)
			}
			if fixture.Target == "noir-recursive" {
				members = append(members, prover.AggregateMember{Bundle: bundle, Statement: request.ProofStatement})
			}
			changed := prover.CloneBundle(bundle)
			changed.PublicInputs[0] = common.Hash{}
			if _, err = backend.Verify(context.Background(), changed, request.ProofStatement, prover.Options{}); err == nil {
				t.Fatal("modified public input accepted")
			}
		})
	}
	if len(members) != 3 {
		t.Fatal("recursive fixtures failed")
	}
	aggregator, err := prover.NewAggregator(backend)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		Name    string
		Members []prover.AggregateMember
		Route   string
	}{{"dedicated", members[:2], "transact_aggregate_1_2"}, {"native-mixed", []prover.AggregateMember{members[0], members[2]}, "transact_aggregate_mixed_1_5_2"}} {
		t.Run(test.Name, func(t *testing.T) {
			bundle, err := aggregator.Aggregate(context.Background(), test.Members)
			if err != nil {
				t.Fatal(err)
			}
			if bundle.CircuitName != test.Route {
				t.Fatal("wrong route", bundle.CircuitName)
			}
		})
	}
}
