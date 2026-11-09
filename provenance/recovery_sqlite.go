package provenance

import (
	"context"
	"database/sql"
	"errors"
)

// SQLiteRecoveryRecords reuses a caller-owned database while keeping provenance ciphertext in its own table.
type SQLiteRecoveryRecords struct{ db *sql.DB }

func NewSQLiteRecoveryRecords(ctx context.Context, db *sql.DB) (*SQLiteRecoveryRecords, error) {
	if db == nil {
		return nil, errors.New("recovery database required")
	}
	if _, e := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS invisi_provenance_recovery (record_key TEXT PRIMARY KEY, revision INTEGER NOT NULL CHECK(revision > 0), data BLOB NOT NULL)"); e != nil {
		return nil, e
	}
	return &SQLiteRecoveryRecords{db}, nil
}
func (s *SQLiteRecoveryRecords) Read(ctx context.Context, key string) (RecoveryCiphertext, bool, error) {
	var out RecoveryCiphertext
	var revision int64
	e := s.db.QueryRowContext(ctx, "SELECT revision,data FROM invisi_provenance_recovery WHERE record_key=?", key).Scan(&revision, &out.Data)
	if errors.Is(e, sql.ErrNoRows) {
		return out, false, nil
	}
	if e != nil {
		return out, false, e
	}
	if revision < 1 || revision > 9007199254740991 {
		return out, false, errors.New("invalid recovery revision")
	}
	out.Revision = uint64(revision)
	return out, true, nil
}
func (s *SQLiteRecoveryRecords) CompareAndSwap(ctx context.Context, key string, old uint64, next RecoveryCiphertext) (bool, error) {
	if old >= 9007199254740991 || next.Revision != old+1 {
		return false, errors.New("invalid recovery revision")
	}
	var result sql.Result
	var e error
	if old == 0 {
		result, e = s.db.ExecContext(ctx, "INSERT INTO invisi_provenance_recovery(record_key,revision,data) VALUES(?,?,?) ON CONFLICT(record_key) DO NOTHING", key, next.Revision, next.Data)
	} else {
		result, e = s.db.ExecContext(ctx, "UPDATE invisi_provenance_recovery SET revision=?,data=? WHERE record_key=? AND revision=?", next.Revision, next.Data, key, old)
	}
	if e != nil {
		return false, e
	}
	n, e := result.RowsAffected()
	return n == 1, e
}
