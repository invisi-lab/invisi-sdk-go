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

func TestNormalizeOutputOpaqueAndConflicts(t *testing.T) {
	raw := json.RawMessage(`{"block":{"number":10,"hash":"0x0000000000000000000000000000000000000000000000000000000000000001","parentHash":"0x0000000000000000000000000000000000000000000000000000000000000002","timestamp":3},"treeId":1,"leafIndex":0,"commitment":"0x0000000000000000000000000000000000000000000000000000000000000003","token":"0x1111111111111111111111111111111111111111","ciphertext":"0x81","clue":"0x` + repeatZero(68) + `","transactionHash":"0x0000000000000000000000000000000000000000000000000000000000000004","logIndex":0,"kind":"Transact"}`)
	out, e := parseNoteOutput(raw)
	if e != nil || len(out.Ciphertext) != 1 {
		t.Fatal(out, e)
	}
	var m map[string]any
	json.Unmarshal(raw, &m)
	delete(m, "logIndex")
	bad, _ := json.Marshal(m)
	if _, e := parseNoteOutput(bad); e == nil {
		t.Fatal("missing logIndex accepted")
	}
}
func repeatZero(n int) string {
	s := ""
	for range n {
		s += "00"
	}
	return s
}

func TestPublicCreditStatesAuthenticateAndRetainRoleEnvelopes(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123}}
	block := CanonicalBlock{Number: 10, Hash: provider.header.Hash(), ParentHash: provider.header.ParentHash, Timestamp: 123}
	dep := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("11"), crypto.Keccak256Hash([]byte{1}), 1}}
	state := func() map[string]any {
		return map[string]any{"block": block, "treeId": 1, "leafIndex": 0, "commitment": common.HexToHash("3"), "token": common.HexToAddress("22"), "transactionHash": common.HexToHash("4"), "logIndex": 0, "kind": "CreditCreated", "deliveries": []any{map[string]any{"ciphertext": "0x81", "clue": "0x" + repeatZero(68)}, map[string]any{"ciphertext": "0x8101", "clue": "0x" + repeatZero(68)}}}
	}
	raw := state()
	rows := []any{raw, raw}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		var result any
		if req.Method == "invisi_getStatus" {
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": dep, "indexedHead": block}
		} else if req.Method == "invisi_getTransactionCreditUpdates" {
			result = map[string]any{"transactionHash": common.HexToHash("4"), "status": "indexed", "indexedHead": block, "blockNumber": block.Number, "blockHash": block.Hash, "receiptStatus": 1, "states": rows, "transitions": []any{}}
		} else {
			t.Error("unexpected method", req.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	c, _ := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &dep})
	c = c.WithProvider(provider)
	scope := DeploymentContext{31337, dep.Protocol.Address}
	tx := common.HexToHash("4")
	page, e := c.GetTransactionCreditUpdates(context.Background(), scope, tx)
	if e != nil || len(page.States) != 1 || len(page.States[0].Deliveries) != 2 {
		t.Fatal(page, e)
	}
	page.States[0].Deliveries[0].Ciphertext[0] = 0
	next, e := c.GetTransactionCreditUpdates(context.Background(), scope, tx)
	if e != nil || next.States[0].Deliveries[0].Ciphertext[0] != 129 {
		t.Fatal("mutable state boundary", e)
	}
	other := state()
	other["token"] = common.HexToAddress("33")
	rows = append(rows, other)
	if _, e = c.GetTransactionCreditUpdates(context.Background(), scope, tx); e == nil {
		t.Fatal("conflicting state accepted")
	}
	rows = []any{state()}
	rows[0].(map[string]any)["deliveries"] = []any{}
	if _, e = c.GetTransactionCreditUpdates(context.Background(), scope, tx); e == nil {
		t.Fatal("missing role envelopes accepted")
	}
}

func TestReceivingAddressBindsOwnerRegistryAndCanonicalMetadataShape(t *testing.T) {
	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	deployment := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{owner, common.HexToHash("0x01"), 1}, Registry: &ContractDeployment{owner, common.HexToHash("0x02"), 2}}
	result := any(map[string]any{"ownerAddress": owner, "viewingPublicKey": "0x0396f42bb298c8a0e74539294780d709799f9838c0b69a134f38c6d968f2921397", "clueKey": "0xc69480518ea88f8cb925fbcd8f5100b1ebc46cc4d8cbe19e0478936567e6b211", "nonce": "1", "block": map[string]any{"number": "2", "hash": common.HexToHash("0x03"), "parentHash": common.HexToHash("0x04"), "timestamp": "5"}})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []string
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Method != "invisi_getReceivingAddress" || len(req.Params) != 1 || req.Params[0] != owner.Hex() {
			t.Error("wrong registration request")
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	c, e := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &deployment})
	if e != nil {
		t.Fatal(e)
	}
	address, e := c.GetReceivingAddress(context.Background(), owner)
	if e != nil || address == nil || address.Block.Timestamp != 5 || address.Nonce != 1 {
		t.Fatal(address, e)
	}
	result.(map[string]any)["ownerAddress"] = common.HexToAddress("0x2222222222222222222222222222222222222222")
	if _, e = c.GetReceivingAddress(context.Background(), owner); e == nil {
		t.Fatal("wrong owner accepted")
	}
	result = nil
	if address, e = c.GetReceivingAddress(context.Background(), owner); e != nil || address != nil {
		t.Fatal(address, e)
	}
	if _, e = c.GetReceivingAddress(context.Background(), common.Address{}); e == nil {
		t.Fatal("zero owner accepted")
	}
}
