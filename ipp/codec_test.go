package ipp

import (
	"encoding/json"
	"strings"
	"testing"
)

func sampleRequest() IPPRequest {
	return IPPRequest{
		V:       1,
		Network: "eip155:84532",
		Type:    MessageTypeRequest,
		Meta: map[string]string{
			MetaKeyMerchantName: "Coffee Shop",
		},
		Payload: IPPPayloadV1{
			IdentityCommitment: "0xabc",
			TokenAddress:       "0xtoken",
			Operator:           "0xoperator",
			TreeID:             "0xtree",
			PaymentTypes:       []PaymentType{PaymentTypeOpen, PaymentTypeBlind},
			Nonce:              strPtr("0xnonce"),
			PaymentCommitment:  strPtr("0xcommitment"),
		},
	}
}

func sampleVoucher() IPPVoucher {
	return IPPVoucher{
		V: 1, Network: "eip155:84532", Type: MessageTypeVoucher,
		Payload: IPPVoucherV1{
			IdentityCommitment: "0xabc",
			PaymentCommitment:  "0xdef",
			TokenAddress:       "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			Operator:           "0x1234567890123456789012345678901234567890",
			Amount:             "100000000",
			Nonce:              "0x123",
			NullifierKey:       "0x456",
			TreeID:             "1",
		},
	}
}

func TestVoucherRoundtripJSON(t *testing.T) {
	v := sampleVoucher()
	encoded, err := EncodeVoucher(v, FormatJSON)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.HasPrefix(encoded, "ipp1:") {
		t.Errorf("prefix: %q", encoded[:10])
	}
	got, err := DecodeVoucher(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Payload.NullifierKey != v.Payload.NullifierKey {
		t.Error("nullifierKey mismatch")
	}
}

func TestVoucherRoundtripCompact(t *testing.T) {
	v := sampleVoucher()
	encoded, _ := EncodeVoucher(v, FormatCompact)
	got, err := DecodeVoucher(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Payload.Operator != v.Payload.Operator {
		t.Error("operator mismatch")
	}
}

func TestDecodeVoucherRejectsRequest(t *testing.T) {
	encoded, _ := Encode(sampleRequest(), FormatJSON)
	if _, err := DecodeVoucher(encoded); err == nil {
		t.Error("expected error decoding a request as voucher")
	}
}

func TestPeekType(t *testing.T) {
	rv, _ := EncodeVoucher(sampleVoucher(), FormatJSON)
	if typ, _ := PeekType(rv); typ != MessageTypeVoucher {
		t.Errorf("got %q", typ)
	}
	rr, _ := Encode(sampleRequest(), FormatJSON)
	if typ, _ := PeekType(rr); typ != MessageTypeRequest {
		t.Errorf("got %q", typ)
	}
}

func minimalRequest() IPPRequest {
	return IPPRequest{
		V:       1,
		Network: "eip155:1",
		Type:    MessageTypeRequest,
		Payload: IPPPayloadV1{
			IdentityCommitment: "0xid",
			TokenAddress:       "0xtoken",
			Operator:           "0xop",
			TreeID:             "0xtree",
			PaymentTypes:       []PaymentType{PaymentTypeOpen},
			Endpoint:           strPtr("https://pay.example.com"),
		},
	}
}

func TestEncodeJSON(t *testing.T) {
	req := sampleRequest()
	encoded, err := Encode(req, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(encoded, "ipp1:") {
		t.Errorf("expected prefix ipp1:, got %q", encoded[:10])
	}
	// The payload after prefix should be valid JSON.
	payload := encoded[len("ipp1:"):]
	var decoded IPPRequest
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if decoded.V != 1 {
		t.Errorf("V = %d, want 1", decoded.V)
	}
}

func TestEncodeCompact(t *testing.T) {
	req := sampleRequest()
	encoded, err := Encode(req, FormatCompact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(encoded, "ipp1:") {
		t.Errorf("expected prefix ipp1:, got %q", encoded[:10])
	}
	// Compact payload should NOT start with '{'.
	payload := encoded[len("ipp1:"):]
	if len(payload) > 0 && payload[0] == '{' {
		t.Error("compact payload should not start with '{'")
	}
}

func TestRoundtripJSON(t *testing.T) {
	req := sampleRequest()
	encoded, err := Encode(req, FormatJSON)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded.V != req.V {
		t.Errorf("V = %d, want %d", decoded.V, req.V)
	}
	if decoded.Network != req.Network {
		t.Errorf("Network = %q, want %q", decoded.Network, req.Network)
	}
	if decoded.Payload.IdentityCommitment != req.Payload.IdentityCommitment {
		t.Errorf("IdentityCommitment = %q, want %q",
			decoded.Payload.IdentityCommitment, req.Payload.IdentityCommitment)
	}
	if *decoded.Payload.Nonce != *req.Payload.Nonce {
		t.Errorf("Nonce = %q, want %q", *decoded.Payload.Nonce, *req.Payload.Nonce)
	}
	if len(decoded.Payload.PaymentTypes) != len(req.Payload.PaymentTypes) {
		t.Errorf("PaymentTypes len = %d, want %d",
			len(decoded.Payload.PaymentTypes), len(req.Payload.PaymentTypes))
	}
}

func TestRoundtripCompact(t *testing.T) {
	req := sampleRequest()
	encoded, err := Encode(req, FormatCompact)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded.V != req.V {
		t.Errorf("V = %d, want %d", decoded.V, req.V)
	}
	if decoded.Network != req.Network {
		t.Errorf("Network = %q, want %q", decoded.Network, req.Network)
	}
	if decoded.Payload.IdentityCommitment != req.Payload.IdentityCommitment {
		t.Errorf("IdentityCommitment = %q, want %q",
			decoded.Payload.IdentityCommitment, req.Payload.IdentityCommitment)
	}
	if *decoded.Payload.PaymentCommitment != *req.Payload.PaymentCommitment {
		t.Errorf("PaymentCommitment = %q, want %q",
			*decoded.Payload.PaymentCommitment, *req.Payload.PaymentCommitment)
	}
}

func TestRoundtripMinimal(t *testing.T) {
	req := minimalRequest()
	for _, format := range []Format{FormatJSON, FormatCompact} {
		t.Run(string(format), func(t *testing.T) {
			encoded, err := Encode(req, format)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}
			decoded, err := Decode(encoded)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if decoded.Payload.Nonce != nil {
				t.Error("Nonce should be nil")
			}
			if decoded.Payload.PaymentCommitment != nil {
				t.Error("PaymentCommitment should be nil")
			}
			if *decoded.Payload.Endpoint != *req.Payload.Endpoint {
				t.Errorf("Endpoint = %q, want %q",
					*decoded.Payload.Endpoint, *req.Payload.Endpoint)
			}
		})
	}
}

func TestEncodeUnknownFormat(t *testing.T) {
	req := sampleRequest()
	_, err := Encode(req, Format("xml"))
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("error = %q, want mention of unknown format", err.Error())
	}
}

func TestDecodeVersionMismatch(t *testing.T) {
	req := sampleRequest()
	encoded, err := Encode(req, FormatJSON)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	// Replace ipp1: with ipp2:
	tampered := "ipp2:" + encoded[len("ipp1:"):]
	_, err = Decode(tampered)
	if err == nil {
		t.Fatal("expected error for version mismatch")
	}
	if !strings.Contains(err.Error(), "version mismatch") {
		t.Errorf("error = %q, want mention of version mismatch", err.Error())
	}
}

func TestCompactBase64URLNoSpecialChars(t *testing.T) {
	req := sampleRequest()
	encoded, err := Encode(req, FormatCompact)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	payload := encoded[len("ipp1:"):]
	// base64url should not contain '+', '/', or '='
	if strings.ContainsAny(payload, "+/=") {
		t.Errorf("compact payload contains non-base64url chars: %q", payload)
	}
}
