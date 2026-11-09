package session

import (
	"context"
	"errors"
	creditsession "github.com/invisi-lab/invisi-sdk-go/credit/session"
	"testing"
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
