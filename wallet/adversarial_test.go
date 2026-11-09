package wallet

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestAcceptRejectsAlteredCryptographicResultsAndContexts(t *testing.T) {
	raw, err := os.ReadFile("../internal/testsupport/testdata/wallet-wire-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Request  Request  `json:"request"`
			Response Response `json:"response"`
		} `json:"cases"`
	}
	json.Unmarshal(raw, &fixture)
	var public PublicResult
	json.Unmarshal(fixture.Cases[0].Response.Result, &public)
	for _, c := range fixture.Cases {
		t.Run(string(c.Request.Method), func(t *testing.T) {
			var result map[string]any
			json.Unmarshal(c.Response.Result, &result)
			switch c.Request.Method {
			case Public:
				result["spendingPublicKey"] = "0x04" + strings.Repeat("00", 64)
			case Detect:
				result["matches"] = []any{nil}
			case DecryptNote, DecryptCredit:
				var p DecryptParams
				json.Unmarshal(c.Request.Params, &p)
				p.Context.TokenAddress = "1"
				c.Request.Params = Encode(p)
			case Authorize:
				result["signature"] = "0x" + strings.Repeat("00", 64)
			}
			c.Response.Result = Encode(result)
			if _, err := Accept(c.Request, c.Response, public.ViewingPublicKey); err == nil {
				t.Fatal("accepted altered cryptographic response or context")
			}
		})
	}
}
func TestDetectRejectsNullBoolean(t *testing.T) {
	q := Request{Version: 1, ID: "null", OwnerAddress: "0x1111111111111111111111111111111111111111", Method: Detect, Params: Encode(DetectParams{Clues: []string{}})}
	r := Response{Version: 1, ID: q.ID, OwnerAddress: q.OwnerAddress, Method: Detect, Result: json.RawMessage(`{"matches":[null]}`)}
	if _, err := Accept(q, r, ""); err == nil {
		t.Fatal("accepted null as boolean")
	}
}
func TestDecodeRejectsTrailingMalformedJSON(t *testing.T) {
	var value struct{}
	if err := Decode(json.RawMessage(`{} trailing`), &value); err == nil {
		t.Fatal("accepted trailing garbage")
	}
}

func TestAcceptRejectsMalformedDeliveryRequest(t *testing.T) {
	raw, err := os.ReadFile("../internal/testsupport/testdata/wallet-wire-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Request  Request  `json:"request"`
			Response Response `json:"response"`
		} `json:"cases"`
	}
	json.Unmarshal(raw, &fixture)
	var public PublicResult
	json.Unmarshal(fixture.Cases[0].Response.Result, &public)
	for _, c := range fixture.Cases {
		if c.Request.Method == DecryptNote || c.Request.Method == DecryptCredit {
			var p DecryptParams
			json.Unmarshal(c.Request.Params, &p)
			p.Envelope = "0x"
			c.Request.Params = Encode(p)
			if _, err := Accept(c.Request, c.Response, public.ViewingPublicKey); err == nil {
				t.Fatal("accepted truncated envelope request", c.Request.Method)
			}
		}
	}
}
