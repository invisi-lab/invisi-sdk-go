package session

import (
	"context"
	"encoding/json"
	"errors"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

func (s *WalletSession) reserveExecution(ctx context.Context, execution *transaction.Execution) error {
	plan := execution.Plan
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
		for _, c := range plan.CreditInputs {
			for _, existing := range state.Credits {
				if existing.Commitment == c.Commitment && (existing.Spent || existing.ReservedBy != "") {
					return errors.New("Credit already spent or reserved")
				}
			}
			current, ok := state.Credits[transaction.CreditKey(c.Commitment, c.Role)]
			if !ok || current.Spent || current.ReservedBy != "" {
				return errors.New("Credit already spent or reserved")
			}
		}
		record := OperationRecord{ID: id, Kind: transaction.PublicOperationKind(plan), Status: "prepared", Deadline: plan.Deadline, Plan: encoded, WorkerID: id}
		for _, n := range plan.Inputs {
			current := state.Notes[n.Commitment]
			current.ReservedBy = id
			state.Notes[n.Commitment] = current
			record.Inputs = append(record.Inputs, n.Commitment)
		}
		for _, c := range plan.CreditInputs {
			for key, current := range state.Credits {
				if current.Commitment == c.Commitment {
					current.ReservedBy = id
					state.Credits[key] = current
				}
			}
		}
		state.Operations[id] = record
		return nil
	})
	return e
}
