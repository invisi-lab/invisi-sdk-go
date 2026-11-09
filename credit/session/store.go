// Package session adds managed postpaid Credit state and journaling without depending on Note transaction plans.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type StoredState struct {
	Delivery              credit.DeliveryState
	Commitment, Nullifier common.Hash
	TreeID                uint64
	LeafIndex             uint16
	Role                  credit.Role
	Consumed, Revoked     bool
	ReservedBy            string
}

func GrantKey(commitment common.Hash, role credit.Role) string {
	return fmt.Sprintf("%s:%d", commitment.Hex(), role)
}
func (s StoredState) key() string { return GrantKey(s.Commitment, s.Role) }

type Journal struct {
	ID, WorkerID, Status string
	Intent               credit.Intent
	Proof                *prover.ProofBundle
	Calldata             []byte
	TransactionHash      common.Hash
	Externalized         bool
}
type State struct {
	Version    int
	Revision   uint64
	Grants     map[string]StoredState
	Operations map[string]Journal
}

func EmptyState() State {
	return State{Version: 2, Grants: map[string]StoredState{}, Operations: map[string]Journal{}}
}

type Store interface {
	ReadSnapshot(context.Context) (State, error)
	Update(context.Context, func(*State) error) error
}

func clone[T any](v T) (T, error) {
	var out T
	raw, err := json.Marshal(v)
	if err != nil {
		return out, err
	}
	defer clear(raw)
	err = json.Unmarshal(raw, &out)
	return out, err
}
func CloneState(s State) (State, error) { return clone(s) }

type MemoryStore struct {
	mu    sync.Mutex
	state State
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{state: EmptyState()} }
func (m *MemoryStore) ReadSnapshot(ctx context.Context) (State, error) {
	if err := ctx.Err(); err != nil {
		return State{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return CloneState(m.state)
}
func (m *MemoryStore) Update(ctx context.Context, change func(*State) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	next, err := CloneState(m.state)
	if err != nil {
		return err
	}
	if err = change(&next); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if m.state.Revision == ^uint64(0) {
		return fmt.Errorf("Credit revision exhausted")
	}
	next.Revision = m.state.Revision + 1
	next, err = CloneState(next)
	if err != nil {
		return err
	}
	m.state = next
	return nil
}
