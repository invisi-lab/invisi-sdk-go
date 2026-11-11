package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestMaxUpdateBlocksExactBoundaryAndEmptyPage(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10000), Time: 123}}
	block := CanonicalBlock{10000, provider.header.Hash(), provider.header.ParentHash, 123}
	deployment := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("0x11"), crypto.Keccak256Hash([]byte{1}), 1}}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		var result any
		if req.Method == "invisi_getStatus" {
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": block}
		} else {
			result = map[string]any{"blockHash": block.Hash, "notes": []any{}, "creditTransitions": []any{}}
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	c, _ := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &deployment})
	c = c.WithProvider(provider)
	scope := DeploymentContext{31337, deployment.Protocol.Address}
	if _, e := c.GetNoteUpdates(context.Background(), scope, ProtocolUpdateRange{FromBlock: 1, ToBlock: 10000}); e != nil {
		t.Fatal(e)
	}
	before := calls
	if _, e := c.GetNoteUpdates(context.Background(), scope, ProtocolUpdateRange{FromBlock: 1, ToBlock: 10001}); e == nil {
		t.Fatal("10001 blocks accepted")
	}
	if calls != before {
		t.Fatal("oversized request reached RPC")
	}
	if _, e := c.GetNoteUpdates(context.Background(), scope, ProtocolUpdateRange{FromBlock: 1, ToBlock: 10000}); e != nil {
		t.Fatal(e)
	}
}
func TestEmptyDeliveryPreservesOriginalLeafPosition(t *testing.T) {
	row := map[string]any{"block": CanonicalBlock{Number: 10}, "treeId": 1, "leafIndex": 7, "commitment": common.HexToHash("0x03"), "token": common.HexToAddress("0x11"), "ciphertext": "0x", "clue": "0x", "transactionHash": common.HexToHash("0x04"), "logIndex": 1, "kind": "Shield"}
	raw, _ := json.Marshal(row)
	out, e := parseNoteOutput(raw)
	if e != nil || out.LeafIndex != 7 || len(out.Ciphertext) != 0 {
		t.Fatal(out, e)
	}
	row["clue"] = "0x" + repeatZero(68)
	raw, _ = json.Marshal(row)
	if _, e = parseNoteOutput(raw); e == nil {
		t.Fatal("unpaired empty delivery accepted")
	}
}

func TestUpdateReadersRejectStaleIndexedHead(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123}}
	block := CanonicalBlock{10, provider.header.Hash(), provider.header.ParentHash, 123}
	deployment := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("0x11"), crypto.Keccak256Hash([]byte{1}), 1}}
	readers := []struct {
		name string
		read func(*ProtocolRpcClient) error
	}{
		{"updates", func(c *ProtocolRpcClient) error {
			_, e := c.GetNoteUpdates(context.Background(), DeploymentContext{31337, deployment.Protocol.Address}, ProtocolUpdateRange{FromBlock: 1, ToBlock: 10})
			return e
		}},
		{"outputs", func(c *ProtocolRpcClient) error {
			_, e := c.GetNoteUpdates(context.Background(), DeploymentContext{31337, deployment.Protocol.Address}, ProtocolUpdateRange{FromBlock: 1, ToBlock: 10})
			return e
		}},
	}
	for _, reader := range readers {
		for _, field := range []string{"hash", "parentHash", "timestamp"} {
			t.Run(reader.name+"/"+field, func(t *testing.T) {
				stale := block
				switch field {
				case "hash":
					stale.Hash = common.HexToHash("0xdead")
				case "parentHash":
					stale.ParentHash = common.HexToHash("0xbeef")
				case "timestamp":
					stale.Timestamp++
				}
				updateCalls := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var req struct {
						ID     uint64
						Method string
					}
					json.NewDecoder(r.Body).Decode(&req)
					var result any
					if req.Method == "invisi_getStatus" {
						result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": stale}
					} else {
						updateCalls++
						result = map[string]any{"blockHash": block.Hash, "notes": []any{}, "creditTransitions": []any{}}
					}
					json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
				}))
				defer server.Close()
				c, e := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &deployment})
				if e != nil {
					t.Fatal(e)
				}
				c = c.WithProvider(provider)
				if e = reader.read(c); e == nil {
					t.Fatal("accepted stale indexed head")
				}
				if updateCalls != 0 {
					t.Fatal("stale indexed head reached updates query")
				}
			})
		}
	}
}
