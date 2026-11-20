package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// HeadersByNumber reads immutable numeric headers in bounded JSON-RPC batches.
// Every response has the same deployment, full-header hash and number checks as
// HeaderByNumber. No result is cached; callers must still fence page commits
// with a fresh canonical anchor. Mutable tags are deliberately not supported.
func (p *TrustedProvider) HeadersByNumber(ctx context.Context, numbers []*big.Int) ([]*types.Header, error) {
	params := make([]any, len(numbers))
	for i, n := range numbers {
		if n == nil || n.Sign() < 0 {
			return nil, errors.New("batch headers require numeric blocks")
		}
		block, err := trustedBlock(n)
		if err != nil {
			return nil, err
		}
		params[i] = []any{trustedRequest{Kind: "header", Block: block}}
	}
	headers := make([]*types.Header, 0, len(numbers))
	for start := 0; start < len(params); start += maxBatchCalls {
		end := min(start+maxBatchCalls, len(params))
		rows, err := p.client.callBatch(ctx, p.scope.ChainID, "invisi_readChainState", params[start:end])
		if err != nil {
			return nil, err
		}
		for i, raw := range rows {
			var out trustedResponse
			if err = json.Unmarshal(raw, &out); err != nil {
				return nil, err
			}
			if err = p.validateResponse(out); err != nil {
				return nil, err
			}
			h := out.Header
			if h == nil {
				return nil, ethereum.NotFound
			}
			if h.Number == nil || !h.Number.IsUint64() || h.Number.Cmp(numbers[start+i]) != 0 || h.Hash() != out.HeaderHash {
				return nil, errors.New("invalid trusted batch header")
			}
			headers = append(headers, h)
		}
	}
	return headers, nil
}
