package session

import (
	"context"
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

type Options struct {
	Invisi        *sdk.Invisi
	Wallet        sdk.WalletReference
	Protocol      *protocol.Protocol
	Store         WalletStore
	Namespace     string
	Viewing       *sdk.ViewingCapability
	Authorization sdk.WalletAuthorizationProvider
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
	history       *sessionNotifications
	refreshGate   chan struct{}
	sends         *sessionSendState
	options       sessionOptions
	wallet        *sdk.Wallet
	protocol      *protocol.Protocol
	scope         WalletScope
	viewing       *sdk.ViewingCapability
	authorization sdk.WalletAuthorizationProvider
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
	runtime, e := transaction.ResolveRuntime(ctx, options.Invisi.GetInfrastructure(), options.Wallet, options.Protocol, options.Namespace)
	if e != nil {
		return nil, e
	}
	viewer := options.Viewing
	if viewer == nil {
		viewer, _ = runtime.Wallet.GetViewer()
	}
	if viewer != nil && transaction.ViewingOwner(viewer) != runtime.Wallet.OwnerAddress() {
		return nil, errors.New("viewing identity mismatch")
	}
	authority := options.Authorization
	if authority == nil {
		authority, _ = runtime.Wallet.ExportAuthorization()
	}
	if authority != nil && authority.OwnerAddress() != runtime.Wallet.OwnerAddress() {
		return nil, errors.New("authorization identity mismatch")
	}
	s := &WalletSession{history: newSessionNotifications(), refreshGate: make(chan struct{}, 1), sends: &sessionSendState{sending: map[string]bool{}}, options: sessionOptions{options.Invisi.GetInfrastructure(), options.Store, options.Signer, options.FundingSigner}, wallet: runtime.Wallet, protocol: options.Protocol, scope: runtime.Scope, viewing: viewer, authorization: authority, rpc: runtime.RPC, runtime: runtime}
	s.options.Store = &notifyingWalletStore{base: s.options.Store, notifications: s.history}
	runtime.Options.Signer = options.Signer
	runtime.Options.FundingSigner = options.FundingSigner
	runtime.Authorization = authority
	runtime.InventoryReader = func(ctx context.Context) (transaction.Inventory, error) {
		state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
		return transaction.Inventory{Notes: state.Notes, Credits: state.Credits, Invalidated: state.Invalidated, ObservedThrough: state.ObservedThrough, CheckedThrough: state.CheckedThrough}, e
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
	mu              sync.Mutex
	metadataMu      sync.RWMutex
	session         *WalletSession
	plan            transaction.Plan
	status          string
	authorization   *core.WalletAuthorizationWitness
	bundle          *prover.ProofBundle
	data            []byte
	transactionHash common.Hash
	workerID        string
}

func operationFromExecution(s *WalletSession, e *transaction.Execution) *Operation {
	return &Operation{session: s, plan: e.Plan, status: e.Status, authorization: e.Authorization, bundle: e.Bundle, data: e.Data, transactionHash: e.TransactionHash, workerID: e.WorkerID}
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
func (s *WalletSession) coreContext() core.ProtocolContext { return s.runtime.CoreContext() }
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
	runtime.Authorization = s.authorization
	runtime.Wallet = s.wallet
	runtime.Protocol = s.protocol
	runtime.Scope = s.scope
	runtime.RPC = s.rpc
	return &runtime
}
