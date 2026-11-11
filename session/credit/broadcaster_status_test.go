package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

func TestKnownExternalizedBroadcasterStatusBypassesNewSendFreshness(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	managed, e := New(Options{Store: store})
	if e != nil {
		t.Fatal(e)
	}
	p, grant := testPlan(t)
	if e = managed.Track(ctx, grant); e != nil {
		t.Fatal(e)
	}
	op, e := managed.Prepare(ctx, "original", p)
	if e != nil {
		t.Fatal(e)
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Method != "broadcaster_getTransactionStatus" {
			t.Errorf("known externalized journal attempted %s", req.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": broadcaster.Status{ID: "original", State: "uncertain"}})
	}))
	defer server.Close()
	client, e := broadcaster.NewClient(broadcaster.Configuration{Endpoint: server.URL, QuoteSigner: common.HexToAddress("0x11"), FeeRecipient: common.HexToAddress("0x22")}, broadcaster.Options{HTTPClient: server.Client()})
	if e != nil {
		t.Fatal(e)
	}
	// Only the durable transport marker is supplied here; this test makes no proof acceptance claim.
	if e = store.Update(ctx, func(s *State) error {
		r := s.Operations[op.id]
		r.Externalized = true
		r.BroadcasterRequestID = "original"
		r.Status = "uncertain"
		r.Intent.Broadcaster = &credit.BroadcasterBinding{Configuration: client.Configuration(), Quote: broadcaster.SignedQuote{Quote: broadcaster.Quote{Execution: "direct"}}}
		s.Operations[op.id] = r
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	managed.validate = func(context.Context, credit.Intent) error { return fmt.Errorf("expired or spent") }
	status, e := op.SubmitBroadcaster(ctx, client)
	if e != nil || status.State != "uncertain" {
		t.Fatal("known status incorrectly required first-send validation", status, e)
	}
}
