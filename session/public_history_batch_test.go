package session

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
)

type batchedHistoryChain struct {
	historyChain
	batches, fences int
	mode            string
}

func (p *batchedHistoryChain) HeadersByNumber(ctx context.Context, numbers []*big.Int) ([]*types.Header, error) {
	p.batches++
	if p.mode == "error" {
		return nil, errors.New("batch unavailable")
	}
	out := make([]*types.Header, len(numbers))
	for i, n := range numbers {
		out[i] = types.CopyHeader(p.headers[n.Uint64()])
	}
	if p.mode == "short" {
		out = out[:len(out)-1]
	}
	if p.mode == "broken" {
		out[1].ParentHash[0] ^= 1
	}
	if p.mode == "wrong" {
		out[1] = out[0]
	}
	return out, nil
}
func (p *batchedHistoryChain) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	p.fences++
	if p.mode == "reorg" && p.batches > 0 && n != nil && n.Uint64() == 10 {
		h := types.CopyHeader(p.headers[10])
		h.Extra = []byte{42}
		return h, nil
	}
	return p.historyChain.HeaderByNumber(ctx, n)
}
func TestPublicHistoryBatchHeadersKeepValidationAndFreshFence(t *testing.T) {
	for _, mode := range []string{"ok", "short", "broken", "wrong", "error", "reorg"} {
		t.Run(mode, func(t *testing.T) {
			s, _ := sharedEmptyFixture(t)
			p := &batchedHistoryChain{historyChain: s.options.Provider.(historyChain), mode: mode}
			h := s.publicHistory
			h.options.Provider = p
			page, err := h.Page(context.Background(), BlockRange{1, 10}, false)
			if p.batches != 1 {
				t.Fatalf("batch header capability unused: %d", p.batches)
			}
			if mode == "ok" {
				if err != nil || len(page.Headers) != 10 || p.fences == 0 {
					t.Fatalf("missing page or fresh canonical fence: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("invalid batch committed")
			}
			state, _ := h.Snapshot(context.Background())
			if len(state.Pages) != 0 {
				t.Fatal("invalid page retained")
			}
		})
	}
}
