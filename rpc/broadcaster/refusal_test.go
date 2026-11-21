package broadcaster

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestSubmissionRefusalRequiresExactRPCCodeAndMachineReason(t *testing.T) {
	for _, tc := range []struct {
		code    int
		message string
		refused bool
	}{
		{-32024, "CREDIT_PERIOD_CHANGED: stale", true}, {-32025, "INSUFFICIENT_EXECUTION_SLACK: short", true}, {-32026, "CREDIT_PERIOD_ENDING: soon", true},
		{-32000, "CREDIT_PERIOD_CHANGED: stale", false}, {-32024, "something CREDIT_PERIOD_CHANGED", false}, {-32024, "CREDIT_PERIOD_ENDING: soon", false},
	} {
		t.Run(tc.message, func(t *testing.T) {
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req struct{ ID uint64 }
				json.NewDecoder(r.Body).Decode(&req)
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": tc.code, "message": tc.message}})
			}))
			defer s.Close()
			c, e := NewClient(Configuration{Endpoint: s.URL, QuoteSigner: common.HexToAddress("1"), FeeRecipient: common.HexToAddress("2")}, Options{HTTPClient: s.Client()})
			if e != nil {
				t.Fatal(e)
			}
			_, e = c.SubmitPrepared(context.Background(), Submission{})
			var refusal *SubmissionRefusedError
			if errors.As(e, &refusal) != tc.refused {
				t.Fatalf("classification: %T %v", e, e)
			}
		})
	}
}

func TestSubmissionRefusalNeverClassifiesTransportOrMalformedResponse(t *testing.T) {
	for _, body := range []string{`CREDIT_PERIOD_CHANGED`, `{"jsonrpc":"2.0","id":1,"error":{"code":-32024,"message":"CREDIT_PERIOD_CHANGED"},"result":{}}`, `{"jsonrpc":"2.0","id":99,"error":{"code":-32024,"message":"CREDIT_PERIOD_CHANGED"}}`, `{"jsonrpc":"2.0","id":1,"result":{"id":"job","state":"invalid"}}`} {
		t.Run(body, func(t *testing.T) {
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
			defer s.Close()
			c, e := NewClient(Configuration{Endpoint: s.URL, QuoteSigner: common.HexToAddress("1"), FeeRecipient: common.HexToAddress("2")}, Options{HTTPClient: s.Client()})
			if e != nil {
				t.Fatal(e)
			}
			_, e = c.SubmitPrepared(context.Background(), Submission{})
			var refusal *SubmissionRefusedError
			if e == nil || errors.As(e, &refusal) {
				t.Fatalf("unsafe malformed response classification: %v", e)
			}
		})
	}
	var refusal *SubmissionRefusedError
	for _, e := range []error{errors.New("CREDIT_PERIOD_CHANGED"), context.DeadlineExceeded, context.Canceled} {
		if errors.As(submissionRefusal(e), &refusal) {
			t.Fatalf("unsafe error classification: %v", e)
		}
	}
}
