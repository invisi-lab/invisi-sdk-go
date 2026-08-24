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
			PaymentTypes:       []PaymentType{PaymentTypeOpen, PaymentTypeBlind},
			Mode:               "single",
			Endpoint:           strPtr("https://pay.example.com/ipp/v1"),
		},
	}
}

func sampleVoucher() IPPVoucher {
	return IPPVoucher{
		V: 1, Network: "eip155:84532", Type: MessageTypeVoucher,
		Payload: IPPVoucherV1{
			ReceivingAddress:  "0xabc",
			PaymentCommitment: "0xdef",
			TokenAddress:      "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			Operator:          "0x1234567890123456789012345678901234567890",
			Amount:            "100000000",
			Nonce:             "0x123",
			NullifierKey:      "0x456",
			TreeID:            "1",
			Endpoint:          "https://pay.merchant.com",
		},
	}
}

func TestVoucherRoundtripJSON(t *testing.T) {
	v := sampleVoucher()
	encoded, err := EncodeVoucher(v, FormatJSON)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.HasPrefix(encoded, "{") {
		t.Errorf("JSON must be a bare document (no prefix), got %q", encoded[:10])
	}
	got, err := DecodeVoucher(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Payload.NullifierKey != v.Payload.NullifierKey {
		t.Error("nullifierKey mismatch")
	}
	if got.Payload.Endpoint != v.Payload.Endpoint {
		t.Errorf("endpoint mismatch: got %q, want %q", got.Payload.Endpoint, v.Payload.Endpoint)
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
	if got.Payload.Endpoint != v.Payload.Endpoint {
		t.Errorf("endpoint mismatch: got %q, want %q", got.Payload.Endpoint, v.Payload.Endpoint)
	}
}

func TestDecodeVoucherRejectsRequest(t *testing.T) {
	encoded, _ := Encode(sampleRequest(), FormatJSON)
	if _, err := DecodeVoucher(encoded); err == nil {
		t.Error("expected error decoding a request as voucher")
	}
}

// TestEncodeVoucherMissingEndpoint verifies EncodeVoucher rejects a voucher
// with an empty Endpoint by calling ValidateIPPVoucher before serializing,
// mirroring how Encode calls ValidateIPPRequest for requests.
func TestEncodeVoucherMissingEndpoint(t *testing.T) {
	v := sampleVoucher()
	v.Payload.Endpoint = ""
	if _, err := EncodeVoucher(v, FormatCompact); err == nil {
		t.Fatal("expected error encoding voucher with missing endpoint (compact)")
	} else if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error = %q, want mention of endpoint", err.Error())
	}
	if _, err := EncodeVoucher(v, FormatJSON); err == nil {
		t.Fatal("expected error encoding voucher with missing endpoint (json)")
	} else if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error = %q, want mention of endpoint", err.Error())
	}
}

// TestDecodeVoucherMissingEndpointJSON verifies DecodeVoucher enforces
// ValidateIPPVoucher on the JSON decode path: a hand-written envelope that
// omits the payload's "endpoint" field must be rejected.
func TestDecodeVoucherMissingEndpointJSON(t *testing.T) {
	const missingEndpointJSON = `{"v":1,"network":"eip155:84532","type":"voucher",` +
		`"payload":{"receiving_address":"0xabc","payment_commitment":"0xdef",` +
		`"token_address":"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",` +
		`"operator":"0x1234567890123456789012345678901234567890",` +
		`"amount":"100000000","nonce":"0x123","nullifier_key":"0x456","tree_id":"1"}}`

	_, err := DecodeVoucher(missingEndpointJSON)
	if err == nil {
		t.Fatal("expected error decoding voucher JSON with missing endpoint")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error = %q, want mention of endpoint", err.Error())
	}
}

// TestDecodeVoucherMissingEndpointCompact verifies DecodeVoucher enforces
// ValidateIPPVoucher on the compact decode path. The invalid envelope is built
// via encodeEnvelope directly (bypassing EncodeVoucher's own validation) so we
// can construct a wire string that should never be produced by this SDK but
// must still be rejected on decode (e.g. one emitted by a buggy peer).
func TestDecodeVoucherMissingEndpointCompact(t *testing.T) {
	v := sampleVoucher()
	v.Payload.Endpoint = ""
	encoded, err := encodeEnvelope(v.V, v, FormatCompact)
	if err != nil {
		t.Fatalf("encodeEnvelope: %v", err)
	}

	_, err = DecodeVoucher(encoded)
	if err == nil {
		t.Fatal("expected error decoding voucher compact with missing endpoint")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error = %q, want mention of endpoint", err.Error())
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
			PaymentTypes:       []PaymentType{PaymentTypeOpen},
			Mode:               "single",
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
	if !strings.HasPrefix(encoded, "{") {
		t.Errorf("JSON must be a bare document (no prefix), got %q", encoded[:10])
	}
	var decoded IPPRequest
	if err := json.Unmarshal([]byte(encoded), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded.V != 1 {
		t.Errorf("V = %d, want 1", decoded.V)
	}
}

// TestDecodeBareJSONFromPeerSDK guards the cross-SDK wire contract: a peer SDK
// (TS) emits JSON as a bare document with no "ippN:" prefix, and Decode must
// accept it. Regression guard against requiring a prefix on the JSON format.
func TestDecodeBareJSONFromPeerSDK(t *testing.T) {
	const bareJSON = `{"v":1,"network":"eip155:84532","type":"request",` +
		`"payload":{"identity_commitment":"0xid","token_address":"0xtoken",` +
		`"operator":"0xop","tree_id":"1","payment_types":["open"],` +
		`"endpoint":"https://pay.example.com"}}`

	got, err := Decode(bareJSON)
	if err != nil {
		t.Fatalf("Decode(bare JSON): %v", err)
	}
	if got.V != 1 || got.Type != MessageTypeRequest {
		t.Errorf("decoded V=%d type=%q, want 1/request", got.V, got.Type)
	}
	if got.Payload.IdentityCommitment != "0xid" {
		t.Errorf("IdentityCommitment = %q, want 0xid", got.Payload.IdentityCommitment)
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
	if decoded.Payload.Mode != req.Payload.Mode {
		t.Errorf("Mode = %q, want %q", decoded.Payload.Mode, req.Payload.Mode)
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
	if decoded.Payload.Mode != req.Payload.Mode {
		t.Errorf("Mode = %q, want %q", decoded.Payload.Mode, req.Payload.Mode)
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
	// Version mismatch is a compact-only check (prefix vs payload); a JSON
	// document has no prefix to disagree with its "v" field.
	encoded, err := Encode(req, FormatCompact)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
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

// TestWireKeysSnakeCase verifies that JSON and msgpack wire keys are snake_case.
func TestWireKeysSnakeCase(t *testing.T) {
	voucher := sampleVoucher()

	// ── JSON wire keys ──────────────────────────────────────────────────
	encoded, err := EncodeVoucher(voucher, FormatJSON)
	if err != nil {
		t.Fatalf("EncodeVoucher JSON: %v", err)
	}
	jsonPayload := encoded

	// Keys that must appear as snake_case in the payload object.
	snakeKeys := []string{
		`"tree_id"`, `"receiving_address"`, `"payment_commitment"`,
		`"token_address"`, `"nullifier_key"`, `"endpoint"`,
	}
	for _, k := range snakeKeys {
		if !strings.Contains(jsonPayload, k) {
			t.Errorf("JSON wire: expected snake_case key %s not found in %s", k, jsonPayload)
		}
	}

	// Keys that must NOT appear (old camelCase forms).
	camelKeys := []string{
		`"treeId"`, `"receivingAddress"`, `"paymentCommitment"`,
		`"tokenAddress"`, `"nullifierKey"`,
	}
	for _, k := range camelKeys {
		if strings.Contains(jsonPayload, k) {
			t.Errorf("JSON wire: stale camelCase key %s still present in %s", k, jsonPayload)
		}
	}

	// Round-trip: decode back and confirm field values survive.
	decoded, err := DecodeVoucher(encoded)
	if err != nil {
		t.Fatalf("DecodeVoucher JSON: %v", err)
	}
	if decoded.Payload.TreeID != voucher.Payload.TreeID {
		t.Errorf("TreeID mismatch after JSON round-trip: got %q, want %q",
			decoded.Payload.TreeID, voucher.Payload.TreeID)
	}
	if decoded.Payload.NullifierKey != voucher.Payload.NullifierKey {
		t.Errorf("NullifierKey mismatch after JSON round-trip: got %q, want %q",
			decoded.Payload.NullifierKey, voucher.Payload.NullifierKey)
	}

	// ── msgpack (compact) wire keys ─────────────────────────────────────
	// Decode msgpack bytes and marshal to JSON to inspect keys.
	encodedCompact, err := EncodeVoucher(voucher, FormatCompact)
	if err != nil {
		t.Fatalf("EncodeVoucher compact: %v", err)
	}
	decodedCompact, err := DecodeVoucher(encodedCompact)
	if err != nil {
		t.Fatalf("DecodeVoucher compact: %v", err)
	}
	// Re-marshal to JSON to read the key names (codec preserves field→tag mapping).
	reEncoded, err := EncodeVoucher(decodedCompact, FormatJSON)
	if err != nil {
		t.Fatalf("re-encode compact decoded: %v", err)
	}
	reJSON := reEncoded
	for _, k := range snakeKeys {
		if !strings.Contains(reJSON, k) {
			t.Errorf("compact→JSON: expected snake_case key %s not found", k)
		}
	}
	if decodedCompact.Payload.TreeID != voucher.Payload.TreeID {
		t.Errorf("TreeID mismatch after compact round-trip: got %q, want %q",
			decodedCompact.Payload.TreeID, voucher.Payload.TreeID)
	}
}
