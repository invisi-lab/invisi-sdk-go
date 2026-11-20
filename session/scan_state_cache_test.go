package session

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
	creditsession "github.com/invisi-labs/invisi-sdk-go/session/credit"
)

type countingStateProvider struct {
	sdk.Provider
	calls    int
	failOnce bool
}

func (p *countingStateProvider) CallContract(ctx context.Context, m ethereum.CallMsg, n *big.Int) ([]byte, error) {
	p.calls++
	if p.failOnce {
		p.failOnce = false
		return nil, errors.New("temporary provider failure")
	}
	return p.Provider.CallContract(ctx, m, n)
}

func TestScanStateChecksReuseExactAnchorButNextScanReadsFresh(t *testing.T) {
	s, _, _ := testSession(t)
	ctx := context.Background()
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	head := s.options.Provider.(sessionProvider).head
	block := protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	provider := &countingStateProvider{Provider: s.options.Provider}
	s.options.Provider = provider
	scan := withScanStateCache(ctx)
	for range 3 {
		if err = s.checkKnownState(scan, &state, block); err != nil {
			t.Fatal(err)
		}
	}
	if provider.calls != 1 {
		t.Fatalf("repeated same-anchor consumption calls %d", provider.calls)
	}
	if err = s.checkKnownState(withScanStateCache(ctx), &state, block); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 2 {
		t.Fatal("next scan reused cached consumption")
	}
	changed := block
	changed.Hash[0] ^= 1
	if err = s.checkKnownState(scan, &state, changed); err == nil {
		t.Fatal("changed canonical anchor accepted")
	}
	if provider.calls != 3 {
		t.Fatal("different hash reused same-height state")
	}
}

func TestScanStateCacheKeepsRevocationChecksAndSeparatesBlocks(t *testing.T) {
	s, _, _ := testSession(t)
	ctx := context.Background()
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	grant := syncGrantState()
	state.PostpaidCredit.Grants["grant"] = creditsession.StoredState{Nullifier: common.HexToHash("0x42"), Role: credit.Owner, Delivery: credit.DeliveryState{State: grant}}
	head := s.options.Provider.(sessionProvider).head
	block := protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	provider := &countingStateProvider{Provider: s.options.Provider}
	s.options.Provider = provider
	scan := withScanStateCache(ctx)
	for range 3 {
		if err = s.checkKnownState(scan, &state, block); err != nil {
			t.Fatal(err)
		}
	}
	if provider.calls != 3 {
		t.Fatalf("expected Note+Credit consumption and revocation bitmap, got %d", provider.calls)
	}
	if state.PostpaidCredit.Grants["grant"].Revoked || state.PostpaidCredit.Grants["grant"].Consumed {
		t.Fatal("cached check altered grant state")
	}
	if err = s.checkKnownState(withScanStateCache(ctx), &state, block); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 6 {
		t.Fatal("later scan did not repeat consumption and revocation checks")
	}
}

func TestScanStateCacheDoesNotRetainErrors(t *testing.T) {
	s, _, _ := testSession(t)
	ctx := context.Background()
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	head := s.options.Provider.(sessionProvider).head
	block := protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	provider := &countingStateProvider{Provider: s.options.Provider, failOnce: true}
	s.options.Provider = provider
	scan := withScanStateCache(ctx)
	if err = s.checkKnownState(scan, &state, block); err == nil {
		t.Fatal("provider failure disappeared")
	}
	if err = s.checkKnownState(scan, &state, block); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 2 {
		t.Fatal("provider failure cached")
	}
}
