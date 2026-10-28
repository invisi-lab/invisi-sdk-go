package sdk

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

func (o *RegistrationOperation) stage(ctx context.Context, work func(*RegistrationOperation) error) (*RegistrationOperation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	lifecycle := o.lifecycle
	lifecycle.mu.Lock()
	if lifecycle.busy || lifecycle.generation != o.generation {
		lifecycle.mu.Unlock()
		return nil, errors.New("registration handle is busy or stale")
	}
	lifecycle.busy = true
	lifecycle.mu.Unlock()
	defer func() { lifecycle.mu.Lock(); lifecycle.busy = false; lifecycle.mu.Unlock() }()
	registration, err := frozen(o.registration)
	if err != nil {
		return nil, err
	}
	contextCopy, err := frozen(o.context)
	if err != nil {
		return nil, err
	}
	next := &RegistrationOperation{lifecycle: lifecycle, generation: o.generation, registry: o.registry, wallet: o.wallet, options: o.options, registration: registration, context: contextCopy, status: o.status, data: append([]byte(nil), o.data...), transactionHash: o.transactionHash}
	if o.authority != nil {
		w := *o.authority
		next.authority = &w
	}
	if o.bundle != nil {
		b := prover.CloneBundle(*o.bundle)
		next.bundle = &b
	}
	if err = work(next); err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	lifecycle.mu.Lock()
	lifecycle.generation++
	next.generation = lifecycle.generation
	lifecycle.mu.Unlock()
	return next, nil
}
func (o *RegistrationOperation) Authorize(ctx context.Context) (*RegistrationOperation, error) {
	return o.stage(ctx, func(next *RegistrationOperation) error { _, err := next.authorizeInternal(ctx); return err })
}
func (o *RegistrationOperation) Prove(ctx context.Context, arguments ...any) (*RegistrationOperation, error) {
	backend, options, err := provingArguments(arguments)
	if err != nil {
		return nil, err
	}
	return o.stage(ctx, func(next *RegistrationOperation) error {
		if backend != nil {
			next.options.Prover = backend
		}
		_, err := next.proveInternal(ctx, options)
		return err
	})
}
func (o *RegistrationOperation) TrackTransaction(ctx context.Context, hash common.Hash) (*RegistrationOperation, error) {
	return o.stage(ctx, func(next *RegistrationOperation) error { _, err := next.trackInternal(ctx, hash); return err })
}

func (o *RegistrationOperation) Reconcile(ctx context.Context) (*RegistrationOperation, error) {
	return o.stage(ctx, func(next *RegistrationOperation) error { _, err := next.reconcileInternal(ctx); return err })
}
func (o *RegistrationOperation) Status() string { o.mu.Lock(); defer o.mu.Unlock(); return o.status }
