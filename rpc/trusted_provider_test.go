package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedProviderPinsFreshHeadersAndReceipts(t *testing.T) {
	d := ProtocolDeployment{ChainID: "1", Protocol: ContractDeployment{Address: common.HexToAddress("1"), RuntimeCodeHash: common.HexToHash("2")}}
	scope := DeploymentContext{1, d.Protocol.Address}
	h := &types.Header{Number: big.NewInt(10), Difficulty: big.NewInt(0)}
	calls := 0
	corrupt := false
	badHash := false
	badReceipt := false
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			ID     uint64           `json:"id"`
			Method string           `json:"method"`
			Params []trustedRequest `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&q)
		if q.Method != "invisi_readChainState" {
			t.Errorf("unexpected method %s", q.Method)
		}
		calls++
		out := trustedResponse{APIVersion: "v1", ProtocolVersion: "1", Deployment: d}
		if q.Params[0].Kind == "header" {
			out.Header = h
			out.HeaderHash = h.Hash()
		}
		if badHash {
			out.HeaderHash = common.HexToHash("99")
		}
		if badReceipt && q.Params[0].Kind == "receipt" {
			out.Receipt = &types.Receipt{Type: 0, Status: 1, TxHash: q.Params[0].TransactionHash, BlockNumber: big.NewInt(10), BlockHash: common.HexToHash("98"), Logs: []*types.Log{}}
		}
		if corrupt {
			out.Deployment.Protocol.RuntimeCodeHash = common.HexToHash("9")
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": out})
	}))
	defer s.Close()
	c, _ := NewProtocolRpcClient(s.URL+"/v1", ClientOptions{Deployment: &d})
	p, e := NewTrustedProvider(c, scope)
	if e != nil {
		t.Fatal(e)
	}
	for range 2 {
		got, e := p.HeaderByNumber(context.Background(), big.NewInt(10))
		if e != nil || got.Hash() != h.Hash() {
			t.Fatal(e)
		}
	}
	if calls != 2 {
		t.Fatal("headers cached")
	}
	if _, e = p.TransactionReceipt(context.Background(), common.HexToHash("3")); !errors.Is(e, ethereum.NotFound) {
		t.Fatal("nil receipt not NotFound", e)
	}
	badHash = true
	if _, e = p.HeaderByNumber(context.Background(), nil); e == nil {
		t.Fatal("accepted forged header hash")
	}
	badHash = false
	badReceipt = true
	if _, e = p.TransactionReceipt(context.Background(), common.HexToHash("3")); e == nil {
		t.Fatal("accepted receipt on other block")
	}
	badReceipt = false
	corrupt = true
	if _, e = p.HeaderByTag(context.Background(), FinalityFinalized); e == nil {
		t.Fatal("accepted changed pins")
	}
	unpinned, _ := NewProtocolRpcClient(s.URL + "/v1")
	if _, e = NewTrustedProvider(unpinned, scope); e == nil {
		t.Fatal("accepted unpinned client")
	}
}
