// Package ipp implements IPP v1 messages and session-bound HTTP recovery without executing transactions.
package ipp

import "time"

type Mode string

const (
	Payment Mode = "payment"
	Credit  Mode = "credit"
)

type Link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}
type Asset struct {
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}
type InitialPaymentTerms struct {
	Amount string `json:"amount"`
}

type CreditTerms struct {
	InitialPayment  *InitialPaymentTerms `json:"initialPayment,omitempty"`
	Spender         string               `json:"spender"`
	AvailableCredit string               `json:"availableCredit"`
	Unlimited       bool                 `json:"unlimited"`
	SingleLimit     string               `json:"singleLimit"`
	DailyLimit      string               `json:"dailyLimit"`
	MonthlyLimit    string               `json:"monthlyLimit"`
	ExpiresAt       string               `json:"expiresAt"`
}
type Option struct {
	ID      string       `json:"id"`
	Method  string       `json:"method"`
	Network string       `json:"network"`
	Asset   Asset        `json:"asset"`
	Amount  string       `json:"amount,omitempty"`
	Credit  *CreditTerms `json:"credit,omitempty"`
}
type Merchant struct {
	Name string `json:"name"`
}
type SessionLinks struct {
	Confirm Link  `json:"confirm"`
	Request *Link `json:"request"`
}
type ExceptionSummary struct {
	Count           uint64 `json:"count"`
	UnresolvedCount uint64 `json:"unresolvedCount"`
}
type PaymentSession struct {
	IPPVersion       int              `json:"ippVersion"`
	Type             string           `json:"type"`
	Mode             Mode             `json:"mode"`
	SessionID        string           `json:"sessionId"`
	Status           string           `json:"status"`
	Merchant         Merchant         `json:"merchant"`
	Description      *string          `json:"description,omitempty"`
	Options          []Option         `json:"options"`
	ActiveRequestID  *string          `json:"activeRequestId"`
	CreatedAt        string           `json:"createdAt"`
	ExpiresAt        string           `json:"expiresAt"`
	Links            SessionLinks     `json:"links"`
	FailureReason    *string          `json:"failureReason"`
	Receipt          *Receipt         `json:"receipt"`
	EvidenceStatus   *string          `json:"evidenceStatus"`
	ExceptionSummary ExceptionSummary `json:"exceptionSummary"`
}
type Checkout struct {
	IPPVersion  int    `json:"ippVersion"`
	Type        string `json:"type"`
	SessionURL  string `json:"sessionUrl"`
	CheckoutURL string `json:"checkoutUrl"`
	ExpiresAt   string `json:"expiresAt"`
}
type CheckoutInput struct {
	IPPVersion  int      `json:"ippVersion"`
	Mode        Mode     `json:"mode"`
	Description *string  `json:"description,omitempty"`
	Options     []Option `json:"options"`
	ExpiresAt   string   `json:"expiresAt"`
}
type Recipient struct {
	Type     string `json:"type"`
	Identity string `json:"identity"`
}
type PaymentDetails struct {
	Network         string    `json:"network"`
	ProtocolAddress string    `json:"protocolAddress"`
	TokenAddress    string    `json:"tokenAddress"`
	Amount          string    `json:"amount"`
	Recipient       Recipient `json:"recipient"`
}
type CreditDetails struct {
	CreditTerms
	Operation       string `json:"operation"`
	Network         string `json:"network"`
	ProtocolAddress string `json:"protocolAddress"`
	TokenAddress    string `json:"tokenAddress"`
	CreditSharedKey string `json:"creditSharedKey"`
}
type RequestLinks struct {
	Credentials Link `json:"credentials"`
}
type Request struct {
	IPPVersion    int             `json:"ippVersion"`
	Type          string          `json:"type"`
	Mode          Mode            `json:"mode"`
	Method        string          `json:"method"`
	SessionID     string          `json:"sessionId"`
	RequestID     string          `json:"requestId"`
	OptionID      string          `json:"optionId"`
	Status        string          `json:"status"`
	Payment       *PaymentDetails `json:"payment,omitempty"`
	Credit        *CreditDetails  `json:"credit,omitempty"`
	CreatedAt     string          `json:"createdAt"`
	ExpiresAt     string          `json:"expiresAt"`
	Links         RequestLinks    `json:"links"`
	FailureReason *string         `json:"failureReason"`
}
type VerificationBlock struct {
	Number    string `json:"number"`
	Hash      string `json:"hash"`
	Timestamp string `json:"timestamp"`
}
type InitialPaymentReceipt struct {
	NoteCommitment string `json:"noteCommitment"`
	GrossAmount    string `json:"grossAmount"`
	NetAmount      string `json:"netAmount"`
	FeeAmount      string `json:"feeAmount"`
	FeeRecipient   string `json:"feeRecipient"`
}
type Receipt struct {
	InitialPayment  *InitialPaymentReceipt `json:"initialPayment,omitempty"`
	Network         string                 `json:"network"`
	ProtocolAddress string                 `json:"protocolAddress"`
	TransactionHash string                 `json:"transactionHash"`
	Commitment      string                 `json:"commitment"`
	BlockNumber     string                 `json:"blockNumber"`
	BlockHash       string                 `json:"blockHash"`
	CompletedAt     string                 `json:"completedAt"`
	CreditID        string                 `json:"creditId,omitempty"`
	CreditStatus    string                 `json:"creditStatus,omitempty"`
	Owner           string                 `json:"owner,omitempty"`
	GrantID         string                 `json:"grantId,omitempty"`

	VerificationBlock *VerificationBlock `json:"verificationBlock,omitempty"`
	ExpiresAt         string             `json:"expiresAt,omitempty"`
}
type Proof struct {
	Type            string `json:"type"`
	TransactionHash string `json:"transactionHash"`
	Commitment      string `json:"commitment"`
	Owner           string `json:"owner,omitempty"`
	GrantID         string `json:"grantId,omitempty"`
}
type Credential struct {
	IPPVersion int    `json:"ippVersion"`
	Type       string `json:"type"`
	Mode       Mode   `json:"mode"`
	Method     string `json:"method"`
	SessionID  string `json:"sessionId"`
	RequestID  string `json:"requestId"`
	Proof      Proof  `json:"proof"`
}
type CredentialAcknowledgement struct {
	IPPVersion int    `json:"ippVersion"`
	Type       string `json:"type"`
	SessionID  string `json:"sessionId"`
	RequestID  string `json:"requestId"`
	Status     string `json:"status"`
}
type CheckoutEntry struct {
	SessionURL string
	Token      string
}

// Error deliberately excludes remote messages and transport details that may contain private values.
type Error struct {
	Code          string
	Retryable     bool
	Status        int
	RetryAfter    time.Duration
	HasRetryAfter bool
}

func (e *Error) Error() string { return "IPP operation failed." }
func invalid() error           { return &Error{Code: "invalid_message"} }

// ErrorMessage is the wire error envelope. HTTP clients expose sanitized Error values instead.
type ErrorMessage struct {
	IPPVersion int          `json:"ippVersion"`
	Type       string       `json:"type"`
	Error      ErrorDetails `json:"error"`
}
type ErrorDetails struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
