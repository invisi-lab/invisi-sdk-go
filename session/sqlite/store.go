// Package sqlite stores only the user's authenticated encrypted wallet state.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/invisi-lab/invisi-sdk-go/session"
	_ "modernc.org/sqlite"
)

type WalletStore struct {
	db    *sql.DB
	codec *session.EncryptedWalletCodec
	mu    sync.Mutex
}

// NewWalletStore accepts a caller-owned database. The caller closes its connection.
func NewWalletStore(ctx context.Context, db *sql.DB, codec *session.EncryptedWalletCodec) (*WalletStore, error) {
	if db == nil || codec == nil {
		return nil, errors.New("database and encryption codec required")
	}
	db.SetMaxOpenConns(1)
	if _, e := db.ExecContext(ctx, "PRAGMA busy_timeout=5000"); e != nil {
		return nil, e
	}
	if _, e := db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); e != nil {
		return nil, e
	}
	if _, e := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS invisi_wallet_state (scope_key TEXT PRIMARY KEY, revision INTEGER NOT NULL, data BLOB NOT NULL)"); e != nil {
		return nil, e
	}
	return &WalletStore{db: db, codec: codec}, nil
}
func (s *WalletStore) ReadSnapshot(ctx context.Context, scope session.WalletScope) (session.WalletState, error) {
	var data []byte
	var revision int64
	e := s.db.QueryRowContext(ctx, "SELECT revision,data FROM invisi_wallet_state WHERE scope_key=?", scope.Key()).Scan(&revision, &data)
	if errors.Is(e, sql.ErrNoRows) {
		return session.NewMemoryWalletStore().ReadSnapshot(ctx, scope)
	}
	if e != nil {
		return session.WalletState{}, e
	}
	state, err := s.codec.Open(scope, data)
	if err != nil {
		return session.WalletState{}, err
	}
	if revision < 1 || state.Revision != uint64(revision) {
		return session.WalletState{}, errors.New("stored wallet revision mismatch")
	}
	return state, nil
}
func (s *WalletStore) Update(ctx context.Context, scope session.WalletScope, update func(*session.WalletState) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var data []byte
	var revision int64
	e = tx.QueryRowContext(ctx, "SELECT revision,data FROM invisi_wallet_state WHERE scope_key=?", scope.Key()).Scan(&revision, &data)
	var state session.WalletState
	if errors.Is(e, sql.ErrNoRows) {
		state, e = session.NewMemoryWalletStore().ReadSnapshot(ctx, scope)
	} else if e == nil {
		state, e = s.codec.Open(scope, data)
	}
	if e != nil {
		return e
	}
	if revision < 0 || state.Revision != uint64(revision) {
		return errors.New("stored wallet revision mismatch")
	}
	if revision == 1<<63-1 {
		return errors.New("wallet revision exhausted")
	}
	if e = update(&state); e != nil {
		return e
	}
	state.Revision = uint64(revision) + 1
	if e := ctx.Err(); e != nil {
		return e
	}
	data, e = s.codec.Seal(scope, state)
	if e != nil {
		return e
	}
	if revision == 0 {
		_, e = tx.ExecContext(ctx, "INSERT INTO invisi_wallet_state(scope_key,revision,data) VALUES(?,1,?)", scope.Key(), data)
	} else {
		var result sql.Result
		result, e = tx.ExecContext(ctx, "UPDATE invisi_wallet_state SET revision=revision+1,data=? WHERE scope_key=? AND revision=?", data, scope.Key(), revision)
		if e == nil {
			changed, x := result.RowsAffected()
			if x != nil {
				return x
			}
			if changed != 1 {
				return errors.New("concurrent wallet update")
			}
		}
	}
	if e != nil {
		return e
	}
	return tx.Commit()
}
