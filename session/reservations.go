package session

import (
	"context"
	"encoding/json"
	"errors"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

func (s *WalletSession) reserveExecution(ctx context.Context, execution *transaction.Execution) error {
	plan := execution.Plan
	cached := execution.Runtime.Snapshot
	if cached != nil && s.publicHistory != nil {
		public, err := s.publicHistory.Snapshot(ctx)
		if err != nil {
			return err
		}
		if public.Generation != cached.Generation {
			return errors.New("wallet requires reset refresh")
		}
	}
	id := plan.ID
	encoded, e := json.Marshal(plan)
	if e != nil {
		return e
	}
	e = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		if state.Invalidated {
			return errors.New("wallet requires reset refresh")
		}
		for _, n := range plan.Inputs {
			current, ok := state.Notes[n.Commitment]
			if !ok || current.Spent || current.ReservedBy != "" {
				return errors.New("input already spent or reserved")
			}
			if !current.Note.Equal(n.Note) || current.TreeID != n.TreeID || current.LeafIndex != n.LeafIndex {
				return errors.New("input changed during preparation")
			}
		}

		record := OperationRecord{ID: id, Kind: transaction.PublicOperationKind(plan), Status: "prepared", Deadline: plan.Deadline, Plan: encoded, WorkerID: id}
		for _, n := range plan.Inputs {
			current := state.Notes[n.Commitment]
			current.ReservedBy = id
			state.Notes[n.Commitment] = current
			record.Inputs = append(record.Inputs, n.Commitment)
		}

		state.Operations[id] = record
		return nil
	})
	if e == nil && cached != nil && s.publicHistory != nil {
		public, err := s.publicHistory.Snapshot(ctx)
		if err != nil {
			return err
		}
		if public.Generation != cached.Generation {
			_ = s.invalidate(ctx)
			return errors.New("wallet requires reset refresh")
		}
	}
	return e
}
