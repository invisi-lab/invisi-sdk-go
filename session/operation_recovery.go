package session

import (
	"context"
	"encoding/json"
	"errors"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

func (s *WalletSession) Resume(ctx context.Context, id string) (*Operation, error) {
	worker, err := transaction.RandomOperationID()
	if err != nil {
		return nil, err
	}
	var operation *Operation
	err = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		record, ok := state.Operations[id]
		if !ok {
			return errors.New("unknown operation")
		}
		if !record.Externalized {
			switch record.Status {
			case "failed", "expired", "confirmed", "reverted":
				return errors.New("local terminal operation cannot be resumed")
			}
		}

		var plan transaction.Plan
		if err := json.Unmarshal(record.Plan, &plan); err != nil {
			return err
		}
		if plan.ID != id || plan.Context.ChainID == nil || !plan.Context.ChainID.IsUint64() || plan.Context.ChainID.Uint64() != s.scope.ChainID || plan.Context.ProtocolAddress != s.scope.ProtocolAddress {
			return errors.New("saved operation deployment mismatch")
		}
		operation = &Operation{session: s, plan: plan, status: record.Status, data: append([]byte(nil), record.CallData...), transactionHash: record.TransactionHash, workerID: worker}
		if len(record.Proof) > 0 {
			if err := json.Unmarshal(record.Proof, &operation.bundle); err != nil {
				return err
			}
		}
		if err := operation.validateRecovered(); err != nil {
			return err
		}
		if operation.status == "authorized" {
			operation.status = "prepared"
			record.Status = "prepared"
		}
		record.WorkerID = worker
		state.Operations[id] = record
		return nil
	})
	if err != nil {
		return nil, err
	}
	return operation.bindBroadcasterQuote(), nil
}
