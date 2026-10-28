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

// validVoucher returns a well-formed mode-3 IPPVoucher for testing.
func validVoucher() IPPVoucher {
	return IPPVoucher{
		V:       IPPCurrentVersion,
		Network: "eip155:84532",
		Type:    MessageTypeVoucher,
		Payload: IPPVoucherV1{
			ReceivingAddress: "0x1111111111111111111111111111111111111111",
			NoteCommitment:   "0xabc",
			TokenAddress:     "0xdef",
			Operator:         "0x1234567890123456789012345678901234567890",
			Amount:           "100",
			Nonce:            "0x01",
			NoteSecret:       "0x02",
			TreeID:           "1",
			Endpoint:         "https://pay.merchant.com",
		},
	}
}

func TestValidateVoucherPasses(t *testing.T) {
	if err := ValidateIPPVoucher(validVoucher()); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidateVoucherInvalidVersion(t *testing.T) {
	v := validVoucher()
	v.V = 0
	err := ValidateIPPVoucher(v)
	if err == nil {
		t.Fatal("expected error for v=0")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("error = %q, want mention of version", err.Error())
	}
}

func TestValidateVoucherInvalidNetwork(t *testing.T) {
	v := validVoucher()
	v.Network = "mainnet"
	err := ValidateIPPVoucher(v)
	if err == nil {
		t.Fatal("expected error for network=\"mainnet\"")
	}
	if !strings.Contains(err.Error(), "network") {
		t.Errorf("error = %q, want mention of network", err.Error())
	}
}

func TestValidateVoucherWrongType(t *testing.T) {
	v := validVoucher()
	v.Type = MessageTypeRequest
	err := ValidateIPPVoucher(v)
	if err == nil {
		t.Fatal("expected error for type=request")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("error = %q, want mention of type", err.Error())
	}
}

func TestValidateVoucherEmptyRequiredFields(t *testing.T) {
	fields := []struct {
		name   string
		mutate func(*IPPVoucher)
	}{
		{"receivingAddress", func(v *IPPVoucher) { v.Payload.ReceivingAddress = "" }},
		{"noteCommitment", func(v *IPPVoucher) { v.Payload.NoteCommitment = "" }},
		{"tokenAddress", func(v *IPPVoucher) { v.Payload.TokenAddress = "" }},
		{"operator", func(v *IPPVoucher) { v.Payload.Operator = "" }},
		{"amount", func(v *IPPVoucher) { v.Payload.Amount = "" }},
		{"nonce", func(v *IPPVoucher) { v.Payload.Nonce = "" }},
		{"noteSecret", func(v *IPPVoucher) { v.Payload.NoteSecret = "" }},
		{"treeId", func(v *IPPVoucher) { v.Payload.TreeID = "" }},
		{"endpoint", func(v *IPPVoucher) { v.Payload.Endpoint = "" }},
	}
	for _, tt := range fields {
		t.Run(tt.name, func(t *testing.T) {
			v := validVoucher()
			tt.mutate(&v)
			err := ValidateIPPVoucher(v)
			if err == nil {
				t.Fatalf("expected error for empty %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.name) {
				t.Errorf("error = %q, want mention of %s", err.Error(), tt.name)
			}
		})
	}
}

func TestValidateVoucherInvalidEndpoint(t *testing.T) {
	v := validVoucher()
	v.Payload.Endpoint = "not a url"
	err := ValidateIPPVoucher(v)
	if err == nil {
		t.Fatal("expected error for endpoint=\"not a url\"")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error = %q, want mention of endpoint", err.Error())
	}
}

func TestValidateVoucherLocalhostEndpointPasses(t *testing.T) {
	// The SDK does not enforce https; production-only TLS is an application-layer
	// concern (see protocol docs), not an SDK validation rule.
	v := validVoucher()
	v.Payload.Endpoint = "http://localhost:3000"
	if err := ValidateIPPVoucher(v); err != nil {
		t.Errorf("expected valid (no scheme enforcement), got: %v", err)
	}
}

func TestValidateVoucherEmptyLeafIndexPasses(t *testing.T) {
	v := validVoucher()
	v.Payload.LeafIndex = ""
	if err := ValidateIPPVoucher(v); err != nil {
		t.Errorf("expected valid (leafIndex optional), got: %v", err)
	}
}
