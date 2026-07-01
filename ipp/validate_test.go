package ipp

import (
	"strings"
	"testing"
)

// validRequest returns a well-formed IPPRequest for testing.
// Mode 1a: nonce + paymentCommitment present (non-interactive).
func validRequest() IPPRequest {
	return IPPRequest{
		V:       1,
		Network: "eip155:84532",
		Payload: IPPPayloadV1{
			IdentityCommitment: "0xabc",
			TokenAddress:       "0xtoken",
			Operator:           "0xoperator",
			TreeID:             "0xtree",
			PaymentTypes:       []PaymentType{PaymentTypeOpen},
			Nonce:              strPtr("0xnonce"),
			PaymentCommitment:  strPtr("0xcommitment"),
		},
	}
}

// validRequestInteractive returns a well-formed IPPRequest for mode 1b
// (interactive: no nonce, endpoint required).
func validRequestInteractive() IPPRequest {
	return IPPRequest{
		V:       1,
		Network: "eip155:84532",
		Payload: IPPPayloadV1{
			IdentityCommitment: "0xabc",
			TokenAddress:       "0xtoken",
			Operator:           "0xoperator",
			TreeID:             "0xtree",
			PaymentTypes:       []PaymentType{PaymentTypeOpen, PaymentTypeBlind},
			Endpoint:           strPtr("https://pay.example.com/ipp/v1"),
		},
	}
}

func TestValidateValid(t *testing.T) {
	if err := ValidateIPPRequest(validRequest()); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestValidateInteractiveValid(t *testing.T) {
	if err := ValidateIPPRequest(validRequestInteractive()); err != nil {
		t.Errorf("expected valid (interactive), got error: %v", err)
	}
}

func TestValidateInvalidVersion(t *testing.T) {
	r := validRequest()
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
			r := validRequest()
			r.Network = tt.network
			err := ValidateIPPRequest(r)
			if err == nil {
				t.Errorf("expected error for network=%q", tt.network)
			}
		})
	}
}

func TestValidateEmptyPaymentTypes(t *testing.T) {
	r := validRequest()
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
	r := validRequest()
	r.Payload.PaymentTypes = []PaymentType{"shielded"}
	err := ValidateIPPRequest(r)
	if err == nil {
		t.Fatal("expected error for invalid paymentType")
	}
	if !strings.Contains(err.Error(), "shielded") {
		t.Errorf("error = %q, want mention of 'shielded'", err.Error())
	}
}

func TestValidateNoncePCMismatch(t *testing.T) {
	// nonce present but no paymentCommitment
	r := validRequest()
	r.Payload.PaymentCommitment = nil
	err := ValidateIPPRequest(r)
	if err == nil {
		t.Fatal("expected error for nonce without paymentCommitment")
	}
	if !strings.Contains(err.Error(), "nonce") {
		t.Errorf("error = %q, want mention of nonce", err.Error())
	}

	// paymentCommitment present but no nonce
	r2 := validRequest()
	r2.Payload.Nonce = nil
	err2 := ValidateIPPRequest(r2)
	if err2 == nil {
		t.Fatal("expected error for paymentCommitment without nonce")
	}
}

func TestValidateNoNonceNoEndpoint(t *testing.T) {
	r := validRequest()
	r.Payload.Nonce = nil
	r.Payload.PaymentCommitment = nil
	r.Payload.Endpoint = nil
	err := ValidateIPPRequest(r)
	if err == nil {
		t.Fatal("expected error when no nonce and no endpoint")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error = %q, want mention of endpoint", err.Error())
	}
}

func TestValidateSingleCreditRequiresEndpoint(t *testing.T) {
	// A credit request needs an endpoint even in single mode (nonce+paymentCommitment
	// present): the payer must reach Authorize/Notify to run the credit line, which
	// the embedded direct-pay fields cannot substitute for. Without it the QR is
	// un-authorizable.
	r := validRequest()
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
	r := validRequest()
	r.Payload.PaymentTypes = []PaymentType{PaymentTypeOpen, PaymentTypeCredit}
	r.Payload.Endpoint = strPtr("https://pay.example.com/ipp/v1")
	if err := ValidateIPPRequest(r); err != nil {
		t.Errorf("expected valid (single+credit with endpoint), got: %v", err)
	}
}

func TestValidateEmptyRequiredFields(t *testing.T) {
	fields := []struct {
		name  string
		mutate func(*IPPRequest)
	}{
		{"identityCommitment", func(r *IPPRequest) { r.Payload.IdentityCommitment = "" }},
		{"tokenAddress", func(r *IPPRequest) { r.Payload.TokenAddress = "" }},
		{"operator", func(r *IPPRequest) { r.Payload.Operator = "" }},
		{"treeId", func(r *IPPRequest) { r.Payload.TreeID = "" }},
	}
	for _, tt := range fields {
		t.Run(tt.name, func(t *testing.T) {
			r := validRequest()
			tt.mutate(&r)
			err := ValidateIPPRequest(r)
			if err == nil {
				t.Errorf("expected error for empty %s", tt.name)
			}
		})
	}
}
