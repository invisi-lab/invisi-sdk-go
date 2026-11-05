package session

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

// ExportTransaction exposes only the frozen public call and keeps managed inputs reserved.
func (o *Operation) ExportTransaction(ctx context.Context) (sdk.TransactionCall, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkLease(ctx); err != nil {
		return sdk.TransactionCall{}, err
	}
	if o.status != "ready" || o.plan.Target == "noir-recursive" || len(o.data) == 0 {
		return sdk.TransactionCall{}, errors.New("ready EVM transaction required")
	}
	if err := o.validateExport(ctx); err != nil {
		return sdk.TransactionCall{}, err
	}
	err := o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		record, ok := state.Operations[o.plan.ID]
		if !ok || record.WorkerID != o.workerID || state.Invalidated {
			return errors.New("stale operation worker")
		}
		record.Externalized = true
		state.Operations[o.plan.ID] = record
		return nil
	})
	if err != nil {
		return sdk.TransactionCall{}, err
	}
	return sdk.TransactionCall{To: o.plan.Context.ProtocolAddress, Data: append([]byte(nil), o.data...), Value: new(big.Int)}, nil
}

// TrackTransaction records an association, not confirmation of successful execution.
func (o *Operation) TrackTransaction(ctx context.Context, hash common.Hash) (*Operation, error) {
	if hash == (common.Hash{}) {
		return nil, errors.New("transaction hash required")
	}
	return o.stage(ctx, func(next *Operation) error {
		if next.status != "ready" && next.status != "submitted" && next.status != "uncertain" {
			return errors.New("ready or uncertain operation required")
		}
		next.transactionHash = hash
		next.status = "uncertain"
		return next.persist(ctx, next.status)
	})
}
