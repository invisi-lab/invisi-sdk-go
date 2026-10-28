package ipp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// strPtr is a helper to create *string values in tests.
func strPtr(s string) *string { return &s }

// newTestServer creates an httptest.Server that serves a fixed JSON response
// at the given path with the given method.
func newTestServer(t *testing.T, method, path string, status int, resp interface{}) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			t.Errorf("expected method %s, got %s", method, r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(resp)
	})
	return httptest.NewServer(mux)
}

func TestClientInfo(t *testing.T) {
	want := InfoResponse{
		MerchantName: strPtr("Coffee Shop"),
		Amount:       strPtr("100000000"),
	}
	srv := newTestServer(t, "POST", "/ipp/v1/info", 200, want)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	got, err := client.Info(InfoRequest{IdentityCommitment: "0xabc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MerchantName == nil || *got.MerchantName != "Coffee Shop" {
		t.Errorf("merchantName = %v, want Coffee Shop", got.MerchantName)
	}
	if got.Amount == nil || *got.Amount != "100000000" {
		t.Errorf("amount = %v, want 100000000", got.Amount)
	}
}

func TestClientPropose(t *testing.T) {
	want := ProposeResponse{
		Status:         "accepted",
		Nonce:          strPtr("0x42"),
		NoteCommitment: strPtr("0xdef"),
	}
	srv := newTestServer(t, "POST", "/ipp/v1/propose", 200, want)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	got, err := client.Propose(ProposeRequest{
		IdentityCommitment: "0xabc",
		Amount:             "100000000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "accepted" {
		t.Errorf("status = %s, want accepted", got.Status)
	}
	if got.Nonce == nil || *got.Nonce != "0x42" {
		t.Errorf("nonce = %v, want 0x42", got.Nonce)
	}
}

func TestClientNotifySubmitted(t *testing.T) {
	want := NotifyResponse{Status: "received"}
	srv := newTestServer(t, "POST", "/ipp/v1/notify", 200, want)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	got, err := client.Notify(NotifyRequest{
		Commitment: "0xabc",
		Action:     "submitted",
		TxHash:     strPtr("0xtx"),
		Network:    strPtr("eip155:84532"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "received" {
		t.Errorf("status = %s, want received", got.Status)
	}
}

func TestClientNotifyCancelled(t *testing.T) {
	want := NotifyResponse{Status: "received"}
	srv := newTestServer(t, "POST", "/ipp/v1/notify", 200, want)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	got, err := client.Notify(NotifyRequest{
		Commitment: "0xabc",
		Action:     "cancelled",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "received" {
		t.Errorf("status = %s, want received", got.Status)
	}
}

func TestClientStatus(t *testing.T) {
	want := StatusResponse{
		Status: "confirmed",
		TxHash: strPtr("0xtx"),
	}
	srv := newTestServer(t, "GET", "/ipp/v1/status/0xpayment", 200, want)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	got, err := client.Status("0xpayment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "confirmed" {
		t.Errorf("status = %s, want confirmed", got.Status)
	}
	if got.TxHash == nil || *got.TxHash != "0xtx" {
		t.Errorf("txHash = %v, want 0xtx", got.TxHash)
	}
}

func TestClientRelaySignatureRejected(t *testing.T) {
	client := NewIPPClient("http://unused.invalid")
	if _, err := client.Relay(RelayRequest{Mode: "signature", Operation: RelayShield}); err == nil {
		t.Fatal("retired signature mode accepted")
	}
}

func TestClientRelayCalldata(t *testing.T) {
	want := RelayResponse{
		Status:  "submitted",
		RelayID: strPtr("relay-456"),
		TxHash:  strPtr("0xtx"),
	}
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/ipp/v1/relay" {
			t.Errorf("path = %s, want /ipp/v1/relay", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	got, err := client.Relay(RelayRequest{
		Mode:            "calldata",
		Operation:       RelayShield,
		ProtocolAddress: "0x1234567890123456789012345678901234567890",
		Network:         "eip155:84532",
		Calldata:        "0xdeadbeef",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "submitted" {
		t.Errorf("status = %s, want submitted", got.Status)
	}
	if got.TxHash == nil || *got.TxHash != "0xtx" {
		t.Errorf("txHash = %v, want 0xtx", got.TxHash)
	}
	if gotBody["mode"] != "calldata" {
		t.Errorf("body.mode = %v, want calldata", gotBody["mode"])
	}
	if gotBody["calldata"] != "0xdeadbeef" {
		t.Errorf("body.calldata = %v, want 0xdeadbeef", gotBody["calldata"])
	}
	// signature-mode fields must be omitted for a calldata-mode request.
	if _, present := gotBody["signature"]; present {
		t.Errorf("body.signature present for calldata mode: %v", gotBody["signature"])
	}
	if _, present := gotBody["tree_id"]; present {
		t.Errorf("body.tree_id present for calldata mode: %v", gotBody["tree_id"])
	}
}

// rawHandlerServer serves a fixed status + raw body, bypassing JSON encoding so
// tests can exercise non-envelope and empty error bodies.
func rawHandlerServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestClientHTTPError(t *testing.T) {
	// The gateway wraps 4xx/5xx in the nested Stripe envelope; a 400 with an
	// invalid_request_error must decode to *InvalidRequestError with its fields.
	body := `{"error":{"type":"invalid_request_error","code":"invalid_owner_identity_commitment","message":"owner_identity_commitment must be bytes32","param":"owner_identity_commitment","request_id":"req_123","details":{"hint":"len"}}}`
	srv := rawHandlerServer(t, 400, body)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	_, err := client.Info(InfoRequest{IdentityCommitment: "0xbad"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// errors.As must unwrap to both the concrete subclass and the base *IPPError.
	var invErr *InvalidRequestError
	if !errors.As(err, &invErr) {
		t.Fatalf("expected *InvalidRequestError, got %T: %v", err, err)
	}
	var base *IPPError
	if !errors.As(err, &base) {
		t.Fatalf("expected errors.As to reach *IPPError, got %T", err)
	}
	if base.StatusCode != 400 {
		t.Errorf("statusCode = %d, want 400", base.StatusCode)
	}
	if base.Code != "invalid_owner_identity_commitment" {
		t.Errorf("code = %q, want invalid_owner_identity_commitment", base.Code)
	}
	if base.Param != "owner_identity_commitment" {
		t.Errorf("param = %q, want owner_identity_commitment", base.Param)
	}
	if base.RequestID != "req_123" {
		t.Errorf("requestID = %q, want req_123", base.RequestID)
	}
	if base.Type != "invalid_request_error" {
		t.Errorf("type = %q, want invalid_request_error", base.Type)
	}
	if got, ok := base.Details["hint"].(string); !ok || got != "len" {
		t.Errorf("details[hint] = %v, want len", base.Details["hint"])
	}
	if len(base.Raw) == 0 {
		t.Error("Raw should retain the unparsed body")
	}
}

// TestClientErrorSubclasses checks that each envelope type / status routes to
// the right typed subclass, and that 5xx and unknown types fold into *ApiError.
func TestClientErrorSubclasses(t *testing.T) {
	cases := []struct {
		name   string
		status int
		typ    string
		assert func(t *testing.T, err error)
	}{
		{"notFound", 404, "not_found_error", func(t *testing.T, err error) {
			var e *NotFoundError
			if !errors.As(err, &e) {
				t.Fatalf("want *NotFoundError, got %T", err)
			}
		}},
		{"conflict", 409, "conflict_error", func(t *testing.T, err error) {
			var e *ConflictError
			if !errors.As(err, &e) {
				t.Fatalf("want *ConflictError, got %T", err)
			}
		}},
		{"rateLimit", 429, "rate_limit_error", func(t *testing.T, err error) {
			var e *RateLimitError
			if !errors.As(err, &e) {
				t.Fatalf("want *RateLimitError, got %T", err)
			}
		}},
		{"precondition412", 412, "invalid_request_error", func(t *testing.T, err error) {
			var e *InvalidRequestError
			if !errors.As(err, &e) {
				t.Fatalf("want *InvalidRequestError, got %T", err)
			}
		}},
		{"serverError", 500, "api_error", func(t *testing.T, err error) {
			var e *ApiError
			if !errors.As(err, &e) {
				t.Fatalf("want *ApiError, got %T", err)
			}
		}},
		// Unknown type on a 503 must fold into the *ApiError catch-all.
		{"unknownTypeCatchAll", 503, "some_future_error", func(t *testing.T, err error) {
			var e *ApiError
			if !errors.As(err, &e) {
				t.Fatalf("want *ApiError catch-all, got %T", err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"error":{"type":"` + tc.typ + `","code":"x","message":"m","request_id":"r"}}`
			srv := rawHandlerServer(t, tc.status, body)
			defer srv.Close()
			client := NewIPPClient(srv.URL)
			_, err := client.Info(InfoRequest{IdentityCommitment: "0x1"})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			tc.assert(t, err)
		})
	}
}

// TestClientNonJSONErrorBody: a 4xx/5xx body that is not the JSON envelope (an
// HTML proxy page) must surface as *ConnectionError, never a typed envelope error.
func TestClientNonJSONErrorBody(t *testing.T) {
	srv := rawHandlerServer(t, 502, "<html><body>Bad Gateway</body></html>")
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	_, err := client.Info(InfoRequest{IdentityCommitment: "0x1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("want *ConnectionError, got %T: %v", err, err)
	}
	if connErr.StatusCode != 502 {
		t.Errorf("statusCode = %d, want 502", connErr.StatusCode)
	}
	var base *IPPError
	if errors.As(err, &base) {
		t.Errorf("non-JSON body must NOT decode to a typed envelope error, got %T", err)
	}
}

// TestClientEmptyErrorBody: an empty 4xx/5xx body is not a parseable envelope
// and must downgrade to *ConnectionError.
func TestClientEmptyErrorBody(t *testing.T) {
	srv := rawHandlerServer(t, 500, "")
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	_, err := client.Info(InfoRequest{IdentityCommitment: "0x1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("want *ConnectionError, got %T: %v", err, err)
	}
}

// TestClientTransportError: when the round-trip itself fails (server closes the
// connection mid-flight), the SDK returns *ConnectionError with no status.
func TestClientTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hijack and close the connection without writing a valid HTTP response.
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("ResponseWriter does not support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("hijack: %v", err)
		}
		conn.Close()
	}))
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	_, err := client.Info(InfoRequest{IdentityCommitment: "0x1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("want *ConnectionError, got %T: %v", err, err)
	}
	if connErr.StatusCode != 0 {
		t.Errorf("statusCode = %d, want 0 for a round-trip failure", connErr.StatusCode)
	}
}

// TestClientProposeRejected: an HTTP 200 business rejection is a normal return
// value carrying RejectCode, NOT a thrown error (channel ②).
func TestClientProposeRejected(t *testing.T) {
	want := ProposeResponse{
		Status:     "rejected",
		Reason:     strPtr("draw amount exceeds remainder"),
		RejectCode: strPtr(string(RejectInsufficientCreditBalance)),
	}
	srv := newTestServer(t, "POST", "/ipp/v1/propose", 200, want)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	got, err := client.Propose(ProposeRequest{IdentityCommitment: "0xabc", Amount: "1"})
	if err != nil {
		t.Fatalf("a 200 rejection must NOT be a thrown error, got: %v", err)
	}
	if got.Status != "rejected" {
		t.Errorf("status = %s, want rejected", got.Status)
	}
	if got.RejectCode == nil || *got.RejectCode != "insufficient_credit_balance" {
		t.Errorf("rejectCode = %v, want insufficient_credit_balance", got.RejectCode)
	}
}

func TestClientAuthorizeRejected(t *testing.T) {
	want := AuthorizeResponse{Status: "rejected", RejectCode: strPtr("merchant_policy_rejected")}
	srv := newTestServer(t, "POST", "/ipp/v1/authorize", 200, want)
	defer srv.Close()
	got, err := NewIPPClient(srv.URL).Authorize(AuthorizeRequest{SpenderIdentityCommitment: "0x1", OwnerIdentityCommitment: "0x2", Amount: "1", Expiry: "10"})
	if err != nil || got.RejectCode == nil || *got.RejectCode != "merchant_policy_rejected" {
		t.Fatalf("business rejection: %+v, %v", got, err)
	}
}

// TestClientRelayStatusFailed: an async relay failure surfaces as a result-state
// field on RelayStatusResponse (channel ③), read as a value, not thrown.
func TestClientRelayStatusFailed(t *testing.T) {
	failed := RelayErrorMerchantDisabledFixture()
	srv := newTestServer(t, "GET", "/ipp/v1/relay/relay-1", 200, failed)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	got, err := client.RelayStatus("relay-1")
	if err != nil {
		t.Fatalf("relay failure is a result-state, not a thrown error, got: %v", err)
	}
	if got.Status != "failed" {
		t.Errorf("status = %s, want failed", got.Status)
	}
	if got.ErrorCode == nil || *got.ErrorCode != RelayMerchantDisabled {
		t.Errorf("errorCode = %v, want relay_merchant_disabled", got.ErrorCode)
	}
}

// RelayErrorMerchantDisabledFixture builds a failed relay-status body with the
// lowercase wire code (regression guard against the old UPPER_SNAKE values).
func RelayErrorMerchantDisabledFixture() RelayStatusResponse {
	code := RelayMerchantDisabled
	return RelayStatusResponse{Status: "failed", ErrorCode: &code}
}

func TestClientTrailingSlash(t *testing.T) {
	// Ensure trailing slashes in the endpoint are handled correctly.
	client := NewIPPClient("https://pay.example.com/")
	if client.baseURL != "https://pay.example.com/ipp/v1" {
		t.Errorf("baseURL = %s, want https://pay.example.com/ipp/v1", client.baseURL)
	}
}
