package ipp

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"
)

func creditRequestMap(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile("testdata/ipp-wire-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct{ Cases []struct{ Name, JSON string } }
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		if c.Name == "credit-request" {
			v, err := DecodeJSON([]byte(c.JSON))
			if err != nil {
				t.Fatal(err)
			}
			return v.(map[string]any)
		}
	}
	t.Fatal("missing credit request")
	return nil
}
func encoded(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func TestInitialPaymentCreditRequestFreezesOnlyNetAmount(t *testing.T) {
	v := creditRequestMap(t)
	c := v["credit"].(map[string]any)
	c["initialPayment"] = map[string]any{"amount": "5"}
	request, err := ParsePaymentRequest(encoded(t, v), origin)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := DecodeJSON(encoded(t, request))
	if err != nil {
		t.Fatal(err)
	}
	if raw.(map[string]any)["credit"].(map[string]any)["initialPayment"].(map[string]any)["amount"] != "5" {
		t.Fatal("initial payment lost during typed round trip")
	}
	for _, bad := range []any{nil, map[string]any{"amount": "0"}, map[string]any{"amount": "01"}, map[string]any{"amount": "-1"}, map[string]any{"amount": new(big.Int).Lsh(big.NewInt(1), 126).String()}, map[string]any{"amount": "5", "recipient": c["spender"]}, map[string]any{"amount": "5", "maxFeeAmount": "1"}} {
		c["initialPayment"] = bad
		if _, err := ParsePaymentRequest(encoded(t, v), origin); err == nil {
			t.Fatalf("accepted invalid initial payment %#v", bad)
		}
	}
}
func initialReceiptMap() map[string]any {
	return map[string]any{"network": "eip155:8453", "protocolAddress": protocolAddress, "transactionHash": hash, "commitment": hash, "blockNumber": "20", "blockHash": hash, "completedAt": "2026-11-17T08:03:00Z", "creditId": "credit_001", "creditStatus": "depleted", "owner": address, "grantId": "1", "verificationBlock": map[string]any{"number": "20", "hash": hash, "timestamp": "2026-11-17T08:03:00Z"}, "expiresAt": "2027-11-17T08:03:00Z", "initialPayment": map[string]any{"noteCommitment": "0x0202020202020202020202020202020202020202020202020202020202020202", "grossAmount": "6", "netAmount": "5", "feeAmount": "1", "feeRecipient": address}}
}
func TestInitialPaymentReceiptChecksGrossNetFeeAndDepleted(t *testing.T) {
	if !validReceipt(initialReceiptMap(), "credit") {
		t.Fatal("valid depleted Credit with successful initial payment rejected")
	}
	for _, change := range []func(map[string]any){
		func(r map[string]any) { delete(r, "initialPayment") },
		func(r map[string]any) { r["initialPayment"].(map[string]any)["grossAmount"] = "7" },
		func(r map[string]any) { r["initialPayment"].(map[string]any)["noteCommitment"] = r["commitment"] },
		func(r map[string]any) {
			r["initialPayment"].(map[string]any)["feeRecipient"] = "0x0000000000000000000000000000000000000000"
		},
		func(r map[string]any) { r["initialPayment"].(map[string]any)["recipient"] = address },
	} {
		r := initialReceiptMap()
		change(r)
		if validReceipt(r, "credit") {
			t.Fatalf("invalid initial receipt accepted: %#v", r)
		}
	}
}

func TestInitialPaymentMatchesFrozenSessionAndReceiptByValue(t *testing.T) {
	raw := creditRequestMap(t)
	raw["credit"].(map[string]any)["initialPayment"] = map[string]any{"amount": "5"}
	r, err := ParsePaymentRequest(encoded(t, raw), origin)
	if err != nil {
		t.Fatal(err)
	}
	// Independent decoding produces different pointers for the same frozen terms.
	var terms CreditTerms
	if err = json.Unmarshal(encoded(t, r.Credit.CreditTerms), &terms); err != nil {
		t.Fatal(err)
	}
	session := &PaymentSession{SessionID: r.SessionID, Mode: Credit, Options: []Option{{ID: r.OptionID, Method: r.Method, Network: r.Credit.Network, Asset: Asset{Address: r.Credit.TokenAddress}, Credit: &terms}}}
	if err = AssertRequestMatchesSession(r, session); err != nil {
		t.Fatal("equal independently decoded terms rejected", err)
	}
	receipt := new(Receipt)
	if err = json.Unmarshal(encoded(t, initialReceiptMap()), receipt); err != nil {
		t.Fatal(err)
	}
	receipt.Network = r.Credit.Network
	receipt.ProtocolAddress = r.Credit.ProtocolAddress
	if err = AssertReceiptMatchesRequest(receipt, r); err != nil {
		t.Fatal(err)
	}
	r.Credit.InitialPayment.Amount = "6"
	if AssertRequestMatchesSession(r, session) == nil {
		t.Fatal("changed frozen net amount accepted")
	}
	if AssertReceiptMatchesRequest(receipt, r) == nil {
		t.Fatal("receipt net mismatch accepted")
	}
	r.Credit.InitialPayment = nil
	if AssertRequestMatchesSession(r, session) == nil {
		t.Fatal("initial payment silently removed from request")
	}
	if AssertReceiptMatchesRequest(receipt, r) == nil {
		t.Fatal("unexpected initial payment accepted")
	}
}

func TestCreditReceiptGrantIDPositiveUint128(t *testing.T) {
	for _, id := range []string{"0", "01", "-1", "340282366920938463463374607431768211456"} {
		receipt := initialReceiptMap()
		receipt["grantId"] = id
		if validReceipt(receipt, "credit") {
			t.Fatalf("accepted invalid grant ID %q", id)
		}
	}
	receipt := initialReceiptMap()
	receipt["grantId"] = "340282366920938463463374607431768211455"
	if !validReceipt(receipt, "credit") {
		t.Fatal("rejected maximum uint128 grant ID")
	}
	receipt["revocationSlot"] = json.Number("0")
	if validReceipt(receipt, "credit") {
		t.Fatal("accepted historical slot field")
	}
}
