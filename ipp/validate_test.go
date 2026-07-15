package ipp

import (
	"strings"
	"testing"
)

// validSingleRequest returns a well-formed single-mode IPPRequest for testing.
func validSingleRequest() IPPRequest {
	endpoint := "https://pay.merchant.com"
	amount := "100"
	return IPPRequest{
		V:       IPPCurrentVersion,
		Network: "eip155:84532",
		Type:    MessageTypeRequest,
		Payload: IPPPayloadV1{
			IdentityCommitment: "0xabc",
			TokenAddress:       "0xdef",
			Operator:           "0x1234567890123456789012345678901234567890",
			PaymentTypes:       []PaymentType{PaymentTypeOpen, PaymentTypeBlind},
			Mode:               "single",
			Amount:             &amount,
			Endpoint:           &endpoint,
		},
	}
}

func TestValidateSinglePasses(t *testing.T) {
	if err := ValidateIPPRequest(validSingleRequest()); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidateMultiWithoutEndpointFails(t *testing.T) {
	req := validSingleRequest()
	req.Payload.Mode = "multi"
	req.Payload.Amount = nil
	req.Payload.Endpoint = nil
	if err := ValidateIPPRequest(req); err == nil {
		t.Fatal("expected error: multi without endpoint")
	}
}

func TestValidateSingleWithoutAmountPassesBaseline(t *testing.T) {
	req := validSingleRequest()
	req.Payload.Amount = nil // amount-per-mode is an encoding-tightening rule, not baseline
	if err := ValidateIPPRequest(req); err != nil {
		t.Fatalf("expected valid (baseline ignores amount), got %v", err)
	}
}

func TestValidateModelessCreditPassesWithoutEndpoint(t *testing.T) {
	// 'Modeless credit' vector: rule fires only for explicit single/multi. Credit
	// codes carry mode='single'; this asserts the validator does not force an
	// endpoint when Mode is empty (untrusted runtime shape).
	req := validSingleRequest()
	req.Payload.Mode = "" // absent
	req.Payload.Amount = nil
	req.Payload.Endpoint = nil
	req.Payload.PaymentTypes = []PaymentType{PaymentTypeCredit}
	creditID := "credit_abc123"
	req.Payload.CreditID = &creditID
	if err := ValidateIPPRequest(req); err != nil {
		t.Fatalf("expected valid (modeless), got %v", err)
	}
}

func TestValidateInvalidVersion(t *testing.T) {
	r := validSingleRequest()
	r.V = 0
	err := ValidateIPPRequest(r)
	if err == nil {
		t.Fatal("expected error for v=0")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("error = %q, want mention of version", err.Error())
	}
}

func TestValidateInvalidNetwork(t *testing.T) {
	tests := []struct {
		name    string
		network string
	}{
		{"wrong prefix", "cosmos:1"},
		{"no digits", "eip155:"},
		{"empty", ""},
		{"no prefix", "84532"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := validSingleRequest()
			r.Network = tt.network
			err := ValidateIPPRequest(r)
			if err == nil {
				t.Errorf("expected error for network=%q", tt.network)
			}
		})
	}
}

func TestValidateEmptyPaymentTypes(t *testing.T) {
	r := validSingleRequest()
	r.Payload.PaymentTypes = nil
	err := ValidateIPPRequest(r)
	if err == nil {
		t.Fatal("expected error for empty paymentTypes")
	}
	if !strings.Contains(err.Error(), "paymentTypes") {
		t.Errorf("error = %q, want mention of paymentTypes", err.Error())
	}
}

func TestValidateInvalidPaymentType(t *testing.T) {
	r := validSingleRequest()
	r.Payload.PaymentTypes = []PaymentType{"shielded"}
	err := ValidateIPPRequest(r)
	if err == nil {
		t.Fatal("expected error for invalid paymentType")
	}
	if !strings.Contains(err.Error(), "shielded") {
		t.Errorf("error = %q, want mention of 'shielded'", err.Error())
	}
}

func TestValidateSingleCreditRequiresEndpoint(t *testing.T) {
	// mode=single requires an endpoint regardless of payment types; credit's own
	// interactive needs (Authorize/Notify) ride along.
	r := validSingleRequest()
	r.Payload.PaymentTypes = []PaymentType{PaymentTypeOpen, PaymentTypeCredit}
	r.Payload.Endpoint = nil
	err := ValidateIPPRequest(r)
	if err == nil {
		t.Fatal("expected error for single+credit without endpoint")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error = %q, want mention of endpoint", err.Error())
	}
}

func TestValidateSingleCreditWithEndpointValid(t *testing.T) {
	r := validSingleRequest()
	r.Payload.PaymentTypes = []PaymentType{PaymentTypeOpen, PaymentTypeCredit}
	r.Payload.Endpoint = strPtr("https://pay.example.com/ipp/v1")
	if err := ValidateIPPRequest(r); err != nil {
		t.Errorf("expected valid (single+credit with endpoint), got: %v", err)
	}
}

func TestValidateEmptyRequiredFields(t *testing.T) {
	fields := []struct {
		name   string
		mutate func(*IPPRequest)
	}{
		{"identityCommitment", func(r *IPPRequest) { r.Payload.IdentityCommitment = "" }},
		{"tokenAddress", func(r *IPPRequest) { r.Payload.TokenAddress = "" }},
		{"operator", func(r *IPPRequest) { r.Payload.Operator = "" }},
	}
	for _, tt := range fields {
		t.Run(tt.name, func(t *testing.T) {
			r := validSingleRequest()
			tt.mutate(&r)
			err := ValidateIPPRequest(r)
			if err == nil {
				t.Errorf("expected error for empty %s", tt.name)
			}
		})
	}
}
