package ipp

import (
	"fmt"
	"strings"
	"testing"
)

func TestSensitiveDiagnosticsRedacted(t *testing.T) {
	c, _ := NewClient(ClientOptions{SessionURL: origin + "/session", Token: "SECRET_SESSION_TOKEN"})
	m, _ := NewMerchantClient(MerchantClientOptions{CreateURL: origin + "/create", Token: "SECRET_MERCHANT_TOKEN"})
	values := []any{c, *c, m, *m, Checkout{CheckoutURL: "https://dapp.test/#token=SECRET_SESSION_TOKEN"}, &CheckoutEntry{Token: "SECRET_SESSION_TOKEN"}, CreditDetails{CreditSharedKey: hash}, &Request{Credit: &CreditDetails{CreditSharedKey: hash}}}
	for _, v := range values {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			s := fmt.Sprintf(format, v)
			for _, secret := range []string{"SECRET_SESSION_TOKEN", "SECRET_MERCHANT_TOKEN", hash} {
				if strings.Contains(s, secret) {
					t.Fatalf("%T %s exposed secret", v, format)
				}
			}
		}
	}
}
