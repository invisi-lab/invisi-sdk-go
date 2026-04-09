// Package ipp provides IPP (Invisible Payment Protocol) interaction types
// and an HTTP client for communicating with IPP-compliant servers.
package ipp

// ── Payer interaction ───────────────────────────────────────────────

// PaymasterConfig holds EIP-4337 paymaster service configuration for
// gasless payments.
type PaymasterConfig struct {
	URL      string                 `json:"url"`
	PolicyID *string                `json:"policyId,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// InfoRequest is sent by the payer to retrieve invoice metadata.
type InfoRequest struct {
	IdentityCommitment string `json:"identityCommitment"`
}

// InfoResponse contains invoice metadata returned by the receiver.
type InfoResponse struct {
	MerchantName *string          `json:"merchantName,omitempty"`
	Description  *string          `json:"description,omitempty"`
	Amount       *string          `json:"amount,omitempty"`
	Currency     *string          `json:"currency,omitempty"`
	ExpiresAt    *string          `json:"expiresAt,omitempty"`
	Paymaster    *PaymasterConfig `json:"paymaster,omitempty"`
}

// ProposeRequest is sent by the payer to propose a payment amount.
type ProposeRequest struct {
	IdentityCommitment string `json:"identityCommitment"`
	Amount             string `json:"amount"`
}

// ProposeResponse is the receiver's reply to a payment proposal.
// Status is "accepted" or "rejected".
type ProposeResponse struct {
	Status            string  `json:"status"`
	Nonce             *string `json:"nonce,omitempty"`
	PaymentCommitment *string `json:"paymentCommitment,omitempty"`
	Reason            *string `json:"reason,omitempty"`
}

// AuthorizeRequest is sent by the payer to authorize a credit line.
type AuthorizeRequest struct {
	IdentityCommitment string `json:"identityCommitment"`
	Amount             string `json:"amount"`
	Expiry             string `json:"expiry"`
}

// AuthorizeResponse is the receiver's reply to a credit authorization.
// Status is "accepted" or "rejected".
type AuthorizeResponse struct {
	Status           string  `json:"status"`
	Nonce            *string `json:"nonce,omitempty"`
	CreditCommitment *string `json:"creditCommitment,omitempty"`
	NullifierKeyHash *string `json:"nullifierKeyHash,omitempty"`
	Reason           *string `json:"reason,omitempty"`
}

// StatusResponse reports the current status of a payment.
// Status is one of: "unknown", "accepted", "cancelled", "submitted",
// "confirmed", "settled".
type StatusResponse struct {
	Status string  `json:"status"`
	TxHash *string `json:"txHash,omitempty"`
	Amount *string `json:"amount,omitempty"`
}

// NotifyRequest is sent by the payer to notify the receiver of a
// payment action. Action is "submitted" or "cancelled".
type NotifyRequest struct {
	Commitment string  `json:"commitment"`
	Action     string  `json:"action"`
	TxHash     *string `json:"txHash,omitempty"`
	Network    *string `json:"network,omitempty"`
}

// NotifyResponse acknowledges a notify request.
// Status is "received", "not_found", or "already_onchain".
type NotifyResponse struct {
	Status string `json:"status"`
}

// ── Receiver operations ─────────────────────────────────────────────

// WithdrawRequest is sent by the receiver to initiate a batch withdrawal.
type WithdrawRequest struct {
	IdentityCommitment string   `json:"identityCommitment"`
	TokenAddress       string   `json:"tokenAddress"`
	TreeID             string   `json:"treeId"`
	PaymentCommitments []string `json:"paymentCommitments"`
}

// WithdrawWitness contains the ZK witness data for a withdrawal proof.
type WithdrawWitness struct {
	MerkleRoot       string        `json:"merkleRoot"`
	Nullifiers       []string      `json:"nullifiers"`
	TotalAmount      string        `json:"totalAmount"`
	ReceivingAddress string        `json:"receivingAddress"`
	MerkleProofs     []interface{} `json:"merkleProofs"`
}

// WithdrawResponse is the server's reply to a withdrawal request.
type WithdrawResponse struct {
	Status  string           `json:"status"`
	Witness *WithdrawWitness `json:"witness,omitempty"`
	Circuit *string          `json:"circuit,omitempty"`
	Reason  *string          `json:"reason,omitempty"`
}

// CreditDrawRequest is sent by the receiver to draw from a credit line.
type CreditDrawRequest struct {
	IdentityCommitment string   `json:"identityCommitment"`
	TokenAddress       string   `json:"tokenAddress"`
	TreeID             string   `json:"treeId"`
	CreditCommitments  []string `json:"creditCommitments"`
	Amounts            []string `json:"amounts"`
}

// CreditDrawWitness contains the ZK witness data for a credit draw proof.
type CreditDrawWitness struct {
	MerkleRoot           string        `json:"merkleRoot"`
	Nullifiers           []string      `json:"nullifiers"`
	TotalAmount          string        `json:"totalAmount"`
	ReceivingAddress     string        `json:"receivingAddress"`
	RemainderCommitments []string      `json:"remainderCommitments"`
	Expiry               string        `json:"expiry"`
	MerkleProofs         []interface{} `json:"merkleProofs"`
}

// CreditDrawResponse is the server's reply to a credit draw request.
type CreditDrawResponse struct {
	Status  string             `json:"status"`
	Witness *CreditDrawWitness `json:"witness,omitempty"`
	Circuit *string            `json:"circuit,omitempty"`
	Reason  *string            `json:"reason,omitempty"`
}

// CreditReclaimRequest is sent by the payer to reclaim an expired credit.
type CreditReclaimRequest struct {
	CreditCommitment string `json:"creditCommitment"`
	TokenAddress     string `json:"tokenAddress"`
	TreeID           string `json:"treeId"`
}

// CreditReclaimWitness contains the ZK witness data for a credit reclaim proof.
type CreditReclaimWitness struct {
	MerkleRoot       string      `json:"merkleRoot"`
	Nullifier        string      `json:"nullifier"`
	Amount           string      `json:"amount"`
	ReceivingAddress string      `json:"receivingAddress"`
	Expiry           string      `json:"expiry"`
	MerkleProof      interface{} `json:"merkleProof"`
}

// CreditReclaimResponse is the server's reply to a credit reclaim request.
type CreditReclaimResponse struct {
	Status  string                `json:"status"`
	Witness *CreditReclaimWitness `json:"witness,omitempty"`
	Circuit *string               `json:"circuit,omitempty"`
	Reason  *string               `json:"reason,omitempty"`
}

// CreditStatusDraw represents a single draw event on a credit line.
type CreditStatusDraw struct {
	Amount    string `json:"amount"`
	Timestamp string `json:"timestamp"`
	TxHash    string `json:"txHash"`
}

// CreditStatusResponse reports the current status of a credit line.
// Status is "active", "expired", or "reclaimed".
type CreditStatusResponse struct {
	Status          string              `json:"status"`
	OriginalAmount  string              `json:"originalAmount"`
	DrawnAmount     string              `json:"drawnAmount"`
	RemainderAmount string              `json:"remainderAmount"`
	Expiry          string              `json:"expiry"`
	Draws           []CreditStatusDraw  `json:"draws"`
}

// ── Transaction relay ───────────────────────────────────────────────

// RelayErrorCode represents error codes returned by the relay service.
type RelayErrorCode string

const (
	RelayErrorInvalidSignature    RelayErrorCode = "invalid_signature"
	RelayErrorInvalidCalldata     RelayErrorCode = "invalid_calldata"
	RelayErrorInvalidWitness      RelayErrorCode = "invalid_witness"
	RelayErrorUnsupportedCircuit  RelayErrorCode = "unsupported_circuit"
	RelayErrorUnsupportedOp       RelayErrorCode = "unsupported_operation"
	RelayErrorInsufficientGas     RelayErrorCode = "insufficient_gas"
	RelayErrorQueueFull           RelayErrorCode = "queue_full"
	RelayErrorRateLimited         RelayErrorCode = "rate_limited"
	RelayErrorInternal            RelayErrorCode = "internal"
)

// RelaySignatureRequest submits a signed transaction for relay.
type RelaySignatureRequest struct {
	Mode               string `json:"mode"` // "signature"
	Operation          string `json:"operation"`
	Network            string `json:"network"`
	Signature          string `json:"signature"`
	PaymentCommitment  string `json:"paymentCommitment"`
	IdentityCommitment string `json:"identityCommitment"`
	TokenAddress       string `json:"tokenAddress"`
	Amount             string `json:"amount"`
	Nonce              string `json:"nonce"`
	TreeID             string `json:"treeId"`
}

// RelayCalldataRequest submits raw calldata for relay.
type RelayCalldataRequest struct {
	Mode      string `json:"mode"` // "calldata"
	Operation string `json:"operation"`
	Network   string `json:"network"`
	Calldata  string `json:"calldata"`
}

// RelayWitnessRequest submits a ZK witness for relay (server proves & submits).
type RelayWitnessRequest struct {
	Mode      string      `json:"mode"` // "witness"
	Operation string      `json:"operation"`
	Network   string      `json:"network"`
	Witness   interface{} `json:"witness"`
	Circuit   string      `json:"circuit"`
}

// RelayProveRequest submits a ZK witness for proof generation only.
type RelayProveRequest struct {
	Mode      string      `json:"mode"` // "prove"
	Operation string      `json:"operation"`
	Witness   interface{} `json:"witness"`
	Circuit   string      `json:"circuit"`
}

// RelayResponse is the server's reply for signature/calldata/witness relay requests.
type RelayResponse struct {
	Status    string          `json:"status"`
	RelayID   *string         `json:"relayId,omitempty"`
	TxHash    *string         `json:"txHash,omitempty"`
	ErrorCode *RelayErrorCode `json:"errorCode,omitempty"`
	Reason    *string         `json:"reason,omitempty"`
}

// RelayProveResponse is the server's reply for prove relay requests.
type RelayProveResponse struct {
	Status    string          `json:"status"`
	RelayID   *string         `json:"relayId,omitempty"`
	Proof     *string         `json:"proof,omitempty"`
	Calldata  *string         `json:"calldata,omitempty"`
	ErrorCode *RelayErrorCode `json:"errorCode,omitempty"`
	Reason    *string         `json:"reason,omitempty"`
}

// RelayStatusResponse reports the current status of a relayed transaction.
type RelayStatusResponse struct {
	Status    string          `json:"status"`
	Proof     *string         `json:"proof,omitempty"`
	Calldata  *string         `json:"calldata,omitempty"`
	TxHash    *string         `json:"txHash,omitempty"`
	ErrorCode *RelayErrorCode `json:"errorCode,omitempty"`
	Reason    *string         `json:"reason,omitempty"`
}

// ── Error ───────────────────────────────────────────────────────────

// ErrorResponse is returned by the server on HTTP 4xx/5xx errors.
type ErrorResponse struct {
	Error   string  `json:"error"`
	Message *string `json:"message,omitempty"`
}
