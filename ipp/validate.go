package ipp

import (
	"fmt"
	"net/url"
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

// ValidateIPPVoucher checks that an IPPVoucher is well-formed.
func ValidateIPPVoucher(v IPPVoucher) error {
	if v.V < 1 {
		return fmt.Errorf("ipp: invalid version: %d (must be >= 1)", v.V)
	}
	if !networkPattern.MatchString(v.Network) {
		return fmt.Errorf("ipp: invalid network: %q (must match eip155:<digits>)", v.Network)
	}
	if v.Type != MessageTypeVoucher {
		return fmt.Errorf("ipp: expected type %q, got %q", MessageTypeVoucher, v.Type)
	}

	p := v.Payload
	if p.ReceivingAddress == "" {
		return fmt.Errorf("ipp: receivingAddress is required")
	}
	if p.NoteCommitment == "" {
		return fmt.Errorf("ipp: noteCommitment is required")
	}
	if p.TokenAddress == "" {
		return fmt.Errorf("ipp: tokenAddress is required")
	}
	if p.Operator == "" {
		return fmt.Errorf("ipp: operator is required")
	}
	if p.Amount == "" {
		return fmt.Errorf("ipp: amount is required")
	}
	if p.Nonce == "" {
		return fmt.Errorf("ipp: nonce is required")
	}
	if p.NoteSecret == "" {
		return fmt.Errorf("ipp: noteSecret is required")
	}
	if p.TreeID == "" {
		return fmt.Errorf("ipp: treeId is required")
	}
	if p.Endpoint == "" {
		return fmt.Errorf("ipp: endpoint is required")
	}

	// The SDK only checks that Endpoint parses as a URL with a scheme and host;
	// enforcing https in production is an application-layer policy, not an SDK
	// rule (see protocol docs on transport security).
	u, err := url.Parse(p.Endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("ipp: invalid endpoint: %q (must be a valid URL)", p.Endpoint)
	}

	return nil
}
