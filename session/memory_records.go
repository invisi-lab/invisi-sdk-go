package session

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
	"sort"
)

type MemoryWalletStoreOptions struct{ RecentRecords int }

func (m *MemoryWalletStore) ReadMetadata(ctx context.Context, scope WalletScope) (WalletState, error) {
	if e := ctx.Err(); e != nil {
		return WalletState{}, e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.states[scope.Key()]
	if !ok {
		v = emptyState()
	}
	ClearWalletRecords(&v)
	return CloneState(v)
}
func (m *MemoryWalletStore) ReadRecords(ctx context.Context, scope WalletScope, q RecordQuery) (RecordPage, error) {
	if e := ctx.Err(); e != nil {
		return RecordPage{}, e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if q.Limit == 0 {
		q.Limit = 50
	}
	if q.Limit < 1 || q.Limit > 256 {
		return RecordPage{}, errors.New("record page limit must be 1..256")
	}
	v := m.states[scope.Key()]
	records, e := WalletRecords(v)
	if e != nil {
		return RecordPage{}, e
	}
	for i := range records {
		locator := RecordLocator(v.Records.IndexSecret, records[i].Kind, records[i].Key)
		records[i].Sequence = v.Records.Active[locator].Sequence
	}

	before := v.Records.NextSequence + 1
	if q.Cursor != "" {
		data, e := base64.RawURLEncoding.DecodeString(q.Cursor)
		if e != nil {
			return RecordPage{}, e
		}
		raw, e := m.recordCodec.OpenRecord(scope, "cursor", data)
		if e != nil {
			return RecordPage{}, e
		}
		var c struct {
			Revision, Before uint64
			Kind             string
		}
		if e = json.Unmarshal(raw, &c); e != nil {
			return RecordPage{}, e
		}
		if c.Revision != v.Revision || c.Kind != q.Kind {
			return RecordPage{}, errors.New("history cursor scope or revision changed")
		}
		before = c.Before
	}
	out := RecordPage{Revision: v.Revision, Items: []WalletRecord{}}
	selected := []WalletRecord{}
	add := func(r WalletRecord) {
		if r.Kind != q.Kind || r.Sequence >= before {
			return
		}
		selected = append(selected, r)
		sort.Slice(selected, func(i, j int) bool { return selected[i].Sequence > selected[j].Sequence })
		if len(selected) > q.Limit+1 {
			selected = selected[:q.Limit+1]
		}
	}
	for _, r := range records {
		add(r)
	}
	for _, r := range m.archives[scope.Key()] {
		add(r)
	}
	if len(selected) > q.Limit {
		selected = selected[:q.Limit]
		raw, _ := json.Marshal(struct {
			Revision, Before uint64
			Kind             string
		}{v.Revision, selected[len(selected)-1].Sequence, q.Kind})
		data, e := m.recordCodec.SealRecord(scope, "cursor", raw)
		if e != nil {
			return out, e
		}
		out.NextCursor = base64.RawURLEncoding.EncodeToString(data)
	}
	out.Items = selected

	raw, e := json.Marshal(out)
	if e != nil {
		return out, e
	}
	e = json.Unmarshal(raw, &out)
	return out, e
}
func (m *MemoryWalletStore) retainRecords(scope WalletScope, v *WalletState) ([]WalletRecord, error) {
	if e := PrepareRecordMetadata(v); e != nil {
		return nil, e
	}
	records, e := WalletRecords(*v)
	if e != nil {
		return nil, e
	}
	for i := range records {
		locator := RecordLocator(v.Records.IndexSecret, records[i].Kind, records[i].Key)
		if old, ok := v.Records.Active[locator]; ok {
			records[i].Sequence = old.Sequence
		} else {
			archived, found := m.archives[scope.Key()][locator]
			if found {
				records[i].Sequence = archived.Sequence
			} else {
				v.Records.NextSequence++
				records[i].Sequence = v.Records.NextSequence
				v.Records.Counts[records[i].Kind]++
			}
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Sequence > records[j].Sequence })
	counts := map[string]int{}
	archive := []WalletRecord{}
	prior := v.Records.Active
	ClearWalletRecords(v)
	v.Records.Active = map[string]RecordManifestEntry{}
	for _, r := range records {
		if IsTerminalRecord(r, v.Scan.CreditReplayTarget != 0) {
			counts[r.Kind]++
			r.Archived = counts[r.Kind] > m.recentRecords
		}
		if r.Archived {
			archive = append(archive, r)
		} else {
			if e = ApplyWalletRecord(v, r); e != nil {
				return nil, e
			}
			v.Records.Active[RecordLocator(v.Records.IndexSecret, r.Kind, r.Key)] = RecordManifestEntry{Kind: r.Kind, Sequence: r.Sequence}
		}
	}
	present := map[string]bool{}
	for _, r := range records {
		present[RecordLocator(v.Records.IndexSecret, r.Kind, r.Key)] = true
	}
	for locator, m := range prior {
		if !present[locator] {
			v.Records.Counts[m.Kind]--
		}
	}

	return archive, nil
}

func (m *MemoryWalletStore) RestoreExternalizedOperations(ctx context.Context, scope WalletScope) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.states[scope.Key()]
	if !ok {
		return nil
	}
	next, e := CloneState(v)
	if e != nil {
		return e
	}
	remaining := map[string]WalletRecord{}
	for locator, r := range m.archives[scope.Key()] {
		restore := false
		switch r.Kind {
		case "operations":
			var o OperationRecord
			if e = json.Unmarshal(r.Value, &o); e != nil {
				return e
			}
			if o.Externalized {
				o.Status = "reorged"
				o.ResolvedAt = nil
				o.WorkerID = ""
				r.Value, e = json.Marshal(o)
				restore = true
			}
		case "credit-operations":
			var o creditsession.Journal
			if e = json.Unmarshal(r.Value, &o); e != nil {
				return e
			}
			if o.Externalized {
				o.Status = "reorged"
				o.WorkerID = ""
				r.Value, e = json.Marshal(o)
				restore = true
			}
		}
		if e != nil {
			return e
		}
		if !restore {
			remaining[locator] = r
			continue
		}
		r.Archived = false
		if e = ApplyWalletRecord(&next, r); e != nil {
			return e
		}
		locator := RecordLocator(next.Records.IndexSecret, r.Kind, r.Key)
		next.Records.Active[locator] = RecordManifestEntry{Kind: r.Kind, Sequence: r.Sequence}
		next.Invalidated = true
	}
	if len(remaining) == len(m.archives[scope.Key()]) {
		return nil
	}
	next.Revision++
	m.states[scope.Key()] = next
	m.archives[scope.Key()] = remaining
	return nil
}
