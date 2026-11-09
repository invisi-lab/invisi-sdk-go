package core

import (
	"encoding/json"
	"testing"
)

func TestOwnerAddressPublicJSON(t *testing.T) {
	var n Note
	if err := json.Unmarshal([]byte(`{"chainId":"1","protocolAddress":"2","ownerAddress":"3","tokenAddress":"4","amount":"5","nonce":"6"}`), &n); err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(n)
	var fields map[string]any
	json.Unmarshal(out, &fields)
	if fields["ownerAddress"] != "3" || len(fields) != 6 {
		t.Fatalf("wrong owner JSON: %s", out)
	}
}

func TestNoteJSONRejectsMissingOwnerAndNoncanonicalFields(t *testing.T) {
	for _, input := range []string{
		`{"chainId":"1","protocolAddress":"2","tokenAddress":"4","amount":"5","nonce":"6"}`,
		`{"chainId":"1","protocolAddress":"2","ownerAddress":"-1","tokenAddress":"4","amount":"5","nonce":"6"}`,
	} {
		var note Note
		if err := json.Unmarshal([]byte(input), &note); err == nil {
			t.Fatal("accepted invalid Note fields", input)
		}
	}
}
