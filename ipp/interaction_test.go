package ipp

import (
	"encoding/json"
	"strings"
	"testing"
)

// AuthorizeResponse must carry the server-delivered credit_nullifier_key.
func TestAuthorizeResponseCreditNullifierKeyRoundTrip(t *testing.T) {
	key := "0x" + strings.Repeat("ab", 32)
	nkh := "0x" + strings.Repeat("cd", 32)
	resp := AuthorizeResponse{
		Status:             "accepted",
		NullifierKeyHash:   &nkh,
		CreditNullifierKey: &key,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"credit_nullifier_key"`) {
		t.Errorf("marshalled response missing credit_nullifier_key: %s", b)
	}
	var back AuthorizeResponse
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.CreditNullifierKey == nil || *back.CreditNullifierKey != key {
		t.Errorf("CreditNullifierKey round-trip failed: got %v", back.CreditNullifierKey)
	}
}

// AuthorizeRequest must not carry nullifier_key_hash or credit_nullifier_key on the wire.
func TestAuthorizeRequestOmitsDeletedFields(t *testing.T) {
	req := AuthorizeRequest{
		IdentityCommitment: "0x01",
		Amount:             "100",
		Expiry:             "1700000000",
		PayerCommitment:    "0x02",
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	for _, banned := range []string{`"nullifier_key_hash"`, `"credit_nullifier_key"`} {
		if strings.Contains(string(b), banned) {
			t.Errorf("AuthorizeRequest must not carry %s; got %s", banned, b)
		}
	}
}

// IPPPayloadV1's optional credit_id (one-time credit code) round-trips through
// both JSON and compact (msgpack) encoding.
func TestIPPPayloadCreditIDRoundTrip(t *testing.T) {
	creditID := "credit-abc-123"
	req := IPPRequest{
		V:       IPPCurrentVersion,
		Network: "eip155:84532",
		Type:    MessageTypeRequest,
		Payload: IPPPayloadV1{
			IdentityCommitment: "0x01",
			TokenAddress:       "0x02",
			Operator:           "0x03",
			Mode:               "single",
			PaymentTypes:       []PaymentType{PaymentTypeCredit},
			CreditID:           &creditID,
			Endpoint:           strPtr("https://pay.example.com/ipp/v1"),
		},
	}
	for _, format := range []Format{FormatJSON, FormatCompact} {
		enc, err := Encode(req, format)
		if err != nil {
			t.Fatalf("Encode(%s): %v", format, err)
		}
		decoded, err := Decode(enc)
		if err != nil {
			t.Fatalf("Decode(%s): %v", format, err)
		}
		if decoded.Payload.CreditID == nil || *decoded.Payload.CreditID != creditID {
			t.Errorf("format %s: CreditID round-trip failed: got %v", format, decoded.Payload.CreditID)
		}
	}
}

// ProposeResponse must carry tree_id (SPEC-029 §3).
func TestProposeResponseTreeIDRoundTrip(t *testing.T) {
	treeID := "7"
	resp := ProposeResponse{Status: "accepted", TreeID: &treeID}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"tree_id"`) {
		t.Errorf("marshalled ProposeResponse missing tree_id: %s", b)
	}
	var back ProposeResponse
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.TreeID == nil || *back.TreeID != treeID {
		t.Errorf("TreeID round-trip failed: got %v", back.TreeID)
	}
}

// AuthorizeResponse must carry tree_id (SPEC-029 §3).
func TestAuthorizeResponseTreeIDRoundTrip(t *testing.T) {
	treeID := "7"
	resp := AuthorizeResponse{Status: "accepted", TreeID: &treeID}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"tree_id"`) {
		t.Errorf("marshalled AuthorizeResponse missing tree_id: %s", b)
	}
}

// IPPPayloadV1.Mode round-trips over JSON and compact (SPEC-029 §1).
func TestIPPPayloadModeRoundTrip(t *testing.T) {
	req := IPPRequest{
		V:       IPPCurrentVersion,
		Network: "eip155:84532",
		Type:    MessageTypeRequest,
		Payload: IPPPayloadV1{
			IdentityCommitment: "0x01",
			TokenAddress:       "0x02",
			Operator:           "0x03",
			PaymentTypes:       []PaymentType{PaymentTypeOpen},
			Mode:               "single",
			Endpoint:           strPtr("https://pay.example.com/ipp/v1"),
		},
	}
	for _, format := range []Format{FormatJSON, FormatCompact} {
		enc, err := Encode(req, format)
		if err != nil {
			t.Fatalf("Encode(%s): %v", format, err)
		}
		decoded, err := Decode(enc)
		if err != nil {
			t.Fatalf("Decode(%s): %v", format, err)
		}
		if decoded.Payload.Mode != "single" {
			t.Errorf("format %s: Mode round-trip failed: got %q", format, decoded.Payload.Mode)
		}
	}
}
