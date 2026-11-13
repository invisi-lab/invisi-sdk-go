package session

import (
	"context"
	"errors"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	"sync"

	"github.com/invisi-lab/invisi-sdk-go/rpc/chain"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

type Options struct {
	PublicHistory *PublicHistory
	Invisi        *sdk.Invisi
	Wallet        *sdk.Wallet
	Protocol      *chain.Protocol
	Store         WalletStore
	Namespace     string
	Signer        *bind.TransactOpts
	FundingSigner sdk.FundingSigner
}
type sessionOptions struct {
	sdk.InvisiOptions
	Store         WalletStore
	Signer        *bind.TransactOpts
	FundingSigner sdk.FundingSigner
}
type WalletSession struct {
	publicHistory *PublicHistory
	history       *sessionNotifications
	refreshGate   chan struct{}
	sends         *sessionSendState
	options       sessionOptions
	wallet        *sdk.Wallet
	protocol      *chain.Protocol
	scope         WalletScope
	rpc           *protocolrpc.ProtocolRpcClient
	runtime       *transaction.Runtime
}

func Connect(ctx context.Context, options Options) (*WalletSession, error) {
	if options.Invisi == nil {
		return nil, errors.New("Invisi infrastructure required")
	}
	if options.Store == nil {
		return nil, errors.New("store required for a managed wallet session")
	}
	var runtime *transaction.Runtime
	var e error
	if options.Wallet != nil && options.Protocol != nil {
		scope := cachedScope(options)
		var state WalletState
		var err error
		if metadata, ok := options.Store.(interface {
			ReadMetadata(context.Context, WalletScope) (WalletState, error)
		}); ok {
			state, err = metadata.ReadMetadata(ctx, scope)
		} else {
			state, err = options.Store.ReadSnapshot(ctx, scope)
		}
		if err != nil {
			return nil, err
		}
		if snapshot := matchingSnapshot(state); snapshot != nil {
			runtime, e = transaction.RestorePreparationRuntime(ctx, options.Invisi.GetInfrastructure(), options.Wallet, options.Protocol, options.Namespace, snapshot)
			if errors.Is(e, transaction.ErrPreparationSnapshotMissing) {
				runtime = nil
				e = nil
			}
		}
	}
	if runtime == nil && e == nil {
		runtime, e = transaction.ResolveRuntime(ctx, options.Invisi.GetInfrastructure(), options.Wallet, options.Protocol, options.Namespace)
	}
	if e != nil {
		return nil, e
	}
	s := &WalletSession{history: newSessionNotifications(), refreshGate: make(chan struct{}, 1), sends: &sessionSendState{sending: map[string]bool{}}, options: sessionOptions{options.Invisi.GetInfrastructure(), options.Store, options.Signer, options.FundingSigner}, wallet: runtime.Wallet, protocol: options.Protocol, scope: runtime.Scope, rpc: runtime.RPC, runtime: runtime}
	s.publicHistory = options.PublicHistory
	if s.publicHistory == nil {
		s.publicHistory, e = sharedPublicHistory(options.Invisi, options.Protocol)
		if e != nil {
			return nil, e
		}
	}
	expectedHistory, e := NewPublicHistory(PublicHistoryOptions{Protocol: options.Protocol, Provider: s.options.Provider, ProtocolRPC: s.rpc})
	if e != nil {
		return nil, e
	}
	if s.publicHistory.key != expectedHistory.key {
		return nil, errors.New("public history deployment mismatch")
	}
	s.options.Store = &notifyingWalletStore{base: s.options.Store, notifications: s.history}
	runtime.Options.Signer = options.Signer
	runtime.Options.FundingSigner = options.FundingSigner
	runtime.InventoryReader = func(ctx context.Context) (transaction.Inventory, error) {
		state, e := s.fencedSnapshot(ctx)
		return transaction.Inventory{Notes: state.Notes, Invalidated: state.Invalidated, ObservedThrough: state.ObservedThrough, CheckedThrough: state.CheckedThrough}, e
	}
	runtime.SummaryReader = func(ctx context.Context) (transaction.Inventory, error) {
		var state WalletState
		var err error
		if metadata, ok := s.options.Store.(interface {
			ReadMetadata(context.Context, WalletScope) (WalletState, error)
		}); ok {
			state, err = metadata.ReadMetadata(ctx, s.scope)
		} else {
			state, err = s.options.Store.ReadSnapshot(ctx, s.scope)
		}
		if err == nil && s.publicHistory != nil {
			public, e := s.publicHistory.Snapshot(ctx)
			err = e
			if err == nil && public.Generation != state.Scan.Generation {
				state.Invalidated = true
			}
		}
		return transaction.Inventory{Invalidated: state.Invalidated, ObservedThrough: state.ObservedThrough, CheckedThrough: state.CheckedThrough}, err
	}
	runtime.OnComplete = s.reserveExecution
	runtime.OnPersist = func(ctx context.Context, e *transaction.Execution, status string) error {
		return operationFromExecution(s, e).persist(ctx, status)
	}
	runtime.OnLease = func(ctx context.Context, e *transaction.Execution) error {
		return operationFromExecution(s, e).checkLease(ctx)
	}
	runtime.OnInvalidate = s.invalidate
	runtime.OnReprice = func(ctx context.Context, e *transaction.Execution) error {
		return operationFromExecution(s, e).commitBroadcasterPlan(ctx)
	}
	runtime.OnExternalize = func(ctx context.Context, e *transaction.Execution) error { return s.externalize(ctx, e) }
	return s, nil
}

type Operation struct {
	broadcaster.OperationTarget `json:"-"`
	mu                          sync.Mutex
	metadataMu                  sync.RWMutex
	session                     *WalletSession
	plan                        transaction.Plan
	status                      string
	authorization               *primitives.WalletAuthorizationWitness
	bundle                      *prover.ProofBundle
	data                        []byte
	transactionHash             common.Hash
	workerID                    string
}

func operationFromExecution(s *WalletSession, e *transaction.Execution) *Operation {
	return (&Operation{session: s, plan: e.Plan, status: e.Status, authorization: e.Authorization, bundle: e.Bundle, data: e.Data, transactionHash: e.TransactionHash, workerID: e.WorkerID}).bindBroadcasterQuote()
}
func (o *Operation) execution() *transaction.Execution {
	r := &transaction.Runtime{}
	if o.session != nil {
		r = o.session.runtimeSnapshot()
	}
	return &transaction.Execution{Runtime: r, Plan: o.plan, Status: o.status, Authorization: o.authorization, Bundle: o.bundle, Data: o.data, TransactionHash: o.transactionHash, WorkerID: o.workerID}
}
func (o *Operation) apply(e *transaction.Execution) {
	o.metadataMu.Lock()
	o.plan = e.Plan
	o.metadataMu.Unlock()
	o.status = e.Status
	o.authorization = e.Authorization
	o.bundle = e.Bundle
	o.data = e.Data
	o.transactionHash = e.TransactionHash
	o.workerID = e.WorkerID
}
func (s *WalletSession) coreContext() primitives.ProtocolContext { return s.runtime.CoreContext() }
func (s *WalletSession) canonical(ctx context.Context, b protocolrpc.CanonicalBlock) error {
	return s.runtimeSnapshot().Canonical(ctx, b)
}
func (s *WalletSession) preparationInventory(ctx context.Context) (WalletState, error) {
	return s.options.Store.ReadSnapshot(ctx, s.scope)
}
func frozen[T any](v T) (T, error) { return transaction.Frozen(v) }

func (s *WalletSession) runtimeSnapshot() *transaction.Runtime {
	var runtime transaction.Runtime
	if s.runtime != nil {
		runtime = *s.runtime
	}
	runtime.Options.InvisiOptions = s.options.InvisiOptions
	runtime.Options.Signer = s.options.Signer
	runtime.Options.FundingSigner = s.options.FundingSigner
	runtime.Wallet = s.wallet
	runtime.Protocol = s.protocol
	runtime.Scope = s.scope
	runtime.RPC = s.rpc
	// Historical readers belong to one preparation, never to signing or submission.
	runtime.Snapshot = nil
	return &runtime
}
