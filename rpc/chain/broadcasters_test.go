package chain

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

func TestDiscoveryReturnsTokenCompatibleRegisteredClient(t *testing.T) {
	var wrongChain atomic.Bool
	var wrongToken atomic.Bool
	registry := protocolrpc.ContractDeployment{Address: common.HexToAddress("0x11"), RuntimeCodeHash: crypto.Keccak256Hash([]byte{1}), FromBlock: 1}
	d := protocolrpc.ProtocolDeployment{ChainID: "97", Protocol: registry, BroadcasterRegistry: &registry}
	token := common.HexToAddress("0x22")
	signer := common.HexToAddress("0x33")
	recipient := common.HexToAddress("0x44")
	block := protocolrpc.CanonicalBlock{Number: 10, Hash: common.HexToHash("0x10"), Timestamp: 100}
	var endpoint string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		json.NewDecoder(r.Body).Decode(&req)
		var result any
		if req.Method == "invisi_getBroadcasters" {
			result = protocolrpc.BroadcasterPage{Block: block, Broadcasters: []protocolrpc.RegisteredBroadcaster{{Operator: common.HexToAddress("0x01"), Endpoint: endpoint + "/bad", QuoteSigner: signer, FeeRecipient: recipient, Enabled: true, Block: block}, {Operator: common.HexToAddress("0x02"), Endpoint: endpoint + "/good", QuoteSigner: signer, FeeRecipient: recipient, Enabled: true, Block: block}}}
		} else {
			var p struct{ TokenAddress common.Address }
			json.Unmarshal(req.Params[0], &p)
			if p.TokenAddress != token {
				t.Error("incorrect token probe")
			}
			feeRecipient := recipient
			if r.URL.Path == "/bad" {
				feeRecipient = common.HexToAddress("0xff")
			}
			chain := "97"
			if wrongChain.Load() {
				chain = "98"
			}
			returnedToken := token
			if wrongToken.Load() {
				returnedToken = common.HexToAddress("0xff")
			}
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "chainId": chain, "protocol": registry.Address.Hex(), "sender": signer.Hex(), "feeRecipient": feeRecipient.Hex(), "submissionHalted": false, "provenance": map[string]any{"enabled": false}, "tokenAddress": returnedToken.Hex(), "supported": true}
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	endpoint = server.URL
	rpc, _ := protocolrpc.NewProtocolRpcClient(endpoint+"/v1", protocolrpc.ClientOptions{HTTPClient: server.Client(), Deployment: &d})
	p, _ := NewProtocol(d)
	p.rpc = rpc
	clients, e := p.GetBroadcasters(context.Background(), BroadcasterOptions{TokenAddress: token, ClientOptions: broadcaster.Options{HTTPClient: server.Client()}})
	if e != nil || len(clients) != 1 {
		t.Fatal(clients, e)
	}
	if _, err := p.GetBroadcasters(context.Background(), BroadcasterOptions{}); err == nil {
		t.Fatal("zero token accepted")
	}
	config := clients[0].Configuration()
	if config.Endpoint != endpoint+"/good" || config.QuoteSigner != signer || config.FeeRecipient != recipient {
		t.Fatal(config)
	}
	wrongChain.Store(true)
	clients, e = p.GetBroadcasters(context.Background(), BroadcasterOptions{TokenAddress: token, ClientOptions: broadcaster.Options{HTTPClient: server.Client()}})
	if e != nil || len(clients) != 0 {
		t.Fatal("cross-chain service accepted", e)
	}
	wrongChain.Store(false)
	wrongToken.Store(true)
	clients, e = p.GetBroadcasters(context.Background(), BroadcasterOptions{TokenAddress: token, ClientOptions: broadcaster.Options{HTTPClient: server.Client()}})
	if e != nil || len(clients) != 0 {
		t.Fatal("different token service accepted", e)
	}
}
