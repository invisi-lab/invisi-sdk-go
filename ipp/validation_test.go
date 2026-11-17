package ipp

import (
	"encoding/json"
	"strings"
	"testing"
)

const address = "0x1111111111111111111111111111111111111111"
const protocolAddress = "0x3333333333333333333333333333333333333333"
const origin = "https://connect.invisipay.fi"
const hash = "0x0101010101010101010101010101010101010101010101010101010101010101"

func sessionJSON() []byte {
	return []byte(`{"ippVersion":1,"type":"payment_session","mode":"payment","sessionId":"ses_test","status":"open","merchant":{"name":"Store"},"options":[{"id":"base_usdc","method":"invisi","network":"eip155:8453","asset":{"address":"` + address + `","symbol":"USDC","decimals":6},"amount":"10000000"}],"activeRequestId":null,"createdAt":"2026-11-17T08:00:00Z","expiresAt":"2026-11-17T08:05:00Z","links":{"confirm":{"href":"` + origin + `/custom/confirm","method":"POST"},"request":null},"failureReason":null,"receipt":null,"evidenceStatus":null,"exceptionSummary":{"count":0,"unresolvedCount":0}}`)
}
func requestJSON() []byte {
	return []byte(`{"ippVersion":1,"type":"payment_request","mode":"payment","method":"invisi","sessionId":"ses_test","requestId":"req_test","optionId":"base_usdc","status":"requires_action","payment":{"network":"eip155:8453","protocolAddress":"` + protocolAddress + `","tokenAddress":"` + address + `","amount":"10000000","recipient":{"type":"identity","identity":"` + hash + `"}},"action":{"type":"invisi_transaction"},"createdAt":"2026-11-17T08:02:00Z","expiresAt":"2026-11-17T08:20:00Z","links":{"credentials":{"href":"` + origin + `/custom/hints","method":"POST"}},"failureReason":null}`)
}
func TestValidationAndIndependentWindows(t *testing.T) {
	s, e := ParsePaymentSession(sessionJSON(), origin)
	if e != nil {
		t.Fatal(e)
	}
	r, e := ParsePaymentRequest(requestJSON(), origin)
	if e != nil {
		t.Fatal(e)
	}
	if e = AssertRequestMatchesSession(r, s); e != nil {
		t.Fatal(e)
	}
	r.Payment.Amount = "2"
	if AssertRequestMatchesSession(r, s) == nil {
		t.Fatal("mismatched amount accepted")
	}
}
func TestStrictMessages(t *testing.T) {
	tests := []string{
		strings.Replace(string(sessionJSON()), `"ippVersion":1`, `"ippVersion":1,"ippVersion":1`, 1),
		strings.Replace(string(sessionJSON()), `"ippVersion":1`, `"ippVersion":1.5`, 1),
		strings.Replace(string(sessionJSON()), `"ippVersion":1`, `"ippVersion":2.0`, 1),
		strings.Replace(string(sessionJSON()), `"count":0`, `"count":-1`, 1),
		strings.Replace(string(sessionJSON()), `"name":"Store"`, `"name":"\ud800"`, 1),
		strings.Replace(string(sessionJSON()), `"name":"Store"`, `"name":"Store","unknown":1`, 1),
		strings.Replace(string(sessionJSON()), `"amount":"10000000"`, `"amount":"01"`, 1),
		strings.Replace(string(sessionJSON()), `"decimals":6`, `"decimals":256`, 1),
		strings.Replace(string(sessionJSON()), `"activeRequestId":null,`, "", 1),
		strings.Replace(string(sessionJSON()), `https://connect.invisipay.fi/custom/confirm`, `https://evil.test/custom/confirm`, 1),
	}
	for i, data := range tests {
		if _, e := ParsePaymentSession([]byte(data), origin); e == nil {
			t.Errorf("accepted invalid case %d", i)
		}
	}
}
func TestCheckoutAndUnsafeURLs(t *testing.T) {
	u, e := CreateCheckoutURL(origin+"/custom/session?route=one", "EXAMPLE_SESSION_TOKEN", "https://dapp.invisipay.fi/checkout")
	if e != nil {
		t.Fatal(e)
	}
	entry, e := ParseCheckoutEntry(u)
	if e != nil || entry.Token != "EXAMPLE_SESSION_TOKEN" {
		t.Fatal(e)
	}
	for _, u := range []string{"http://x.test/", "https://X.test/", "https://x.test", "https://x.test:443/", "https://x.test/a/../b", "https://x.test/?api_key=x", "https://x.test/#token=x", "https://x.test/%zz", "https://u@x.test/"} {
		if ValidateAPIURL(u, "") == nil {
			t.Error("accepted", u)
		}
	}
	c := Checkout{IPPVersion: 1, Type: "checkout", SessionURL: entry.SessionURL, CheckoutURL: u, ExpiresAt: "2026-11-17T08:05:00Z"}
	b, _ := json.Marshal(c)
	if _, e = ParseCheckout(b); e != nil {
		t.Fatal(e)
	}
}

func TestURLCanonicalHostAndEscapes(t *testing.T) {
	for _, u := range []string{"https://[0:0:0:0:0:0:0:1]/a", "https://127.1/a", "https://x.test/?v='", "https://x.test/%ff", "https://x.test:/a"} {
		if ValidateAPIURL(u, "") == nil {
			t.Error("noncanonical URL accepted", u)
		}
	}
	for _, u := range []string{"https://[::1]/a", "https://127.0.0.1/a", "https://x.test:0/a", "https://x.test/?route=a;b"} {
		if ValidateAPIURL(u, "") != nil {
			t.Error("canonical URL rejected", u)
		}
	}
}

func TestTypedInputDoesNotRepairInvalidUnicode(t *testing.T) {
	s, _ := ParsePaymentSession(sessionJSON(), origin)
	bad := string([]byte{255})
	input := CheckoutInput{IPPVersion: 1, Mode: Payment, Options: s.Options, ExpiresAt: s.ExpiresAt, Description: &bad}
	if _, e := EncodeJSON(input); e == nil {
		t.Fatal("invalid UTF8 silently replaced")
	}
	r, _ := ParsePaymentRequest(requestJSON(), origin)
	r.SessionID = bad
	if _, e := EncodeJSON(r); e == nil {
		t.Fatal("invalid request identity repaired")
	}
}
