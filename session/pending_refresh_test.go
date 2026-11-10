package session

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestPendingRefreshDoesNotAdvanceUntilApplied(t *testing.T) {
	s, _ := refreshFixture(t, false, nil, nil)
	ctx := context.Background()
	before, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	pending, err := s.Refresh(ctx, RefreshOptions{FromBlock: 1})
	if err != nil {
		t.Fatal(err)
	}
	after, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if after.Revision != before.Revision || after.DiscoveryThrough != before.DiscoveryThrough {
		t.Fatal("preparation changed real inventory")
	}
	if _, err = pending.Decrypt(ctx, testDeliveryDecryptor(s)); err != nil {
		t.Fatal(err)
	}
	if _, err = pending.Decrypt(ctx, testDeliveryDecryptor(s)); err == nil {
		t.Fatal("pending refresh applied twice")
	}
}

func TestPendingRefreshRejectsChangedInventory(t *testing.T) {
	s, _ := refreshFixture(t, false, nil, nil)
	ctx := context.Background()
	pending, err := s.Refresh(ctx, RefreshOptions{FromBlock: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err = pending.Decrypt(ctx, testDeliveryDecryptor(s)); err == nil {
		t.Fatal("pending refresh overwrote newer inventory")
	}
}

func TestPendingRefreshFormattingDoesNotExposeStoredPlaintext(t *testing.T) {
	pending := &PendingWalletRefresh{base: WalletState{Operations: map[string]OperationRecord{"private-operation": {Plan: []byte(`{"private":"secret"}`)}}}}
	if strings.Contains(fmt.Sprintf("%+v %#v", pending, pending), "private-operation") {
		t.Fatal("pending handle formatting leaked stored operation")
	}
}

func TestPendingRefreshRejectsConcurrentApply(t *testing.T) {
	s, _ := refreshFixture(t, false, nil, nil)
	ctx := context.Background()
	pending, err := s.Refresh(ctx, RefreshOptions{FromBlock: 1})
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	provider := &pauseSpentProvider{Provider: s.options.Provider, entered: entered, release: release}
	s.options.Provider = provider
	s.rpc = s.rpc.WithProvider(provider)
	done := make(chan error, 1)
	go func() { _, e := pending.Decrypt(ctx, testDeliveryDecryptor(s)); done <- e }()
	<-entered
	if _, err = pending.Decrypt(ctx, testDeliveryDecryptor(s)); err == nil {
		close(release)
		t.Fatal("concurrent pending apply accepted")
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
}
