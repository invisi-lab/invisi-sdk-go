package sdk

import (
	"errors"
	"fmt"
)

type InvisiErrorCode string

const (
	PreparationFailed       InvisiErrorCode = "PREPARATION_FAILED"
	MissingCapability       InvisiErrorCode = "MISSING_CAPABILITY"
	InvalidRecipient        InvisiErrorCode = "INVALID_RECIPIENT"
	EVMAddressNotRegistered InvisiErrorCode = "EVM_ADDRESS_NOT_REGISTERED"
	RecipientLookupFailed   InvisiErrorCode = "RECIPIENT_LOOKUP_FAILED"
)

// InvisiError is a stable preparation boundary; applications use Code and structured details.
type InvisiError struct {
	Code        InvisiErrorCode
	Operation   string
	Address     string
	OutputIndex *int
	Field       string
	Capability  string
	Cause       error
}

func (e *InvisiError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return string(e.Code)
}
func (e *InvisiError) Unwrap() error     { return e.Cause }
func (e *InvisiError) ErrorCode() string { return string(e.Code) }

type MissingCapabilityError struct{ Capability string }

func (e *MissingCapabilityError) Error() string {
	return fmt.Sprintf("missing capability: %s", e.Capability)
}
func wrapPreparation(operation string, err error) error {
	if err == nil {
		return nil
	}
	var invisi *InvisiError
	if errors.As(err, &invisi) {
		copy := *invisi
		if copy.Operation == "" {
			copy.Operation = operation
		}
		return &copy
	}
	result := &InvisiError{Code: PreparationFailed, Operation: operation, Cause: err}
	var missing *MissingCapabilityError
	if errors.As(err, &missing) {
		result.Code = MissingCapability
		result.Capability = missing.Capability
	}
	return result
}
func outputFailure(err error, index int, address, field string) error {
	var existing *InvisiError
	if errors.As(err, &existing) {
		copy := *existing
		copy.OutputIndex = &index
		if copy.Address == "" {
			copy.Address = address
		}
		if copy.Field == "" {
			copy.Field = field
		}
		return &copy
	}
	return &InvisiError{Code: PreparationFailed, Cause: err, Address: address, OutputIndex: &index, Field: field}
}
