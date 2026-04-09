package ipp

import "testing"

func TestIPPRequestAllFields(t *testing.T) {
	req := IPPRequest{
		V:       1,
		Network: "eip155:84532",
		Meta: map[string]string{
			MetaKeyMerchantName: "Coffee Shop",
			MetaKeyOrderID:      "order-123",
		},
		Payload: IPPPayloadV1{
			IdentityCommitment: "0xabc",
			TokenAddress:       "0xtoken",
			Operator:           "0xoperator",
			TreeID:             "0xtree",
			PaymentTypes:       []PaymentType{PaymentTypeOpen, PaymentTypeBlind},
			Amount:             strPtr("1000000"),
			Nonce:              strPtr("0xnonce"),
			PaymentCommitment:  strPtr("0xcommitment"),
			Endpoint:           strPtr("https://pay.example.com"),
		},
	}

	if req.V != 1 {
		t.Errorf("V = %d, want 1", req.V)
	}
	if req.Network != "eip155:84532" {
		t.Errorf("Network = %s, want eip155:84532", req.Network)
	}
	if len(req.Meta) != 2 {
		t.Errorf("Meta len = %d, want 2", len(req.Meta))
	}
	if req.Payload.IdentityCommitment != "0xabc" {
		t.Errorf("IdentityCommitment = %s, want 0xabc", req.Payload.IdentityCommitment)
	}
	if len(req.Payload.PaymentTypes) != 2 {
		t.Errorf("PaymentTypes len = %d, want 2", len(req.Payload.PaymentTypes))
	}
	if *req.Payload.Amount != "1000000" {
		t.Errorf("Amount = %s, want 1000000", *req.Payload.Amount)
	}
	if *req.Payload.Nonce != "0xnonce" {
		t.Errorf("Nonce = %s, want 0xnonce", *req.Payload.Nonce)
	}
	if *req.Payload.PaymentCommitment != "0xcommitment" {
		t.Errorf("PaymentCommitment = %s, want 0xcommitment", *req.Payload.PaymentCommitment)
	}
}

func TestIPPRequestMinimal(t *testing.T) {
	req := IPPRequest{
		V:       1,
		Network: "eip155:1",
		Payload: IPPPayloadV1{
			IdentityCommitment: "0xabc",
			TokenAddress:       "0xtoken",
			Operator:           "0xoperator",
			TreeID:             "0xtree",
			PaymentTypes:       []PaymentType{PaymentTypeOpen},
			Endpoint:           strPtr("https://pay.example.com"),
		},
	}

	if req.Meta != nil {
		t.Errorf("Meta = %v, want nil", req.Meta)
	}
	if req.Payload.Amount != nil {
		t.Errorf("Amount = %v, want nil", req.Payload.Amount)
	}
	if req.Payload.Nonce != nil {
		t.Errorf("Nonce = %v, want nil", req.Payload.Nonce)
	}
	if req.Payload.PaymentCommitment != nil {
		t.Errorf("PaymentCommitment = %v, want nil", req.Payload.PaymentCommitment)
	}
}

func TestPaymentTypeConstants(t *testing.T) {
	if PaymentTypeOpen != "open" {
		t.Errorf("PaymentTypeOpen = %q, want open", PaymentTypeOpen)
	}
	if PaymentTypeBlind != "blind" {
		t.Errorf("PaymentTypeBlind = %q, want blind", PaymentTypeBlind)
	}
	if PaymentTypeCredit != "credit" {
		t.Errorf("PaymentTypeCredit = %q, want credit", PaymentTypeCredit)
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want json", FormatJSON)
	}
	if FormatCompact != "compact" {
		t.Errorf("FormatCompact = %q, want compact", FormatCompact)
	}
}

func TestIPPCurrentVersion(t *testing.T) {
	if IPPCurrentVersion != 1 {
		t.Errorf("IPPCurrentVersion = %d, want 1", IPPCurrentVersion)
	}
}
