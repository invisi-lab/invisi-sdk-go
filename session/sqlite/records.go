package sqlite

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/invisi-lab/invisi-sdk-go/session"
	"reflect"
	"sort"
)

type rowReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (s *WalletStore) loadRecords(ctx context.Context, reader rowReader, scope session.WalletScope, v *session.WalletState) error {
	if v.Records.Version == 0 {
		return nil
	}
	if e := session.PrepareRecordMetadata(v); e != nil {
		return e
	}
	rows, e := reader.QueryContext(ctx, "SELECT locator,kind,sequence,data FROM invisi_wallet_records WHERE scope_key=? AND archived=0", scope.Key())
	if e != nil {
		return e
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var locator, kind string
		var seq uint64
		var data []byte
		if e = rows.Scan(&locator, &kind, &seq, &data); e != nil {
			return e
		}
		m, ok := v.Records.Active[locator]
		if !ok || m.Kind != kind || m.Sequence != seq || m.Digest != sha256.Sum256(data) {
			return errors.New("wallet active record inventory authentication failed")
		}
		raw, e := s.codec.OpenRecord(scope, locator, data)
		if e != nil {
			return e
		}
		var r session.WalletRecord
		e = json.Unmarshal(raw, &r)
		clear(raw)
		if e != nil {
			return e
		}
		if r.Kind != kind || r.Sequence != seq || session.RecordLocator(v.Records.IndexSecret, r.Kind, r.Key) != locator {
			return errors.New("wallet record locator mismatch")
		}
		if e = session.ApplyWalletRecord(v, r); e != nil {
			return e
		}
		seen[locator] = true
	}
	if e = rows.Err(); e != nil {
		return e
	}
	if len(seen) != len(v.Records.Active) {
		return errors.New("wallet active record missing")
	}
	return nil
}
func (s *WalletStore) writeRecords(ctx context.Context, tx *sql.Tx, scope session.WalletScope, v *session.WalletState, previous session.WalletState) error {
	if previous.Records.Version != 0 && !reflect.DeepEqual(v.Records, previous.Records) {
		return errors.New("wallet storage metadata must be preserved")
	}
	if e := session.PrepareRecordMetadata(v); e != nil {
		return e
	}
	records, e := session.WalletRecords(*v)
	if e != nil {
		return e
	}
	oldRecords, e := session.WalletRecords(previous)
	if e != nil {
		return e
	}
	old := map[string]session.WalletRecord{}
	for _, r := range oldRecords {
		old[session.RecordLocator(v.Records.IndexSecret, r.Kind, r.Key)] = r
	}
	prior := previous.Records.Active
	if prior == nil {
		prior = map[string]session.RecordManifestEntry{}
	}
	for i := range records {
		r := &records[i]
		locator := session.RecordLocator(v.Records.IndexSecret, r.Kind, r.Key)
		if m, ok := prior[locator]; ok {
			r.Sequence = m.Sequence
		} else {
			var sequence uint64
			err := tx.QueryRowContext(ctx, "SELECT sequence FROM invisi_wallet_records WHERE scope_key=? AND locator=?", scope.Key(), locator).Scan(&sequence)
			if err == nil {
				r.Sequence = sequence
			} else if errors.Is(err, sql.ErrNoRows) {
				v.Records.NextSequence++
				r.Sequence = v.Records.NextSequence
				v.Records.Counts[r.Kind]++
			} else {
				return err
			}
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Sequence > records[j].Sequence })
	recent := map[string]int{}
	manifest := map[string]session.RecordManifestEntry{}
	active := *v
	session.ClearWalletRecords(&active)
	for _, r := range records {
		locator := session.RecordLocator(v.Records.IndexSecret, r.Kind, r.Key)
		terminal := session.IsTerminalRecord(r, v.Scan.CreditReplayTarget != 0)
		if terminal {
			recent[r.Kind]++
			r.Archived = recent[r.Kind] > s.recentRecords
		}
		raw, e := json.Marshal(r)
		if e != nil {
			return e
		}
		m, existing := prior[locator]
		unchanged := false
		if old, ok := old[locator]; ok {
			old.Sequence = r.Sequence
			old.Archived = r.Archived
			old.Generation = r.Generation
			b, _ := json.Marshal(old)
			unchanged = bytes.Equal(b, raw) && existing && !r.Archived
		}
		digest := m.Digest
		if !unchanged {
			data, e := s.codec.SealRecord(scope, locator, raw)
			clear(raw)
			if e != nil {
				return e
			}
			digest = sha256.Sum256(data)
			_, e = tx.ExecContext(ctx, "INSERT INTO invisi_wallet_records(scope_key,locator,kind,sequence,archived,data) VALUES(?,?,?,?,?,?) ON CONFLICT(scope_key,locator) DO UPDATE SET archived=excluded.archived,data=excluded.data", scope.Key(), locator, r.Kind, r.Sequence, r.Archived, data)
			if e != nil {
				return e
			}
		}
		if !r.Archived {
			manifest[locator] = session.RecordManifestEntry{Kind: r.Kind, Sequence: r.Sequence, Digest: digest}
			if e = session.ApplyWalletRecord(&active, r); e != nil {
				return e
			}
		}
	}
	present := map[string]bool{}
	for _, r := range records {
		present[session.RecordLocator(v.Records.IndexSecret, r.Kind, r.Key)] = true
	}
	for locator, m := range prior {
		if !present[locator] {
			if _, e = tx.ExecContext(ctx, "DELETE FROM invisi_wallet_records WHERE scope_key=? AND locator=?", scope.Key(), locator); e != nil {
				return e
			}
			v.Records.Counts[m.Kind]--
		}
	}
	active.Records = v.Records
	active.Records.Active = manifest
	*v = active
	return nil
}
func (s *WalletStore) ReadMetadata(ctx context.Context, scope session.WalletScope) (session.WalletState, error) {
	return s.readMetadata(ctx, s.db, scope)
}

func (s *WalletStore) readMetadata(ctx context.Context, reader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, scope session.WalletScope) (session.WalletState, error) {
	var data []byte
	var revision uint64
	e := reader.QueryRowContext(ctx, "SELECT revision,data FROM invisi_wallet_state WHERE scope_key=?", scope.Key()).Scan(&revision, &data)
	if errors.Is(e, sql.ErrNoRows) {
		return session.NewMemoryWalletStore().ReadSnapshot(ctx, scope)
	}
	if e != nil {
		return session.WalletState{}, e
	}
	v, e := s.codec.Open(scope, data)
	if e != nil {
		return v, e
	}
	if v.Revision != revision {
		return v, errors.New("stored wallet revision mismatch")
	}
	return v, nil
}

type recordCursor struct {
	Revision, Before uint64
	Kind             string
}

func (s *WalletStore) ReadRecords(ctx context.Context, scope session.WalletScope, q session.RecordQuery) (session.RecordPage, error) {
	if q.Limit == 0 {
		q.Limit = 50
	}
	if q.Limit < 1 || q.Limit > 256 {
		return session.RecordPage{}, errors.New("record page limit must be 1..256")
	}
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if e != nil {
		return session.RecordPage{}, e
	}
	defer tx.Rollback()
	v, e := s.readMetadata(ctx, tx, scope)
	if e != nil {
		return session.RecordPage{}, e
	}
	if v.Records.Version != 1 {
		return session.RecordPage{}, errors.New("wallet must be migrated by a successful write before paging")
	}
	cursor := recordCursor{Revision: v.Revision, Before: v.Records.NextSequence + 1, Kind: q.Kind}
	if q.Cursor != "" {
		data, e := base64.RawURLEncoding.DecodeString(q.Cursor)
		if e != nil {
			return session.RecordPage{}, e
		}
		raw, e := s.codec.OpenRecord(scope, "cursor", data)
		if e != nil {
			return session.RecordPage{}, e
		}
		e = json.Unmarshal(raw, &cursor)
		clear(raw)
		if e != nil {
			return session.RecordPage{}, e
		}
		if cursor.Revision != v.Revision || cursor.Kind != q.Kind {
			return session.RecordPage{}, errors.New("history cursor scope or revision changed")
		}
	}
	var count uint64
	if e = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM invisi_wallet_records WHERE scope_key=? AND kind=?", scope.Key(), q.Kind).Scan(&count); e != nil {
		return session.RecordPage{}, e
	}
	if count != v.Records.Counts[q.Kind] {
		return session.RecordPage{}, errors.New("wallet record count mismatch")
	}
	rows, e := tx.QueryContext(ctx, "SELECT locator,sequence,archived,data FROM invisi_wallet_records WHERE scope_key=? AND kind=? AND sequence<? ORDER BY sequence DESC LIMIT ?", scope.Key(), q.Kind, cursor.Before, q.Limit+1)
	if e != nil {
		return session.RecordPage{}, e
	}
	defer rows.Close()
	out := session.RecordPage{Revision: v.Revision, Items: []session.WalletRecord{}}
	for rows.Next() {
		var locator string
		var sequence uint64
		var archived bool
		var data []byte
		if e = rows.Scan(&locator, &sequence, &archived, &data); e != nil {
			return out, e
		}
		if len(out.Items) == q.Limit {
			last := out.Items[len(out.Items)-1]
			raw, _ := json.Marshal(recordCursor{v.Revision, last.Sequence, q.Kind})
			sealed, e := s.codec.SealRecord(scope, "cursor", raw)
			if e != nil {
				return out, e
			}
			out.NextCursor = base64.RawURLEncoding.EncodeToString(sealed)
			break
		}
		raw, e := s.codec.OpenRecord(scope, locator, data)
		if e != nil {
			return out, e
		}
		var r session.WalletRecord
		e = json.Unmarshal(raw, &r)
		clear(raw)
		if e != nil {
			return out, e
		}
		if r.Kind != q.Kind || r.Sequence != sequence || r.Archived != archived || session.RecordLocator(v.Records.IndexSecret, r.Kind, r.Key) != locator {
			return out, errors.New("wallet history routing authentication failed")
		}
		out.Items = append(out.Items, r)
	}
	return out, rows.Err()
}
