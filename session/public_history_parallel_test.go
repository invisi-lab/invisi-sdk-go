package session

import (
	"context"
	"errors"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

type parallelHistoryChain struct {
	historyChain
	active, peak   atomic.Int32
	broken, failed uint64
}

func (p *parallelHistoryChain) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	active := p.active.Add(1)
	defer p.active.Add(-1)
	for old := p.peak.Load(); active > old; old = p.peak.Load() {
		if p.peak.CompareAndSwap(old, active) {
			break
		}
	}
	time.Sleep(5 * time.Millisecond)
	if n != nil && n.Uint64() == p.failed {
		return nil, errors.New("header unavailable")
	}
	header, err := p.historyChain.HeaderByNumber(ctx, n)
	if err == nil && n != nil && n.Uint64() == p.broken {
		header = types.CopyHeader(header)
		header.ParentHash[0] ^= 1
	}
	return header, err
}

func TestPublicHistoryParallelHeaders(t *testing.T) {
	for _, mode := range []string{"ordered", "broken", "failed"} {
		t.Run(mode, func(t *testing.T) {
			s, _ := sharedEmptyFixture(t)
			provider := &parallelHistoryChain{historyChain: s.options.Provider.(historyChain)}
			if mode == "broken" {
				provider.broken = 9
			}
			if mode == "failed" {
				provider.failed = 9
			}
			history := s.publicHistory
			history.options.Provider = provider
			page, err := history.Page(context.Background(), BlockRange{FromBlock: 1, ToBlock: 10}, false)
			if mode == "ordered" {
				if err != nil {
					t.Fatal(err)
				}
				if peak := provider.peak.Load(); peak <= 1 || peak > 8 {
					t.Fatalf("parallel header reads: %d, want 2..8", peak)
				}
				if len(page.Headers) != 10 {
					t.Fatalf("headers: %d", len(page.Headers))
				}
				for i, header := range page.Headers {
					if header.Number != uint64(i+1) {
						t.Fatalf("header %d: %d", i, header.Number)
					}
				}
			} else {
				if err == nil {
					t.Fatal("invalid page accepted")
				}
				state, err := history.Snapshot(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if len(state.Pages) != 0 || len(state.Coverage) != 0 {
					t.Fatal("failed page was committed")
				}
			}
			if provider.active.Load() != 0 {
				t.Fatal("header readers still running")
			}
		})
	}
}
