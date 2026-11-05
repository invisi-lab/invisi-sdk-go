package transaction

import (
	"errors"
	"math/big"
)

// NoteOutputExecutionOptions are payer choices, never written into the original IPR.
type NoteOutputExecutionOptions struct {
	Nonce         *big.Int
	AddressLookup string
}

func NewNoteOutput(input any, options ...NoteOutputExecutionOptions) (NoteOutput, error) {
	if len(options) > 1 {
		return NoteOutput{}, errors.New("expected at most one output execution option")
	}
	var original PaymentRequest
	var encoded string
	var e error
	switch v := input.(type) {
	case PaymentRequest:
		original = v
		encoded, e = EncodePaymentRequest(v)
	case string:
		encoded = v
		original, e = DecodePaymentRequest(v)
	default:
		return NoteOutput{}, errors.New("expected PaymentRequest or IPR")
	}
	if e != nil {
		return NoteOutput{}, e
	}
	// Decode to detach all mutable integers, recipient fields and application bytes.
	request, e := DecodePaymentRequest(encoded)
	if e != nil {
		return NoteOutput{}, e
	}
	if v, ok := original.Recipient.(EVMRecipient); ok {
		r := request.Recipient.(EVMRecipient)
		r.AddressLookup = v.AddressLookup
		request.Recipient = r
	}
	var execution NoteOutputExecutionOptions
	if len(options) == 1 {
		execution = options[0]
	}
	if execution.Nonce != nil && !CanonicalField(execution.Nonce) {
		return NoteOutput{}, errors.New("nonce must be a canonical BN254 field element")
	}
	if execution.AddressLookup != "" && execution.AddressLookup != "chain" && execution.AddressLookup != "protocol-rpc" {
		return NoteOutput{}, errors.New("invalid execution addressLookup")
	}
	choose := func(fixed *big.Int) (*big.Int, error) {
		if execution.Nonce == nil {
			return fixed, nil
		}
		if fixed != nil && fixed.Cmp(execution.Nonce) != 0 {
			return nil, errors.New("fixed request nonce cannot be overridden")
		}
		return new(big.Int).Set(execution.Nonce), nil
	}
	switch r := request.Recipient.(type) {
	case EVMRecipient:
		if execution.AddressLookup != "" {
			if r.Envelope != nil && !*r.Envelope {
				return NoteOutput{}, errors.New("addressLookup requires an encrypted EVM output")
			}
			r.AddressLookup = execution.AddressLookup
		}
		r.Nonce, e = choose(r.Nonce)
		request.Recipient = r
	case InvisiRecipient:
		if execution.AddressLookup != "" {
			return NoteOutput{}, errors.New("addressLookup requires an encrypted EVM output")
		}
		r.Nonce, e = choose(r.Nonce)
		request.Recipient = r
	case IdentityRecipient:
		if execution.Nonce != nil || execution.AddressLookup != "" {
			return NoteOutput{}, errors.New("identity output rejects execution options")
		}
	}
	if e != nil {
		return NoteOutput{}, e
	}
	output, e := createNoteOutput(request)
	if e != nil {
		return NoteOutput{}, e
	}
	output.saved.IPP = encoded
	return output, nil
}
