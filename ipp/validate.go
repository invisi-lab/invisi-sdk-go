package ipp

import (
	"fmt"
	"regexp"
)

var networkPattern = regexp.MustCompile(`^eip155:\d+$`)

// ValidateIPPRequest checks that an IPPRequest is well-formed.
func ValidateIPPRequest(r IPPRequest) error {
	if r.V < 1 {
		return fmt.Errorf("ipp: invalid version: %d (must be >= 1)", r.V)
	}
	if !networkPattern.MatchString(r.Network) {
		return fmt.Errorf("ipp: invalid network: %q (must match eip155:<digits>)", r.Network)
	}

	p := r.Payload
	if p.IdentityCommitment == "" {
		return fmt.Errorf("ipp: identityCommitment is required")
	}
	if p.TokenAddress == "" {
		return fmt.Errorf("ipp: tokenAddress is required")
	}
	if p.Operator == "" {
		return fmt.Errorf("ipp: operator is required")
	}
	if len(p.PaymentTypes) == 0 {
		return fmt.Errorf("ipp: paymentTypes must not be empty")
	}
	for _, pt := range p.PaymentTypes {
		if !pt.Valid() {
			return fmt.Errorf("ipp: invalid paymentType: %q", pt)
		}
	}

	// A single/multi request delivers payment materials via propose, so it must
	// carry an endpoint. The rule fires only for an explicit single/multi (a
	// credit code's one-time semantics ride credit_id, not mode; amount-per-mode
	// / credit⇒credit_id are encoding-tightening rules the encoder guarantees,
	// not SDK baseline).
	if p.Mode == "single" || p.Mode == "multi" {
		if p.Endpoint == nil || *p.Endpoint == "" {
			return fmt.Errorf("ipp: endpoint is required when mode is 'single' or 'multi' (materials delivered via propose)")
		}
	}

	return nil
}
