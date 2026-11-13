package broadcaster

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
)

type Configuration struct {
	Endpoint                  string
	QuoteSigner, FeeRecipient common.Address
}
type Options struct{ HTTPClient *http.Client }
type Client struct {
	configuration Configuration
	http          *http.Client
	sequence      atomic.Uint64
}
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string { return fmt.Sprintf("broadcaster RPC (%d): %s", e.Code, e.Message) }

type Fee struct {
	FeeRecipient common.Address `json:"feeRecipient"`
	FeeAmount    string         `json:"feeAmount"`
	MaxGas       string         `json:"maxGas"`
}
type Status struct {
	Confirmation    *Confirmation `json:"confirmation,omitempty"`
	ID              string        `json:"id"`
	State           string        `json:"state"`
	TransactionHash string        `json:"transactionHash,omitempty"`
}

func NewClient(config Configuration, options ...Options) (*Client, error) {
	u, e := url.Parse(config.Endpoint)
	if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || config.QuoteSigner == (common.Address{}) || config.FeeRecipient == (common.Address{}) || len(options) > 1 {
		return nil, errors.New("invalid broadcaster configuration: HTTPS and nonzero registered addresses required")
	}
	h := &http.Client{Timeout: 30 * time.Second}
	if len(options) == 1 && options[0].HTTPClient != nil {
		copy := *options[0].HTTPClient
		h = &copy
	}
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{configuration: config, http: h}, nil
}
func (c *Client) Configuration() Configuration { return c.configuration }
func (c *Client) call(ctx context.Context, method string, params any, result any) error {
	id := c.sequence.Add(1)
	args := []any{}
	if params != nil {
		args = append(args, params)
	}
	b, e := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": args})
	if e != nil {
		return e
	}
	req, e := http.NewRequestWithContext(ctx, "POST", c.configuration.Endpoint, bytes.NewReader(b))
	if e != nil {
		return e
	}
	req.Header.Set("Content-Type", "application/json")
	r, e := c.http.Do(req)
	if e != nil {
		return e
	}
	defer r.Body.Close()
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return fmt.Errorf("broadcaster HTTP %d", r.StatusCode)
	}
	b, e = io.ReadAll(io.LimitReader(r.Body, 65537))
	if e != nil {
		return e
	}
	if len(b) > 65536 {
		return errors.New("broadcaster response too large")
	}
	var envelope map[string]json.RawMessage
	if json.Unmarshal(b, &envelope) != nil {
		return errors.New("invalid broadcaster JSON-RPC response")
	}
	var version string
	var rid uint64
	if json.Unmarshal(envelope["jsonrpc"], &version) != nil || version != "2.0" || json.Unmarshal(envelope["id"], &rid) != nil || rid != id {
		return errors.New("invalid broadcaster JSON-RPC response")
	}
	res, hasResult := envelope["result"]
	errRaw, hasError := envelope["error"]
	if hasResult == hasError {
		return errors.New("invalid broadcaster JSON-RPC response")
	}
	if hasError {
		var er struct {
			Code    *int    `json:"code"`
			Message *string `json:"message"`
		}
		if json.Unmarshal(errRaw, &er) != nil || er.Code == nil || er.Message == nil {
			return errors.New("invalid broadcaster JSON-RPC error")
		}
		return &RPCError{*er.Code, *er.Message}
	}
	return json.Unmarshal(res, result)
}
func (c *Client) EstimateQuote(ctx context.Context, data []byte, schedules ...*Schedule) (Fee, error) {
	if _, e := DecodeIntent(data); e != nil {
		return Fee{}, e
	}
	p := map[string]any{}
	if len(data) > 0 {
		p["data"] = "0x" + common.Bytes2Hex(data)
	}
	if e := addSchedule(p, schedules); e != nil {
		return Fee{}, e
	}
	var f Fee
	e := c.call(ctx, "broadcaster_estimateQuote", p, &f)
	if e != nil {
		return f, e
	}
	amount, e := decimal(f.FeeAmount)
	if e != nil {
		return f, e
	}
	gas, e := decimal(f.MaxGas)
	if e != nil {
		return f, e
	}
	if f.FeeRecipient != c.configuration.FeeRecipient || amount.Sign() <= 0 || amount.BitLen() > 128 || gas.Sign() <= 0 {
		return f, errors.New("invalid broadcaster fee")
	}
	return f, nil
}
func (c *Client) GetQuote(ctx context.Context, data []byte, trust Trust, schedules ...*Schedule) (SignedQuote, error) {
	var s SignedQuote
	params := map[string]any{"data": "0x" + common.Bytes2Hex(data)}
	if e := addSchedule(params, schedules); e != nil {
		return s, e
	}
	_, e := DecodeIntent(data)
	if e != nil {
		return s, e
	}
	if e := c.call(ctx, "broadcaster_getQuote", params, &s); e != nil {
		return s, e
	}
	if !sameSchedule(s.Quote.Schedule, schedules) {
		return s, errors.New("broadcaster schedule mismatch")
	}
	trust.Signer = c.configuration.QuoteSigner
	if e := VerifyQuote(s, trust, data, uint64(time.Now().Unix())); e != nil {
		return s, e
	}
	if s.Quote.FeeRecipient != c.configuration.FeeRecipient {
		return s, errors.New("unregistered fee recipient")
	}
	trust.ChainID = new(big.Int).Set(trust.ChainID)
	s.Context = &trust
	return s, nil
}
func checkStatus(s Status, id string) error {
	if s.ID == "" || len(s.ID) > 256 || (id != "" && s.ID != id) {
		return errors.New("invalid broadcaster status")
	}
	switch s.State {
	case "pending", "submitted", "confirmed", "reverted", "uncertain", "queued", "proving", "rejected", "expired", "scheduled", "preparing", "cancelled":
	default:
		return errors.New("invalid broadcaster state")
	}
	if s.Confirmation != nil {
		switch s.State {
		case "queued", "proving", "rejected", "expired", "scheduled", "preparing", "cancelled":
			return errors.New("unexpected broadcaster confirmation")
		}
		if s.Confirmation.Level != "pending" && s.TransactionHash == "" {
			return errors.New("confirmation transaction hash missing")
		}
	}
	if s.TransactionHash != "" {
		b := common.FromHex(s.TransactionHash)
		if len(s.TransactionHash) != 66 || len(b) != 32 || !strings.HasPrefix(s.TransactionHash, "0x") {
			return errors.New("invalid broadcaster transaction hash")
		}
	}
	return nil
}
func (c *Client) Submit(ctx context.Context, data []byte, s SignedQuote, sidecars ...[]provenance.ProofSidecar) (Status, error) {
	var selected []provenance.ProofSidecar
	if len(sidecars) > 1 {
		return Status{}, errors.New("at most one sidecar list required")
	}
	if len(sidecars) == 1 {
		selected = sidecars[0]
	}
	return c.SubmitMember(ctx, data, s, selected, nil)
}

func (c *Client) Status(ctx context.Context, id string) (Status, error) {
	var out Status
	if id == "" || len(id) > 256 {
		return out, errors.New("invalid broadcaster submission ID")
	}
	e := c.call(ctx, "broadcaster_getTransactionStatus", map[string]any{"id": id}, &out)
	if e != nil {
		return out, e
	}
	return out, checkStatus(out, id)
}
