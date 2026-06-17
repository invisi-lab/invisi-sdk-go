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

// IPPRequest is the top-level envelope for an IPP payment request.
type IPPRequest struct {
	V       int               `json:"v" msgpack:"v"`
	Network string            `json:"network" msgpack:"network"`
	Meta    map[string]string `json:"meta,omitempty" msgpack:"meta,omitempty"`
	Payload IPPPayloadV1      `json:"payload" msgpack:"payload"`
}

// IPPPayloadV1 is the v1 payload within an IPP request.
type IPPPayloadV1 struct {
	IdentityCommitment string        `json:"identityCommitment" msgpack:"identityCommitment"`
	TokenAddress       string        `json:"tokenAddress" msgpack:"tokenAddress"`
	Operator           string        `json:"operator" msgpack:"operator"`
	TreeID             string        `json:"treeId" msgpack:"treeId"`
	PaymentTypes       []PaymentType `json:"paymentTypes" msgpack:"paymentTypes"`
	Amount             *string       `json:"amount,omitempty" msgpack:"amount,omitempty"`
	Nonce              *string       `json:"nonce,omitempty" msgpack:"nonce,omitempty"`
	PaymentCommitment  *string       `json:"paymentCommitment,omitempty" msgpack:"paymentCommitment,omitempty"`
	Endpoint           *string       `json:"endpoint,omitempty" msgpack:"endpoint,omitempty"`
}

// Format identifies the wire format used to encode/decode an IPP request.
type Format string

const (
	FormatJSON    Format = "json"
	FormatCompact Format = "compact"
)

// IPPCurrentVersion is the current IPP protocol version.
const IPPCurrentVersion = 1
