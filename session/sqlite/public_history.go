package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/session"
	"sync"
)

type PublicHistoryStore struct {
	db        *sql.DB
	mu        sync.Mutex
	retention session.PublicRetention
}

func NewPublicHistoryStore(ctx context.Context, db *sql.DB, retention session.PublicRetention) (*PublicHistoryStore, error) {
	if db == nil {
		return nil, errors.New("database required")
	}
	if _, e := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS invisi_public_history (scope_key TEXT PRIMARY KEY,revision INTEGER NOT NULL,data BLOB NOT NULL)"); e != nil {
		return nil, e
	}
	if _, e := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS invisi_public_history_pages (scope_key TEXT NOT NULL,page_key TEXT NOT NULL,data BLOB NOT NULL,PRIMARY KEY(scope_key,page_key))"); e != nil {
		return nil, e
	}
	return &PublicHistoryStore{db: db, retention: retention.Defaults()}, nil
}
func (s *PublicHistoryStore) Retention() session.PublicRetention { return s.retention }
func (s *PublicHistoryStore) Read(ctx context.Context, key string) (session.PublicHistoryState, error) {
	var data []byte
	var rev uint64
	e := s.db.QueryRowContext(ctx, "SELECT revision,data FROM invisi_public_history WHERE scope_key=?", key).Scan(&rev, &data)
	if errors.Is(e, sql.ErrNoRows) {
		return session.PublicHistoryState{}, nil
	}
	var v session.PublicHistoryState
	if e != nil {
		return v, e
	}
	if e = json.Unmarshal(data, &v); e != nil {
		return v, e
	}
	if rev != v.Revision {
		return v, errors.New("public history revision mismatch")
	}
	return v, nil
}
func (s *PublicHistoryStore) Update(ctx context.Context, key string, fn func(*session.PublicHistoryState) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var data []byte
	var rev uint64
	e = tx.QueryRowContext(ctx, "SELECT revision,data FROM invisi_public_history WHERE scope_key=?", key).Scan(&rev, &data)
	var v session.PublicHistoryState
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return e
	}
	if len(data) > 0 {
		if e = json.Unmarshal(data, &v); e != nil {
			return e
		}
		if v.Revision != rev {
			return errors.New("public history revision mismatch")
		}
	}
	if e = fn(&v); e != nil {
		return e
	}
	v.Revision = rev + 1
	keep := map[string]bool{}
	for i := range v.Pages {
		ref := &v.Pages[i]
		pkey := fmt.Sprintf("%d:%d", ref.Range.FromBlock, ref.Range.ToBlock)
		keep[pkey] = true
		if len(ref.Data) > 0 {
			if _, e = tx.ExecContext(ctx, "INSERT INTO invisi_public_history_pages(scope_key,page_key,data) VALUES(?,?,?) ON CONFLICT(scope_key,page_key) DO UPDATE SET data=excluded.data", key, pkey, []byte(ref.Data)); e != nil {
				return e
			}
		}
		ref.Data = nil
	}
	rows, e := tx.QueryContext(ctx, "SELECT page_key FROM invisi_public_history_pages WHERE scope_key=?", key)
	if e != nil {
		return e
	}
	remove := []string{}
	for rows.Next() {
		var pkey string
		if e = rows.Scan(&pkey); e != nil {
			rows.Close()
			return e
		}
		if !keep[pkey] {
			remove = append(remove, pkey)
		}
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, pkey := range remove {
		if _, e = tx.ExecContext(ctx, "DELETE FROM invisi_public_history_pages WHERE scope_key=? AND page_key=?", key, pkey); e != nil {
			return e
		}
	}
	data, e = json.Marshal(v)
	if e != nil {
		return e
	}
	if rev == 0 {
		_, e = tx.ExecContext(ctx, "INSERT INTO invisi_public_history(scope_key,revision,data) VALUES(?,?,?)", key, v.Revision, data)
	} else {
		var r sql.Result
		r, e = tx.ExecContext(ctx, "UPDATE invisi_public_history SET revision=?,data=? WHERE scope_key=? AND revision=?", v.Revision, data, key, rev)
		if e == nil {
			n, x := r.RowsAffected()
			if x != nil {
				return x
			}
			if n != 1 {
				return errors.New("concurrent public history update")
			}
		}
	}
	if e != nil {
		return e
	}
	return tx.Commit()
}

func (s *PublicHistoryStore) ReadPayload(ctx context.Context, key string, ref session.PublicPageReference) (json.RawMessage, error) {
	var data []byte
	e := s.db.QueryRowContext(ctx, "SELECT data FROM invisi_public_history_pages WHERE scope_key=? AND page_key=?", key, fmt.Sprintf("%d:%d", ref.Range.FromBlock, ref.Range.ToBlock)).Scan(&data)
	if e != nil {
		return nil, e
	}
	if sha256.Sum256(data) != ref.Digest {
		return nil, errors.New("public page digest mismatch")
	}
	return json.RawMessage(data), nil
}
