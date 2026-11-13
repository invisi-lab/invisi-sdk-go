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
	status, e := client.Submit(ctx, op)
	if e != nil || status.State != "uncertain" {
		t.Fatal("known status incorrectly required first-send validation", status, e)
	}
}

func TestUnknownExternalizedBroadcasterReplaysFrozenRequestWithoutFreshness(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	managed, err := New(Options{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	p, g := testPlan(t)
	if err = managed.Track(ctx, g); err != nil {
		t.Fatal(err)
	}
	op, err := managed.Prepare(ctx, "frozen", p)
	if err != nil {
		t.Fatal(err)
	}
	frozen := broadcaster.Submission{Data: "0x010203", Signature: "frozen-signature", Quote: broadcaster.Quote{Execution: "direct", Deadline: "1"}}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []broadcaster.Submission
		}
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			t.Fatal(e)
		}
		if req.Method != "broadcaster_submitTransaction" || len(req.Params) != 1 || req.Params[0].Data != frozen.Data || req.Params[0].Signature != frozen.Signature || req.Params[0].Quote.Deadline != "1" {
			t.Errorf("replay changed: %+v", req)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": broadcaster.Status{ID: "frozen-id", State: "uncertain"}})
	}))
	defer server.Close()
	client, err := broadcaster.NewClient(broadcaster.Configuration{Endpoint: server.URL, QuoteSigner: common.HexToAddress("0x11"), FeeRecipient: common.HexToAddress("0x22")}, broadcaster.Options{HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Update(ctx, func(s *State) error {
		r := s.Operations[op.id]
		r.Externalized = true
		r.Status = "uncertain"
		r.Calldata = []byte{1, 2, 3}
		r.BroadcasterRequest = &frozen
		r.Intent.Broadcaster = &credit.BroadcasterBinding{Configuration: client.Configuration(), Quote: broadcaster.SignedQuote{Quote: broadcaster.Quote{Execution: "direct"}}}
		s.Operations[op.id] = r
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	managed.validate = func(context.Context, credit.Intent) error { return fmt.Errorf("expired, spent, provider unavailable") }
	status, err := client.Submit(ctx, op)
	if err != nil || status.ID != "frozen-id" {
		t.Fatal(status, err)
	}
}

func TestCreditCancellationPersistsAcknowledgementBeforeCanonicalRelease(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	checks, cancels := 0, 0
	managed, err := New(Options{Store: store, CheckCancellation: func(context.Context, credit.Intent) error {
		checks++
		if checks == 1 {
			return fmt.Errorf("canonical provider unavailable")
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	p, g := testPlan(t)
	if err = managed.Track(ctx, g); err != nil {
		t.Fatal(err)
	}
	op, err := managed.Prepare(ctx, "scheduled", p)
	if err != nil {
		t.Fatal(err)
	}
	token, hash, err := broadcaster.NewCancelToken()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cancels++
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Method != "broadcaster_cancelTransaction" || cancels != 1 {
			t.Errorf("unexpected relay call: %s/%d", req.Method, cancels)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": broadcaster.Status{ID: "scheduled-id", State: "cancelled"}})
	}))
	defer server.Close()
	client, err := broadcaster.NewClient(broadcaster.Configuration{Endpoint: server.URL, QuoteSigner: common.HexToAddress("0x11"), FeeRecipient: common.HexToAddress("0x22")}, broadcaster.Options{HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Update(ctx, func(s *State) error {
		r := s.Operations[op.id]
		r.Externalized = true
		r.Status = "uncertain"
		r.BroadcasterRequestID = "scheduled-id"
		r.Intent.Broadcaster = &credit.BroadcasterBinding{CancelToken: token, Configuration: client.Configuration(), Quote: broadcaster.SignedQuote{Quote: broadcaster.Quote{Execution: "direct", Schedule: &broadcaster.Schedule{PublishAfter: "2000000000", ExecuteBefore: "2000000300", CancelTokenHash: hash}}}}
		s.Operations[op.id] = r
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	managed.validate = func(context.Context, credit.Intent) error { return fmt.Errorf("expired") }
	if _, err = client.Cancel(ctx, op); err == nil {
		t.Fatal("canonical failure released inputs")
	}
	state, err := store.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Operations[op.id].CancellationConfirmed || state.Grants[g.key()].ReservedBy != op.id {
		t.Fatal("ack lost or input released")
	}
	server.Close()
	status, err := client.Cancel(ctx, op)
	if err != nil || status.State != "cancelled" {
		t.Fatal(status, err)
	}
	state, err = store.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Grants[g.key()].ReservedBy != "" || state.Operations[op.id].Status != "cancelled" || cancels != 1 || checks != 2 {
		t.Fatal("reconciliation did not finish locally", cancels, checks)
	}
	if _, err = client.Cancel(ctx, op); err != nil || checks != 2 {
		t.Fatal("completed cancellation not idempotent", err)
	}
	if _, err = op.ExportSigningRequest(ctx); err == nil {
		t.Fatal("cancelled operation reused released inputs")
	}
}
