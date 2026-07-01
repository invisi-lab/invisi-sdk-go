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
	if p.TreeID == "" {
		return fmt.Errorf("ipp: treeId is required")
	}
	if len(p.PaymentTypes) == 0 {
		return fmt.Errorf("ipp: paymentTypes must not be empty")
	}
	for _, pt := range p.PaymentTypes {
		if !pt.Valid() {
			return fmt.Errorf("ipp: invalid paymentType: %q", pt)
		}
	}

	hasNonce := p.Nonce != nil
	hasPC := p.PaymentCommitment != nil

	if hasNonce != hasPC {
		return fmt.Errorf("ipp: nonce and paymentCommitment must both be present or both absent")
	}

	// Endpoint is required whenever the payer must reach the receiver
	// interactively: no nonce (interactive/multi mode), or any credit payment
	// type. A credit payer runs Authorize/Notify to draw the credit line, which
	// the embedded direct-pay nonce+paymentCommitment cannot substitute for, so
	// even a single-mode credit request needs the endpoint.
	if p.Endpoint == nil && (!hasNonce || ContainsCredit(p.PaymentTypes)) {
		return fmt.Errorf("ipp: endpoint is required for interactive or credit requests")
	}

	return nil
}
