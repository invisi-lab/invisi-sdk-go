package wallet

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAcceptNativeResponsesAndRejectTampering(t *testing.T) {
	raw, e := os.ReadFile("../native/testdata/wallet-wire-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Cases []struct {
			Request  Request  `json:"request"`
			Response Response `json:"response"`
		} `json:"cases"`
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	var public PublicResult
	json.Unmarshal(f.Cases[0].Response.Result, &public)
	for _, c := range f.Cases {
		t.Run(string(c.Request.Method), func(t *testing.T) {
			if _, e := Accept(c.Request, c.Response, public.ViewingPublicKey); e != nil {
				t.Fatal(e)
			}
			bad := c.Response
			bad.ID = "replayed"
			if _, e := Accept(c.Request, bad, public.ViewingPublicKey); e == nil {
				t.Fatal("accepted replay")
			}
		})
	}
	for _, c := range f.Cases {
		if c.Request.Method == Authorize {
			var r AuthorizeResult
			json.Unmarshal(c.Response.Result, &r)
			r.Digest = "0x" + string(make([]byte, 64))
			c.Response.Result = Encode(r)
			if _, e := Accept(c.Request, c.Response, public.ViewingPublicKey); e == nil {
				t.Fatal("accepted wrong digest")
			}
		}
	}
}
