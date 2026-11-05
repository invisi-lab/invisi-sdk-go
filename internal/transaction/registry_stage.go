package transaction

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
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
	registration, err := Frozen(o.registration)
	if err != nil {
		return nil, err
	}
	contextCopy, err := Frozen(o.context)
	if err != nil {
		return nil, err
	}
	next := &RegistrationOperation{lifecycle: lifecycle, generation: o.generation, registry: o.registry, wallet: o.wallet, options: o.options, registration: registration, context: contextCopy, status: o.status, data: append([]byte(nil), o.data...), direct: o.direct, transactionHash: o.transactionHash}
	if o.receiptBlock != nil {
		block := *o.receiptBlock
		next.receiptBlock = &block
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

// WithSignature verifies an externally obtained EOA or ERC-1271 signature.
func (o *RegistrationOperation) WithSignature(ctx context.Context, signature []byte) (*RegistrationOperation, error) {
	signature = append([]byte(nil), signature...)
	return o.stage(ctx, func(next *RegistrationOperation) error {
		_, err := next.withSignatureInternal(ctx, signature)
		return err
	})
}

// AuthorizeDirect creates an empty-signature call that only the owner may submit.
func (o *RegistrationOperation) AuthorizeDirect(ctx context.Context) (*RegistrationOperation, error) {
	return o.stage(ctx, func(next *RegistrationOperation) error {
		if next.status != "prepared" {
			return errors.New("registration must be prepared")
		}
		current, err := next.currentContext(ctx)
		if err != nil {
			return err
		}
		call, err := protocol.BuildRegistrationCalldata(current, next.registration, nil, protocol.RegistrationOptions{Update: next.options.Update})
		if err != nil {
			return err
		}
		next.data = append([]byte(nil), call.Data...)
		next.status = "ready"
		next.direct = true
		return nil
	})
}

// ExportRegistrationAuthorization exports a fresh, JSON-safe request without accessing private capabilities.
func (o *RegistrationOperation) ExportRegistrationAuthorization(ctx context.Context) (core.SigningRequest, error) {
	if err := ctx.Err(); err != nil {
		return core.SigningRequest{}, err
	}
	o.lifecycle.mu.Lock()
	defer o.lifecycle.mu.Unlock()
	if o.lifecycle.busy || o.lifecycle.generation != o.generation {
		return core.SigningRequest{}, errors.New("registration handle is busy or stale")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.status != "prepared" {
		return core.SigningRequest{}, errors.New("registration must be prepared")
	}
	current, err := o.currentContext(ctx)
	if err != nil {
		return core.SigningRequest{}, err
	}
	typed, err := protocol.RegistrationTypedData(current, o.registration, protocol.RegistrationOptions{Update: o.options.Update})
	if err != nil {
		return core.SigningRequest{}, err
	}
	if err = ctx.Err(); err != nil {
		return core.SigningRequest{}, err
	}
	return core.NewSigningRequest(typed)
}

func (o *RegistrationOperation) TrackTransaction(ctx context.Context, hash common.Hash) (*RegistrationOperation, error) {
	return o.stage(ctx, func(next *RegistrationOperation) error { _, err := next.trackInternal(ctx, hash); return err })
}

func (o *RegistrationOperation) Reconcile(ctx context.Context) (*RegistrationOperation, error) {
	return o.stage(ctx, func(next *RegistrationOperation) error { _, err := next.reconcileInternal(ctx); return err })
}

func (o *RegistrationOperation) Status() string { o.mu.Lock(); defer o.mu.Unlock(); return o.status }
