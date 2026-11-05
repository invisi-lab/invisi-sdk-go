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
	if fields["ownerAddress"] == nil || fields["walletIdentity"] != nil {
		t.Fatalf("wrong owner JSON: %s", out)
	}
}

func TestNoteRejectsLegacyWalletIdentityJSON(t *testing.T) {
	var note Note
	if err := json.Unmarshal([]byte(`{"chainId":"1","protocolAddress":"2","walletIdentity":"3","tokenAddress":"4","amount":"5","nonce":"6"}`), &note); err == nil {
		t.Fatal("legacy walletIdentity accepted")
	}
}
