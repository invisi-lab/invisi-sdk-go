// Package ipp provides IPP (Invisible Payment Protocol) interaction types
// and an HTTP client for communicating with IPP-compliant servers.
package ipp

// ── Payer interaction ───────────────────────────────────────────────

// PaymasterConfig holds EIP-4337 paymaster service configuration for
// gasless payments.
type PaymasterConfig struct {
	URL      string                 `json:"url"`
	PolicyID *string                `json:"policy_id,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// InfoRequest is sent by the payer to retrieve invoice metadata.
type InfoRequest struct {
	IdentityCommitment string `json:"identity_commitment"`
}

// InfoResponse contains invoice metadata returned by the receiver.
type InfoResponse struct {
	MerchantName *string          `json:"merchant_name,omitempty"`
	Description  *string          `json:"description,omitempty"`
	Amount       *string          `json:"amount,omitempty"`
	Currency     *string          `json:"currency,omitempty"`
	ExpiresAt    *string          `json:"expires_at,omitempty"`
	Paymaster    *PaymasterConfig `json:"paymaster,omitempty"`
}

// ProposeRequest is sent by the payer to propose a payment amount.
type ProposeRequest struct {
	IdentityCommitment string `json:"identity_commitment"`
	Amount             string `json:"amount"`
}

// ProposeResponse is the receiver's reply to a payment proposal.
// Status is "accepted" or "rejected".
type ProposeResponse struct {
	Status         string  `json:"status"`
	Nonce          *string `json:"nonce,omitempty"`
	NoteCommitment *string `json:"note_commitment,omitempty"`
	Reason         *string `json:"reason,omitempty"`
	// RejectCode is a machine-readable rejection reason (lowercase snake, no type
	// prefix), populated only when Status == "rejected". A rejection is a normal
	// HTTP 200 business outcome, not an HTTP error — callers branch on it without
	// catching a thrown error. See RejectCode constants in codes.go.
	RejectCode *string `json:"reject_code,omitempty"`
}

// AuthorizeRequest authorizes a Credit with independent spender and owner identities.
type AuthorizeRequest struct {
	SpenderIdentityCommitment string `json:"spender_identity_commitment"`
	Amount                    string `json:"amount"`
	Expiry                    string `json:"expiry"`
	OwnerIdentityCommitment   string `json:"owner_identity_commitment"`
	SubscriptionID            string `json:"subscription_id,omitempty"`
}

// AuthorizeResponse carries the random, independently generated shared secret.
// Verify CreditSharedKeyHash locally before using CreditSharedKey.
type AuthorizeResponse struct {
	Status              string  `json:"status"`
	Nonce               *string `json:"nonce,omitempty"`
	CreditCommitment    *string `json:"credit_commitment,omitempty"`
	CreditSharedKeyHash *string `json:"credit_shared_key_hash,omitempty"`
	CreditSharedKey     *string `json:"credit_shared_key,omitempty"`
	Reason              *string `json:"reason,omitempty"`
	RejectCode          *string `json:"reject_code,omitempty"`
}

// StatusResponse reports the current status of a payment.
// Status is one of: "unknown", "accepted", "cancelled", "submitted",
// "confirmed", "settled".
type StatusResponse struct {
	Status string  `json:"status"`
	TxHash *string `json:"tx_hash,omitempty"`
	Amount *string `json:"amount,omitempty"`
}

// NotifyRequest is sent by the payer to notify the receiver of a
// payment action. Action is "submitted" or "cancelled".
type NotifyRequest struct {
	Commitment string  `json:"commitment"`
	Action     string  `json:"action"`
	TxHash     *string `json:"tx_hash,omitempty"`
	Network    *string `json:"network,omitempty"`
}

// NotifyResponse acknowledges a notify request.
// Status is "received", "not_found", or "already_onchain".
type NotifyResponse struct {
	Status string `json:"status"`
}

// ── Receiver operations ─────────────────────────────────────────────

// WithdrawRequest identifies the contract instance and notes to locate.
type WithdrawRequest struct {
	Network         string   `json:"network"`
	TokenAddress    string   `json:"token_address"`
	Operator        string   `json:"operator"`
	TreeID          string   `json:"tree_id"`
	NoteCommitments []string `json:"note_commitments"`
}

// MerkleMembershipProof contains only public Merkle membership data.
type MerkleMembershipProof struct {
	LeafIndex  string   `json:"leaf_index"`
	MerklePath []string `json:"merkle_path"`
}

// WithdrawWitness contains public Merkle data. TotalAmount is advisory;
// derive the spend balance, secrets and nullifiers from local notes.
type WithdrawWitness struct {
	MerkleRoot   string                  `json:"merkle_root"`
	MerkleProofs []MerkleMembershipProof `json:"merkle_proofs"`
	TotalAmount  string                  `json:"total_amount"`
}

type WithdrawResponse struct {
	Status  string           `json:"status"`
	Witness *WithdrawWitness `json:"witness,omitempty"`
	Reason  *string          `json:"reason,omitempty"`
}

// RelayOperation is an operation supported by the current protocol.
type RelayOperation string

const (
	RelayShield       RelayOperation = "shield"
	RelayTransact     RelayOperation = "transact"
	RelayCreateCredit RelayOperation = "create-credit"
	RelayCreditDraw   RelayOperation = "credit-draw"
	RelayCancelCredit RelayOperation = "cancel-credit"
	RelayTopUpCredit  RelayOperation = "top-up-credit"
)

// RelayCalldataRequest submits calldata for an explicit contract instance.
type RelayCalldataRequest struct {
	Mode            string         `json:"mode"`
	Operation       RelayOperation `json:"operation"`
	Network         string         `json:"network"`
	ProtocolAddress string         `json:"protocol_address"`
	Calldata        string         `json:"calldata"`
}

// RelayWitnessRequest explicitly delegates proof generation and submission.
type RelayWitnessRequest struct {
	Mode      string         `json:"mode"`
	Operation RelayOperation `json:"operation"`
	Network   string         `json:"network"`
	Witness   interface{}    `json:"witness"`
	Circuit   string         `json:"circuit"`
}

// RelayProveRequest explicitly delegates proof generation only.
type RelayProveRequest struct {
	Mode      string         `json:"mode"`
	Operation RelayOperation `json:"operation"`
	Witness   interface{}    `json:"witness"`
	Circuit   string         `json:"circuit"`
}

// RelayRequest represents calldata, witness or prove mode. Relay validates the
// mode-specific field set and circuit routing before sending a request.
type RelayRequest struct {
	Mode            string         `json:"mode"`
	Operation       RelayOperation `json:"operation"`
	Network         string         `json:"network,omitempty"`
	ProtocolAddress string         `json:"protocol_address,omitempty"`
	Calldata        string         `json:"calldata,omitempty"`
	Witness         interface{}    `json:"witness,omitempty"`
	Circuit         string         `json:"circuit,omitempty"`
}

// RelayResponse is the synchronous reply for a relay request, including prove mode.
//
// This is the SYNCHRONOUS enqueue reply: on success it carries status="queued"
// plus a relay_id. The asynchronous execution outcome (including any failure
// error_code) is read later from RelayStatusResponse via RelayStatus();
// synchronous request failures arrive as a typed HTTP envelope error instead, so
// this reply carries no error_code field.
type RelayResponse struct {
	Proof    *string `json:"proof,omitempty"`
	Calldata *string `json:"calldata,omitempty"`
	Status   string  `json:"status"`
	RelayID  *string `json:"relay_id,omitempty"`
	TxHash   *string `json:"tx_hash,omitempty"`
	Reason   *string `json:"reason,omitempty"`
}

// RelayProveResponse is the server's reply for prove relay requests. Like
// RelayResponse this is the synchronous enqueue reply and carries no error_code;
// the async prove outcome is read from RelayStatusResponse via RelayStatus().
type RelayProveResponse struct {
	Status   string  `json:"status"`
	RelayID  *string `json:"relay_id,omitempty"`
	Proof    *string `json:"proof,omitempty"`
	Calldata *string `json:"calldata,omitempty"`
	Reason   *string `json:"reason,omitempty"`
}

// RelayStatusResponse reports the current status of a relayed transaction.
//
// This is where an asynchronous relay FAILURE surfaces: when Status == "failed"
// the ErrorCode field carries the relay machine code (see RelayErrorCode in
// codes.go), mirroring Stripe's Payout.failure_code result-state pattern. It is
// a normal HTTP 200 result-state read, not a thrown error.
type RelayStatusResponse struct {
	Status    string          `json:"status"`
	Proof     *string         `json:"proof,omitempty"`
	Calldata  *string         `json:"calldata,omitempty"`
	TxHash    *string         `json:"tx_hash,omitempty"`
	ErrorCode *RelayErrorCode `json:"error_code,omitempty"`
	Reason    *string         `json:"reason,omitempty"`
}
