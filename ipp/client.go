package ipp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// IPPClient is an HTTP client for the IPP interaction protocol.
type IPPClient struct {
	baseURL    string
	httpClient *http.Client
	// headers are sent on every request, applied before the per-request
	// Content-Type / Accept so a caller cannot accidentally clobber those.
	headers map[string]string
}

// Option customises an IPPClient at construction time.
type Option func(*IPPClient)

// WithHTTPClient overrides the default *http.Client (e.g. to set a timeout,
// custom transport, or proxy). Passing nil is a no-op.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *IPPClient) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithHeaders registers a set of headers sent on every request the client
// makes. This is a generic escape hatch for deployment-specific gateway
// requirements (auth tokens, routing hints, etc.); the SDK stays agnostic to
// any particular header's meaning. Repeated calls merge, last value wins.
// Content-Type / Accept set per request take precedence over these.
func WithHeaders(h map[string]string) Option {
	return func(c *IPPClient) {
		if c.headers == nil {
			c.headers = make(map[string]string, len(h))
		}
		for k, v := range h {
			c.headers[k] = v
		}
	}
}

// NewIPPClient creates a new IPP client pointing at the given endpoint.
// The endpoint should be the server root (e.g. "https://pay.example.com");
// the /ipp/v1 prefix is appended automatically.
func NewIPPClient(endpoint string, opts ...Option) *IPPClient {
	trimmed := strings.TrimRight(endpoint, "/")
	c := &IPPClient{
		baseURL:    trimmed + "/ipp/v1",
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// applyHeaders sets the client's per-call default headers on req. Callers set
// Content-Type / Accept afterwards so request-specific values win.
func (c *IPPClient) applyHeaders(req *http.Request) {
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
}

// ── Payer endpoints ─────────────────────────────────────────────────

// Info retrieves invoice metadata.
func (c *IPPClient) Info(params InfoRequest) (InfoResponse, error) {
	return post[InfoResponse](c, "info", params)
}

// Propose proposes a payment amount.
func (c *IPPClient) Propose(params ProposeRequest) (ProposeResponse, error) {
	return post[ProposeResponse](c, "propose", params)
}

// Authorize authorizes a credit line.
func (c *IPPClient) Authorize(params AuthorizeRequest) (AuthorizeResponse, error) {
	return post[AuthorizeResponse](c, "authorize", params)
}

// Status queries the current status of a payment by its commitment.
func (c *IPPClient) Status(noteCommitment string) (StatusResponse, error) {
	return get[StatusResponse](c, "status/"+noteCommitment)
}

// Notify sends a payment action notification (submitted or cancelled).
func (c *IPPClient) Notify(params NotifyRequest) (NotifyResponse, error) {
	return post[NotifyResponse](c, "notify", params)
}

// ── Receiver endpoints ──────────────────────────────────────────────

// Withdraw initiates a batch withdrawal.
func (c *IPPClient) Withdraw(params WithdrawRequest) (WithdrawResponse, error) {
	return post[WithdrawResponse](c, "withdraw", params)
}

// ── Relay endpoints ─────────────────────────────────────────────────

// Relay submits a transaction for relaying. Mode selects calldata submission,
// remote proving and submission (witness), or remote proving only (prove).
func (c *IPPClient) Relay(params RelayRequest) (RelayResponse, error) {
	return post[RelayResponse](c, "relay", params)
}

// RelayStatus queries the status of a relayed transaction.
func (c *IPPClient) RelayStatus(relayID string) (RelayStatusResponse, error) {
	return get[RelayStatusResponse](c, "relay/"+relayID)
}

// ── Internal helpers ────────────────────────────────────────────────

func post[T any](c *IPPClient, path string, body interface{}) (T, error) {
	var zero T

	if err := validatePublicRequest(body); err != nil {
		return zero, err
	}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return zero, fmt.Errorf("ipp: marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/"+path, bytes.NewReader(jsonBytes))
	if err != nil {
		return zero, fmt.Errorf("ipp: create request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	return doRequest[T](c, req)
}

func get[T any](c *IPPClient, path string) (T, error) {
	var zero T

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/"+path, nil)
	if err != nil {
		return zero, fmt.Errorf("ipp: create request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Accept", "application/json")

	return doRequest[T](c, req)
}

func doRequest[T any](c *IPPClient, req *http.Request) (T, error) {
	var zero T

	// Channel ⓪ — transport-level failure (network error, timeout): the
	// round-trip never produced an HTTP response, so this is a *ConnectionError,
	// not an envelope error.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, &ConnectionError{Message: "http request failed", Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, &ConnectionError{
			Message:    "reading response body failed",
			StatusCode: resp.StatusCode,
			Err:        err,
		}
	}

	// Channel ① — non-2xx HTTP: unwrap either documented envelope shape as a
	// typed error. A body that is not the envelope (HTML page, empty, etc.)
	// downgrades to a *ConnectionError inside newEnvelopeError.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, newEnvelopeError(resp.StatusCode, respBody)
	}

	// Channel ② — HTTP 200 business rejection (status == "rejected") is a normal
	// success value here, not an error: it is decoded into the response struct's
	// RejectCode/Reason fields and returned with a nil error so the caller
	// branches on it. Channel ③ relay failures likewise surface as result-state
	// fields on RelayStatusResponse, not as errors.
	var result T
	if err := json.Unmarshal(respBody, &result); err != nil {
		return zero, &ConnectionError{
			Message:    "decoding 2xx response body failed",
			StatusCode: resp.StatusCode,
			Raw:        respBody,
			Err:        err,
		}
	}
	if err := validatePublicResponse(result); err != nil {
		return zero, &ConnectionError{Message: "invalid 2xx IPP response", StatusCode: resp.StatusCode, Raw: respBody, Err: err}
	}
	return result, nil
}
