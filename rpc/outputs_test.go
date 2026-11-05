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
	raw := json.RawMessage(`{"block":{"number":10,"hash":"0x0000000000000000000000000000000000000000000000000000000000000001","parentHash":"0x0000000000000000000000000000000000000000000000000000000000000002","timestamp":3},"treeId":1,"leafIndex":0,"commitment":"0x0000000000000000000000000000000000000000000000000000000000000003","token":"0x1111111111111111111111111111111111111111","ciphertext":"0x81","clue":"0x` + repeatZero(68) + `","transactionHash":"0x0000000000000000000000000000000000000000000000000000000000000004","logIndex":0,"kind":"CreditCreated"}`)
	out, e := parseOutput(raw)
	if e != nil || len(out.Ciphertext) != 1 {
		t.Fatal(out, e)
	}
	var m map[string]any
	json.Unmarshal(raw, &m)
	delete(m, "logIndex")
	bad, _ := json.Marshal(m)
	if _, e := parseOutput(bad); e == nil {
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

func TestOutputPagesAuthenticateAndPreserveCreditRoles(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123, ParentHash: common.HexToHash("0x09")}}
	block := CanonicalBlock{10, provider.header.Hash(), provider.header.ParentHash, 123}
	deployment := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("0x1111111111111111111111111111111111111111"), crypto.Keccak256Hash([]byte{0x60, 0}), 1}}
	clue := "0x" + repeatZero(68)
	output := func(ciphertext string) map[string]any {
		return map[string]any{"block": block, "treeId": "1", "leafIndex": "0", "commitment": common.HexToHash("0x03"), "token": deployment.Protocol.Address, "ciphertext": ciphertext, "clue": clue, "transactionHash": common.HexToHash("0x04"), "logIndex": "0", "kind": "CreditCreated"}
	}
	rows := []any{output("0x81"), output("0x8101"), output("0x81")}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		var result any
		if req.Method == "invisi_getStatus" {
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
		} else {
			result = map[string]any{"blockHash": block.Hash, "outputs": rows, "creditTransitions": []any{}}
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	c, e := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &deployment})
	if e != nil {
		t.Fatal(e)
	}
	c = c.WithProvider(provider)
	d := DeploymentContext{31337, deployment.Protocol.Address}
	r := ProtocolUpdateRange{FromBlock: 10, ToBlock: 10}
	page, e := c.GetCreditOutputs(context.Background(), d, r)
	if e != nil || len(page.Outputs) != 2 {
		t.Fatal(page, e)
	}
	page.Outputs[0].Ciphertext[0] = 0
	next, e := c.GetCreditOutputs(context.Background(), d, r)
	if e != nil || next.Outputs[0].Ciphertext[0] != 129 {
		t.Fatal("mutable output boundary", e)
	}
	rows = append(rows, output("0x8102"))
	if _, e = c.GetCreditOutputs(context.Background(), d, r); e == nil {
		t.Fatal("third role accepted")
	}
	rows = []any{output("0x81"), output("0x8101")}
	rows[1].(map[string]any)["token"] = common.HexToAddress("0x2222222222222222222222222222222222222222")
	if _, e = c.GetCreditOutputs(context.Background(), d, r); e == nil {
		t.Fatal("conflicting leaf accepted")
	}
	rows = []any{output("0x01")}
	funds, e := c.GetOutputs(context.Background(), d, r)
	if e != nil || len(funds.Outputs) != 1 {
		t.Fatal("funds candidate in credit event lost", e)
	}
	forged := block
	forged.Timestamp++
	rows[0].(map[string]any)["block"] = forged
	if _, e = c.GetEncryptedOutputs(context.Background(), d, r); e == nil {
		t.Fatal("forged timestamp accepted")
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
