package sdk

import (
	"context"
	"errors"
	"math/big"
	"sync"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type TransactionCall = protocol.TransactionCall

type transactionLifecycle struct {
	mu         sync.Mutex
	generation uint64
	busy       bool
}

// PreparedTransaction is a frozen local transaction handle. It has no Store, submission or recovery API.
// Successful stages replace the handle. A failed or canceled stage leaves it available for retry.
type PreparedTransaction struct {
	operation  *Operation
	lifecycle  *transactionLifecycle
	generation uint64
}

func detached(op *Operation, err error) (*PreparedTransaction, error) {
	if err != nil {
		return nil, err
	}
	if len(op.plan.CreditInputs) != len(op.session.explicitInventory.Credits) {
		return nil, errors.New("Credit inputs must exactly match the requested operation")
	}
	return &PreparedTransaction{operation: op, lifecycle: &transactionLifecycle{}}, nil
}
func (p *PreparedTransaction) String() string   { return "PreparedTransaction([private plan redacted])" }
func (p *PreparedTransaction) GoString() string { return p.String() }
func (p *PreparedTransaction) MarshalJSON() ([]byte, error) {
	return []byte(`{"type":"PreparedTransaction"}`), nil
}
func (p *PreparedTransaction) Preview() OperationPreview { return p.operation.Preview() }
func (p *PreparedTransaction) stage(ctx context.Context, perform func(*Operation) error) (*PreparedTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	l := p.lifecycle
	l.mu.Lock()
	if l.busy || l.generation != p.generation {
		l.mu.Unlock()
		return nil, errors.New("transaction handle is busy or stale")
	}
	l.busy = true
	l.mu.Unlock()
	defer func() { l.mu.Lock(); l.busy = false; l.mu.Unlock() }()
	plan, err := frozen(p.operation.plan)
	if err != nil {
		return nil, err
	}
	session := *p.operation.session
	next := &Operation{session: &session, plan: plan, status: p.operation.status, data: append([]byte(nil), p.operation.data...)}
	if p.operation.authorization != nil {
		w := *p.operation.authorization
		next.authorization = &w
	}
	if p.operation.bundle != nil {
		bundle := prover.CloneBundle(*p.operation.bundle)
		next.bundle = &bundle
	}
	if err = perform(next); err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	l.mu.Lock()
	l.generation++
	generation := l.generation
	l.mu.Unlock()
	return &PreparedTransaction{operation: next, lifecycle: l, generation: generation}, nil
}
func (p *PreparedTransaction) Prove(ctx context.Context, options ...TransactionProveOptions) (*PreparedTransaction, error) {
	if len(options) > 1 {
		return nil, errors.New("expected at most one transaction prove options value")
	}
	opt := TransactionProveOptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	return p.stage(ctx, func(next *Operation) error {
		if opt.Prover != nil {
			next.session.options.Prover = opt.Prover
		}
		_, err := next.proveInternal(ctx, opt)
		return err
	})
}

func (p *PreparedTransaction) Ready(ctx context.Context) (*PreparedTransaction, error) {
	return p.stage(ctx, func(o *Operation) error { _, err := o.readyInternal(ctx); return err })
}
func (p *PreparedTransaction) current(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.lifecycle.busy || p.lifecycle.generation != p.generation {
		return errors.New("transaction handle is busy or stale")
	}
	return nil
}
func (p *PreparedTransaction) ExportTransaction(ctx context.Context) (TransactionCall, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if err := p.current(ctx); err != nil {
		return TransactionCall{}, err
	}
	o := p.operation
	if o.status != "ready" || o.plan.Target == "noir-recursive" || len(o.data) == 0 {
		return TransactionCall{}, errors.New("ready EVM transaction required")
	}
	return TransactionCall{To: o.plan.Context.ProtocolAddress, Data: append([]byte(nil), o.data...), Value: new(big.Int)}, nil
}
func (p *PreparedTransaction) ExportProof(ctx context.Context) (prover.ProofBundle, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if err := p.current(ctx); err != nil {
		return prover.ProofBundle{}, err
	}
	if p.operation.bundle == nil {
		return prover.ProofBundle{}, errors.New("proof is not ready")
	}
	return prover.CloneBundle(*p.operation.bundle), nil
}
func (p *PreparedTransaction) ExportAggregateMember(ctx context.Context) (OperationAggregateMember, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if err := p.current(ctx); err != nil {
		return OperationAggregateMember{}, err
	}
	return p.operation.ExportAggregateMember(ctx)
}
