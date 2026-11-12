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
	block := CanonicalBlock{Number: 10, Hash: provider.header.Hash(), ParentHash: provider.header.ParentHash, Timestamp: 123}
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
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": head}
		case "invisi_getTransactionNoteUpdates", "invisi_getTransactionCreditUpdates":
			if len(req.Params) != 1 || string(req.Params[0]) != `"`+hash.Hex()+`"` {
				t.Errorf("wrong transaction parameters: %s", req.Params)
			}
			value := map[string]any{"transactionHash": hash, "status": "indexed", "blockNumber": 10, "blockHash": block.Hash, "receiptStatus": 1, "indexedHead": head, "notes": []any{}, "transitions": []any{}, "states": []any{}}
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
					value["notes"] = nil
					value["transitions"] = nil
					value["states"] = nil
				}
				if state == "not_found" {
					delete(value, "blockNumber")
					delete(value, "blockHash")
					delete(value, "receiptStatus")
				}
			}, false)
			result, err := client.GetTransactionNoteUpdates(t.Context(), context, hash)
			if err != nil {
				t.Fatal(err)
			}
			if string(result.Status) != state {
				t.Fatalf("wrong state: %+v", result)
			}
			if state == "indexed" && (result.Notes == nil) {
				t.Fatal("indexed arrays cannot be null")
			}
			if state != "indexed" && (result.Notes != nil) {
				t.Fatal("unindexed arrays must be null")
			}
		})
	}
}
func TestTransactionUpdatesStaleCheckpoint(t *testing.T) {
	client, scope, hash := transactionClient(t, func(value map[string]any) {
		value["status"] = "pending_index"
		value["notes"] = nil
		value["transitions"] = nil
		value["states"] = nil
	}, true)
	result, err := client.GetTransactionNoteUpdates(t.Context(), scope, hash)
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
					v["notes"] = nil
				case "pending_arrays":
					v["status"] = "pending_index"
				case "not_found_metadata":
					v["status"] = "not_found"
					v["notes"] = nil
					v["transitions"] = nil
				case "wrong_output_transaction":
					v["notes"] = []any{map[string]any{"transactionHash": common.HexToHash("0x99")}}
				case "conflicting_transition":
					event := func(amount string) any {
						return map[string]any{"block": v["indexedHead"], "transactionHash": v["transactionHash"], "name": "CreditDraw", "logIndex": 1, "data": map[string]any{"amount": amount}}
					}
					v["transitions"] = []any{event("1"), event("2")}
				}
			}, false)
			var err error
			if name == "conflicting_transition" {
				_, err = client.GetTransactionCreditUpdates(t.Context(), scope, hash)
			} else {
				_, err = client.GetTransactionNoteUpdates(t.Context(), scope, hash)
			}
			if err == nil {
				t.Fatal("accepted invalid transaction updates")
			}
		})
	}
	client, scope, hash := transactionClient(t, func(map[string]any) {}, false)
	if _, err := client.GetTransactionNoteUpdates(t.Context(), scope, common.Hash{}); err == nil {
		t.Fatal("accepted zero hash")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.GetTransactionNoteUpdates(ctx, scope, hash); err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestCreditTransactionPreservesDeliveriesAndPrecision(t *testing.T) {
	client, scope, hash := transactionClient(t, func(v map[string]any) {
		state := map[string]any{"block": v["indexedHead"], "treeId": 1, "leafIndex": 0, "commitment": common.HexToHash("3"), "token": common.HexToAddress("22"), "transactionHash": v["transactionHash"], "logIndex": 1, "kind": "CreditDraw", "deliveries": []any{map[string]any{"ciphertext": "0x81", "clue": "0x" + repeatZero(68)}, map[string]any{"ciphertext": "0x8101", "clue": "0x" + repeatZero(68)}}}
		v["states"] = []any{state, state}
		event := map[string]any{"block": v["indexedHead"], "transactionHash": v["transactionHash"], "name": "CreditDraw", "logIndex": 2, "data": map[string]any{"amount": "12345678901234567890"}}
		v["transitions"] = []any{event, event}
	}, false)
	result, e := client.GetTransactionCreditUpdates(t.Context(), scope, hash)
	if e != nil {
		t.Fatal(e)
	}
	if len(result.States) != 1 || len(result.States[0].Deliveries) != 2 || len(result.Transitions) != 1 {
		t.Fatal(result)
	}
	if string(result.Transitions[0].Data) != `{"amount":"12345678901234567890"}` {
		t.Fatal("precision lost")
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
	if _, err := client.GetTransactionNoteUpdates(t.Context(), scope, hash); err == nil {
		t.Fatal("accepted reorganization during query")
	}
}
