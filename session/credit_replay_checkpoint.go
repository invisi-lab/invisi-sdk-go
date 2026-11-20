package session

import (
	"context"
	"errors"
)

// CreditReplayCheckpoint only advances an unfinished replay over a validated
// empty Credit window. It cannot add private records or complete an obligation.
// The optional adapter capability avoids decrypting every retained wallet record
// for a cursor-only update; adapters without it keep the full atomic update path.
type CreditReplayCheckpoint struct {
	Revision, Generation, ReplayRevision, Target, Cursor uint64
}

func (c CreditReplayCheckpoint) Apply(v *WalletState) error {
	if v.Invalidated || v.Revision != c.Revision || v.Scan.Generation != c.Generation || v.Scan.CreditReplayRevision != c.ReplayRevision || v.Scan.CreditReplayTarget != c.Target {
		return errors.New("wallet changed during Credit replay checkpoint")
	}
	if c.Target == 0 || c.Cursor >= c.Target || c.Cursor <= v.Scan.CreditReplayCursor || c.Cursor < v.Scan.CreditReplayFrom {
		return errors.New("invalid unfinished Credit replay checkpoint")
	}
	v.Scan.CreditReplayCursor = c.Cursor
	return nil
}

type creditReplayCheckpointStore interface {
	ReadMetadata(context.Context, WalletScope) (WalletState, error)
	CommitCreditReplayCheckpoint(context.Context, WalletScope, CreditReplayCheckpoint) error
}

func (m *MemoryWalletStore) CommitCreditReplayCheckpoint(ctx context.Context, scope WalletScope, c CreditReplayCheckpoint) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.states[scope.Key()]
	if !ok {
		return errors.New("wallet checkpoint requires persisted state")
	}
	if err := c.Apply(&v); err != nil {
		return err
	}
	if v.Revision == ^uint64(0) {
		return errors.New("wallet revision exhausted")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	v.Revision++
	m.states[scope.Key()] = v
	return nil
}

func (s *WalletSession) checkpointEmptyCreditReplay(ctx context.Context, page PublicPage, required, replayRevision, cursor uint64) (bool, error) {
	store, ok := s.options.Store.(creditReplayCheckpointStore)
	if !ok || cursor >= required {
		return false, nil
	}
	for _, tx := range page.Transactions {
		if len(tx.Batch.CreditStates) > 0 || len(tx.Batch.CreditNullifiers) > 0 {
			return false, nil
		}
	}
	before, err := store.ReadMetadata(ctx, s.scope)
	if err != nil {
		return false, err
	}
	if before.Scan.CreditReplayRevision != replayRevision || before.Scan.CreditReplayTarget != required {
		return false, nil
	}
	public, err := s.publicHistory.Snapshot(ctx)
	if err != nil {
		return false, err
	}
	if public.Generation != before.Scan.Generation {
		return false, errors.New("public history generation changed")
	}
	c := CreditReplayCheckpoint{Revision: before.Revision, Generation: before.Scan.Generation, ReplayRevision: replayRevision, Target: required, Cursor: cursor}
	if err = store.CommitCreditReplayCheckpoint(ctx, s.scope, c); err != nil {
		return false, err
	}
	public, err = s.publicHistory.Snapshot(ctx)
	if err != nil {
		return false, err
	}
	if public.Generation != c.Generation {
		_ = s.invalidate(context.WithoutCancel(ctx))
		return false, errors.New("public history generation changed during Credit replay checkpoint")
	}
	return true, nil
}
