package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"math/big"
	"testing"
)

func TestPublicRetentionKeepsCoverageAndNewestPages(t *testing.T) {
	s := NewMemoryPublicHistoryStore(PublicRetention{MaxBytes: 1024 * 1024, MaxPages: 1, MaxPageBytes: 1024 * 1024})
	ctx := context.Background()
	for _, n := range []uint64{100, 200, 50} {
		p := PublicPage{Range: BlockRange{n, n}, Block: protocolrpc.CanonicalBlock{Number: n, Hash: common.BigToHash(new(big.Int).SetUint64(n))}}
		if err := PutPublicPage(ctx, s, "deployment", p, 0); err != nil {
			t.Fatal(err)
		}
	}
	state, e := s.Read(ctx, "deployment")
	if e != nil {
		t.Fatal(e)
	}
	if len(state.Pages) != 1 || state.Pages[0].Range.FromBlock != 200 {
		t.Fatalf("retained %+v", state.Pages)
	}
	if len(state.Coverage) != 3 {
		t.Fatalf("coverage lost %+v", state.Coverage)
	}
}
func TestPublicPageGenerationFence(t *testing.T) {
	s := NewMemoryPublicHistoryStore(PublicRetention{})
	ctx := context.Background()
	if e := s.Update(ctx, "x", func(v *PublicHistoryState) error { v.Generation++; return nil }); e != nil {
		t.Fatal(e)
	}
	if e := PutPublicPage(ctx, s, "x", PublicPage{Range: BlockRange{1, 1}, Block: protocolrpc.CanonicalBlock{Number: 1, Hash: common.HexToHash("0x1")}}, 0); e == nil {
		t.Fatal("stale generation published")
	}
}

func TestSharedDownloadCancellationDoesNotCancelOtherWaiter(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	s, calls := refreshFixture(t, false, entered, release)
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	second := make(chan error, 1)
	go func() { _, e := s.publicHistory.Page(ctx, BlockRange{1, 10}, false); first <- e }()
	<-entered
	go func() { _, e := s.publicHistory.Page(context.Background(), BlockRange{1, 10}, false); second <- e }()
	cancel()
	if e := <-first; !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	close(release)
	if e := <-second; e != nil {
		t.Fatal(e)
	}
	if calls.Load() != 1 {
		t.Fatal("concurrent waiter redownloaded")
	}
}
