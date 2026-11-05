package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/session"
)

func TestSQLiteEncryptedStateReopensAndRollsBack(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "wallet.sqlite")
	codec, e := session.NewEncryptedWalletCodec(make([]byte, 32))
	if e != nil {
		t.Fatal(e)
	}
	db, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	store, e := NewWalletStore(ctx, db, codec)
	if e != nil {
		t.Fatal(e)
	}
	scope := session.WalletScope{ChainID: 1}
	hash := common.HexToHash("0x01")
	if e = store.Update(ctx, scope, func(state *session.WalletState) error {
		state.Notes[hash] = session.StoredNote{Commitment: hash}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	db.Close()
	db, e = sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	store, e = NewWalletStore(ctx, db, codec)
	if e != nil {
		t.Fatal(e)
	}
	if e = store.Update(ctx, scope, func(state *session.WalletState) error { delete(state.Notes, hash); return errors.New("rollback") }); e == nil {
		t.Fatal("expected rollback")
	}
	state, e := store.ReadSnapshot(ctx, scope)
	if e != nil || len(state.Notes) != 1 {
		t.Fatal(state, e)
	}
	var raw []byte
	if e = db.QueryRow("SELECT data FROM invisi_wallet_state WHERE scope_key=?", scope.Key()).Scan(&raw); e != nil {
		t.Fatal(e)
	}
	if len(raw) == 0 || raw[0] == '{' {
		t.Fatal("wallet stored plaintext JSON")
	}
}

func TestSQLiteRevisionMatchesAuthenticatedStateAndRejectsMismatch(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "revision.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	codec, err := session.NewEncryptedWalletCodec(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewWalletStore(ctx, db, codec)
	if err != nil {
		t.Fatal(err)
	}
	scope := session.WalletScope{ChainID: 97}
	for expected := uint64(1); expected <= 2; expected++ {
		if err = store.Update(ctx, scope, func(state *session.WalletState) error { state.Revision = 999; return nil }); err != nil {
			t.Fatal(err)
		}
		state, e := store.ReadSnapshot(ctx, scope)
		if e != nil || state.Revision != expected {
			t.Fatalf("revision mismatch: %d, %v", state.Revision, e)
		}
	}
	if _, err = db.ExecContext(ctx, "UPDATE invisi_wallet_state SET revision=3 WHERE scope_key=?", scope.Key()); err != nil {
		t.Fatal(err)
	}
	if _, err = store.ReadSnapshot(ctx, scope); err == nil {
		t.Fatal("unauthenticated revision mismatch accepted")
	}
	if err = store.Update(ctx, scope, func(state *session.WalletState) error { return nil }); err == nil {
		t.Fatal("updated from mismatched revision")
	}
}
