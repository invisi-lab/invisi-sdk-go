package session

import (
	"context"
	"errors"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

// stage claims persisted work before invoking local capabilities. Failed local work
// restores its original record only while this worker still owns the claim.
func (o *Operation) stage(ctx context.Context, work func(*Operation) error, commitIsFinal ...bool) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := o.checkLease(ctx); err != nil {
		return nil, err
	}
	plan, err := frozen(o.plan)
	if err != nil {
		return nil, err
	}
	session := *o.session
	next := &Operation{session: &session, plan: plan, status: o.status, data: append([]byte(nil), o.data...), transactionHash: o.transactionHash, workerID: o.workerID}
	if o.authorization != nil {
		witness := *o.authorization
		next.authorization = &witness
	}
	if o.bundle != nil {
		bundle := prover.CloneBundle(*o.bundle)
		next.bundle = &bundle
	}
	var original OperationRecord
	worker, err := transaction.RandomOperationID()
	if err != nil {
		return nil, err
	}
	err = o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		record, ok := state.Operations[o.plan.ID]
		if !ok || state.Invalidated || record.WorkerID != o.workerID {
			return errors.New("stale operation worker")
		}
		original, err = frozen(record)
		if err != nil {
			return err
		}
		record.WorkerID = worker
		state.Operations[o.plan.ID] = record
		return nil
	})
	if err != nil {
		return nil, err
	}
	next.workerID = worker
	if err = work(next); err == nil && (len(commitIsFinal) == 0 || !commitIsFinal[0]) {
		err = ctx.Err()
	}
	if err != nil {
		cleanupErr := o.session.options.Store.Update(context.WithoutCancel(ctx), o.session.scope, func(state *WalletState) error {
			record, ok := state.Operations[o.plan.ID]
			if ok && record.WorkerID == next.workerID {
				state.Operations[o.plan.ID] = original
			}
			return nil
		})
		return nil, errors.Join(err, cleanupErr)
	}
	return next.bindBroadcasterQuote(), nil
}

func (o *Operation) Prove(ctx context.Context, options ...sdk.TransactionProveOptions) (*Operation, error) {
	if len(options) > 1 {
		return nil, errors.New("expected at most one transaction prove options value")
	}
	opt := sdk.TransactionProveOptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	return o.stage(ctx, func(next *Operation) error {
		if opt.Prover != nil {
			next.session.options.Prover = opt.Prover
		}
		_, err := next.proveInternal(ctx, opt)
		return err
	})
}

func (o *Operation) Ready(ctx context.Context) (*Operation, error) {
	return o.stage(ctx, func(next *Operation) error { _, err := next.readyInternal(ctx); return err })
}

func (o *Operation) AuthorizeFunding(ctx context.Context) (*Operation, error) {
	return o.stage(ctx, func(next *Operation) error { _, err := next.authorizeFundingInternal(ctx); return err })
}

func (o *Operation) Reconcile(ctx context.Context, options ...ReconcileOptions) (*Operation, error) {
	// Reconciliation commits inventory and its record atomically. Cancellation after
	// that commit must not roll back only the record.
	return o.stage(ctx, func(next *Operation) error { _, err := next.reconcileInternal(ctx, options...); return err }, true)
}

func (o *Operation) Status() string { o.mu.Lock(); defer o.mu.Unlock(); return o.status }
