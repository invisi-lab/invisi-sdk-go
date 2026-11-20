package session

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
	creditsession "github.com/invisi-labs/invisi-sdk-go/session/credit"
)

type delayedStateCheckProvider struct {
	sdk.Provider
	calls int
}

func (p *delayedStateCheckProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, n *big.Int) ([]byte, error) {
	p.calls++
	time.Sleep(100 * time.Microsecond)
	return p.Provider.CallContract(ctx, msg, n)
}

func BenchmarkWalletFixedAnchorStateChecks(b *testing.B) {
	s, _, _ := testSession(b)
	ctx := context.Background()
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		b.Fatal(err)
	}
	grant := syncGrantState()
	state.PostpaidCredit.Grants["grant"] = creditsession.StoredState{Nullifier: common.HexToHash("0x42"), Role: credit.Owner, Delivery: credit.DeliveryState{State: grant}}
	head := s.options.Provider.(sessionProvider).head
	block := protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	provider := &delayedStateCheckProvider{Provider: s.options.Provider}
	s.options.Provider = provider
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scan := withScanStateCache(ctx)
		for range 32 {
			if err = s.checkKnownState(scan, &state, block); err != nil {
				b.Fatal(err)
			}
		}
	}
	b.ReportMetric(float64(provider.calls)/float64(b.N), "contract-reads/op")
	b.ReportAllocs()
}
