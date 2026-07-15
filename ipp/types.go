package ipp

// PaymentType identifies the payment mode in an IPP request.
type PaymentType string

const (
	PaymentTypeOpen   PaymentType = "open"
	PaymentTypeBlind  PaymentType = "blind"
	PaymentTypeCredit PaymentType = "credit"
)

// AllPaymentTypes is the single source of truth for the exhaustive set of valid PaymentType values.
// Consumers that need to enumerate valid types (e.g. for error messages) should reference this slice
// rather than duplicating the list.
var AllPaymentTypes = []PaymentType{PaymentTypeOpen, PaymentTypeBlind, PaymentTypeCredit}

// Valid reports whether t is a recognised PaymentType.
// This is the single source of truth for PaymentType validity; all validation must go through Valid().
func (t PaymentType) Valid() bool {
	switch t {
	case PaymentTypeOpen, PaymentTypeBlind, PaymentTypeCredit:
		return true
	default:
		return false
	}
}

// ContainsCredit reports whether credit is among the payment types. It is the
// single source for the credit-membership check: a credit payment type requires
// credit_id on the wire (IPP 6.1 encoding-tightening rule) so the payer can reach
// Authorize/Notify to run the credit line. SPEC-029 moved the endpoint
// requirement itself onto Mode (see ValidateIPPRequest); this helper remains for
// the credit_id-presence check.
func ContainsCredit(pts []PaymentType) bool {
	for _, pt := range pts {
		if pt == PaymentTypeCredit {
			return true
		}
	}
	return false
}

// IPPMessageType is the discriminator field for IPP envelopes.
type IPPMessageType string

const (
	MessageTypeRequest IPPMessageType = "request"
	MessageTypeVoucher IPPMessageType = "voucher"
)

// IPPRequest is the top-level envelope for an IPP payment request.
type IPPRequest struct {
	V       int               `json:"v" msgpack:"v"`
	Network string            `json:"network" msgpack:"network"`
	Type    IPPMessageType    `json:"type" msgpack:"type"`
	Meta    map[string]string `json:"meta,omitempty" msgpack:"meta,omitempty"`
	Payload IPPPayloadV1      `json:"payload" msgpack:"payload"`
}

// IPPVoucher is the top-level envelope for an IPP payment voucher (mode 3).
type IPPVoucher struct {
	V       int               `json:"v" msgpack:"v"`
	Network string            `json:"network" msgpack:"network"`
	Type    IPPMessageType    `json:"type" msgpack:"type"`
	Meta    map[string]string `json:"meta,omitempty" msgpack:"meta,omitempty"`
	Payload IPPVoucherV1      `json:"payload" msgpack:"payload"`
}

// IPPVoucherV1 is the v1 payload for an IPP payment voucher.
// It carries the nullifier key (secret) alongside the payment details.
type IPPVoucherV1 struct {
	// ReceivingAddress: with NullifierKey it recomputes identityCommitment /
	// paymentCommitment and is the settle circuit's receiving_address public input.
	ReceivingAddress  string `json:"receiving_address" msgpack:"receiving_address"`
	PaymentCommitment string `json:"payment_commitment" msgpack:"payment_commitment"`
	TokenAddress      string `json:"token_address" msgpack:"token_address"`
	Operator          string `json:"operator" msgpack:"operator"`
	Amount            string `json:"amount" msgpack:"amount"`
	Nonce             string `json:"nonce" msgpack:"nonce"`
	NullifierKey      string `json:"nullifier_key" msgpack:"nullifier_key"`
	TreeID            string `json:"tree_id" msgpack:"tree_id"`
	LeafIndex         string `json:"leaf_index,omitempty" msgpack:"leaf_index,omitempty"`
}

// IPPPayloadV1 is the v1 payload within an IPP request.
type IPPPayloadV1 struct {
	IdentityCommitment string        `json:"identity_commitment" msgpack:"identity_commitment"`
	TokenAddress       string        `json:"token_address" msgpack:"token_address"`
	Operator           string        `json:"operator" msgpack:"operator"`
	PaymentTypes       []PaymentType `json:"payment_types" msgpack:"payment_types"`
	// Mode is the receive-request mode (SPEC-029 §1): "single" = fixed amount, single-use (1a);
	// "multi" = many payers / many times (1b/2). Distinct from RelayRequest.Mode.
	Mode     string  `json:"mode" msgpack:"mode"`
	Amount   *string `json:"amount,omitempty" msgpack:"amount,omitempty"`
	Endpoint *string `json:"endpoint,omitempty" msgpack:"endpoint,omitempty"`
	// CreditID carries the one-time credit code's credit_id: required on the wire
	// when PaymentTypes includes 'credit' (IPP 6.1 encoding-tightening rule). It is
	// a public business ID (not a secret) and does not violate the QR secret
	// isolation invariant. Kept out of SDK baseline validation (same tier as the
	// endpoint rule).
	CreditID *string `json:"credit_id,omitempty" msgpack:"credit_id,omitempty"`
	// SPEC-029: payment materials (nonce / payment_commitment / tree_id) are no longer
	// prefilled in the request payload — nonce / payment_commitment are delivered via
	// propose; tree_id via Propose/AuthorizeResponse.
}

// Format identifies the wire format used to encode/decode an IPP request.
type Format string

const (
	FormatJSON    Format = "json"
	FormatCompact Format = "compact"
)

// IPPCurrentVersion is the current IPP protocol version.
const IPPCurrentVersion = 1
