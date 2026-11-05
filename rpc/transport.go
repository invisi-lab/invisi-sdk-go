package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

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
