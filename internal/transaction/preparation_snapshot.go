package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"math/big"
	"sync"
)

var ErrPreparationSnapshotMissing = errors.New("synchronized preparation data unavailable")

// PreparationSnapshot is historical evidence, never permission to submit a transaction.
type PreparationSnapshot struct {
	Version                         int
	Scope                           WalletScope
	Deployment                      protocolrpc.ProtocolDeployment
	Block                           protocolrpc.CanonicalBlock
	Generation                      uint64
	ObservedThrough, CheckedThrough *protocolrpc.CanonicalBlock
	Headers                         map[uint64]*types.Header
	Code                            map[common.Address][]byte
	Calls                           map[string][]byte
	Memberships                     map[common.Hash]protocolrpc.Membership
}

func (s *PreparationSnapshot) Clone() (*PreparationSnapshot, error) {
	if s == nil {
		return nil, nil
	}
	copy, e := Frozen(*s)
	return &copy, e
}
func SnapshotCallKey(m ethereum.CallMsg) string { raw, _ := json.Marshal(m); return string(raw) }
func (s *PreparationSnapshot) Match(scope WalletScope, d protocolrpc.ProtocolDeployment) bool {
	if s == nil || s.Version != 1 || s.Scope != scope {
		return false
	}
	h := s.Headers[s.Block.Number]
	if h == nil || h.Number == nil || !h.Number.IsUint64() || h.Number.Uint64() != s.Block.Number || h.Hash() != s.Block.Hash || h.ParentHash != s.Block.ParentHash || h.Time != s.Block.Timestamp {
		return false
	}
	a, _ := json.Marshal(d)
	b, _ := json.Marshal(s.Deployment)
	return string(a) == string(b)
}
func (s *PreparationSnapshot) Membership(input protocolrpc.MembershipInput, block protocolrpc.CanonicalBlock) (protocolrpc.Membership, error) {
	m, ok := s.Memberships[input.Commitment]
	if !ok || m.TreeID != input.TreeID || m.LeafIndex != input.LeafIndex || m.Block != block || (input.Root != nil && *input.Root != m.Root) {
		return m, ErrPreparationSnapshotMissing
	}
	current, e := primitives.HexToElement(input.Commitment.Hex())
	if e != nil {
		return m, e
	}
	if len(m.Path) != primitives.TreeDepth || m.Leaf != input.Commitment {
		return m, errors.New("invalid cached membership binding")
	}
	for i, h := range m.Path {
		sibling, e := primitives.HexToElement(h.Hex())
		if e != nil {
			return m, e
		}
		if input.LeafIndex>>i&1 == 0 {
			current, e = primitives.Poseidon2Hash(current, sibling)
		} else {
			current, e = primitives.Poseidon2Hash(sibling, current)
		}
		if e != nil {
			return m, e
		}
	}
	if FieldHash(current) != m.Root {
		return m, errors.New("invalid cached membership path")
	}
	return Frozen(m)
}
func (s *PreparationSnapshot) Provider(live Provider) Provider {
	return &snapshotProvider{Provider: live, snapshot: s}
}

type snapshotProvider struct {
	Provider
	snapshot *PreparationSnapshot
}

func (p *snapshotProvider) ChainID(ctx context.Context) (*big.Int, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	return new(big.Int).SetUint64(p.snapshot.Scope.ChainID), nil
}
func (p *snapshotProvider) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	number := p.snapshot.Block.Number
	if n != nil {
		if !n.IsUint64() {
			return nil, ErrPreparationSnapshotMissing
		}
		number = n.Uint64()
	}
	h := p.snapshot.Headers[number]
	if h == nil {
		return nil, ErrPreparationSnapshotMissing
	}
	return types.CopyHeader(h), nil
}
func (p *snapshotProvider) CodeAt(ctx context.Context, a common.Address, n *big.Int) ([]byte, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if n != nil && (!n.IsUint64() || n.Uint64() != p.snapshot.Block.Number) {
		return nil, ErrPreparationSnapshotMissing
	}
	v, ok := p.snapshot.Code[a]
	if !ok {
		return nil, ErrPreparationSnapshotMissing
	}
	return append([]byte(nil), v...), nil
}
func (p *snapshotProvider) CallContract(ctx context.Context, m ethereum.CallMsg, n *big.Int) ([]byte, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if n != nil && (!n.IsUint64() || n.Uint64() != p.snapshot.Block.Number) {
		return nil, ErrPreparationSnapshotMissing
	}
	v, ok := p.snapshot.Calls[SnapshotCallKey(m)]
	if !ok {
		return nil, ErrPreparationSnapshotMissing
	}
	return append([]byte(nil), v...), nil
}
func (p *snapshotProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return nil, ErrPreparationSnapshotMissing
}

// SnapshotRecorder pins reads to one real header. It is private to a sync, not shared between accounts.
type SnapshotRecorder struct {
	Provider
	Snapshot *PreparationSnapshot
	mu       sync.Mutex
	reads    *preparationReads
}

func NewSnapshotRecorder(p Provider, snapshot *PreparationSnapshot) *SnapshotRecorder {
	return &SnapshotRecorder{Provider: p, Snapshot: snapshot, reads: newPreparationReads()}
}
func (r *SnapshotRecorder) ChainID(ctx context.Context) (*big.Int, error) {
	n, e := r.Provider.ChainID(ctx)
	if e == nil && (n == nil || !n.IsUint64() || n.Uint64() != r.Snapshot.Scope.ChainID) {
		return nil, errors.New("preparation chain identity changed")
	}
	return n, e
}
func (r *SnapshotRecorder) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	if n == nil {
		n = new(big.Int).SetUint64(r.Snapshot.Block.Number)
	}
	h, e := r.Provider.HeaderByNumber(ctx, n)
	if e == nil && h != nil && h.Number != nil {
		r.mu.Lock()
		r.Snapshot.Headers[h.Number.Uint64()] = types.CopyHeader(h)
		r.mu.Unlock()
	}
	return h, e
}
func (r *SnapshotRecorder) CodeAt(ctx context.Context, a common.Address, n *big.Int) ([]byte, error) {
	if n != nil && (!n.IsUint64() || n.Uint64() != r.Snapshot.Block.Number) {
		return nil, ErrPreparationSnapshotMissing
	}
	return r.reads.read(ctx, "code:"+a.Hex(), func() ([]byte, error) {
		r.mu.Lock()
		cached, ok := r.Snapshot.Code[a]
		r.mu.Unlock()
		if ok {
			return append([]byte(nil), cached...), nil
		}
		v, e := r.Provider.CodeAt(ctx, a, new(big.Int).SetUint64(r.Snapshot.Block.Number))
		if e == nil {
			r.mu.Lock()
			r.Snapshot.Code[a] = append([]byte(nil), v...)
			r.mu.Unlock()
		}
		return v, e
	})
}
func (r *SnapshotRecorder) CallContract(ctx context.Context, m ethereum.CallMsg, n *big.Int) ([]byte, error) {
	if n != nil && (!n.IsUint64() || n.Uint64() != r.Snapshot.Block.Number) {
		return nil, ErrPreparationSnapshotMissing
	}
	k := SnapshotCallKey(m)
	return r.reads.read(ctx, "call:"+k, func() ([]byte, error) {
		r.mu.Lock()
		v, ok := r.Snapshot.Calls[k]
		r.mu.Unlock()
		if ok {
			return append([]byte(nil), v...), nil
		}
		v, e := r.Provider.CallContract(ctx, m, new(big.Int).SetUint64(r.Snapshot.Block.Number))
		if e == nil {
			r.RecordCall(m, v)
		}
		return v, e
	})
}
func (r *SnapshotRecorder) RecordMembership(m protocolrpc.Membership) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m.Path = append([]common.Hash(nil), m.Path...)
	r.Snapshot.Memberships[m.Leaf] = m
}
func (r *SnapshotRecorder) String() string {
	return fmt.Sprintf("PreparationRecorder(%d)", r.Snapshot.Block.Number)
}

func (r *SnapshotRecorder) RecordCall(m ethereum.CallMsg, value []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Snapshot.Calls[SnapshotCallKey(m)] = append([]byte(nil), value...)
}
