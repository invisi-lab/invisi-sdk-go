package payer

import (
	"math/big"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

func TestBuildVoucher(t *testing.T) {
	v, err := BuildVoucher(BuildVoucherParams{
		Network:           "eip155:84532",
		ReceivingAddress:  "0xabc",
		PaymentCommitment: "0xdef",
		TokenAddress:      "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		Operator:          "0x1234567890123456789012345678901234567890",
		Amount:            big.NewInt(100_000_000),
		Nonce:             "0x123",
		NullifierKey:      "0x456",
		TreeID:            "1",
		Endpoint:          "https://pay.merchant.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != ipp.MessageTypeVoucher {
		t.Errorf("type = %q", v.Type)
	}
	if v.Payload.Amount != "100000000" {
		t.Errorf("amount = %q", v.Payload.Amount)
	}
	if v.Payload.Endpoint != "https://pay.merchant.com" {
		t.Errorf("endpoint = %q", v.Payload.Endpoint)
	}

	encoded, _ := ipp.EncodeVoucher(v, ipp.FormatCompact)
	got, err := ipp.DecodeVoucher(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got.Payload.NullifierKey != "0x456" {
		t.Error("roundtrip mismatch")
	}
	if got.Payload.Endpoint != "https://pay.merchant.com" {
		t.Errorf("roundtrip endpoint mismatch: got %q", got.Payload.Endpoint)
	}
}

// TestBuildVoucherMissingEndpoint documents current behavior when Endpoint is
// omitted: BuildVoucher performs pure local assembly with no field-presence
// validation (it only guards against a nil Amount), and EncodeVoucher does not
// yet call any voucher validator (unlike Encode, which calls ValidateIPPRequest).
// So today this path assembles and encodes an Endpoint-less voucher without
// error. This is a known, temporary gap: a parallel task chain is adding
// ipp.ValidateIPPVoucher and wiring it into EncodeVoucher/DecodeVoucher, at
// which point a missing Endpoint will start failing at the EncodeVoucher step.
// This test pins today's actual behavior so it fails loudly (rather than
// silently going stale) once that validation lands, forcing this test to be
// updated to assert the new error path.
func TestBuildVoucherMissingEndpoint(t *testing.T) {
	v, err := BuildVoucher(BuildVoucherParams{
		Network:           "eip155:84532",
		ReceivingAddress:  "0xabc",
		PaymentCommitment: "0xdef",
		TokenAddress:      "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		Operator:          "0x1234567890123456789012345678901234567890",
		Amount:            big.NewInt(100_000_000),
		Nonce:             "0x123",
		NullifierKey:      "0x456",
		TreeID:            "1",
		// Endpoint intentionally omitted.
	})
	if err != nil {
		t.Fatalf("BuildVoucher unexpectedly errored on missing Endpoint: %v", err)
	}
	if v.Payload.Endpoint != "" {
		t.Fatalf("expected empty Endpoint, got %q", v.Payload.Endpoint)
	}

	if _, err := ipp.EncodeVoucher(v, ipp.FormatCompact); err != nil {
		t.Fatalf("EncodeVoucher unexpectedly errored on missing Endpoint: %v", err)
	}
}
