package ipp

// This file is the single seam for IPP machine codes across the three error
// channels (HTTP envelope `code`, business `reject_code`, async relay
// `error_code`). It is intentionally the only place convenience constants live
// so a future generated `error-codes.json` can replace it wholesale.
//
// All code values are open string types: the server is free to add codes the
// SDK has not seen, and the SDK must surface them verbatim rather than reject
// or crash. The constants below name the values known at authoring time; never
// switch exhaustively on them without a default branch.
//
// All values are lowercase snake_case with no type prefix, matching the wire
// form so the same literal appears in the envelope `code`, the `reject_code`
// body field, and the relay `error_code` field.

// ── Channel ① — HTTP envelope error `type` ──────────────────────────

// ErrorType is the Stripe-style envelope `type` field. It selects which typed
// error subclass the SDK constructs. Open string: an unknown type folds into
// the catch-all *ApiError.
type ErrorType string

const (
	ErrorTypeInvalidRequest ErrorType = "invalid_request_error"
	ErrorTypeNotFound       ErrorType = "not_found_error"
	ErrorTypeConflict       ErrorType = "conflict_error"
	ErrorTypeRateLimit      ErrorType = "rate_limit_error"
	ErrorTypeApi            ErrorType = "api_error"
)

// ── Channel ② — business `reject_code` (HTTP 200) ───────────────────

// RejectCode is the machine-readable reason on a true business rejection
// (HTTP 200, status == "rejected"). Open string: protocol-level catalogue may
// grow without an SDK change.
type RejectCode string

const (
	// RejectInsufficientCreditBalance: draw amount exceeds the credit remainder.
	RejectInsufficientCreditBalance RejectCode = "insufficient_credit_balance"
	// RejectPendingDrawExists: the credit already has a pending draw.
	RejectPendingDrawExists RejectCode = "pending_draw_exists"
	// RejectCreditAlreadyCancelled: the credit has already been cancelled.
	RejectCreditAlreadyCancelled RejectCode = "credit_already_cancelled"
	// RejectNoPreallocatedCredential: a single-mode request has no preallocated credential.
	RejectNoPreallocatedCredential RejectCode = "no_preallocated_credential"
)

// ── Channel ③ — async relay `error_code` (HTTP 200) ─────────────────

// RelayErrorCode is the machine code carried on a failed asynchronous relay,
// read from RelayStatusResponse.ErrorCode when Status == "failed". Open string:
// the relay worker may emit codes the SDK has not enumerated.
type RelayErrorCode string

const (
	// RelayBadCalldata: calldata was not valid 0x-hex.
	RelayBadCalldata RelayErrorCode = "relay_bad_calldata"
	// RelayResolveFailed: commitment → PI → PR → operator resolution failed.
	RelayResolveFailed RelayErrorCode = "relay_resolve_failed"
	// RelaySubmitFailed: the on-chain submission deterministically failed.
	RelaySubmitFailed RelayErrorCode = "relay_submit_failed"
	// RelaySignFailed: signature-mode calldata assembly failed.
	RelaySignFailed RelayErrorCode = "relay_sign_failed"
	// RelayMerchantDisabled: the merchant has relay_enabled = false.
	RelayMerchantDisabled RelayErrorCode = "relay_merchant_disabled"
	// RelayUnknownMode: an unsupported relay mode was requested.
	RelayUnknownMode RelayErrorCode = "relay_unknown_mode"
	// RelayMissingWitness: the witness payload was missing.
	RelayMissingWitness RelayErrorCode = "relay_missing_witness"
	// RelayKMSSignFailed: the KMS signing operation failed.
	RelayKMSSignFailed RelayErrorCode = "relay_kms_sign_failed"
)
