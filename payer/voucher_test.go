package payer

import (
	"math/big"
	"strings"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

func TestBuildVoucher(t *testing.T) {
	v, err := BuildVoucher(BuildVoucherParams{
		Network:          "eip155:84532",
		ReceivingAddress: "0xabc",
		NoteCommitment:   "0xdef",
		TokenAddress:     "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		Operator:         "0x1234567890123456789012345678901234567890",
		Amount:           big.NewInt(100_000_000),
		Nonce:            "0x123",
		NoteSecret:       "0x456",
		TreeID:           "1",
		Endpoint:         "https://pay.merchant.com",
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
	if got.Payload.NoteSecret != "0x456" {
		t.Error("roundtrip mismatch")
	}
	if got.Payload.Endpoint != "https://pay.merchant.com" {
		t.Errorf("roundtrip endpoint mismatch: got %q", got.Payload.Endpoint)
	}
}

// TestBuildVoucherMissingEndpoint documents the split of responsibilities for a
// missing Endpoint: BuildVoucher performs pure local assembly with no
// field-presence validation (it only guards against a nil Amount), so it still
// assembles an Endpoint-less voucher without error. EncodeVoucher, however,
// calls ipp.ValidateIPPVoucher before serializing (mirroring how ipp.Encode
// calls ipp.ValidateIPPRequest for requests), so encoding that same
// Endpoint-less voucher fails with an error mentioning "endpoint".
func TestBuildVoucherMissingEndpoint(t *testing.T) {
	v, err := BuildVoucher(BuildVoucherParams{
		Network:          "eip155:84532",
		ReceivingAddress: "0xabc",
		NoteCommitment:   "0xdef",
		TokenAddress:     "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		Operator:         "0x1234567890123456789012345678901234567890",
		Amount:           big.NewInt(100_000_000),
		Nonce:            "0x123",
		NoteSecret:       "0x456",
		TreeID:           "1",
		// Endpoint intentionally omitted.
	})
	if err != nil {
		t.Fatalf("BuildVoucher unexpectedly errored on missing Endpoint: %v", err)
	}
	if v.Payload.Endpoint != "" {
		t.Fatalf("expected empty Endpoint, got %q", v.Payload.Endpoint)
	}

	_, err = ipp.EncodeVoucher(v, ipp.FormatCompact)
	if err == nil {
		t.Fatal("expected EncodeVoucher to reject a voucher with missing Endpoint")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error = %q, want mention of endpoint", err.Error())
	}
}
