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
	Status            string  `json:"status"`
	Nonce             *string `json:"nonce,omitempty"`
	PaymentCommitment *string `json:"payment_commitment,omitempty"`
	// TreeID is the target Payment Tree id (SPEC-029 §3); returned when accepted. Lifecycle see 6.1.
	TreeID *string `json:"tree_id,omitempty"`
	Reason *string `json:"reason,omitempty"`
	// RejectCode is a machine-readable rejection reason (lowercase snake, no type
	// prefix), populated only when Status == "rejected". A rejection is a normal
	// HTTP 200 business outcome, not an HTTP error — callers branch on it without
	// catching a thrown error. See RejectCode constants in codes.go.
	RejectCode *string `json:"reject_code,omitempty"`
}

// AuthorizeRequest is sent by the payer to authorize a credit line.
//
// PayerCommitment is a payer-derived value: the anonymous payer derives it
// locally from payerMasterKey + creditId and sends it so the server folds it
// into the credit commitment. The payer does not send nullifier_key_hash or
// credit_nullifier_key — the server derives creditNullifierKey from
// receiverMasterKey and returns it (and its hash) in AuthorizeResponse.
type AuthorizeRequest struct {
	IdentityCommitment string `json:"identity_commitment"`
	Amount             string `json:"amount"`
	Expiry             string `json:"expiry"`
	PayerCommitment    string `json:"payer_commitment,omitempty"`
	// SubscriptionID links this authorization to a subscription; when set, the
	// server persists it and forwards it to the subscription side on the
	// deposit-confirmed event. Omitted for a standalone authorization.
	SubscriptionID string `json:"subscription_id,omitempty"`
}

// AuthorizeResponse is the receiver's reply to a credit authorization.
// Status is "accepted" or "rejected".
type AuthorizeResponse struct {
	Status string  `json:"status"`
	Nonce  *string `json:"nonce,omitempty"`
	// TreeID is the target Payment Tree id (SPEC-029 §3); returned when accepted. Lifecycle see 6.1.
	TreeID           *string `json:"tree_id,omitempty"`
	CreditCommitment *string `json:"credit_commitment,omitempty"`
	NullifierKeyHash *string `json:"nullifier_key_hash,omitempty"`
	// CreditNullifierKey (0x bytes32 hex) is the server-derived shared secret
	// Poseidon2(D_CREDIT_NKEY, receiverMasterKey, creditId), delivered to the payer
	// over TLS on an accepted authorize (IPP 6.2 authorize). The payer must verify
	// NullifierKeyHash == Poseidon2(D_NKHASH, CreditNullifierKey) before deposit.
	CreditNullifierKey *string `json:"credit_nullifier_key,omitempty"`
	Reason             *string `json:"reason,omitempty"`
	// RejectCode is a machine-readable rejection reason (lowercase snake, no type
	// prefix), populated only when Status == "rejected".
	RejectCode *string `json:"reject_code,omitempty"`
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

// WithdrawRequest is sent by the receiver to initiate a batch withdrawal.
type WithdrawRequest struct {
	// IdentityCommitment is advisory: the server locates leaves by
	// PaymentCommitments and ignores it; mode-3 vouchers omit it.
	IdentityCommitment string   `json:"identity_commitment,omitempty"`
	TokenAddress       string   `json:"token_address"`
	TreeID             string   `json:"tree_id"`
	PaymentCommitments []string `json:"payment_commitments"`
}

// WithdrawWitness contains server-authoritative merkle data for a withdrawal.
// Withdraw is non-custodial (the server has no nullifier key): nullifiers and the
// receiving address are client-supplied circuit inputs, not returned here.
type WithdrawWitness struct {
	MerkleRoot   string        `json:"merkle_root"`
	MerkleProofs []interface{} `json:"merkle_proofs"`
	TotalAmount  string        `json:"total_amount"`
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
	IdentityCommitment string   `json:"identity_commitment"`
	TokenAddress       string   `json:"token_address"`
	TreeID             string   `json:"tree_id"`
	CreditCommitments  []string `json:"credit_commitments"`
	Amounts            []string `json:"amounts"`
}

// CreditDrawWitness contains the ZK witness data for a credit draw proof.
type CreditDrawWitness struct {
	MerkleRoot           string        `json:"merkle_root"`
	Nullifiers           []string      `json:"nullifiers"`
	TotalAmount          string        `json:"total_amount"`
	ReceivingAddress     string        `json:"receiving_address"`
	RemainderCommitments []string      `json:"remainder_commitments"`
	Expiry               string        `json:"expiry"`
	MerkleProofs         []interface{} `json:"merkle_proofs"`
}

// CreditDrawResponse is the server's reply to a credit draw request.
type CreditDrawResponse struct {
	Status  string             `json:"status"`
	Witness *CreditDrawWitness `json:"witness,omitempty"`
	Circuit *string            `json:"circuit,omitempty"`
	Reason  *string            `json:"reason,omitempty"`
	// RejectCode carries a true business rejection (e.g. insufficient_credit_balance,
	// pending_draw_exists) on an HTTP 200 response; populated only when Status ==
	// "rejected". Syntax/resource/internal failures arrive as HTTP errors instead.
	RejectCode *string `json:"reject_code,omitempty"`
}

// CreditReclaimRequest is sent by the payer to reclaim an expired credit.
type CreditReclaimRequest struct {
	CreditCommitment string `json:"credit_commitment"`
	TokenAddress     string `json:"token_address"`
	TreeID           string `json:"tree_id"`
}

// CreditReclaimWitness contains the ZK witness data for a credit reclaim proof.
type CreditReclaimWitness struct {
	MerkleRoot       string      `json:"merkle_root"`
	Nullifier        string      `json:"nullifier"`
	Amount           string      `json:"amount"`
	ReceivingAddress string      `json:"receiving_address"`
	Expiry           string      `json:"expiry"`
	MerkleProof      interface{} `json:"merkle_proof"`
}

// CreditReclaimResponse is the server's reply to a credit reclaim request.
type CreditReclaimResponse struct {
	Status  string                `json:"status"`
	Witness *CreditReclaimWitness `json:"witness,omitempty"`
	Circuit *string               `json:"circuit,omitempty"`
	Reason  *string               `json:"reason,omitempty"`
	// RejectCode carries a true business rejection (e.g. credit_already_reclaimed,
	// credit_not_expired) on an HTTP 200 response; populated only when Status ==
	// "rejected". Syntax/resource/internal failures arrive as HTTP errors instead.
	RejectCode *string `json:"reject_code,omitempty"`
}

// CreditStatusDraw represents a single draw event on a credit line.
type CreditStatusDraw struct {
	Amount    string `json:"amount"`
	Timestamp string `json:"timestamp"`
	TxHash    string `json:"tx_hash"`
}

// CreditStatusResponse reports the current status of a credit line.
// Status is "active", "expired", or "reclaimed".
type CreditStatusResponse struct {
	Status          string             `json:"status"`
	OriginalAmount  string             `json:"original_amount"`
	DrawnAmount     string             `json:"drawn_amount"`
	RemainderAmount string             `json:"remainder_amount"`
	Expiry          string             `json:"expiry"`
	Draws           []CreditStatusDraw `json:"draws"`
}

// ── Transaction relay ───────────────────────────────────────────────

// RelayErrorCode (an open string type) and its convenience constants live in
// codes.go alongside the other wire machine-code seams. It is read from the
// asynchronous relay status poll's error_code field — see RelayStatusResponse.

// RelaySignatureRequest submits a signed transaction for relay.
type RelaySignatureRequest struct {
	Mode               string `json:"mode"` // "signature"
	Operation          string `json:"operation"`
	Network            string `json:"network"`
	Signature          string `json:"signature"`
	PaymentCommitment  string `json:"payment_commitment"`
	IdentityCommitment string `json:"identity_commitment"`
	TokenAddress       string `json:"token_address"`
	Amount             string `json:"amount"`
	Nonce              string `json:"nonce"`
	TreeID             string `json:"tree_id"`
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

// RelayRequest is the single payload accepted by POST /ipp/v1/relay across all
// four modes. The TS SDK models this as a discriminated union on Mode; Go has no
// union type, so — like RelaySignatureRequest et al. above — every mode's fields
// live on one struct and the unused ones are omitted on the wire via omitempty.
// Only Mode is always sent; it selects which subset the server reads.
//
// Wire-type notes (the gateway transcodes proto3 → JSON): tree_id and permit_v
// are uint32 → JSON numbers; sig_nonce / deadline / permit_deadline are proto
// uint64, which the gateway serializes as JSON strings, so they are strings here.
type RelayRequest struct {
	Mode      string `json:"mode"` // "signature" / "calldata" / "witness" / "prove"
	Operation string `json:"operation,omitempty"`
	Network   string `json:"network,omitempty"`

	// mode=signature
	Signature          string `json:"signature,omitempty"`
	PaymentCommitment  string `json:"payment_commitment,omitempty"`
	IdentityCommitment string `json:"identity_commitment,omitempty"`
	TokenAddress       string `json:"token_address,omitempty"`
	Amount             string `json:"amount,omitempty"`
	Nonce              string `json:"nonce,omitempty"`
	TreeID             uint32 `json:"tree_id,omitempty"`

	// mode=calldata
	Calldata string `json:"calldata,omitempty"`

	// mode=witness / mode=prove
	WitnessJSON string `json:"witness_json,omitempty"`
	Circuit     string `json:"circuit,omitempty"`

	// mode=signature: EIP-712 PayOpenWithSig + EIP-2612 permit structured params;
	// the relay assembles payOpenWithSig calldata from these. PayerAddress is
	// carried explicitly (not recovered from the signature) so ERC-1271 contract
	// wallets work.
	SigNonce       string `json:"sig_nonce,omitempty"`
	Deadline       string `json:"deadline,omitempty"`
	RefundLink     string `json:"refund_link,omitempty"`
	PermitToken    string `json:"permit_token,omitempty"`
	PermitDeadline string `json:"permit_deadline,omitempty"`
	PermitV        uint32 `json:"permit_v,omitempty"`
	PermitR        string `json:"permit_r,omitempty"`
	PermitS        string `json:"permit_s,omitempty"`
	PayerAddress   string `json:"payer_address,omitempty"`
}

// RelayResponse is the server's reply for signature/calldata/witness relay requests.
//
// This is the SYNCHRONOUS enqueue reply: on success it carries status="queued"
// plus a relay_id. The asynchronous execution outcome (including any failure
// error_code) is read later from RelayStatusResponse via RelayStatus();
// synchronous request failures arrive as a typed HTTP envelope error instead, so
// this reply carries no error_code field.
type RelayResponse struct {
	Status  string  `json:"status"`
	RelayID *string `json:"relay_id,omitempty"`
	TxHash  *string `json:"tx_hash,omitempty"`
	Reason  *string `json:"reason,omitempty"`
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
