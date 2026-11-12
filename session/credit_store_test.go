package session

import (
	"context"
	"errors"
	"testing"

	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

func TestCreditAdapterUsesAtomicEncryptedWalletStore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryWalletStore()
	wallet := &WalletSession{scope: WalletScope{ChainID: 1}, options: sessionOptions{Store: store}}
	adapter := creditStoreAdapter{session: wallet}
	before, err := adapter.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err = adapter.Update(ctx, func(state *creditsession.State) error {
		state.Operations["revoke"] = creditsession.Journal{ID: "revoke", Status: "prepared"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	after, err := adapter.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if after.Revision != before.Revision+1 || after.Operations["revoke"].ID != "revoke" {
		t.Fatal("Credit journal was not atomically persisted")
	}
	sentinel := errors.New("abort")
	if err = adapter.Update(ctx, func(state *creditsession.State) error { delete(state.Operations, "revoke"); return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	snapshot, err := adapter.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Revision != after.Revision || snapshot.Operations["revoke"].ID != "revoke" {
		t.Fatal("failed update changed Credit journal")
	}
}

func TestCreditReservationRejectsPublicGenerationChangeDuringCommit(t *testing.T) {
	ctx := context.Background()
	wallet, _, _ := testSession(t)
	adapter := creditStoreAdapter{session: wallet}
	err := adapter.Update(ctx, func(state *creditsession.State) error {
		state.Operations["stale"] = creditsession.Journal{ID: "stale", Status: "prepared"}
		return wallet.publicHistory.options.Store.Update(ctx, wallet.publicHistory.key, func(public *PublicHistoryState) error { public.Generation++; return nil })
	})
	if err == nil {
		t.Fatal("reservation succeeded across public history reset")
	}
	state, e := wallet.options.Store.ReadSnapshot(ctx, wallet.scope)
	if e != nil {
		t.Fatal(e)
	}
	if !state.Invalidated {
		t.Fatal("reservation was usable after history generation changed")
	}
}
