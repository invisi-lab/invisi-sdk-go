package ipp

import (
	"encoding/json"
	"fmt"
)

// The gateway wraps every REST 4xx/5xx in a Stripe-style nested envelope:
//
//	{"error":{"type","code","message","param","request_id","details"}}
//
// Decoding therefore unwraps the outer {"error":{...}} layer first, then
// populates a typed error selected by the inner `type` / HTTP status. A 4xx/5xx
// body that does NOT parse as this envelope (or a transport failure before any
// response) is not an envelope error at all — it becomes a *ConnectionError.

// errorEnvelope is the outer {"error":{...}} wrapper the gateway always emits.
type errorEnvelope struct {
	Error envelopeBody `json:"error"`
}

// envelopeBody is the inner Stripe-style error object.
type envelopeBody struct {
	Type      string         `json:"type"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Param     string         `json:"param"`
	RequestID string         `json:"request_id"`
	Details   map[string]any `json:"details"`
}

// IPPError is the base type for every HTTP envelope error. The typed subclasses
// (InvalidRequestError, NotFoundError, ConflictError, RateLimitError, ApiError)
// embed it by pointer so errors.As unwraps through to *IPPError as well as to
// the concrete subclass.
//
// Code is an open string (see codes.go): it is surfaced verbatim and may be a
// value the SDK has never seen.
type IPPError struct {
	// Type is the envelope discriminator (e.g. "invalid_request_error").
	Type string
	// Code is the machine-readable code (lowercase snake, no type prefix).
	Code string
	// Message is the human-readable, gateway-sanitized message.
	Message string
	// Param names the offending request field for syntax errors ("" if none).
	Param string
	// RequestID correlates the failure with server-side logs.
	RequestID string
	// StatusCode is the HTTP status that carried the envelope.
	StatusCode int
	// Details carries client-consumable structured context (may be nil).
	Details map[string]any
	// Raw is the unparsed response body, retained for diagnostics.
	Raw []byte
}

func (e *IPPError) Error() string {
	// Prefer the most specific identifier available; the gateway sanitizes
	// Message for 5xx, so code/type still distinguish those.
	switch {
	case e.Code != "":
		return fmt.Sprintf("ipp: %s (code=%s, status=%d)", msgOr(e.Message, e.Code), e.Code, e.StatusCode)
	case e.Type != "":
		return fmt.Sprintf("ipp: %s (type=%s, status=%d)", msgOr(e.Message, e.Type), e.Type, e.StatusCode)
	default:
		return fmt.Sprintf("ipp: HTTP %d: %s", e.StatusCode, e.Message)
	}
}

func msgOr(msg, fallback string) string {
	if msg != "" {
		return msg
	}
	return fallback
}

// Each subclass embeds *IPPError for field access and additionally exposes it
// via Unwrap so errors.As reaches BOTH the concrete subclass and the base
// *IPPError. Embedding alone is not enough — errors.As does not descend into
// embedded fields, only the Unwrap chain.

// InvalidRequestError is a client-side syntax / precondition failure: HTTP 400
// (malformed or missing fields, bad batch construction, illegal amounts) and
// HTTP 412 (failed precondition, e.g. an expired receive request).
type InvalidRequestError struct{ *IPPError }

func (e *InvalidRequestError) Unwrap() error { return e.IPPError }

// NotFoundError is a missing-resource failure: HTTP 404 (receive request,
// credit, or commitment does not exist).
type NotFoundError struct{ *IPPError }

func (e *NotFoundError) Unwrap() error { return e.IPPError }

// ConflictError is a state-conflict failure: HTTP 409.
type ConflictError struct{ *IPPError }

func (e *ConflictError) Unwrap() error { return e.IPPError }

// RateLimitError signals the caller is being throttled: HTTP 429.
type RateLimitError struct{ *IPPError }

func (e *RateLimitError) Unwrap() error { return e.IPPError }

// ApiError is a server-side / internal failure: HTTP 5xx. It is also the
// catch-all for any unrecognized envelope type or status code, so callers that
// type-switch always have a terminal case.
type ApiError struct{ *IPPError }

func (e *ApiError) Unwrap() error { return e.IPPError }

// ConnectionError is NOT an HTTP envelope error. It is raised for transport-level
// problems: the HTTP round-trip itself failed (network error, timeout), or a
// 4xx/5xx body could not be unmarshaled into the nested envelope (e.g. a proxy
// returned an HTML error page or an empty body). Err wraps the underlying cause
// when present.
type ConnectionError struct {
	// Message describes the transport-level failure.
	Message string
	// StatusCode is the HTTP status when the failure was a non-envelope error
	// body (0 when the round-trip never produced a response).
	StatusCode int
	// Raw is the response body when one was read (nil on a round-trip failure).
	Raw []byte
	// Err is the underlying transport / decode error, if any.
	Err error
}

func (e *ConnectionError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("ipp: connection error: HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("ipp: connection error: %s", e.Message)
}

// Unwrap exposes the underlying transport error to errors.Is / errors.As.
func (e *ConnectionError) Unwrap() error { return e.Err }

// newEnvelopeError unwraps the nested {"error":{...}} envelope from a 4xx/5xx
// response body and returns the typed error selected by the inner `type` and
// HTTP status. If the body is not a parseable envelope (transport-level garbage
// reached a 4xx/5xx path), it returns a *ConnectionError so callers never see a
// half-populated envelope error.
func newEnvelopeError(status int, body []byte) error {
	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return &ConnectionError{
			Message:    "4xx/5xx body is not a JSON error envelope",
			StatusCode: status,
			Raw:        body,
			Err:        err,
		}
	}
	// A well-formed JSON document that carries none of the envelope fields (e.g.
	// `{}` or an unrelated object) is not the contract either — treat it as a
	// non-envelope body rather than fabricating an empty typed error.
	b := env.Error
	if b.Type == "" && b.Code == "" && b.Message == "" {
		return &ConnectionError{
			Message:    "4xx/5xx body did not contain an error envelope",
			StatusCode: status,
			Raw:        body,
		}
	}

	base := &IPPError{
		Type:       b.Type,
		Code:       b.Code,
		Message:    b.Message,
		Param:      b.Param,
		RequestID:  b.RequestID,
		StatusCode: status,
		Details:    b.Details,
		Raw:        body,
	}
	return classify(base)
}

// classify selects the typed subclass from the envelope type then the HTTP
// status. Status is the tiebreaker so a missing/unknown type still routes by
// the transport signal; anything unrecognized folds into the *ApiError
// catch-all. 401/403 are anonymous on IPP and have no dedicated subclass — they
// fold into the catch-all if ever seen.
func classify(base *IPPError) error {
	switch ErrorType(base.Type) {
	case ErrorTypeInvalidRequest:
		return &InvalidRequestError{base}
	case ErrorTypeNotFound:
		return &NotFoundError{base}
	case ErrorTypeConflict:
		return &ConflictError{base}
	case ErrorTypeRateLimit:
		return &RateLimitError{base}
	case ErrorTypeApi:
		return &ApiError{base}
	}

	switch base.StatusCode {
	case 400, 412:
		return &InvalidRequestError{base}
	case 404:
		return &NotFoundError{base}
	case 409:
		return &ConflictError{base}
	case 429:
		return &RateLimitError{base}
	default:
		// 5xx, plus 401/403 and any other unrecognized status.
		return &ApiError{base}
	}
}
