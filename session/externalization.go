package session

import (
	"context"
	"errors"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

func (s *WalletSession) externalize(ctx context.Context, e *transaction.Execution) error {
	return s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		r, ok := state.Operations[e.Plan.ID]
		if !ok || r.WorkerID != e.WorkerID || state.Invalidated {
			return errors.New("stale operation worker")
		}
		r.Externalized = true
		r.Status = "uncertain"
		state.Operations[e.Plan.ID] = r
		return nil
	})
}
