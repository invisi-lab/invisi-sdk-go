// Package sqlite stores only the user's authenticated encrypted wallet state.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/invisi-lab/invisi-sdk-go/session"
)

type WalletStore struct {
	db            *sql.DB
	codec         *session.EncryptedWalletCodec
	mu            sync.Mutex
	recentRecords int
}

// NewWalletStore accepts a caller-owned database. The caller closes its connection.
type WalletStoreOptions struct{ RecentRecords int }

func NewWalletStore(ctx context.Context, db *sql.DB, codec *session.EncryptedWalletCodec, options ...WalletStoreOptions) (*WalletStore, error) {
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
	if _, e := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS invisi_wallet_records (scope_key TEXT NOT NULL,locator TEXT NOT NULL,kind TEXT NOT NULL,sequence INTEGER NOT NULL,archived INTEGER NOT NULL,data BLOB NOT NULL,PRIMARY KEY(scope_key,locator))"); e != nil {
		return nil, e
	}
	if _, e := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS invisi_wallet_records_page ON invisi_wallet_records(scope_key,kind,sequence DESC)"); e != nil {
		return nil, e
	}
	recent := 128
	if len(options) > 0 {
		recent = options[0].RecentRecords
	}
	if recent < 0 {
		return nil, errors.New("recentRecords must be nonnegative")
	}
	return &WalletStore{db: db, codec: codec, recentRecords: recent}, nil
}
func (s *WalletStore) ReadSnapshot(ctx context.Context, scope session.WalletScope) (session.WalletState, error) {
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if e != nil {
		return session.WalletState{}, e
	}
	defer tx.Rollback()
	var data []byte
	var revision int64
	e = tx.QueryRowContext(ctx, "SELECT revision,data FROM invisi_wallet_state WHERE scope_key=?", scope.Key()).Scan(&revision, &data)
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
	if err = s.loadRecords(ctx, tx, scope, &state); err != nil {
		return session.WalletState{}, err
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
	if e = s.loadRecords(ctx, tx, scope, &state); e != nil {
		return e
	}
	previous, e := session.CloneState(state)
	if e != nil {
		return e
	}
	if e = update(&state); e != nil {
		return e
	}
	state.Revision = uint64(revision) + 1
	if e := ctx.Err(); e != nil {
		return e
	}
	if e = s.writeRecords(ctx, tx, scope, &state, previous); e != nil {
		return e
	}
	metadata, e := session.CloneState(state)
	if e != nil {
		return e
	}
	session.ClearWalletRecords(&metadata)
	data, e = s.codec.Seal(scope, metadata)
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
