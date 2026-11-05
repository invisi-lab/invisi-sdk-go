package protocol

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func TestDiscoveryReturnsTokenCompatibleRegisteredClient(t *testing.T) {
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
			var p struct{ Token common.Address }
			json.Unmarshal(req.Params[0], &p)
			if p.Token != token {
				t.Error("incorrect token probe")
			}
			feeRecipient := recipient
			if r.URL.Path == "/bad" {
				feeRecipient = common.HexToAddress("0xff")
			}
			result = broadcaster.Fee{FeeRecipient: feeRecipient, FeeAmount: "99999", MaxGas: "10000"}
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
	config := clients[0].Configuration()
	if config.Endpoint != endpoint+"/good" || config.QuoteSigner != signer || config.FeeRecipient != recipient {
		t.Fatal(config)
	}
}
