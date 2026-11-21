package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
)

type membershipBatchProvider struct {
	testProvider
	reorganized     *atomic.Bool
	rootReorganized *atomic.Bool
}

func (p membershipBatchProvider) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	h, err := p.testProvider.HeaderByNumber(ctx, n)
	copy := types.CopyHeader(h)
	if n != nil && n.Uint64() == 9 {
		copy.Number = big.NewInt(9)
		copy.Time = 122
		if p.rootReorganized.Load() {
			copy.Extra = []byte{98}
		}
		return copy, err
	}
	if p.reorganized.Load() {
		copy.Extra = []byte{99}
	}
	return copy, err
}

type membershipBatchFixture struct {
	oldRoot, rootReorgAfterProof, rootReorganized atomic.Bool
	client                                        *ProtocolRpcClient
	scope                                         DeploymentContext
	block                                         CanonicalBlock
	inputs                                        []MembershipInput
	status, roots, proofs                         atomic.Int32
	corrupt, reorgAfterProof, reorganized         atomic.Bool
}

func newMembershipBatchFixture(t testing.TB) *membershipBatchFixture {
	t.Helper()
	f := &membershipBatchFixture{}
	h := &types.Header{Number: big.NewInt(10), Time: 123}
	f.block = CanonicalBlock{Number: 10, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
	address := common.HexToAddress("0x1234")
	d := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{address, crypto.Keccak256Hash([]byte{0x60, 0}), 1}}
	f.scope = DeploymentContext{31337, address}
	leaves := []fr.Element{primitives.ElementFromUint64(7), primitives.ElementFromUint64(11)}
	paths := [][]common.Hash{make([]common.Hash, primitives.TreeDepth), make([]common.Hash, primitives.TreeDepth)}
	paths[0][0], paths[1][0] = common.Hash(leaves[1].Bytes()), common.Hash(leaves[0].Bytes())
	root, err := primitives.Poseidon2Hash(leaves[0], leaves[1])
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < primitives.TreeDepth; i++ {
		root, err = primitives.Poseidon2Hash(root, primitives.ElementFromUint64(0))
		if err != nil {
			t.Fatal(err)
		}
	}
	anchor := rootAnchor{TreeID: 1, Root: common.Hash(root.Bytes()), LeafCount: 2, Block: f.block}
	for i, leaf := range leaves {
		f.inputs = append(f.inputs, MembershipInput{TreeID: 1, LeafIndex: uint32(i), Commitment: common.Hash(leaf.Bytes())})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
			t.Error(err)
			return
		}
		var out any
		responseAnchor := anchor
		if f.oldRoot.Load() {
			old := types.CopyHeader(h)
			old.Number = big.NewInt(9)
			old.Time = 122
			responseAnchor.Block = CanonicalBlock{Number: 9, Hash: old.Hash(), ParentHash: old.ParentHash, Timestamp: old.Time}
		}
		switch q.Method {
		case "invisi_getStatus":
			f.status.Add(1)
			out = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": d, "indexedHead": f.block}
		case "invisi_getRoots":
			f.roots.Add(1)
			out = []rootAnchor{responseAnchor}
		case "invisi_getMembership":
			f.proofs.Add(1)
			var index uint32
			_ = json.Unmarshal(q.Params[1], &index)
			path := append([]common.Hash(nil), paths[index]...)
			if f.corrupt.Load() && index == 1 {
				path[1] = common.HexToHash("0x01")
			}
			out = map[string]any{"anchor": responseAnchor, "leafIndex": index, "commitment": f.inputs[index].Commitment, "siblings": path}
			if f.rootReorgAfterProof.Load() && index == 1 {
				f.rootReorganized.Store(true)
			}
			if f.reorgAfterProof.Load() && index == 1 {
				f.reorganized.Store(true)
			}
		default:
			t.Errorf("unexpected method %s", q.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": out})
	}))
	t.Cleanup(server.Close)
	c, err := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &d})
	if err != nil {
		t.Fatal(err)
	}
	f.client = c.WithProvider(membershipBatchProvider{testProvider{h}, &f.reorganized, &f.rootReorganized})
	return f
}

func TestMembershipBatchSharesEvidenceAndKeepsInputOrder(t *testing.T) {
	f := newMembershipBatchFixture(t)
	for range 2 {
		got, err := f.client.GetMemberships(context.Background(), f.scope, f.inputs, f.block)
		if err != nil || len(got) != 2 {
			t.Fatal(got, err)
		}
		for i, m := range got {
			if m.Leaf != f.inputs[i].Commitment || m.LeafIndex != uint32(i) || m.Block != f.block {
				t.Fatal("membership order or anchor changed")
			}
		}
	}
	if f.status.Load() != 2 || f.roots.Load() != 2 || f.proofs.Load() != 4 {
		t.Fatalf("batch did not share evidence or reused it across calls: status=%d roots=%d proofs=%d", f.status.Load(), f.roots.Load(), f.proofs.Load())
	}
}

func TestMembershipBatchRejectsTamperedMemberAndReorganizationWithoutPartialResults(t *testing.T) {
	for _, kind := range []string{"path", "reorg"} {
		t.Run(kind, func(t *testing.T) {
			f := newMembershipBatchFixture(t)
			f.corrupt.Store(kind == "path")
			f.reorgAfterProof.Store(kind == "reorg")
			got, err := f.client.GetMemberships(context.Background(), f.scope, f.inputs, f.block)
			if err == nil || len(got) != 0 {
				t.Fatalf("partial/unverified results escaped: %v %v", got, err)
			}
		})
	}
}

func TestMembershipBatchRejectsConflictingRootsAndOversizedOrCanceledWork(t *testing.T) {
	f := newMembershipBatchFixture(t)
	a, b := common.HexToHash("0x01"), common.HexToHash("0x02")
	inputs := append([]MembershipInput(nil), f.inputs...)
	inputs[0].Root, inputs[1].Root = &a, &b
	if _, err := f.client.GetMemberships(context.Background(), f.scope, inputs, f.block); err == nil {
		t.Fatal("conflicting roots accepted")
	}
	if _, err := f.client.GetMemberships(context.Background(), f.scope, make([]MembershipInput, 257), f.block); err == nil {
		t.Fatal("unbounded batch accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := f.client.GetMemberships(ctx, f.scope, f.inputs, f.block); err == nil {
		t.Fatal("canceled batch accepted")
	}
}

func BenchmarkMembershipBatch(b *testing.B) {
	for _, batch := range []bool{false, true} {
		name := "individual"
		if batch {
			name = "batch"
		}
		b.Run(name, func(b *testing.B) {
			f := newMembershipBatchFixture(b)
			b.ResetTimer()
			for range b.N {
				if batch {
					if _, err := f.client.GetMemberships(context.Background(), f.scope, f.inputs, f.block); err != nil {
						b.Fatal(err)
					}
				} else {
					for _, input := range f.inputs {
						if _, err := f.client.GetMembership(context.Background(), f.scope, input, f.block); err != nil {
							b.Fatal(err)
						}
					}
				}
			}
			b.ReportMetric(float64(f.status.Load())/float64(b.N), "status/op")
			b.ReportMetric(float64(f.roots.Load())/float64(b.N), "roots/op")
		})
	}
}

func TestMembershipBatchEndFencesOlderRootAnchor(t *testing.T) {
	f := newMembershipBatchFixture(t)
	f.oldRoot.Store(true)
	f.rootReorgAfterProof.Store(true)
	got, err := f.client.GetMemberships(context.Background(), f.scope, f.inputs, f.block)
	if err == nil || len(got) != 0 {
		t.Fatal("older root anchor changed while fixed target stayed unchanged", err)
	}
}
