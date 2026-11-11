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
	for _, name := range []string{"GetNoteUpdates", "GetCreditTransitions", "GetTransactionNoteUpdates", "GetTransactionCreditUpdates", "GetNote", "GetCreditState", "IsNoteSpent", "IsCreditConsumed", "GetCreditAllocation", "GetCreditRevocationBitmap"} {
		if _, ok := typ.MethodByName(name); !ok {
			t.Errorf("missing typed method %s", name)
		}
	}

}

func TestGrantQueriesPreserveUint64ShardAndUint256Bitmap(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123}}
	block := CanonicalBlock{10, provider.header.Hash(), provider.header.ParentHash, 123}
	dep := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("11"), crypto.Keccak256Hash([]byte{1}), 1}}
	owner, token := common.HexToAddress("22"), common.HexToAddress("33")
	bad := false
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
		case "invisi_getCreditAllocation":
			result = map[string]any{"block": block, "shard": "18446744073709551615", "slot": 65535}
		case "invisi_getCreditRevocationBitmap":
			bits := new(big.Int).Lsh(big.NewInt(1), 255).String()
			if bad {
				bits = "01"
			}
			result = map[string]any{"block": block, "bitmap": bits}
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
	allocation, e := client.GetCreditAllocation(t.Context(), scope, owner, token, block)
	if e != nil || allocation.Shard != ^uint64(0) || allocation.Slot != 65535 {
		t.Fatal(allocation, e)
	}
	bitmap, e := client.GetCreditRevocationBitmap(t.Context(), scope, owner, token, ^uint64(0), 255, block)
	if e != nil || bitmap.Bitmap.Bit(255) != 1 {
		t.Fatal(bitmap, e)
	}
	if string(seen["invisi_getCreditRevocationBitmap"][2]) != `"18446744073709551615"` {
		t.Fatal("shard precision lost")
	}
	if spent, e := client.IsNoteSpent(t.Context(), scope, common.HexToHash("1"), block); e != nil || !spent {
		t.Fatal(spent, e)
	}
	if used, e := client.IsCreditConsumed(t.Context(), scope, common.HexToHash("2"), block); e != nil || !used {
		t.Fatal(used, e)
	}
	bad = true
	if _, e = client.GetCreditRevocationBitmap(t.Context(), scope, owner, token, 1, 0, block); e == nil {
		t.Fatal("noncanonical bitmap accepted")
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
