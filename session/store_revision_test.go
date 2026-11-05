package session

import (
	"context"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func TestStoreRevisionAndDiscoveryAreAtomicCopies(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryWalletStore()
	scope := WalletScope{ChainID: 97}
	if err := store.Update(ctx, scope, func(state *WalletState) error {
		state.Revision = 999
		state.DiscoveryThrough = &protocolrpc.CanonicalBlock{Number: 12, Hash: common.HexToHash("0x12")}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	state, err := store.ReadSnapshot(ctx, scope)
	if err != nil || state.Revision != 1 || state.DiscoveryThrough.Number != 12 {
		t.Fatalf("first revision: %+v, %v", state, err)
	}
	state.DiscoveryThrough.Number = 123
	if err := store.Update(ctx, scope, func(state *WalletState) error {
		state.DiscoveryThrough.Number = 999
		return errors.New("rollback")
	}); err == nil {
		t.Fatal("expected rollback")
	}
	again, err := store.ReadSnapshot(ctx, scope)
	if err != nil || again.Revision != 1 || again.DiscoveryThrough.Number != 12 {
		t.Fatalf("rollback changed state: %+v, %v", again, err)
	}
	if err := store.Update(ctx, scope, func(state *WalletState) error { return nil }); err != nil {
		t.Fatal(err)
	}
	again, _ = store.ReadSnapshot(ctx, scope)
	if again.Revision != 2 {
		t.Fatalf("revision = %d", again.Revision)
	}
}
