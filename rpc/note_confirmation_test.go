package rpc

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestTrustedNoteRejectsMissingEnvelope(t *testing.T) {
	for _, raw := range []string{`null`, `{}`} {
		if _, err := parseTrustedNote(json.RawMessage(raw), ProtocolDeployment{}, common.HexToHash("0x07"), "safe"); err == nil {
			t.Fatal("accepted", raw)
		}
	}
}

func TestTrustedNoteProofAndBindings(t *testing.T) {
	d := ProtocolDeployment{ChainID: "1", Protocol: ContractDeployment{Address: common.HexToAddress("0x01"), RuntimeCodeHash: common.HexToHash("0x02"), FromBlock: 1}}
	b := CanonicalBlock{Number: 10, Hash: common.HexToHash("0x10"), ParentHash: common.HexToHash("0x09"), Timestamp: 100}
	leaf := common.HexToHash("0x07")
	current, _ := primitives.HexToElement(leaf.Hex())
	path := make([]common.Hash, primitives.TreeDepth)
	for range path {
		current, _ = primitives.Poseidon2Hash(current, primitives.ElementFromUint64(0))
	}
	fixture := func() map[string]any {
		return map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": d, "indexedHead": b, "status": "confirmed", "reason": nil, "output": map[string]any{"block": b, "treeId": 1, "leafIndex": 0, "commitment": leaf, "token": common.HexToAddress("0x03"), "transactionHash": common.HexToHash("0x04"), "logIndex": 0, "kind": "Shield", "ciphertext": "0x", "clue": "0x"}, "membership": map[string]any{"anchor": map[string]any{"treeId": 1, "root": common.Hash(current.Bytes()), "leafCount": 1, "block": b}, "leafIndex": 0, "commitment": leaf, "siblings": path}, "confirmation": map[string]any{"finality": "safe", "verifiedAt": b, "finalityHead": b}}
	}
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"valid", nil}, {"wrong finality", func(m map[string]any) { m["confirmation"].(map[string]any)["finality"] = "finalized" }},
		{"forged proof", func(m map[string]any) {
			p := append([]common.Hash(nil), path...)
			p[0] = common.HexToHash("0x01")
			m["membership"].(map[string]any)["siblings"] = p
		}},
		{"wrong deployment", func(m map[string]any) { dd := d; dd.ChainID = "2"; m["deployment"] = dd }},
		{"conflicting hash", func(m map[string]any) {
			bb := b
			bb.Hash = common.HexToHash("0x99")
			m["confirmation"].(map[string]any)["verifiedAt"] = bb
		}},
		{"pending partial", func(m map[string]any) { m["status"] = "pending"; m["reason"] = "not_observed" }},
		{"bad count", func(m map[string]any) { m["membership"].(map[string]any)["anchor"].(map[string]any)["leafCount"] = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := fixture()
			if tc.mutate != nil {
				tc.mutate(m)
			}
			raw, _ := json.Marshal(m)
			_, err := parseTrustedNote(raw, d, leaf, "safe")
			if (err != nil) != (tc.mutate != nil) {
				t.Fatal(err)
			}
		})
	}
	// The trusted API performs exactly one transport request and requires no provider.
	var response any = fixture()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Method != "invisi_getNote" || len(req.Params) != 2 || string(req.Params[1]) != `{"finality":"safe"}` {
			t.Error("incorrect request", req)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": response})
	}))
	defer server.Close()
	c, err := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &d})
	if err != nil {
		t.Fatal(err)
	}
	scope := DeploymentContext{1, d.Protocol.Address}
	if _, err = c.GetNoteFromTrustedRPC(context.Background(), scope, leaf, "safe"); err != nil {
		t.Fatal(err)
	}
	response = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": d, "indexedHead": nil, "status": "pending", "reason": "pending_index", "output": nil, "membership": nil, "confirmation": nil}
	if got, err := c.GetNoteFromTrustedRPC(context.Background(), scope, leaf, "safe"); err != nil || got.Status != "pending" {
		t.Fatal(got, err)
	}
	if _, err = c.GetNoteFromTrustedRPC(context.Background(), scope, leaf, "latest"); err == nil {
		t.Fatal("accepted latest")
	}
	if calls != 2 {
		t.Fatal("unexpected calls", calls)
	}
	unpinned, _ := NewProtocolRpcClient(server.URL)
	unpinned.deployment = &d
	if _, err = unpinned.GetNoteFromTrustedRPC(context.Background(), scope, leaf, "safe"); err == nil {
		t.Fatal("accepted discovered deployment")
	}
}

func TestTrustedNoteServerFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/note-confirmation.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Deployment   ProtocolDeployment
		Output       struct{ Commitment common.Hash }
		Confirmation struct{ Finality string }
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		if req.Method != "invisi_getNote" || len(req.Params) != 2 {
			t.Error("unexpected request", req)
		} else {
			var selector map[string]string
			json.Unmarshal(req.Params[1], &selector)
			if selector["finality"] != "finalized" || len(selector) != 1 {
				t.Error("incorrect selector", selector)
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": json.RawMessage(raw)})
	}))
	defer server.Close()
	client, err := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &fixture.Deployment})
	if err != nil {
		t.Fatal(err)
	}
	// Every possible chain-provider operation panics; the trusted path must use HTTP only.
	client = client.WithProvider(noTrustedChainAccess{})
	result, err := client.GetNoteFromTrustedRPC(context.Background(), DeploymentContext{31337, fixture.Deployment.Protocol.Address}, fixture.Output.Commitment, fixture.Confirmation.Finality)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "confirmed" || result.Output.Commitment != fixture.Output.Commitment || result.Confirmation.Finality != "finalized" || calls != 1 {
		t.Fatalf("incorrect fixture result: %+v, calls %d", result, calls)
	}
}

type noTrustedChainAccess struct{}

func (noTrustedChainAccess) ChainID(context.Context) (*big.Int, error) {
	panic("unexpected chain access")
}
func (noTrustedChainAccess) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	panic("unexpected chain access")
}
func (noTrustedChainAccess) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	panic("unexpected chain access")
}
func (noTrustedChainAccess) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	panic("unexpected chain access")
}
