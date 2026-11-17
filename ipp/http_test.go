package ipp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(code int, b []byte) *http.Response {
	return &http.Response{StatusCode: code, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(string(b)))}
}
func TestSessionHTTPAndRecovery(t *testing.T) {
	n := 0
	h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		n++
		if r.Header.Get("Authorization") != "Bearer SESSION_TOKEN" || r.Header.Get("Cookie") != "" {
			t.Fatal("wrong authorization")
		}
		switch n {
		case 1:
			if r.Method != "GET" {
				t.Fatal(r.Method)
			}
			return response(200, sessionJSON()), nil
		case 2:
			if r.Method != "POST" || r.URL.Path != "/custom/confirm" || r.Header.Get("Idempotency-Key") != "confirm-key" {
				t.Fatal("wrong confirm")
			}
			res := response(201, requestJSON())
			res.Header.Set("Location", origin+"/chosen/request")
			return res, nil
		case 3:
			if r.URL.Path != "/chosen/request" || r.Method != "GET" {
				t.Fatal("wrong recovery URL")
			}
			return response(200, requestJSON()), nil
		}
		t.Fatal("unexpected request")
		return nil, nil
	})}
	c, e := NewClient(ClientOptions{SessionURL: origin + "/custom/session", Token: "SESSION_TOKEN", HTTPClient: h})
	if e != nil {
		t.Fatal(e)
	}
	s, e := c.ReadSession(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	s.Options[0].Amount = "99"
	if _, e = c.Confirm(context.Background(), "base_usdc", ""); e == nil {
		t.Fatal("missing key")
	}
	if _, e = c.Confirm(context.Background(), "base_usdc", "confirm-key"); e != nil {
		t.Fatal(e)
	}
	if _, e = c.ReadRequest(context.Background()); e != nil {
		t.Fatal(e)
	}
	if n != 3 {
		t.Fatal(n)
	}
}
func TestHTTPRejectsUnsafeResponseAndSanitizes(t *testing.T) {
	for _, test := range []struct {
		name string
		res  *http.Response
	}{{"redirect", response(302, sessionJSON())}, {"oversize", response(200, []byte(strings.Repeat("x", MaxResponseBytes+1)))}, {"error", response(503, []byte(`{"ippVersion":1,"type":"error","error":{"code":"unavailable","message":"SESSION_TOKEN private","retryable":true}}`))}} {
		t.Run(test.name, func(t *testing.T) {
			h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) { return test.res, nil })}
			c, e := NewClient(ClientOptions{SessionURL: origin + "/custom/session", Token: "SESSION_TOKEN", HTTPClient: h})
			if e != nil {
				t.Fatal(e)
			}
			_, e = c.ReadSession(context.Background())
			if e == nil || strings.Contains(e.Error(), "SESSION_TOKEN") || strings.Contains(e.Error(), "private") {
				t.Fatal(e)
			}
		})
	}
}
func TestPollCancellableGETOnlyAndRetryAfter(t *testing.T) {
	n := 0
	h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		n++
		if r.Method != "GET" {
			t.Fatal("poll mutated")
		}
		res := response(200, sessionJSON())
		res.Header.Set("Retry-After", "5")
		return res, nil
	})}
	c, _ := NewClient(ClientOptions{SessionURL: origin + "/custom/session", Token: "SESSION_TOKEN", HTTPClient: h})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	_, e := c.WaitForResult(ctx, WaitOptions{Interval: time.Millisecond, Timeout: time.Second})
	if e == nil || n != 1 {
		t.Fatalf("%v polls=%d", e, n)
	}
	var ie *Error
	if !errors.As(e, &ie) || ie.Code != "aborted" {
		t.Fatal(e)
	}
}
func TestImmutableTermsAndTerminalReceipt(t *testing.T) {
	s, _ := ParsePaymentSession(sessionJSON(), origin)
	n := 0
	h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		n++
		out := clone(s)
		if n == 2 {
			out.Options[0].Amount = "2"
		}
		b, _ := json.Marshal(out)
		return response(200, b), nil
	})}
	c, _ := NewClient(ClientOptions{SessionURL: origin + "/custom/session", Token: "SESSION_TOKEN", HTTPClient: h})
	if _, e := c.ReadSession(context.Background()); e != nil {
		t.Fatal(e)
	}
	if _, e := c.ReadSession(context.Background()); e == nil {
		t.Fatal("terms mutation accepted")
	}
}
func TestMerchantReplayAndPaymentRequired(t *testing.T) {
	s, _ := ParsePaymentSession(sessionJSON(), origin)
	u, _ := CreateCheckoutURL(origin+"/session", "SESSION_TOKEN", "https://dapp.invisipay.fi/checkout")
	co := &Checkout{IPPVersion: 1, Type: "checkout", SessionURL: origin + "/session", CheckoutURL: u, ExpiresAt: s.ExpiresAt}
	res, e := PaymentRequiredResponse(s, co)
	if e != nil {
		t.Fatal(e)
	}
	c, e := ClientFromPaymentRequired(res, HTTPOptions{})
	if e != nil || c == nil {
		t.Fatal(e)
	}
	input := CheckoutInput{IPPVersion: 1, Mode: Payment, Options: s.Options, ExpiresAt: s.ExpiresAt}
	h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "merchant.test" || r.Method != "POST" || r.Header.Get("Idempotency-Key") != "create-key" {
			t.Fatal("wrong creation")
		}
		b, _ := json.Marshal(co)
		return response(201, b), nil
	})}
	m, e := NewMerchantClient(MerchantClientOptions{CreateURL: "https://merchant.test/create", Token: "MERCHANT_TOKEN", HTTPOptions: HTTPOptions{HTTPClient: h}})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.CreateCheckout(context.Background(), input, "create-key"); e != nil {
		t.Fatal(e)
	}
}

func TestCancellationAndDeclaredResponseBound(t *testing.T) {
	h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })}
	c, _ := NewClient(ClientOptions{SessionURL: origin + "/session", Token: "SESSION_TOKEN", HTTPClient: h})
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, e := c.ReadSession(ctx)
	var ie *Error
	if !errors.As(e, &ie) || ie.Code != "aborted" {
		t.Fatalf("cancellation: %v", e)
	}
	for _, length := range []string{"1048577", "garbage"} {
		res := response(200, sessionJSON())
		res.Header.Set("Content-Length", length)
		if _, e = readResponse(res); e == nil {
			t.Fatal("accepted declared length", length)
		}
	}
}
func TestErrorStringNeverReflectsRemoteCode(t *testing.T) {
	if strings.Contains((&Error{Code: "private_secret"}).Error(), "private_secret") {
		t.Fatal("untrusted error code echoed")
	}
}
func TestMetadataNegativeZero(t *testing.T) {
	data := strings.Replace(string(sessionJSON()), `"count":0`, `"count":-0`, 1)
	if _, e := ParsePaymentSession([]byte(data), origin); e != nil {
		t.Fatal(e)
	}
}

func TestTerminalStateAndReceiptImmutable(t *testing.T) {
	s, _ := ParsePaymentSession(sessionJSON(), origin)
	id := "req_test"
	evidence := "confirmed"
	s.ActiveRequestID = &id
	s.Status = "succeeded"
	s.EvidenceStatus = &evidence
	s.Links.Request = &Link{Href: origin + "/read", Method: "GET"}
	s.Receipt = &Receipt{Network: "eip155:8453", ProtocolAddress: protocolAddress, TransactionHash: hash, Commitment: hash, BlockNumber: "38000000", BlockHash: hash, CompletedAt: "2026-11-17T08:03:00Z"}
	for _, change := range []func(*PaymentSession){func(v *PaymentSession) { v.Status = "processing"; v.Receipt = nil; v.EvidenceStatus = nil }, func(v *PaymentSession) { v.Receipt.BlockNumber = "38000001" }} {
		n := 0
		h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			n++
			v := clone(s)
			if n > 1 {
				change(v)
			}
			b, _ := json.Marshal(v)
			return response(200, b), nil
		})}
		c, _ := NewClient(ClientOptions{SessionURL: origin + "/session", Token: "SESSION_TOKEN", HTTPClient: h})
		if _, e := c.ReadSession(context.Background()); e != nil {
			t.Fatal(e)
		}
		if _, e := c.ReadSession(context.Background()); e == nil {
			t.Fatal("terminal state or receipt changed")
		}
	}
}
func TestCredentialOptionalKeyAndAssociation(t *testing.T) {
	n := 0
	h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		n++
		switch n {
		case 1:
			return response(200, sessionJSON()), nil
		case 2:
			res := response(201, requestJSON())
			res.Header.Set("Location", origin+"/request")
			return res, nil
		case 3:
			if r.URL.Path != "/custom/hints" || r.Method != "POST" || r.Header.Get("Idempotency-Key") != "" {
				t.Fatal("credential endpoint/key")
			}
			return response(202, []byte(`{"ippVersion":1,"type":"credential_acknowledgement","sessionId":"ses_test","requestId":"req_test","status":"accepted"}`)), nil
		}
		t.Fatal("unexpected call")
		return nil, nil
	})}
	c, _ := NewClient(ClientOptions{SessionURL: origin + "/session", Token: "SESSION_TOKEN", HTTPClient: h})
	if _, e := c.Confirm(context.Background(), "base_usdc", "confirm-key"); e != nil {
		t.Fatal(e)
	}
	credential := Credential{IPPVersion: 1, Type: "credential", Mode: Payment, Method: "invisi", SessionID: "wrong", RequestID: "req_test", Proof: Proof{Type: "note", TransactionHash: hash, Commitment: hash}}
	if _, e := c.SubmitCredential(context.Background(), credential, ""); e == nil {
		t.Fatal("unrelated credential accepted")
	}
	credential.SessionID = "ses_test"
	if _, e := c.SubmitCredential(context.Background(), credential, ""); e != nil {
		t.Fatal(e)
	}
	if n != 3 {
		t.Fatal(n)
	}
}
func TestRetryableErrorPollingRecoversByGET(t *testing.T) {
	n := 0
	h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		n++
		if r.Method != "GET" {
			t.Fatal("poll mutated")
		}
		if n == 1 {
			res := response(503, []byte(`{"ippVersion":1,"type":"error","error":{"code":"busy","message":"private SESSION_TOKEN","retryable":true}}`))
			res.Header.Set("Retry-After", "0")
			return res, nil
		}
		s, _ := ParsePaymentSession(sessionJSON(), origin)
		s.Status = "failed"
		reason := "expired"
		s.FailureReason = &reason
		b, _ := json.Marshal(s)
		return response(200, b), nil
	})}
	c, _ := NewClient(ClientOptions{SessionURL: origin + "/session", Token: "SESSION_TOKEN", HTTPClient: h})
	s, e := c.WaitForResult(context.Background(), WaitOptions{Timeout: time.Second})
	if e != nil || s.Status != "failed" || n != 2 {
		t.Fatal(s, e, n)
	}
}

func creditMessages(t *testing.T) (*PaymentSession, *Request) {
	t.Helper()
	data, e := os.ReadFile("testdata/ipp-wire-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var corpus struct {
		Cases []struct {
			Kind  string
			Valid bool
			JSON  string
		}
	}
	if e = json.Unmarshal(data, &corpus); e != nil {
		t.Fatal(e)
	}
	var session *PaymentSession
	var request *Request
	for _, v := range corpus.Cases {
		if !v.Valid {
			continue
		}
		switch v.Kind {
		case "session":
			candidate, e := ParsePaymentSession([]byte(v.JSON), origin)
			if e == nil && candidate.Mode == Credit && candidate.Status == "open" {
				session = candidate
			}
		case "request":
			candidate, e := ParsePaymentRequest([]byte(v.JSON), origin)
			if e == nil && candidate.Mode == Credit && candidate.Status == "requires_action" {
				request = candidate
			}
		}
		if session != nil && request != nil {
			break
		}
	}
	if session == nil || request == nil {
		t.Fatal("missing valid credit corpus fixtures")
	}
	return session, request
}
func TestRecoveryRejectsKnownCreditKeyInSessionLinks(t *testing.T) {
	for _, field := range []string{"request", "confirm"} {
		for _, encoded := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/encoded=%v", field, encoded), func(t *testing.T) {
				s, request := creditMessages(t)
				key := request.Credit.CreditSharedKey
				if encoded {
					key = "%30" + key[1:]
				}
				poisoned := clone(s)
				poisoned.ActiveRequestID = &request.RequestID
				poisoned.Status = request.Status
				poisoned.Links.Request = &Link{Href: origin + "/original", Method: "GET"}
				if field == "request" {
					poisoned.Links.Request.Href = origin + "/recovery/" + key
				} else {
					poisoned.Links.Confirm.Href = origin + "/confirm/" + key
				}
				calls := 0
				h := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					if strings.Contains(r.URL.Path, request.Credit.CreditSharedKey) {
						t.Fatal("private credit key was sent in URL")
					}
					switch calls {
					case 1:
						b, _ := json.Marshal(s)
						return response(200, b), nil
					case 2:
						b, _ := json.Marshal(request)
						res := response(201, b)
						res.Header.Set("Location", origin+"/original")
						return res, nil
					case 3:
						b, _ := json.Marshal(poisoned)
						return response(200, b), nil
					case 4:
						if r.Method != "GET" || r.URL.String() != origin+"/original" {
							t.Fatal("rejected links changed recovery URL")
						}
						b, _ := json.Marshal(request)
						return response(200, b), nil
					}
					t.Fatal("unexpected request")
					return nil, nil
				})}
				client, e := NewClient(ClientOptions{SessionURL: origin + "/session", Token: "SESSION_TOKEN", HTTPClient: h})
				if e != nil {
					t.Fatal(e)
				}
				if _, e = client.Confirm(context.Background(), request.OptionID, "confirm-key"); e != nil {
					t.Fatal(e)
				}
				if _, e = client.ReadSession(context.Background()); e == nil {
					t.Fatal("accepted private credit key in session link")
				}
				if _, e = client.ReadRequest(context.Background()); e != nil {
					t.Fatal(e)
				}
				if calls != 4 {
					t.Fatal(calls)
				}
			})
		}
	}
}
func TestNewCreditRequestRejectsPrivateKeyInExistingSessionLink(t *testing.T) {
	s, r := creditMessages(t)
	s.Links.Confirm.Href = origin + "/confirm/" + r.Credit.CreditSharedKey
	c, _ := NewClient(ClientOptions{SessionURL: origin + "/session", Token: "SESSION_TOKEN"})
	c.session = s
	b, _ := json.Marshal(r)
	if _, e := c.acceptRequest(b); e == nil {
		t.Fatal("new key did not screen existing link")
	}
}
