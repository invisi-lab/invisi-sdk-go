package wallet

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBatchBindingAndErrors(t *testing.T) {
	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	a, _ := Prepare("a", owner, Detect, DetectParams{Clues: []string{}})
	b, _ := Prepare("b", owner, Detect, DetectParams{Clues: []string{}})
	response := func(q Request) Response {
		return Response{Version: 1, ID: q.ID, OwnerAddress: q.OwnerAddress, Method: q.Method, Result: Encode(DetectResult{Matches: []bool{}})}
	}
	ar, br := response(a), response(b)
	ordered, err := ValidateResponses([]Request{a, b}, []Response{br, ar})
	if err != nil || ordered[0].ID != "a" {
		t.Fatalf("reordered valid batch: %v", err)
	}
	for name, rs := range map[string][]Response{"missing": {ar}, "duplicate": {ar, ar}, "unknown": {ar, response(Request{ID: "unknown"})}} {
		t.Run(name, func(t *testing.T) {
			if _, err := ValidateResponses([]Request{a, b}, rs); err == nil {
				t.Fatal("accepted invalid response batch")
			}
		})
	}
	both := ar
	both.Error = &ResponseError{Code: "WALLET_CANCELLED"}
	if err := ValidateResponse(a, both); err == nil {
		t.Fatal("accepted result and error")
	}
	failure := ar
	failure.Result = nil
	failure.Error = &ResponseError{Code: "WALLET_CANCELLED"}
	if err := ValidateResponse(a, failure); err != nil {
		t.Fatal(err)
	}
	if _, err := Accept(a, failure, ""); err == nil {
		t.Fatal("accepted failed wallet response")
	}
	if err := ValidateRequests([]Request{a, a}); err == nil {
		t.Fatal("accepted duplicate requests")
	}
	other := b
	other.OwnerAddress = "0x2222222222222222222222222222222222222222"
	if err := ValidateRequests([]Request{a, other}); err == nil {
		t.Fatal("accepted mixed owners")
	}
	if err := ValidateRequests(make([]Request, 1025)); err == nil {
		t.Fatal("accepted oversized batch")
	}
	if err := ValidateRequests(nil); err != nil {
		t.Fatal(err)
	}
	huge := a
	huge.Params = json.RawMessage(`{"clues":[],"padding":"` + strings.Repeat("a", 16<<20) + `"}`)
	if err := ValidateRequests([]Request{huge}); err == nil {
		t.Fatal("accepted oversized transport")
	}
}

func TestResponseJSONRejectsAmbiguousOrSecretFields(t *testing.T) {
	for _, raw := range []string{
		`{"version":1,"id":"a","ownerAddress":"0x1111111111111111111111111111111111111111","method":"wallet.public","result":{},"error":null}`,
		`{"version":1,"id":"a","ownerAddress":"0x1111111111111111111111111111111111111111","method":"wallet.public","result":{},"privateKey":"secret"}`,
		`{"version":1,"id":"a","ownerAddress":"0x1111111111111111111111111111111111111111","method":"wallet.public","error":{"code":"WALLET_LOCKED","message":"private details"}}`,
	} {
		var r Response
		if err := json.Unmarshal([]byte(raw), &r); err == nil {
			t.Fatal("accepted ambiguous or unknown response JSON")
		}
	}
}

func TestBatchLimitsAndIndependentAcceptance(t *testing.T) {
	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	requests := make([]Request, 1024)
	for i := range requests {
		q, err := Prepare(fmt.Sprintf("item-%d", i), owner, Detect, DetectParams{Clues: []string{}})
		if err != nil {
			t.Fatal(err)
		}
		requests[i] = q
	}
	if err := ValidateRequests(requests); err != nil {
		t.Fatal(err)
	}
	q := requests[0]
	good := Response{Version: 1, ID: q.ID, OwnerAddress: q.OwnerAddress, Method: q.Method, Result: Encode(DetectResult{Matches: []bool{}})}
	if values, err := AcceptBatch([]Request{q}, []Response{good}, ""); err != nil || len(values) != 1 {
		t.Fatalf("valid batch: %v", err)
	}
	bad := good
	bad.Result = Encode(DetectResult{Matches: []bool{true}})
	if values, err := AcceptBatch([]Request{q}, []Response{bad}, ""); err == nil || values != nil {
		t.Fatal("malformed result returned partial acceptance")
	}
	for _, code := range []string{NotRecipient, "WALLET_LOCKED", "WALLET_CANCELLED"} {
		failure := good
		failure.Result = nil
		failure.Error = &ResponseError{Code: code}
		if values, err := AcceptBatch([]Request{q}, []Response{failure}, ""); err == nil || values != nil {
			t.Fatal("batch acceptance swallowed item failure")
		}
	}
}

func TestCanonicalErrorCodesAndExactNoteContext(t *testing.T) {
	q, _ := Prepare("code", common.HexToAddress("0x1111111111111111111111111111111111111111"), Public, struct{}{})
	for _, code := range []string{"lowercase", "1CODE", "_CODE", "BAD CODE"} {
		r := Response{Version: 1, ID: q.ID, OwnerAddress: q.OwnerAddress, Method: q.Method, Error: &ResponseError{Code: code}}
		if err := ValidateResponse(q, r); err == nil {
			t.Fatal("accepted noncanonical error code", code)
		}
	}
	raw, err := os.ReadFile("../internal/testsupport/testdata/wallet-wire-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Request Request `json:"request"`
		} `json:"cases"`
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, c := range fixture.Cases {
		if c.Request.Method != DecryptNote {
			continue
		}
		var p map[string]any
		if err = json.Unmarshal(c.Request.Params, &p); err != nil {
			t.Fatal(err)
		}
		p["context"].(map[string]any)["role"] = ""
		c.Request.Params = Encode(p)
		if err = ValidateRequest(c.Request); err == nil {
			t.Fatal("accepted role field on Note context")
		}
	}
}
