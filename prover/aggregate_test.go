package prover

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

type aggregateBackend struct{ calls int }

func (b *aggregateBackend) Prove(_ context.Context, r ProvingRequest, _ Options) (ProofBundle, error) {
	b.calls++
	proof := make([]byte, 32)
	proof[31] = 1
	hash := sha256.Sum256(proof)
	return ProofBundle{FormatVersion: 1, CircuitName: r.CircuitName, Target: r.Target, BBVersion: "6.0.0", ArtifactSHA256: r.ArtifactSHA256, VKSHA256: r.VKSHA256, Proof: proof, ProofSHA256: hex.EncodeToString(hash[:]), PublicInputs: append([]common.Hash(nil), r.ExpectedPublicInputs...)}, nil
}
func (b *aggregateBackend) Verify(context.Context, ProofBundle, ProofStatement, Options) (bool, error) {
	return true, nil
}
func TestAggregateAssemblesAuthenticatedRecursiveSchemaAndRejectsDuplicateSpends(t *testing.T) {
	raw, e := os.ReadFile("../protocol/protocol-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Fixtures struct {
			AggregateStatements struct {
				Members [][]common.Hash `json:"members"`
			} `json:"aggregateStatements"`
		} `json:"fixtures"`
	}
	if e = json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}
	backend := &aggregateBackend{}
	aggregator, e := NewAggregator(backend)
	if e != nil {
		t.Fatal(e)
	}
	members := []AggregateMember{}
	for _, row := range fixture.Fixtures.AggregateStatements.Members[:4] {
		statement, e := Statement("transact_user_slots14", "noir-recursive", row)
		if e != nil {
			t.Fatal(e)
		}
		proof := make([]byte, 458*32)
		hash := sha256.Sum256(proof)
		members = append(members, AggregateMember{Statement: statement, Bundle: ProofBundle{FormatVersion: 1, CircuitName: statement.CircuitName, Target: statement.Target, BBVersion: "6.0.0", ArtifactSHA256: statement.ArtifactSHA256, VKSHA256: statement.VKSHA256, PublicInputs: row, Proof: proof, ProofSHA256: hex.EncodeToString(hash[:])}})
	}
	result, e := aggregator.Aggregate(context.Background(), members)
	if e != nil || len(result.PublicInputs) != 229 || backend.calls != 1 {
		t.Fatal(result.CircuitName, e)
	}
	members[1] = members[0]
	if _, e = aggregator.Aggregate(context.Background(), members); e == nil {
		t.Fatal("duplicate spend aggregated")
	}
	if backend.calls != 1 {
		t.Fatal("proving started before duplicate rejection")
	}
}
