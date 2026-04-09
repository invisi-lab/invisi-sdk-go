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
