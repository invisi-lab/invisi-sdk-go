package ipp

import (
	"encoding/json"
	"strings"
	"testing"
)

// AuthorizeResponse must carry the server-delivered credit_shared_key.
func TestAuthorizeResponseCreditSharedKeyRoundTrip(t *testing.T) {
	key := "0x" + strings.Repeat("ab", 32)
	hash := "0x" + strings.Repeat("cd", 32)
	resp := AuthorizeResponse{
		Status:              "accepted",
		CreditSharedKeyHash: &hash,
		CreditSharedKey:     &key,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"credit_shared_key"`) {
		t.Errorf("marshalled response missing credit_shared_key: %s", b)
	}
	var back AuthorizeResponse
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.CreditSharedKey == nil || *back.CreditSharedKey != key {
		t.Errorf("CreditSharedKey round-trip failed: got %v", back.CreditSharedKey)
	}
}

// AuthorizeRequest must not carry credit_shared_key_hash or credit_shared_key on the wire.
func TestAuthorizeRequestOmitsDeletedFields(t *testing.T) {
	req := AuthorizeRequest{
		SpenderIdentityCommitment: "0x01",
		Amount:                    "100",
		Expiry:                    "1700000000",
		OwnerIdentityCommitment:   "0x02",
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	for _, banned := range []string{`"credit_shared_key_hash"`, `"credit_shared_key"`} {
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

func TestFutureTreeIDRejected(t *testing.T) {
	for _, value := range []interface{}{&ProposeResponse{}, &AuthorizeResponse{}} {
		if err := json.Unmarshal([]byte(`{"status":"accepted","tree_id":"7"}`), value); err == nil {
			t.Fatalf("future tree_id accepted by %T", value)
		}
	}
}

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
