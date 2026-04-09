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
}

// NewIPPClient creates a new IPP client pointing at the given endpoint.
// The endpoint should be the server root (e.g. "https://pay.example.com");
// the /ipp/v1 prefix is appended automatically.
func NewIPPClient(endpoint string) *IPPClient {
	trimmed := strings.TrimRight(endpoint, "/")
	return &IPPClient{
		baseURL:    trimmed + "/ipp/v1",
		httpClient: &http.Client{},
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
func (c *IPPClient) Status(paymentCommitment string) (StatusResponse, error) {
	return get[StatusResponse](c, "status/"+paymentCommitment)
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

// CreditDraw draws from a credit line.
func (c *IPPClient) CreditDraw(params CreditDrawRequest) (CreditDrawResponse, error) {
	return post[CreditDrawResponse](c, "credit-draw", params)
}

// CreditReclaim reclaims an expired credit.
func (c *IPPClient) CreditReclaim(params CreditReclaimRequest) (CreditReclaimResponse, error) {
	return post[CreditReclaimResponse](c, "credit-reclaim", params)
}

// CreditStatus queries the current status of a credit line.
func (c *IPPClient) CreditStatus(creditCommitment string) (CreditStatusResponse, error) {
	return get[CreditStatusResponse](c, "credit/"+creditCommitment)
}

// ── Relay endpoints ─────────────────────────────────────────────────

// RelayStatus queries the status of a relayed transaction.
func (c *IPPClient) RelayStatus(relayID string) (RelayStatusResponse, error) {
	return get[RelayStatusResponse](c, "relay/"+relayID)
}

// ── Internal helpers ────────────────────────────────────────────────

// IPPError wraps an ErrorResponse returned by the server.
type IPPError struct {
	StatusCode int
	Err        ErrorResponse
}

func (e *IPPError) Error() string {
	return e.Err.Error
}

func post[T any](c *IPPClient, path string, body interface{}) (T, error) {
	var zero T

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return zero, fmt.Errorf("ipp: marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/"+path, bytes.NewReader(jsonBytes))
	if err != nil {
		return zero, fmt.Errorf("ipp: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return doRequest[T](c, req)
}

func get[T any](c *IPPClient, path string) (T, error) {
	var zero T

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/"+path, nil)
	if err != nil {
		return zero, fmt.Errorf("ipp: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	return doRequest[T](c, req)
}

func doRequest[T any](c *IPPClient, req *http.Request) (T, error) {
	var zero T

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("ipp: http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("ipp: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil && errResp.Error != "" {
			return zero, &IPPError{StatusCode: resp.StatusCode, Err: errResp}
		}
		return zero, fmt.Errorf("ipp: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result T
	if err := json.Unmarshal(respBody, &result); err != nil {
		return zero, fmt.Errorf("ipp: unmarshal response: %w", err)
	}
	return result, nil
}
