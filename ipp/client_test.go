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
		Status:            "accepted",
		Nonce:             strPtr("0x42"),
		PaymentCommitment: strPtr("0xdef"),
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

func TestClientRelaySignature(t *testing.T) {
	want := RelayResponse{
		Status:  "queued",
		RelayID: strPtr("relay-123"),
	}
	// Capture the request body to assert the signature-mode fields land on the
	// wire under their snake_case names with the right JSON types (tree_id is a
	// number; sig_nonce / deadline are strings).
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
		Mode:               "signature",
		Operation:          "pay-open",
		Network:            "eip155:84532",
		Signature:          "0xsig",
		PaymentCommitment:  "0xpay",
		IdentityCommitment: "0xid",
		TokenAddress:       "0xusdc",
		Amount:             "100000000",
		Nonce:              "0x1",
		TreeID:             1,
		SigNonce:           "7",
		Deadline:           "1750000000",
		PayerAddress:       "0xpayer",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "queued" {
		t.Errorf("status = %s, want queued", got.Status)
	}
	if got.RelayID == nil || *got.RelayID != "relay-123" {
		t.Errorf("relayID = %v, want relay-123", got.RelayID)
	}
	if gotBody["mode"] != "signature" {
		t.Errorf("body.mode = %v, want signature", gotBody["mode"])
	}
	if gotBody["signature"] != "0xsig" {
		t.Errorf("body.signature = %v, want 0xsig", gotBody["signature"])
	}
	// tree_id is uint32 → JSON number (decodes into Go as float64).
	if tid, ok := gotBody["tree_id"].(float64); !ok || tid != 1 {
		t.Errorf("body.tree_id = %v (%T), want 1 (number)", gotBody["tree_id"], gotBody["tree_id"])
	}
	// sig_nonce / deadline are proto uint64 → JSON strings.
	if gotBody["sig_nonce"] != "7" {
		t.Errorf("body.sig_nonce = %v, want \"7\"", gotBody["sig_nonce"])
	}
	if gotBody["deadline"] != "1750000000" {
		t.Errorf("body.deadline = %v, want \"1750000000\"", gotBody["deadline"])
	}
	if gotBody["payer_address"] != "0xpayer" {
		t.Errorf("body.payer_address = %v, want 0xpayer", gotBody["payer_address"])
	}
	// calldata-mode field must be absent for a signature-mode request.
	if _, present := gotBody["calldata"]; present {
		t.Errorf("body.calldata present for signature mode: %v", gotBody["calldata"])
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
		Mode:      "calldata",
		Operation: "pay-open",
		Network:   "eip155:84532",
		Calldata:  "0xdeadbeef",
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

func TestClientHTTPError(t *testing.T) {
	errResp := ErrorResponse{
		Error:   "invalid_commitment",
		Message: strPtr("Bad commitment"),
	}
	srv := newTestServer(t, "POST", "/ipp/v1/info", 400, errResp)
	defer srv.Close()

	client := NewIPPClient(srv.URL)
	_, err := client.Info(InfoRequest{IdentityCommitment: "0xbad"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var ippErr *IPPError
	if !errors.As(err, &ippErr) {
		t.Fatalf("expected *IPPError, got %T: %v", err, err)
	}
	if ippErr.StatusCode != 400 {
		t.Errorf("statusCode = %d, want 400", ippErr.StatusCode)
	}
	if ippErr.Error() != "invalid_commitment" {
		t.Errorf("error = %s, want invalid_commitment", ippErr.Error())
	}
}

func TestClientTrailingSlash(t *testing.T) {
	// Ensure trailing slashes in the endpoint are handled correctly.
	client := NewIPPClient("https://pay.example.com/")
	if client.baseURL != "https://pay.example.com/ipp/v1" {
		t.Errorf("baseURL = %s, want https://pay.example.com/ipp/v1", client.baseURL)
	}
}
