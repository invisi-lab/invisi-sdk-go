package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	"math/big"
	"reflect"
	"sort"
	"sync"
)

type preparationCapture struct {
	mu    sync.Mutex
	paths map[common.Hash]protocolrpc.Membership
}

func (c *preparationCapture) record(m protocolrpc.Membership) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.paths) >= 4096 {
		for key := range c.paths {
			delete(c.paths, key)
			break
		}
	}
	m.Path = append([]common.Hash(nil), m.Path...)
	c.paths[m.Leaf] = m
}
func (c *preparationCapture) get(k common.Hash) (protocolrpc.Membership, bool) {
	if c == nil {
		return protocolrpc.Membership{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.paths[k]
	m.Path = append([]common.Hash(nil), m.Path...)
	return m, ok
}

func (s *WalletSession) cachedPreparation(ctx context.Context) (*transaction.PreparationSnapshot, error) {
	state, e := s.metadataSnapshot(ctx)
	if e != nil {
		return nil, e
	}
	if state.Invalidated {
		return nil, errors.New("wallet requires reset scan")
	}
	snap := state.PreparationSnapshot
	if snap == nil {
		return nil, nil
	}
	if !snap.Match(s.scope, s.protocol.Deployment()) {
		return nil, errors.New("preparation snapshot scope or deployment mismatch")
	}
	if snap.Generation != state.Scan.Generation || !reflect.DeepEqual(snap.ObservedThrough, state.ObservedThrough) || !reflect.DeepEqual(snap.CheckedThrough, state.CheckedThrough) {
		return nil, nil
	}
	if s.publicHistory != nil {
		public, e := s.publicHistory.Snapshot(ctx)
		if e != nil {
			return nil, e
		}
		if public.Generation != snap.Generation {
			return nil, errors.New("wallet requires reset scan")
		}
	}
	return snap.Clone()
}
func (s *WalletSession) preparationRuntime(ctx context.Context) (*transaction.Runtime, error) {
	r := s.runtimeSnapshot()
	snap, e := s.cachedPreparation(ctx)
	if e != nil {
		return nil, e
	}
	r.Snapshot = snap
	return r, nil
}
func (s *WalletSession) prepareWithSnapshot(ctx context.Context, work func(*transaction.Runtime) (*transaction.Execution, error)) (*Operation, error) {
	r, e := s.preparationRuntime(ctx)
	if e != nil {
		return nil, e
	}
	execution, e := work(r)
	if errors.Is(e, transaction.ErrPreparationSnapshotMissing) {
		execution, e = work(s.runtimeSnapshot())
	}
	if e != nil {
		return nil, e
	}
	return operationFromExecution(s, execution), nil
}
func newPreparationSnapshot(s *WalletSession, v WalletState, b protocolrpc.CanonicalBlock, h *types.Header) *transaction.PreparationSnapshot {
	return &transaction.PreparationSnapshot{Version: 1, Scope: s.scope, Deployment: s.protocol.Deployment(), Block: b, Generation: v.Scan.Generation, ObservedThrough: v.ObservedThrough, CheckedThrough: &b, Headers: map[uint64]*types.Header{b.Number: types.CopyHeader(h)}, Code: map[common.Address][]byte{}, Calls: map[string][]byte{}, Memberships: map[common.Hash]protocolrpc.Membership{}}
}

// Consumption and bitmap reads are captured by the state refresh itself. Only missing immutable paths are fetched.
func (s *WalletSession) finishPreparationSnapshot(ctx context.Context, before *WalletState, r *transaction.SnapshotRecorder) error {
	snap := r.Snapshot
	b := snap.Block
	if e := s.protocol.ValidateDeployment(ctx, r); e != nil {
		return e
	}
	for _, point := range []*protocolrpc.CanonicalBlock{before.ObservedThrough, before.CheckedThrough} {
		if point != nil {
			h, e := r.HeaderByNumber(ctx, new(big.Int).SetUint64(point.Number))
			if e != nil {
				return e
			}
			if h == nil || h.Hash() != point.Hash {
				return errors.New("preparation observation changed")
			}
		}
	}
	type asset struct {
		commitment common.Hash
		tree       uint64
		index      uint32
		token      common.Address
	}
	assets := []asset{}
	seen := map[common.Hash]bool{}
	tokens := map[common.Address]bool{}
	owners := map[common.Address]map[common.Address]bool{}
	for _, n := range before.Notes {
		if n.Spent {
			continue
		}
		token := common.BigToAddress(primitives.ElementToBigInt(n.Note.TokenAddress))
		if !seen[n.Commitment] {
			assets = append(assets, asset{n.Commitment, n.TreeID, n.LeafIndex, token})
			seen[n.Commitment] = true
		}
		tokens[token] = true
	}
	for _, g := range before.PostpaidCredit.Grants {
		if g.Consumed || g.Revoked {
			continue
		}
		n := g.Delivery.State
		token := common.BigToAddress(primitives.ElementToBigInt(n.TokenAddress))
		if !seen[g.Commitment] {
			assets = append(assets, asset{g.Commitment, g.TreeID, uint32(g.LeafIndex), token})
			seen[g.Commitment] = true
		}
		tokens[token] = true
		if owners[token] == nil {
			owners[token] = map[common.Address]bool{}
		}
		owners[token][common.BigToAddress(primitives.ElementToBigInt(n.OwnerWalletIdentity))] = true
	}
	caller, e := contracts.NewInvisiProtocolCaller(s.scope.ProtocolAddress, r)
	if e != nil {
		return e
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(b.Number)}
	for token := range tokens {
		code, e := r.CodeAt(ctx, token, opts.BlockNumber)
		if e != nil {
			return e
		}
		if len(code) == 0 {
			return errors.New("token has no code")
		}
		if _, e = caller.Nonces(opts, s.scope.OwnerAddress); e != nil {
			return e
		}
		if owners[token] == nil {
			owners[token] = map[common.Address]bool{}
		}
		owners[token][s.scope.OwnerAddress] = true
		for owner := range owners[token] {
			if _, e = caller.CreditAllocation(opts, owner, token); e != nil {
				return e
			}
		}
	}
	// The highest leaf must choose the root first: an older root cannot include newer leaves.
	sort.Slice(assets, func(i, j int) bool {
		if assets[i].tree != assets[j].tree {
			return assets[i].tree < assets[j].tree
		}
		return assets[i].index > assets[j].index
	})
	roots := map[uint64]common.Hash{}
	var mu sync.Mutex
	work := func(ctx context.Context, i int) error {
		a := assets[i]
		m, ok := s.publicHistory.preparationCapture.get(a.commitment)
		if !ok && before.PreparationSnapshot != nil {
			m, ok = before.PreparationSnapshot.Memberships[a.commitment]
		}
		mu.Lock()
		root, pinned := roots[a.tree]
		mu.Unlock()
		if ok && m.TreeID == a.tree && m.LeafIndex == a.index && (!pinned || root == m.Root) {
			if m.Block == b {
				ok = true
			} else {
				known, e := caller.RootHistory(&bind.CallOpts{Context: ctx, BlockNumber: opts.BlockNumber}, new(big.Int).SetUint64(a.tree), m.Root)
				if e != nil {
					return e
				}
				ok = known
			}
		} else {
			ok = false
		}
		if !ok {
			input := protocolrpc.MembershipInput{TreeID: a.tree, LeafIndex: a.index, Commitment: a.commitment}
			if pinned {
				input.Root = &root
			}
			var e error
			m, e = s.rpc.GetMembership(ctx, s.protocol.Context(), input, b)
			if e != nil {
				return e
			}

		}
		abi, _ := protocol.ProtocolABI()
		method := abi.Methods["rootHistory"]
		data, e := abi.Pack("rootHistory", new(big.Int).SetUint64(a.tree), m.Root)
		if e != nil {
			return e
		}
		value, e := method.Outputs.Pack(true)
		if e != nil {
			return e
		}
		r.RecordCall(ethereum.CallMsg{To: &s.scope.ProtocolAddress, Data: data}, value)
		if pinned && m.Root != root {
			return errors.New("cached membership root conflict")
		}
		m.Block = b
		local := *snap
		local.Memberships = map[common.Hash]protocolrpc.Membership{a.commitment: m}
		if _, e := local.Membership(protocolrpc.MembershipInput{TreeID: a.tree, LeafIndex: a.index, Commitment: a.commitment}, b); e != nil {
			return e
		}
		mu.Lock()
		snap.Memberships[a.commitment] = m
		roots[a.tree] = m.Root
		mu.Unlock()
		return nil
	}
	first := map[uint64]int{}
	leaders, rest := []int{}, []int{}
	for i, a := range assets {
		if _, ok := first[a.tree]; ok {
			rest = append(rest, i)
		} else {
			first[a.tree] = i
			leaders = append(leaders, i)
		}
	}
	for _, batch := range [][]int{leaders, rest} {
		if _, e := transaction.MapPreparation(ctx, len(batch), func(ctx context.Context, j int) (struct{}, error) { return struct{}{}, work(ctx, batch[j]) }); e != nil {
			return e
		}
	}
	if _, e = r.ChainID(ctx); e != nil {
		return e
	}
	h, e := s.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(b.Number))
	if e != nil {
		return e
	}
	if h == nil || h.Hash() != b.Hash {
		return errors.New("preparation canonical anchor changed")
	}
	snap.ObservedThrough = before.ObservedThrough
	snap.CheckedThrough = before.CheckedThrough
	copy, e := snap.Clone()
	if e != nil {
		return e
	}
	before.PreparationSnapshot = copy
	return nil
}
func cachedScope(options Options) transaction.WalletScope {
	namespace := options.Namespace
	if namespace == "" {
		namespace = "default"
	}
	return transaction.WalletScope{Namespace: namespace, ChainID: options.Protocol.Context().ChainID, ProtocolAddress: options.Protocol.Context().ProtocolAddress, OwnerAddress: options.Wallet.OwnerAddress()}
}
func matchingSnapshot(state WalletState) *transaction.PreparationSnapshot {
	p := state.PreparationSnapshot
	if state.Invalidated || p == nil || p.Generation != state.Scan.Generation || !reflect.DeepEqual(p.ObservedThrough, state.ObservedThrough) || !reflect.DeepEqual(p.CheckedThrough, state.CheckedThrough) {
		return nil
	}
	return p
}
