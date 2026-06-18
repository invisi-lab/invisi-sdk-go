package payer

import (
	"math/big"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

func TestBuildVoucher(t *testing.T) {
	v, err := BuildVoucher(BuildVoucherParams{
		Network:            "eip155:84532",
		IdentityCommitment: "0xabc",
		PaymentCommitment:  "0xdef",
		TokenAddress:       "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		Operator:           "0x1234567890123456789012345678901234567890",
		Amount:             big.NewInt(100_000_000),
		Nonce:              "0x123",
		NullifierKey:       "0x456",
		TreeID:             "1",
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

	encoded, _ := ipp.EncodeVoucher(v, ipp.FormatCompact)
	got, err := ipp.DecodeVoucher(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got.Payload.NullifierKey != "0x456" {
		t.Error("roundtrip mismatch")
	}
}
