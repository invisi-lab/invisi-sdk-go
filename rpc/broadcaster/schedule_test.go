package broadcaster

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestScheduledTypedDataMatchesSettlementVector(t *testing.T) {
	raw, e := os.ReadFile("testdata/scheduled-quote.json")
	if e != nil {
		t.Fatal(e)
	}
	var vector struct {
		Quote  Quote  `json:"quote"`
		Digest string `json:"digest"`
	}
	if e = json.Unmarshal(raw, &vector); e != nil {
		t.Fatal(e)
	}
	td, e := QuoteTypedData(vector.Quote, Trust{ChainID: big.NewInt(31337), Protocol: common.HexToAddress("0x1111111111111111111111111111111111111111"), Signer: common.HexToAddress("0x1111111111111111111111111111111111111111")})
	if e != nil {
		t.Fatal(e)
	}
	digest, _, e := apitypes.TypedDataAndHash(td)
	if e != nil || common.BytesToHash(digest).Hex() != vector.Digest {
		t.Fatal(common.BytesToHash(digest), vector.Digest, e)
	}
}
func TestScheduledStatesAndConfirmation(t *testing.T) {
	for _, state := range []string{"scheduled", "preparing", "cancelled"} {
		s := Status{ID: "id", State: state}
		if e := checkStatus(s, "id"); e != nil {
			t.Fatal(e)
		}
		s.Confirmation = &Confirmation{}
		if checkStatus(s, "id") == nil {
			t.Fatal("accepted confirmation", state)
		}
	}
}

func TestCancelRPCUsesSecretAndRequiresConfirmedState(t *testing.T) {
	token, _, e := NewCancelToken()
	if e != nil {
		t.Fatal(e)
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64              `json:"id"`
			Method string              `json:"method"`
			Params []map[string]string `json:"params"`
		}
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			t.Error(e)
		}
		if req.Method != "broadcaster_cancelTransaction" || len(req.Params) != 1 || req.Params[0]["id"] != "job" || req.Params[0]["cancelToken"] != token {
			t.Errorf("incorrect cancellation request: %+v", req)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]string{"id": "job", "state": "cancelled"}})
	}))
	defer server.Close()
	c, e := NewClient(Configuration{Endpoint: server.URL, QuoteSigner: common.HexToAddress("0x1"), FeeRecipient: common.HexToAddress("0x2")}, Options{HTTPClient: server.Client()})
	if e != nil {
		t.Fatal(e)
	}
	status, e := c.CancelTransaction(context.Background(), "job", token)
	if e != nil || status.State != "cancelled" {
		t.Fatal(status, e)
	}
	if _, e = c.CancelTransaction(context.Background(), "job", "bad"); e == nil {
		t.Fatal("invalid secret accepted")
	}
}
