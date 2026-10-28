package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestSQLiteEncryptedStateReopensAndRollsBack(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "wallet.sqlite")
	codec, e := sdk.NewEncryptedWalletCodec(make([]byte, 32))
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
	scope := sdk.WalletScope{ChainID: 1}
	hash := common.HexToHash("0x01")
	if e = store.Update(ctx, scope, func(state *sdk.WalletState) error { state.Notes[hash] = sdk.StoredNote{Commitment: hash}; return nil }); e != nil {
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
	if e = store.Update(ctx, scope, func(state *sdk.WalletState) error { delete(state.Notes, hash); return errors.New("rollback") }); e == nil {
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
