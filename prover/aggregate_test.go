package prover

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

type aggregateBackend struct {
	calls    int
	onVerify func()
}

func (b *aggregateBackend) Prove(_ context.Context, r ProvingRequest, _ Options) (ProofBundle, error) {
	b.calls++
	proof := make([]byte, 32)
	proof[31] = 1
	hash := sha256.Sum256(proof)
	return ProofBundle{FormatVersion: 1, CircuitName: r.CircuitName, Target: r.Target, BBVersion: "6.0.0", ArtifactSHA256: r.ArtifactSHA256, VKSHA256: r.VKSHA256, Proof: proof, ProofSHA256: hex.EncodeToString(hash[:]), PublicInputs: append([]common.Hash(nil), r.ExpectedPublicInputs...)}, nil
}
func (b *aggregateBackend) Verify(context.Context, ProofBundle, ProofStatement, Options) (bool, error) {
	if b.onVerify != nil {
		f := b.onVerify
		b.onVerify = nil
		f()
	}
	return true, nil
}

// Public schema vectors use a verifying double; native tests exercise actual proofs.
func aggregateMember(t *testing.T, capacity, id, token int) AggregateMember {
	t.Helper()
	row := make([]common.Hash, 17+3*capacity)
	put := func(at, value int) { row[at] = common.BigToHash(big.NewInt(int64(value))) }
	put(0, 56)
	put(1, 17)
	put(2, token)
	put(3, 1)
	put(4, capacity-1)
	put(5, 1)
	put(6, 1)
	put(6+capacity, 7)
	for i := 0; i < capacity; i++ {
		put(6+2*capacity+i, id+i)
	}
	put(10+3*capacity, 123)
	if capacity == 1 {
		put(8+3*capacity, 17)
		put(9+3*capacity, 1)
	}
	request := core.WalletAuthorizationRequest{Context: core.ProtocolContext{ChainID: core.ElementFromUint64(56), ProtocolAddress: core.ElementFromUint64(17)}, Operation: core.WalletTransact, Capacity: capacity}
	for _, word := range row[:len(row)-2] {
		request.Words = append(request.Words, core.BigIntToElement(word.Big()))
	}
	digest, err := core.ComputeWalletAuthorizationDigest(request)
	if err != nil {
		t.Fatal(err)
	}
	row[len(row)-2] = common.BytesToHash(digest[:16])
	row[len(row)-1] = common.BytesToHash(digest[16:])
	statement, err := Statement(fmt.Sprintf("transact_user_slots%d", capacity), "noir-recursive", row)
	if err != nil {
		t.Fatal(err)
	}
	proof := make([]byte, 458*32)
	hash := sha256.Sum256(proof)
	return AggregateMember{statementBundle(statement, proof, hash), statement}
}
func statementBundle(s ProofStatement, proof []byte, hash [32]byte) ProofBundle {
	return ProofBundle{FormatVersion: 1, CircuitName: s.CircuitName, Target: s.Target, BBVersion: "6.0.0", ArtifactSHA256: s.ArtifactSHA256, VKSHA256: s.VKSHA256, Proof: proof, ProofSHA256: hex.EncodeToString(hash[:]), PublicInputs: append([]common.Hash(nil), s.ExpectedPublicInputs...)}
}
func TestAllThirtyFourAggregateRoutesAndMixedAssets(t *testing.T) {
	backend := &aggregateBackend{}
	aggregator, _ := NewAggregator(backend)
	for c := 1; c <= 10; c++ {
		for _, n := range []int{1, 2, 5} {
			members := []AggregateMember{}
			for i := 0; i < n; i++ {
				members = append(members, aggregateMember(t, c, 1000+i*100, 34+i))
			}
			got, err := aggregator.Aggregate(context.Background(), members)
			if err != nil || got.CircuitName != fmt.Sprintf("transact_aggregate_%d_%d", c, n) || len(got.PublicInputs) != (17+3*c)*n+1 {
				t.Fatalf("C%d N%d: %v", c, n, err)
			}
		}
	}
	for _, n := range []int{2, 3, 4, 5} {
		members := []AggregateMember{}
		for i := 0; i < n; i++ {
			members = append(members, aggregateMember(t, i+1, 2000+i*100, 34+i))
		}
		got, err := aggregator.Aggregate(context.Background(), members)
		if err != nil || len(got.PublicInputs) != 33*n+5 || got.CircuitName != fmt.Sprintf("transact_aggregate_mixed_1_5_%d", n) {
			t.Fatal(n, err)
		}
		for i := 0; i < n; i++ {
			if got.PublicInputs[i].Big().Int64() != int64(i+1) {
				t.Fatal("capacity tag order")
			}
			for j := 17 + 3*(i+1); j < 32; j++ {
				if got.PublicInputs[n+i*32+j] != (common.Hash{}) {
					t.Fatal("mixed padding")
				}
			}
		}
	}
	if backend.calls != 34 {
		t.Fatal("missing route")
	}
}
func TestAggregateSnapshotsAndRejectsDuplicatesWrongAuthAndUnsupportedRoutes(t *testing.T) {
	backend := &aggregateBackend{}
	aggregator, _ := NewAggregator(backend)
	members := []AggregateMember{aggregateMember(t, 2, 1000, 34), aggregateMember(t, 2, 2000, 35)}
	backend.onVerify = func() { members[1].Bundle.Proof[0] = 255 }
	if _, err := aggregator.Aggregate(context.Background(), members); err != nil {
		t.Fatal(err)
	}
	members[1].Bundle.Proof[0] = 0
	for _, bad := range [][]AggregateMember{{members[0], members[0]}, {aggregateMember(t, 6, 1000, 34), aggregateMember(t, 1, 2000, 34)}, {aggregateMember(t, 6, 1000, 34), aggregateMember(t, 6, 2000, 34), aggregateMember(t, 6, 3000, 34)}} {
		if _, err := aggregator.Aggregate(context.Background(), bad); err == nil {
			t.Fatal("unsupported or duplicate members accepted")
		}
	}
	changed := aggregateMember(t, 2, 4000, 34)
	changed.Bundle.PublicInputs[len(changed.Bundle.PublicInputs)-1] = common.Hash{}
	changed.Statement.ExpectedPublicInputs = append([]common.Hash(nil), changed.Bundle.PublicInputs...)
	if _, err := aggregator.Aggregate(context.Background(), []AggregateMember{changed}); err == nil {
		t.Fatal("authorization mismatch accepted")
	}
	if backend.calls != 1 {
		t.Fatal("proving started before validation")
	}
}
