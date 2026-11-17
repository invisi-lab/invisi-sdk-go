package ipp

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSharedConformance(t *testing.T) {
	b, e := os.ReadFile("testdata/ipp-wire-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var corpus struct {
		Origin string `json:"origin"`
		Cases  []struct {
			Name  string `json:"name"`
			Kind  string `json:"kind"`
			Valid bool   `json:"valid"`
			JSON  string `json:"json"`
		} `json:"cases"`
	}
	if e = json.Unmarshal(b, &corpus); e != nil {
		t.Fatal(e)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var e error
			switch c.Kind {
			case "session":
				_, e = ParsePaymentSession([]byte(c.JSON), corpus.Origin)
			case "request":
				_, e = ParsePaymentRequest([]byte(c.JSON), corpus.Origin)
			case "checkout":
				_, e = ParseCheckout([]byte(c.JSON))
			case "input":
				_, e = ParseCheckoutInput([]byte(c.JSON))
			case "credential":
				_, e = ParseCredential([]byte(c.JSON), nil)
			default:
				t.Fatal(c.Kind)
			}
			if (e == nil) != c.Valid {
				t.Fatalf("valid=%v, error=%v", c.Valid, e)
			}
		})
	}
}
