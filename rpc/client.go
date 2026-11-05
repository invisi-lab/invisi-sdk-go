package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ClientOptions struct {
	HTTPClient *http.Client
	Headers    http.Header
	Deployment *ProtocolDeployment
}
type ProtocolRpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *ProtocolRpcError) Error() string {
	return fmt.Sprintf("protocol RPC (%d): %s", e.Code, e.Message)
}

type HTTPError struct {
	StatusCode int
	RetryAfter string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("protocol RPC HTTP %d", e.StatusCode) }

type ProtocolRpcClient struct {
	url        url.URL
	http       *http.Client
	headers    http.Header
	sequence   atomic.Uint64
	mu         sync.Mutex
	deployment *ProtocolDeployment
	provider   ChainProvider
}

func NewProtocolRpcClient(endpoint string, options ...ClientOptions) (*ProtocolRpcClient, error) {
	u, e := url.Parse(endpoint)
	if e != nil || u == nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
		return nil, errors.New("invalid Protocol RPC URL")
	}
	if len(options) > 1 {
		return nil, errors.New("expected at most one client option")
	}
	var opt ClientOptions
	if len(options) == 1 {
		opt = options[0]
	}
	client := opt.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	clone := *client
	clone.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c := &ProtocolRpcClient{url: *u, http: &clone, headers: opt.Headers.Clone()}
	if opt.Deployment != nil {
		if e := opt.Deployment.Validate(); e != nil {
			return nil, e
		}
		d := cloneDeployment(*opt.Deployment)
		if d.Registry != nil {
			r := *d.Registry
			d.Registry = &r
		}
		c.deployment = &d
	}
	return c, nil
}
func (c *ProtocolRpcClient) WithProvider(provider ChainProvider) *ProtocolRpcClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	var d *ProtocolDeployment
	if c.deployment != nil {
		copy := cloneDeployment(*c.deployment)
		d = &copy
	}
	return &ProtocolRpcClient{url: c.url, http: c.http, headers: c.headers.Clone(), deployment: d, provider: provider}
}
func (c *ProtocolRpcClient) request(ctx context.Context, target url.URL, method string, body []byte, result any) error {
	req, e := http.NewRequestWithContext(ctx, method, target.String(), bytes.NewReader(body))
	if e != nil {
		return e
	}
	req.Header = c.headers.Clone()
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	response, e := c.http.Do(req)
	if e != nil {
		return e
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &HTTPError{response.StatusCode, response.Header.Get("Retry-After")}
	}
	raw, e := io.ReadAll(io.LimitReader(response.Body, 32<<20+1))
	if e != nil {
		return e
	}
	if len(raw) > 32<<20 {
		return errors.New("RPC response too large")
	}
	return json.Unmarshal(raw, result)
}
func (c *ProtocolRpcClient) endpoint(chain uint64) (url.URL, error) {
	target := c.url
	path := strings.TrimSuffix(target.Path, "/")
	if strings.HasSuffix(path, "/v1") {
		target.Path = path + "/chains/" + strconv.FormatUint(chain, 10)
	} else if !strings.HasSuffix(path, "/v1/chains/"+strconv.FormatUint(chain, 10)) {
		return target, errors.New("RPC endpoint chain mismatch")
	}
	return target, nil
}
func (c *ProtocolRpcClient) call(ctx context.Context, chain uint64, method string, params any, result any) error {
	target, e := c.endpoint(chain)
	if e != nil {
		return e
	}
	id := c.sequence.Add(1)
	body, e := json.Marshal(struct {
		JSONRPC string `json:"jsonrpc"`
		ID      uint64 `json:"id"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}{"2.0", id, method, params})
	if e != nil {
		return e
	}
	var response struct {
		JSONRPC string            `json:"jsonrpc"`
		ID      uint64            `json:"id"`
		Result  json.RawMessage   `json:"result"`
		Error   *ProtocolRpcError `json:"error"`
	}
	if e = c.request(ctx, target, "POST", body, &response); e != nil {
		return e
	}
	if response.JSONRPC != "2.0" || response.ID != id {
		return errors.New("invalid JSON-RPC envelope")
	}
	if response.Error != nil {
		return response.Error
	}
	if response.Result == nil {
		return errors.New("missing RPC result")
	}
	return json.Unmarshal(response.Result, result)
}
func (c *ProtocolRpcClient) GetChains(ctx context.Context) ([]Chain, error) {
	target := c.url
	if !strings.HasSuffix(strings.TrimSuffix(target.Path, "/"), "/v1") {
		return nil, errors.New("chain discovery requires service /v1 URL")
	}
	var response struct {
		APIVersion string  `json:"apiVersion"`
		Chains     []Chain `json:"chains"`
	}
	if e := c.request(ctx, target, "GET", nil, &response); e != nil {
		return nil, e
	}
	if response.APIVersion != "v1" || response.Chains == nil {
		return nil, errors.New("unsupported service response")
	}
	seen := map[string]bool{}
	for i := range response.Chains {
		ch := &response.Chains[i]
		n, e := strconv.ParseUint(ch.ChainID, 10, 64)
		if e != nil || n == 0 || strconv.FormatUint(n, 10) != ch.ChainID || seen[ch.ChainID] || ch.ProtocolVersion != "1" || ch.RPCPath != "/v1/chains/"+ch.ChainID {
			return nil, errors.New("invalid chain discovery entry")
		}
		seen[ch.ChainID] = true
		ch.RPCURL = target.String()
		u := target
		u.Path = strings.TrimSuffix(u.Path, "/") + "/chains/" + ch.ChainID
		ch.RPCURL = u.String()
	}
	return response.Chains, nil
}
func (c *ProtocolRpcClient) GetDeployment(ctx context.Context, chain uint64) (ProtocolDeployment, error) {
	var response struct {
		APIVersion      string             `json:"apiVersion"`
		ProtocolVersion string             `json:"protocolVersion"`
		Deployment      ProtocolDeployment `json:"deployment"`
	}
	if e := c.call(ctx, chain, "invisi_getDeployment", []any{}, &response); e != nil {
		return ProtocolDeployment{}, e
	}
	d := response.Deployment
	if e := d.Validate(); e != nil {
		return d, e
	}
	if response.APIVersion != "v1" || response.ProtocolVersion != "1" || d.ChainID != strconv.FormatUint(chain, 10) {
		return d, errors.New("deployment version or chain mismatch")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.deployment != nil {
		a, _ := json.Marshal(c.deployment)
		b, _ := json.Marshal(d)
		if !bytes.Equal(a, b) {
			return d, errors.New("deployment pins mismatch")
		}
	}
	copy := cloneDeployment(d)
	c.deployment = &copy
	return cloneDeployment(d), nil
}
func (c *ProtocolRpcClient) context(context DeploymentContext) (ProtocolDeployment, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.deployment == nil || c.deployment.ChainID != strconv.FormatUint(context.ChainID, 10) || c.deployment.Protocol.Address != context.ProtocolAddress {
		return ProtocolDeployment{}, errors.New("deployment not loaded or context mismatch")
	}
	return cloneDeployment(*c.deployment), nil
}

func cloneDeployment(d ProtocolDeployment) ProtocolDeployment {
	if d.Registry != nil {
		registry := *d.Registry
		d.Registry = &registry
	}
	if d.BroadcasterRegistry != nil {
		r := *d.BroadcasterRegistry
		d.BroadcasterRegistry = &r
	}
	return d
}
