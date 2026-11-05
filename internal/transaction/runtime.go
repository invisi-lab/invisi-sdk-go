package transaction

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type Inventory struct {
	Notes                           map[common.Hash]NoteRecord
	Credits                         map[string]CreditRecord
	Invalidated                     bool
	ObservedThrough, CheckedThrough *protocolrpc.CanonicalBlock
}

func EmptyInventory() Inventory {
	return Inventory{Notes: map[common.Hash]NoteRecord{}, Credits: map[string]CreditRecord{}}
}

type RuntimeOptions struct {
	InvisiOptions
	FundingSigner FundingSigner
	Signer        *bind.TransactOpts
}
type Runtime struct {
	Options           RuntimeOptions
	Wallet            *Wallet
	Protocol          *protocol.Protocol
	Scope             WalletScope
	RPC               *protocolrpc.ProtocolRpcClient
	Authorization     WalletAuthorizationProvider
	ExplicitInventory *Inventory
	ChangeAddress     string
	InventoryReader   func(context.Context) (Inventory, error)
	OnComplete        func(context.Context, *Execution) error
	OnPersist         func(context.Context, *Execution, string) error
	OnLease           func(context.Context, *Execution) error
	OnInvalidate      func(context.Context) error
	OnReprice         func(context.Context, *Execution) error
	OnExternalize     func(context.Context, *Execution) error
}

func (i *Invisi) GetInfrastructure() InvisiOptions {
	o := i.options
	o.GasBenchmarks = CopyGasProfiles(o.GasBenchmarks)
	return o
}
func SnapshotWallet(ref WalletReference) (*Wallet, error) {
	if ref == nil {
		return nil, &MissingCapabilityError{Capability: "wallet"}
	}
	if w, ok := ref.(*Wallet); ok && w == nil {
		return nil, &MissingCapabilityError{Capability: "wallet"}
	}
	p := ref.publicWallet()
	if p == nil {
		return nil, &MissingCapabilityError{Capability: "public wallet"}
	}
	w := &Wallet{PublicWallet: &PublicWallet{identity: p.identity, address: p.address, keys: core.ReceivingKeys{ViewingPublicKey: append([]byte(nil), p.keys.ViewingPublicKey...), ClueKey: append([]byte(nil), p.keys.ClueKey...)}}}
	if local, ok := ref.(*Wallet); ok {
		w.material = local.material
	}
	return w, nil
}
func ResolveRuntime(ctx context.Context, infrastructure InvisiOptions, wallet WalletReference, pool *protocol.Protocol, namespace string) (*Runtime, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if pool == nil || infrastructure.Provider == nil || infrastructure.ProtocolRPC == nil {
		return nil, &MissingCapabilityError{Capability: "wallet, protocol, Chain RPC and Protocol RPC"}
	}
	w, e := SnapshotWallet(wallet)
	if e != nil {
		return nil, e
	}
	if e = pool.ValidateDeployment(ctx, infrastructure.Provider); e != nil {
		return nil, e
	}
	scope := WalletScope{ChainID: pool.Context().ChainID, ProtocolAddress: pool.Context().ProtocolAddress, OwnerAddress: w.identity, Namespace: namespace}
	if scope.Namespace == "" {
		scope.Namespace = "default"
	}
	deployment, e := infrastructure.ProtocolRPC.GetDeployment(ctx, scope.ChainID)
	if e != nil {
		return nil, e
	}
	expected, _ := json.Marshal(pool.Deployment())
	actual, _ := json.Marshal(deployment)
	if string(expected) != string(actual) {
		return nil, errors.New("Protocol RPC deployment mismatch")
	}
	return &Runtime{Options: RuntimeOptions{InvisiOptions: infrastructure}, Wallet: w, Protocol: pool, Scope: scope, RPC: infrastructure.ProtocolRPC.WithProvider(infrastructure.Provider)}, nil
}
func (r *Runtime) Invalidate(ctx context.Context) error {
	if r.OnInvalidate != nil {
		return r.OnInvalidate(ctx)
	}
	return ctx.Err()
}
func (e *Execution) Persist(ctx context.Context, status string) error {
	if e.Runtime.OnPersist != nil {
		return e.Runtime.OnPersist(ctx, e, status)
	}
	return ctx.Err()
}
func (e *Execution) CheckLease(ctx context.Context) error {
	if e.Runtime.OnLease != nil {
		return e.Runtime.OnLease(ctx, e)
	}
	return ctx.Err()
}
func (e *Execution) CommitBroadcasterPlan(ctx context.Context) error {
	if e.Runtime.OnReprice != nil {
		return e.Runtime.OnReprice(ctx, e)
	}
	return ctx.Err()
}

func ViewingOwner(v *ViewingCapability) common.Address { return v.identity }
func PrivateMaterial(w *Wallet) *core.InvisiWallet     { return w.material }
