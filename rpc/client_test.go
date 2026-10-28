package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

type testProvider struct{ header *types.Header }

func (p testProvider) ChainID(context.Context) (*big.Int, error) { return big.NewInt(31337), nil }
func (p testProvider) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	return p.header, nil
}
func (p testProvider) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return []byte{0x60, 0}, nil
}
func (p testProvider) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	return common.LeftPadBytes([]byte{1}, 32), nil
}
func TestProtocolRPCAuthenticatesIndexedPaths(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 123, ParentHash: common.HexToHash("0x09")}}
	block := CanonicalBlock{10, provider.header.Hash(), provider.header.ParentHash, 123}
	address := common.HexToAddress("0x1111111111111111111111111111111111111111")
	deployment := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{address, crypto.Keccak256Hash([]byte{0x60, 0}), 1}}
	leaf := core.ElementFromUint64(7)
	root := leaf
	path := make([]common.Hash, 16)
	for range 16 {
		var e error
		root, e = core.Poseidon2Hash(root, core.ElementFromUint64(0))
		if e != nil {
			t.Fatal(e)
		}
	}
	anchor := rootAnchor{TreeID: 1, Root: common.Hash(root.Bytes()), LeafCount: 1, Block: block}
	var corrupt atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "tenant=test" {
			t.Error("query not preserved")
		}
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(map[string]any{"apiVersion": "v1", "chains": []any{map[string]any{"chainId": "31337", "protocolVersion": "1", "rpcPath": "/v1/chains/31337"}}})
			return
		}
		var request struct {
			ID     uint64
			Method string
		}
		if e := json.NewDecoder(r.Body).Decode(&request); e != nil {
			t.Error(e)
		}
		var result any
		switch request.Method {
		case "invisi_getDeployment":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment}
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
		case "invisi_getRoots":
			result = []rootAnchor{anchor}
		case "invisi_getMembership":
			siblings := append([]common.Hash(nil), path...)
			if corrupt.Load() {
				siblings[0] = common.HexToHash("0x01")
			}
			result = map[string]any{"anchor": anchor, "leafIndex": 0, "commitment": common.Hash(leaf.Bytes()), "siblings": siblings}
		default:
			t.Error("unexpected method", request.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
	}))
	defer server.Close()
	rpc, e := NewProtocolRpcClient(server.URL + "/prefix/v1?tenant=test")
	if e != nil {
		t.Fatal(e)
	}
	chains, e := rpc.GetChains(context.Background())
	if e != nil || len(chains) != 1 {
		t.Fatal(chains, e)
	}
	if _, e = rpc.GetDeployment(context.Background(), 31337); e != nil {
		t.Fatal(e)
	}
	verified := rpc.WithProvider(provider)
	context := DeploymentContext{31337, address}
	m, e := verified.GetMembership(contextpkg(), context, MembershipInput{TreeID: 1, LeafIndex: 0, Commitment: common.Hash(leaf.Bytes())}, block)
	if e != nil || m.Root != anchor.Root {
		t.Fatal(m, e)
	}
	corrupt.Store(true)
	if _, e = verified.GetMembership(contextpkg(), context, MembershipInput{TreeID: 1, LeafIndex: 0, Commitment: common.Hash(leaf.Bytes())}, block); e == nil {
		t.Fatal("forged membership accepted")
	}
}
func contextpkg() context.Context { return context.Background() }
func TestHTTPOverloadCarriesRetryGuidance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Retry-After", "1"); w.WriteHeader(429) }))
	defer server.Close()
	rpc, e := NewProtocolRpcClient(server.URL + "/v1")
	if e != nil {
		t.Fatal(e)
	}
	_, e = rpc.GetChains(context.Background())
	var overload *HTTPError
	if !errors.As(e, &overload) || overload.StatusCode != 429 || overload.RetryAfter != "1" {
		t.Fatal(e)
	}
}
