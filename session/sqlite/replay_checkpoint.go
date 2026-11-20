package sqlite

import (
	"context"
	"errors"

	"github.com/invisi-labs/invisi-sdk-go/session"
)

// CommitCreditReplayCheckpoint seals the existing authenticated metadata without
// loading or rewriting record rows. The final replay page always uses Update,
// which reauthenticates the active record manifest and restores reservations.
func (s *WalletStore) CommitCreditReplayCheckpoint(ctx context.Context, scope session.WalletScope, checkpoint session.CreditReplayCheckpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	state, err := s.readMetadata(ctx, tx, scope)
	if err != nil {
		return err
	}
	if err = checkpoint.Apply(&state); err != nil {
		return err
	}
	if state.Revision == 1<<63-1 {
		return errors.New("wallet revision exhausted")
	}
	state.Revision++
	data, err := s.codec.Seal(scope, state)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, "UPDATE invisi_wallet_state SET revision=revision+1,data=? WHERE scope_key=? AND revision=?", data, scope.Key(), checkpoint.Revision)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return errors.New("concurrent wallet checkpoint")
	}
	return tx.Commit()
}
