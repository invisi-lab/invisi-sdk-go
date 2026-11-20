package rpc

import (
	"context"
	"encoding/json"
	"errors"
)

// The deployed Protocol RPC accepts at most 20 calls in a JSON-RPC batch.
const maxBatchCalls = 20

type batchRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      uint64 `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// callBatch authenticates every envelope before returning any results. It does
// not retry individual failures: a partially observed history page is unusable.
func (c *ProtocolRpcClient) callBatch(ctx context.Context, chain uint64, method string, params []any) ([]json.RawMessage, error) {
	if len(params) == 0 || len(params) > maxBatchCalls {
		return nil, errors.New("invalid RPC batch size")
	}
	target, err := c.endpoint(chain)
	if err != nil {
		return nil, err
	}
	requests := make([]batchRequest, len(params))
	positions := make(map[uint64]int, len(params))
	for i, p := range params {
		id := c.sequence.Add(1)
		requests[i] = batchRequest{"2.0", id, method, p}
		positions[id] = i
	}
	body, err := json.Marshal(requests)
	if err != nil {
		return nil, err
	}
	var responses []struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      uint64          `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   json.RawMessage `json:"error"`
	}
	if err = c.request(ctx, target, "POST", body, &responses); err != nil {
		return nil, err
	}
	if len(responses) != len(requests) {
		return nil, errors.New("incomplete RPC batch")
	}
	results := make([]json.RawMessage, len(params))
	seen := make(map[uint64]bool, len(params))
	for _, response := range responses {
		i, ok := positions[response.ID]
		if response.JSONRPC != "2.0" || !ok || seen[response.ID] {
			return nil, errors.New("invalid RPC batch envelope")
		}
		seen[response.ID] = true
		if response.Error != nil {
			if response.Result != nil {
				return nil, errors.New("ambiguous RPC batch response")
			}
			var remote struct {
				Code    *int            `json:"code"`
				Message *string         `json:"message"`
				Data    json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(response.Error, &remote); err != nil || remote.Code == nil || remote.Message == nil {
				return nil, errors.New("invalid RPC batch error")
			}
			return nil, &ProtocolRpcError{Code: *remote.Code, Message: *remote.Message, Data: remote.Data}
		}
		if response.Result == nil {
			return nil, errors.New("missing RPC batch result")
		}
		results[i] = response.Result
	}
	return results, nil
}
