package provenance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionResponseMustMatchPinnedProtocol(t *testing.T) {
	const protocol = "0x1111111111111111111111111111111111111111"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID uint64 `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"scope": map[string]string{"chainId": "31337", "protocolAddress": protocol}}})
	}))
	defer server.Close()
	for _, test := range []struct {
		protocol string
		valid    bool
	}{{protocol, true}, {"0", false}, {"0x2222222222222222222222222222222222222222", false}} {
		client := Client{Endpoint: server.URL, Trust: TrustContext{Scope: Scope{ChainID: "31337", ProtocolAddress: test.protocol}}}
		_, err := client.GetVersion(context.Background())
		if (err == nil) != test.valid {
			t.Fatalf("protocol %s: %v", test.protocol, err)
		}
	}
}
