package rpc

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestIndependentBusinessMethodSurface(t *testing.T) {
	typ := reflect.TypeOf((*ProtocolRpcClient)(nil))
	for _, name := range []string{"GetNoteUpdates", "GetCreditTransitions", "GetTransactionNoteUpdates", "GetTransactionCreditUpdates", "GetNote", "GetCreditState", "IsNoteSpent", "IsCreditConsumed", "GetCreditStatus"} {
		if _, ok := typ.MethodByName(name); !ok {
			t.Errorf("missing typed method %s", name)
		}
	}

}

func TestGrantQueriesPreserveUint128IDAndValidateStatus(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123}}
	block := CanonicalBlock{Number: 10, Hash: provider.header.Hash(), ParentHash: provider.header.ParentHash, Timestamp: 123}
	dep := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("11"), crypto.Keccak256Hash([]byte{1}), 1}}
	owner, token := common.HexToAddress("22"), common.HexToAddress("33")
	var observation any = 2
	omitStatus := false
	seen := map[string][]json.RawMessage{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		json.NewDecoder(r.Body).Decode(&req)
		seen[req.Method] = req.Params
		var result any
		switch req.Method {
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": dep, "indexedHead": block}
		case "invisi_getCreditStatus":
			result = map[string]any{"block": block, "status": observation}
			if omitStatus {
				result = map[string]any{"block": block}
			}
		case "invisi_isNoteSpent", "invisi_isCreditConsumed":
			result = true
		default:
			t.Error("unexpected method", req.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	client, _ := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &dep})
	client = client.WithProvider(provider)
	scope := DeploymentContext{31337, dep.Protocol.Address}
	id := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	status, e := client.GetCreditStatus(t.Context(), scope, owner, token, id, block)
	if e != nil || status.Status != 2 {
		t.Fatal(status, e)
	}
	if string(seen["invisi_getCreditStatus"][2]) != `"340282366920938463463374607431768211455"` {
		t.Fatal("grant ID precision lost")
	}
	if spent, e := client.IsNoteSpent(t.Context(), scope, common.HexToHash("1"), block); e != nil || !spent {
		t.Fatal(spent, e)
	}
	if used, e := client.IsCreditConsumed(t.Context(), scope, common.HexToHash("2"), block); e != nil || !used {
		t.Fatal(used, e)
	}
	for _, invalid := range []any{nil, 3, true, "0", -1, json.Number("0.0")} {
		observation = invalid
		if _, e = client.GetCreditStatus(t.Context(), scope, owner, token, id, block); e == nil {
			t.Fatalf("malformed status accepted: %#v", invalid)
		}
	}
	observation = 0
	omitStatus = true
	if _, e = client.GetCreditStatus(t.Context(), scope, owner, token, id, block); e == nil {
		t.Fatal("missing status treated as unknown")
	}

}

func TestCreditTransactionAvailabilityAndInvalidTransitions(t *testing.T) {
	for _, state := range []string{"not_found", "pending_index", "indexed"} {
		t.Run(state, func(t *testing.T) {
			client, scope, hash := transactionClient(t, func(v map[string]any) {
				v["status"] = state
				if state != "indexed" {
					v["notes"] = nil
					v["states"] = nil
					v["transitions"] = nil
				}
				if state == "not_found" {
					delete(v, "blockNumber")
					delete(v, "blockHash")
					delete(v, "receiptStatus")
				}
			}, false)
			r, e := client.GetTransactionCreditUpdates(t.Context(), scope, hash)
			if e != nil {
				t.Fatal(e)
			}
			if state == "indexed" && (r.States == nil || r.Transitions == nil) {
				t.Fatal("indexed collections null")
			}
			if state != "indexed" && (r.States != nil || r.Transitions != nil) {
				t.Fatal("unindexed collections present")
			}
		})
	}
	for _, name := range []string{"NotAnEvent", "Shield", "Transact"} {
		t.Run(name, func(t *testing.T) {
			client, scope, hash := transactionClient(t, func(v map[string]any) {
				v["transitions"] = []any{map[string]any{"block": v["indexedHead"], "transactionHash": v["transactionHash"], "logIndex": 1, "name": name, "data": map[string]any{}}}
			}, false)
			if _, e := client.GetTransactionCreditUpdates(t.Context(), scope, hash); e == nil {
				t.Fatal("non-Credit transition accepted")
			}
		})
	}
}

func TestRemoteDetectionMethodsAreAbsent(t *testing.T) {
	typ := reflect.TypeOf(&ProtocolRpcClient{})
	for _, name := range []string{"DetectNotes", "DetectCreditStates"} {
		if _, ok := typ.MethodByName(name); ok {
			t.Fatalf("remote method remains: %s", name)
		}
	}
}
