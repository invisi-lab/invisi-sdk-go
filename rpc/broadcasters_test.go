package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestBroadcasterDiscoveryPinsPagesAndRegistration(t *testing.T) {
	registry := ContractDeployment{common.HexToAddress("0x11"), crypto.Keccak256Hash([]byte{1}), 2}
	deployment := ProtocolDeployment{ChainID: "97", Protocol: registry, BroadcasterRegistry: &registry}
	block := CanonicalBlock{Number: 10, Hash: common.HexToHash("0x10"), Timestamp: 100}
	enabled := true
	entry := RegisteredBroadcaster{Operator: common.HexToAddress("0x22"), Endpoint: "https://relay.example/rpc", MetadataURI: "", QuoteSigner: common.HexToAddress("0x33"), FeeRecipient: common.HexToAddress("0x44"), Enabled: true, Block: block}
	page := BroadcasterPage{block, []RegisteredBroadcaster{entry}, &entry.Operator}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Method != "invisi_getBroadcasters" || len(req.Params) != 1 {
			t.Error("bad discovery request")
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": page})
	}))
	defer server.Close()
	c, _ := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &deployment})
	scope := DeploymentContext{97, registry.Address}
	got, e := c.GetBroadcasters(context.Background(), scope, BroadcasterQuery{Enabled: &enabled, Limit: 1})
	if e != nil || len(got.Broadcasters) != 1 {
		t.Fatal(got, e)
	}
	if _, e = c.GetBroadcasters(context.Background(), scope, BroadcasterQuery{After: &entry.Operator}); e == nil {
		t.Fatal("unanchored cursor accepted")
	}
	other := block
	other.Hash = common.HexToHash("0xff")
	if _, e = c.GetBroadcasters(context.Background(), scope, BroadcasterQuery{Block: &other}); e == nil {
		t.Fatal("different page anchor accepted")
	}
	page.Broadcasters[0].QuoteSigner = common.Address{}
	if _, e = c.GetBroadcasters(context.Background(), scope, BroadcasterQuery{}); e == nil {
		t.Fatal("zero quote signer accepted")
	}
}

func TestBroadcasterDiscoveryStringAnchorAndRequiredMetadata(t *testing.T) {
	registry := ContractDeployment{common.HexToAddress("0x11"), crypto.Keccak256Hash([]byte{1}), 2}
	d := ProtocolDeployment{ChainID: "97", Protocol: registry, BroadcasterRegistry: &registry}
	block := map[string]any{"number": "10", "hash": common.HexToHash("0x10"), "parentHash": common.HexToHash("0x09"), "timestamp": "100"}
	entry := map[string]any{"operator": common.HexToAddress("0x01"), "endpoint": "https://relay.example/rpc", "metadataURI": "", "quoteSigner": common.HexToAddress("0x02"), "feeRecipient": common.HexToAddress("0x03"), "enabled": true, "block": block}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ ID uint64 }
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"block": block, "broadcasters": []any{entry}}})
	}))
	defer server.Close()
	c, _ := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &d})
	scope := DeploymentContext{97, registry.Address}
	page, e := c.GetBroadcasters(context.Background(), scope, BroadcasterQuery{})
	if e != nil || page.Block.Number != 10 || page.Block.Timestamp != 100 {
		t.Fatal(page, e)
	}
	delete(entry, "metadataURI")
	if _, e = c.GetBroadcasters(context.Background(), scope, BroadcasterQuery{}); e == nil {
		t.Fatal("missing metadata URI accepted")
	}
	entry["metadataURI"] = nil
	if _, e = c.GetBroadcasters(context.Background(), scope, BroadcasterQuery{}); e == nil {
		t.Fatal("null metadata URI accepted")
	}
}
