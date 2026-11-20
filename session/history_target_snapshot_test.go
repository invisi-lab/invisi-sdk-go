package session

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
)

type laggingScanChain struct{ historyChain }

func (p laggingScanChain) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	if n == nil {
		return p.headers[5], nil
	}
	return p.historyChain.HeaderByNumber(ctx, n)
}
func TestScanObservationCannotPrecedeIndexedTarget(t *testing.T) {
	s, _ := sharedEmptyFixture(t)
	s.options.Provider = laggingScanChain{s.options.Provider.(historyChain)}
	if _, _, err := s.historyTarget(context.Background(), ScanOptions{}); err == nil {
		t.Fatal("accepted state observation before discovered assets")
	}
	public, err := s.publicHistory.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if public.KnownHead != nil || public.IndexedHead != nil {
		t.Fatal("invalid observation was persisted")
	}
}
