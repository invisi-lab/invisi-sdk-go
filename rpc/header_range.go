package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const maxHeaderRange = 256

type headerRangeRequest struct {
	From uint64 `json:"fromBlock"`
	To   uint64 `json:"toBlock"`
}

// Geth permits absent zero-valued seal and miner fields in JSON; persisted
// historical headers must contain them to represent the complete original RLP.
type historicalHeader types.Header

func (h *historicalHeader) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	for _, name := range []string{"miner", "mixHash", "nonce"} {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			return errors.New("incomplete historical header")
		}
	}
	// RLP fills gaps before later fork fields with zero values. Hash equality
	// alone therefore cannot detect missing zero-valued numeric fields.
	optional := []string{"baseFeePerGas", "withdrawalsRoot", "blobGasUsed", "excessBlobGas", "parentBeaconBlockRoot", "requestsHash", "slotNumber"}
	later := false
	for i := len(optional) - 1; i >= 0; i-- {
		value, present := fields[optional[i]]
		present = present && string(value) != "null"
		if later && !present {
			return errors.New("incomplete historical fork fields")
		}
		later = later || present
	}
	return json.Unmarshal(raw, (*types.Header)(h))
}

type headerRangeResponse struct {
	APIVersion      string             `json:"apiVersion"`
	ProtocolVersion string             `json:"protocolVersion"`
	Deployment      ProtocolDeployment `json:"deployment"`
	Range           struct {
		From      *uint64     `json:"fromBlock"`
		To        *uint64     `json:"toBlock"`
		BlockHash common.Hash `json:"blockHash"`
	} `json:"range"`
	Headers []struct {
		Header     *historicalHeader `json:"header"`
		HeaderHash common.Hash       `json:"headerHash"`
	} `json:"headers"`
}

// historicalHeaderRange checks the envelope before classifying remote errors.
// An ambiguous or malformed envelope must never enable the legacy fallback.
func (p *TrustedProvider) historicalHeaderRange(ctx context.Context, from, to uint64) (headerRangeResponse, error) {
	var out headerRangeResponse
	c := p.client
	target, err := c.endpoint(p.scope.ChainID)
	if err != nil {
		return out, err
	}
	id := c.sequence.Add(1)
	body, err := json.Marshal(batchRequest{"2.0", id, "invisi_getBlockHeaders", []any{headerRangeRequest{from, to}}})
	if err != nil {
		return out, err
	}
	var response struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      uint64          `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   json.RawMessage `json:"error"`
	}
	if err = c.request(ctx, target, "POST", body, &response); err != nil {
		return out, err
	}
	if response.JSONRPC != "2.0" || response.ID != id {
		return out, errors.New("invalid header range RPC envelope")
	}
	if response.Error != nil {
		if response.Result != nil {
			return out, errors.New("ambiguous header range RPC response")
		}
		var remote struct {
			Code    *int            `json:"code"`
			Message *string         `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		if err = json.Unmarshal(response.Error, &remote); err != nil || remote.Code == nil || remote.Message == nil {
			return out, errors.New("invalid header range RPC error")
		}
		return out, &ProtocolRpcError{Code: *remote.Code, Message: *remote.Message, Data: remote.Data}
	}
	if response.Result == nil {
		return out, errors.New("missing header range RPC result")
	}
	err = json.Unmarshal(response.Result, &out)
	return out, err
}

func (p *TrustedProvider) historicalHeaders(ctx context.Context, numbers []*big.Int) ([]*types.Header, error) {
	headers := make([]*types.Header, 0, len(numbers))
	for start := 0; start < len(numbers); start += maxHeaderRange {
		end := min(start+maxHeaderRange, len(numbers))
		from, to := numbers[start].Uint64(), numbers[end-1].Uint64()
		out, err := p.historicalHeaderRange(ctx, from, to)
		if err != nil {
			return nil, err
		}
		if err = p.validateResponse(trustedResponse{APIVersion: out.APIVersion, ProtocolVersion: out.ProtocolVersion, Deployment: out.Deployment}); err != nil {
			return nil, err
		}
		if out.Range.From == nil || out.Range.To == nil || *out.Range.From != from || *out.Range.To != to || len(out.Headers) != end-start {
			return nil, errors.New("invalid trusted header range")
		}
		for i, row := range out.Headers {
			h := (*types.Header)(row.Header)
			if h == nil || h.Number == nil || h.Difficulty == nil || !h.Number.IsUint64() || h.Number.Cmp(numbers[start+i]) != 0 || h.Hash() != row.HeaderHash {
				return nil, errors.New("invalid trusted historical header")
			}
			if len(headers) > 0 && h.ParentHash != headers[len(headers)-1].Hash() {
				return nil, errors.New("discontinuous trusted historical headers")
			}
			headers = append(headers, h)
		}
		if headers[len(headers)-1].Hash() != out.Range.BlockHash {
			return nil, errors.New("trusted historical range anchor mismatch")
		}
	}
	return headers, nil
}
