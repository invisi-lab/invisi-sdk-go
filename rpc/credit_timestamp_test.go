package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestUpdatesAuthenticateTransitionBlockTimestamp(t *testing.T) {
	provider := testProvider{&types.Header{Number: big.NewInt(10), Time: 2000000000, ParentHash: common.HexToHash("0x09")}}
	block := CanonicalBlock{10, provider.header.Hash(), provider.header.ParentHash, provider.header.Time}
	deployment := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{common.HexToAddress("0x1111111111111111111111111111111111111111"), crypto.Keccak256Hash([]byte{0x60, 0}), 1}}
	var corrupt atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64
			Method string
		}
		if e := json.NewDecoder(r.Body).Decode(&request); e != nil {
			t.Error(e)
			return
		}
		var result any
		if request.Method == "invisi_getStatus" {
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
		} else {
			claimed := block
			if corrupt.Load() {
				claimed.Timestamp++
			}
			result = []any{map[string]any{"transactionHash": common.HexToHash("0x1234"), "logIndex": 1, "block": claimed, "name": "CreditDraw", "data": map[string]any{}}}
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
	}))
	defer server.Close()
	c, e := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &deployment})
	if e != nil {
		t.Fatal(e)
	}
	c = c.WithProvider(provider)
	ctx := DeploymentContext{31337, deployment.Protocol.Address}
	updates, e := c.GetCreditTransitions(context.Background(), ctx, ProtocolUpdateRange{FromBlock: 10, ToBlock: 10})
	if e != nil {
		t.Fatal(e)
	}
	if len(updates.Transitions) != 1 || updates.Transitions[0].Block.Timestamp != provider.header.Time {
		t.Fatal("timestamp not retained")
	}
	corrupt.Store(true)
	if _, e = c.GetCreditTransitions(context.Background(), ctx, ProtocolUpdateRange{FromBlock: 10, ToBlock: 10}); e == nil {
		t.Fatal("forged transition timestamp accepted")
	}
}
