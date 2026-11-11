package session

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

type OperationRecord struct {
	ResolvedAt       *protocolrpc.CanonicalBlock
	ID, Kind, Status string
	TransactionHash  common.Hash
	Inputs           []common.Hash
	CallData         []byte
	Deadline         uint64
	Plan             json.RawMessage
	Proof            json.RawMessage
	WorkerID         string
	Externalized     bool
}

type WalletState struct {
	Scan             PrivateScanState
	Records          RecordMetadata
	FormatVersion    int
	PostpaidCredit   creditsession.State
	Revision         uint64
	DiscoveryThrough *protocolrpc.CanonicalBlock
	Notes            map[common.Hash]StoredNote
	Operations       map[string]OperationRecord
	ObservedThrough  *protocolrpc.CanonicalBlock
	Invalidated      bool
	CheckedThrough   *protocolrpc.CanonicalBlock
}

func emptyState() WalletState {
	return WalletState{FormatVersion: 2, PostpaidCredit: creditsession.EmptyState(), Notes: map[common.Hash]StoredNote{}, Operations: map[string]OperationRecord{}}
}

func CloneState(s WalletState) (WalletState, error) {
	raw, e := json.Marshal(s)
	if e != nil {
		return WalletState{}, e
	}
	var result WalletState
	e = json.Unmarshal(raw, &result)
	return result, e
}

type WalletStore interface {
	ReadSnapshot(context.Context, WalletScope) (WalletState, error)
	Update(context.Context, WalletScope, func(*WalletState) error) error
}

type MemoryWalletStore struct {
	mu            sync.Mutex
	states        map[string]WalletState
	archives      map[string]map[string]WalletRecord
	recentRecords int
	recordCodec   *EncryptedWalletCodec
}

func NewMemoryWalletStore(options ...MemoryWalletStoreOptions) *MemoryWalletStore {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	codec, _ := NewEncryptedWalletCodec(key)
	clear(key)
	recent := 128
	if len(options) > 0 {
		recent = options[0].RecentRecords
	}
	return &MemoryWalletStore{states: map[string]WalletState{}, archives: map[string]map[string]WalletRecord{}, recentRecords: recent, recordCodec: codec}
}

func (m *MemoryWalletStore) ReadSnapshot(ctx context.Context, scope WalletScope) (WalletState, error) {
	if e := ctx.Err(); e != nil {
		return WalletState{}, e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.states[scope.Key()]
	if !ok {
		s = emptyState()
	}
	return CloneState(s)
}

func (m *MemoryWalletStore) Update(ctx context.Context, scope WalletScope, update func(*WalletState) error) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.states[scope.Key()]
	if !ok {
		s = emptyState()
	}
	clone, e := CloneState(s)
	if e != nil {
		return e
	}
	if e = update(&clone); e != nil {
		return e
	}
	if s.Records.Version != 0 && !reflect.DeepEqual(clone.Records, s.Records) {
		return errors.New("wallet storage metadata must be preserved")
	}
	if s.Revision >= 1<<63-1 {
		return errors.New("wallet revision exhausted")
	}
	clone.Revision = s.Revision + 1
	if e := ctx.Err(); e != nil {
		return e
	}
	saved, e := CloneState(clone)
	if e != nil {
		return e
	}
	if m.archives != nil {
		archive, e := m.retainRecords(scope, &saved)
		if e != nil {
			return e
		}
		existing := m.archives[scope.Key()]
		if existing == nil {
			existing = map[string]WalletRecord{}
		}
		for locator := range saved.Records.Active {
			delete(existing, locator)
		}
		for _, r := range archive {
			existing[RecordLocator(saved.Records.IndexSecret, r.Kind, r.Key)] = r
		}
		m.archives[scope.Key()] = existing
	}
	m.states[scope.Key()] = saved
	return nil
}

type EncryptedWalletCodec struct{ aead cipher.AEAD }

func NewEncryptedWalletCodec(key []byte) (*EncryptedWalletCodec, error) {
	if len(key) != 32 {
		return nil, errors.New("wallet storage key must contain 32 bytes")
	}
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	aead, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	return &EncryptedWalletCodec{aead}, nil
}

func (c *EncryptedWalletCodec) Seal(scope WalletScope, state WalletState) ([]byte, error) {
	state.FormatVersion = 2
	if state.PostpaidCredit.Version == 0 {
		state.PostpaidCredit = creditsession.EmptyState()
	}
	raw, e := json.Marshal(state)
	if e != nil {
		return nil, e
	}
	defer clear(raw)
	nonce := make([]byte, c.aead.NonceSize())
	if _, e = rand.Read(nonce); e != nil {
		return nil, e
	}
	return c.aead.Seal(nonce, nonce, raw, []byte("Invisi.WalletStore.v1:"+scope.Key())), nil
}

func (c *EncryptedWalletCodec) decode(scope WalletScope, data []byte) (WalletState, error) {
	if len(data) < c.aead.NonceSize()+c.aead.Overhead() {
		return WalletState{}, errors.New("invalid encrypted wallet state")
	}
	nonce := data[:c.aead.NonceSize()]
	plain, e := c.aead.Open(nil, nonce, data[len(nonce):], []byte("Invisi.WalletStore.v1:"+scope.Key()))
	if e != nil {
		return WalletState{}, errors.New("wallet state authentication failed")
	}
	defer clear(plain)
	s := emptyState()
	e = json.Unmarshal(plain, &s)
	if e != nil {
		return WalletState{}, e
	}

	return s, nil
}

func normalizeWalletVersion(s WalletState) (WalletState, error) {
	if s.FormatVersion != 0 && s.FormatVersion != 2 {
		return WalletState{}, errors.New("unsupported wallet storage version")
	}
	if s.PostpaidCredit.Version == 0 {
		s.PostpaidCredit = creditsession.EmptyState()
	}
	if s.PostpaidCredit.Version != 2 {
		return WalletState{}, errors.New("unsupported Credit storage schema")
	}
	s.FormatVersion = 2
	return s, nil
}
func (c *EncryptedWalletCodec) Open(scope WalletScope, data []byte) (WalletState, error) {
	state, err := c.decode(scope, data)
	if err != nil {
		return WalletState{}, err
	}
	return normalizeWalletVersion(state)
}

// restoreReservations preserves pending spends when rebuilding the local inventory.
func restoreReservations(state *WalletState) error {
	for id, record := range state.Operations {
		if record.Status == "confirmed" || record.Status == "expired" || (record.Status == "failed" && (!record.Externalized || record.ResolvedAt != nil)) {
			continue
		}
		var plan transaction.Plan
		if err := json.Unmarshal(record.Plan, &plan); err != nil {
			return err
		}
		for _, input := range plan.Inputs {
			if note, ok := state.Notes[input.Commitment]; ok && !note.Spent {
				if note.ReservedBy != "" && note.ReservedBy != id {
					return errors.New("conflicting persisted note reservations")
				}
				note.ReservedBy = id
				state.Notes[input.Commitment] = note
			}
		}

	}
	return nil
}

type WalletScope = transaction.WalletScope
type StoredNote = transaction.NoteRecord
