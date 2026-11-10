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
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func TestGetNote(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123}}
	block := CanonicalBlock{10, provider.header.Hash(), provider.header.ParentHash, 123}
	dep := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("11"), crypto.Keccak256Hash([]byte{1}), 1}}
	leaf := primitives.ElementFromUint64(1)
	root := leaf
	path := make([]common.Hash, primitives.TreeDepth)
	for i := range path {
		var e error
		root, e = primitives.Poseidon2Hash(root, primitives.ElementFromUint64(0))
		if e != nil {
			t.Fatal(e)
		}
		path[i] = common.Hash{}
	}
	anchor := rootAnchor{TreeID: 1, LeafCount: 1, Root: common.Hash(root.Bytes()), Block: block}
	mode := "valid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		proof := map[string]any{"anchor": anchor, "leafIndex": 0, "commitment": common.Hash(leaf.Bytes()), "siblings": path}
		var result any
		switch req.Method {
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": dep, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
		case "invisi_getRoots":
			result = []rootAnchor{anchor}
		case "invisi_getMembership":
			result = proof
		case "invisi_getNote":
			if mode == "missing" {
				result = nil
				break
			}
			output := map[string]any{"block": block, "treeId": 1, "leafIndex": 0, "commitment": common.Hash(leaf.Bytes()), "token": common.HexToAddress("22"), "ciphertext": "0x", "clue": "0x", "transactionHash": common.HexToHash("33"), "logIndex": 0, "kind": "Shield"}
			if mode == "bad" {
				proof["leafIndex"] = 1
			}
			result = map[string]any{"output": output, "membership": proof}
		default:
			t.Errorf("unexpected method %s", req.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	client, _ := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &dep})
	client = client.WithProvider(provider)
	scope := DeploymentContext{31337, dep.Protocol.Address}
	result, e := client.GetNote(context.Background(), scope, common.Hash(leaf.Bytes()), block)
	if e != nil || result == nil {
		t.Fatal(result, e)
	}
	if result.Membership.Root != anchor.Root || result.Output.LeafIndex != 0 {
		t.Fatal("wrong membership")
	}
	mode = "missing"
	if value, e := client.GetNote(context.Background(), scope, common.Hash(leaf.Bytes()), block); e != nil || value != nil {
		t.Fatal(value, e)
	}
	mode = "bad"
	if _, e = client.GetNote(context.Background(), scope, common.Hash(leaf.Bytes()), block); e == nil {
		t.Fatal("wrong position accepted")
	}
}

func TestGetCreditState(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123}}
	block := CanonicalBlock{10, provider.header.Hash(), provider.header.ParentHash, 123}
	dep := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("11"), crypto.Keccak256Hash([]byte{1}), 1}}
	leaf := primitives.ElementFromUint64(1)
	root := leaf
	path := make([]common.Hash, primitives.TreeDepth)
	for i := range path {
		var e error
		root, e = primitives.Poseidon2Hash(root, primitives.ElementFromUint64(0))
		if e != nil {
			t.Fatal(e)
		}
		path[i] = common.Hash{}
	}
	anchor := rootAnchor{TreeID: 1, LeafCount: 1, Root: common.Hash(root.Bytes()), Block: block}
	mode := "valid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		proof := map[string]any{"anchor": anchor, "leafIndex": 0, "commitment": common.Hash(leaf.Bytes()), "siblings": path}
		var result any
		switch req.Method {
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": dep, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
		case "invisi_getRoots":
			result = []rootAnchor{anchor}
		case "invisi_getMembership":
			result = proof
		case "invisi_getCreditState":
			if mode == "missing" {
				result = nil
				break
			}
			output := map[string]any{"block": block, "treeId": 1, "leafIndex": 0, "commitment": common.Hash(leaf.Bytes()), "token": common.HexToAddress("22"), "ciphertext": "0x", "clue": "0x", "transactionHash": common.HexToHash("33"), "logIndex": 0, "kind": "CreditCreated", "deliveries": []any{map[string]any{"ciphertext": "0x81", "clue": "0x" + repeatZero(68)}, map[string]any{"ciphertext": "0x8101", "clue": "0x" + repeatZero(68)}}}
			if mode == "bad" {
				proof["leafIndex"] = 1
			}
			result = map[string]any{"state": output, "membership": proof}
		default:
			t.Errorf("unexpected method %s", req.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	client, _ := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &dep})
	client = client.WithProvider(provider)
	scope := DeploymentContext{31337, dep.Protocol.Address}
	result, e := client.GetCreditState(context.Background(), scope, common.Hash(leaf.Bytes()), block)
	if e != nil || result == nil {
		t.Fatal(result, e)
	}
	if result.Membership.Root != anchor.Root || result.State.LeafIndex != 0 {
		t.Fatal("wrong membership")
	}
	mode = "missing"
	if value, e := client.GetCreditState(context.Background(), scope, common.Hash(leaf.Bytes()), block); e != nil || value != nil {
		t.Fatal(value, e)
	}
	mode = "bad"
	if _, e = client.GetCreditState(context.Background(), scope, common.Hash(leaf.Bytes()), block); e == nil {
		t.Fatal("wrong position accepted")
	}
}
