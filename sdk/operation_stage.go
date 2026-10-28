package sdk

import (
	"context"
	"errors"

	"github.com/invisi-lab/invisi-sdk-go/prover"
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
	managed := o.session.explicitInventory == nil
	if managed {
		worker, err := randomOperationID()
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
	}
	if err = work(next); err == nil && (len(commitIsFinal) == 0 || !commitIsFinal[0]) {
		err = ctx.Err()
	}
	if err != nil {
		if managed {
			cleanupErr := o.session.options.Store.Update(context.WithoutCancel(ctx), o.session.scope, func(state *WalletState) error {
				record, ok := state.Operations[o.plan.ID]
				if ok && record.WorkerID == next.workerID {
					state.Operations[o.plan.ID] = original
				}
				return nil
			})
			return nil, errors.Join(err, cleanupErr)
		}
		return nil, err
	}
	return next, nil
}

func (o *Operation) Authorize(ctx context.Context) (*Operation, error) {
	return o.stage(ctx, func(next *Operation) error { _, err := next.authorizeInternal(ctx); return err })
}
func (o *Operation) Prove(ctx context.Context, arguments ...any) (*Operation, error) {
	backend, options, err := provingArguments(arguments)
	if err != nil {
		return nil, err
	}
	return o.stage(ctx, func(next *Operation) error {
		if backend != nil {
			next.session.options.Prover = backend
		}
		_, err := next.proveInternal(ctx, options)
		return err
	})
}
func (o *Operation) Ready(ctx context.Context) (*Operation, error) {
	return o.stage(ctx, func(next *Operation) error { _, err := next.readyInternal(ctx); return err })
}
func (o *Operation) AuthorizeFunding(ctx context.Context, signer ...FundingSigner) (*Operation, error) {
	if len(signer) > 1 {
		return nil, errors.New("expected at most one funding signer")
	}
	return o.stage(ctx, func(next *Operation) error {
		if len(signer) == 1 {
			next.session.options.FundingSigner = signer[0]
		}
		_, err := next.authorizeFundingInternal(ctx)
		return err
	})
}

func provingArguments(arguments []any) (Prover, ProverOptions, error) {
	var backend Prover
	var options ProverOptions
	seenOptions := false
	for _, argument := range arguments {
		switch value := argument.(type) {
		case Prover:
			if backend != nil || seenOptions {
				return nil, options, errors.New("expected prover followed by options")
			}
			backend = value
		case ProverOptions:
			if seenOptions {
				return nil, options, errors.New("expected at most one prover options value")
			}
			options, seenOptions = value, true
		default:
			return nil, options, errors.New("expected prover or prover options")
		}
	}
	return backend, options, nil
}

func (o *Operation) Reconcile(ctx context.Context, options ...ReconcileOptions) (*Operation, error) {
	// Reconciliation commits inventory and its record atomically. Cancellation after
	// that commit must not roll back only the record.
	return o.stage(ctx, func(next *Operation) error { _, err := next.reconcileInternal(ctx, options...); return err }, true)
}
func (o *Operation) Status() string { o.mu.Lock(); defer o.mu.Unlock(); return o.status }
