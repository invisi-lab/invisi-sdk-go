package wallet

import (
	"encoding/json"
	"testing"
)

func TestResponseBindsRequest(t *testing.T) {
	req := Request{Version: 1, ID: "one", OwnerAddress: "0x1111111111111111111111111111111111111111", Method: Public, Params: json.RawMessage(`{}`)}
	good := Response{Version: 1, ID: req.ID, OwnerAddress: req.OwnerAddress, Method: req.Method, Result: json.RawMessage(`{}`)}
	if err := ValidateResponse(req, good); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Response){func(r *Response) { r.ID = "two" }, func(r *Response) { r.Version = 2 }, func(r *Response) { r.Method = Detect }, func(r *Response) { r.OwnerAddress = "0x2222222222222222222222222222222222222222" }, func(r *Response) { r.Result = nil }} {
		bad := good
		mutate(&bad)
		if ValidateResponse(req, bad) == nil {
			t.Fatal("accepted unrelated or malformed response")
		}
	}
}

func TestPortableCanonicalStrings(t *testing.T) {
	for _, id := range []string{"", "has space", "é", string(make([]byte, 129))} {
		q := Request{Version: 1, ID: id, OwnerAddress: "0x1111111111111111111111111111111111111111", Method: Public, Params: Encode(struct{}{})}
		if ValidateRequest(q) == nil {
			t.Fatal("accepted invalid id", id)
		}
	}
	if _, err := Bytes("0xAB", 1); err == nil {
		t.Fatal("accepted uppercase hex")
	}
}
