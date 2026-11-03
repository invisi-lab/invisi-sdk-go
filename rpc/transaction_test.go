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

func transactionClient(t *testing.T, mutate func(map[string]any), stale bool) (*ProtocolRpcClient, DeploymentContext, common.Hash) {
	t.Helper()
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123, ParentHash: common.HexToHash("0x09")}}
	block := CanonicalBlock{10, provider.header.Hash(), provider.header.ParentHash, 123}
	deployment := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("0x1111111111111111111111111111111111111111"), crypto.Keccak256Hash([]byte{0x60, 0}), 1}}
	hash := common.HexToHash("0x123")
	head := block
	if stale {
		head.Hash = common.HexToHash("0xff")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		var result any
		switch req.Method {
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": head, "checkedBits": 8, "detectionEnabled": true}
		case "invisi_getTransactionUpdates":
			if len(req.Params) != 1 || string(req.Params[0]) != `"`+hash.Hex()+`"` {
				t.Errorf("wrong transaction parameters: %s", req.Params)
			}
			value := map[string]any{"transactionHash": hash, "status": "indexed", "blockNumber": 10, "blockHash": block.Hash, "receiptStatus": 1, "indexedHead": head, "outputs": []any{}, "creditTransitions": []any{}}
			mutate(value)
			result = value
		default:
			t.Errorf("unexpected method %s", req.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	client, err := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &deployment})
	if err != nil {
		t.Fatal(err)
	}
	return client.WithProvider(provider), DeploymentContext{31337, deployment.Protocol.Address}, hash
}

func TestTransactionUpdatesAvailability(t *testing.T) {
	for _, state := range []string{"not_found", "pending_index", "indexed"} {
		t.Run(state, func(t *testing.T) {
			client, context, hash := transactionClient(t, func(value map[string]any) {
				value["status"] = state
				if state != "indexed" {
					value["outputs"] = nil
					value["creditTransitions"] = nil
				}
				if state == "not_found" {
					delete(value, "blockNumber")
					delete(value, "blockHash")
					delete(value, "receiptStatus")
				}
			}, false)
			result, err := client.GetTransactionUpdates(t.Context(), context, hash)
			if err != nil {
				t.Fatal(err)
			}
			if string(result.Status) != state {
				t.Fatalf("wrong state: %+v", result)
			}
			if state == "indexed" && (result.Outputs == nil || result.CreditTransitions == nil) {
				t.Fatal("indexed arrays cannot be null")
			}
			if state != "indexed" && (result.Outputs != nil || result.CreditTransitions != nil) {
				t.Fatal("unindexed arrays must be null")
			}
		})
	}
}
func TestTransactionUpdatesStaleCheckpoint(t *testing.T) {
	client, scope, hash := transactionClient(t, func(value map[string]any) {
		value["status"] = "pending_index"
		value["outputs"] = nil
		value["creditTransitions"] = nil
	}, true)
	result, err := client.GetTransactionUpdates(t.Context(), scope, hash)
	if err != nil || result.Status != TransactionPendingIndex {
		t.Fatal(result, err)
	}
	if _, err = client.GetStatus(t.Context(), scope); err == nil {
		t.Fatal("getStatus accepted stale checkpoint")
	}
}
func TestTransactionUpdatesValidation(t *testing.T) {
	for _, name := range []string{"unknown_status", "wrong_transaction", "wrong_block", "unsafe_height", "bad_receipt_status", "missing_head", "null_indexed_outputs", "pending_arrays", "not_found_metadata", "wrong_output_transaction", "conflicting_transition"} {
		t.Run(name, func(t *testing.T) {
			client, scope, hash := transactionClient(t, func(v map[string]any) {
				switch name {
				case "unknown_status":
					v["status"] = "other"
				case "wrong_transaction":
					v["transactionHash"] = common.HexToHash("0x99")
				case "wrong_block":
					v["blockHash"] = common.HexToHash("0x99")
				case "unsafe_height":
					v["blockNumber"] = uint64(1 << 53)
				case "bad_receipt_status":
					v["receiptStatus"] = 2
				case "missing_head":
					delete(v, "indexedHead")
				case "null_indexed_outputs":
					v["outputs"] = nil
				case "pending_arrays":
					v["status"] = "pending_index"
				case "not_found_metadata":
					v["status"] = "not_found"
					v["outputs"] = nil
					v["creditTransitions"] = nil
				case "wrong_output_transaction":
					v["outputs"] = []any{map[string]any{"transactionHash": common.HexToHash("0x99")}}
				case "conflicting_transition":
					event := func(amount string) any {
						return map[string]any{"block": v["indexedHead"], "transactionHash": v["transactionHash"], "name": "CreditDraw", "logIndex": 1, "data": map[string]any{"amount": amount}}
					}
					v["creditTransitions"] = []any{event("1"), event("2")}
				}
			}, false)
			if _, err := client.GetTransactionUpdates(t.Context(), scope, hash); err == nil {
				t.Fatal("accepted invalid transaction updates")
			}
		})
	}
	client, scope, hash := transactionClient(t, func(map[string]any) {}, false)
	if _, err := client.GetTransactionUpdates(t.Context(), scope, common.Hash{}); err == nil {
		t.Fatal("accepted zero hash")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.GetTransactionUpdates(ctx, scope, hash); err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestTransactionUpdatesPreserveDeliveriesAndDeduplicate(t *testing.T) {
	client, scope, hash := transactionClient(t, func(v map[string]any) {
		output := func(ciphertext string, leaf int, commitment string) any {
			return map[string]any{"block": v["indexedHead"], "treeId": 1, "leafIndex": leaf, "commitment": common.HexToHash(commitment), "token": "0x2222222222222222222222222222222222222222", "ciphertext": ciphertext, "clue": "0x" + repeatZero(68), "transactionHash": v["transactionHash"], "logIndex": 1, "kind": "CreditDraw"}
		}
		funds := output("0x01", 1, "0x4")
		v["outputs"] = []any{output("0x81", 0, "0x3"), output("0x8101", 0, "0x3"), funds, funds}
		event := map[string]any{"block": v["indexedHead"], "transactionHash": v["transactionHash"], "name": "CreditDraw", "logIndex": 2, "data": map[string]any{"amount": "12345678901234567890"}}
		v["creditTransitions"] = []any{event, event}
	}, false)
	result, err := client.GetTransactionUpdates(t.Context(), scope, hash)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outputs) != 3 || len(result.CreditTransitions) != 1 {
		t.Fatalf("unexpected candidates: %+v", result)
	}
	if string(result.CreditTransitions[0].Data) != `{"amount":"12345678901234567890"}` {
		t.Fatal("event integer precision lost")
	}
}

type transactionReorgProvider struct {
	testProvider
	reads int
}

func (p *transactionReorgProvider) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	p.reads++
	if p.reads >= 3 {
		copy := *p.header
		copy.Time++
		return &copy, nil
	}
	return p.header, nil
}
func TestTransactionUpdatesRechecksCanonicalBlock(t *testing.T) {
	client, scope, hash := transactionClient(t, func(map[string]any) {}, false)
	provider := &transactionReorgProvider{testProvider: client.provider.(testProvider)}
	client = client.WithProvider(provider)
	if _, err := client.GetTransactionUpdates(t.Context(), scope, hash); err == nil {
		t.Fatal("accepted reorganization during query")
	}
}
