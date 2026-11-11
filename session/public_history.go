package session

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/protocol/receipts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/chain"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
	"sort"
	"sync"
)

var ErrPublicHistoryReorg = errors.New("public history reorg; account reset required")

type PublicRetention struct{ MaxBytes, MaxPages, MaxPageBytes int }

func (r PublicRetention) Defaults() PublicRetention {
	if r.MaxBytes == 0 {
		r.MaxBytes = 128 * 1024 * 1024
	}
	if r.MaxPages == 0 {
		r.MaxPages = 4096
	}
	if r.MaxPageBytes == 0 {
		r.MaxPageBytes = 8 * 1024 * 1024
	}
	return r
}

type PublicTransaction struct {
	PositionValidated bool
	Receipt           *types.Receipt
	Header            *types.Header
	Batch             receipts.Batch
}
type PublicPage struct {
	Range        BlockRange
	Block        protocolrpc.CanonicalBlock
	Headers      []protocolrpc.CanonicalBlock
	Transactions []PublicTransaction
}
type PublicPageReference struct {
	Range  BlockRange
	Digest [32]byte
	Bytes  int
	Data   json.RawMessage `json:"data,omitempty"`
}
type PublicHistoryState struct {
	KnownHead, IndexedHead *protocolrpc.CanonicalBlock
	Revision, Generation   uint64
	Coverage               []CoverageRange
	Pages                  []PublicPageReference
}
type PublicHistoryStore interface {
	Read(context.Context, string) (PublicHistoryState, error)
	Update(context.Context, string, func(*PublicHistoryState) error) error
	Retention() PublicRetention
}
type PublicHistoryPayloadStore interface {
	ReadPayload(context.Context, string, PublicPageReference) (json.RawMessage, error)
}
type MemoryPublicHistoryStore struct {
	mu        sync.Mutex
	states    map[string]PublicHistoryState
	payloads  map[string]map[string]json.RawMessage
	retention PublicRetention
}

func NewMemoryPublicHistoryStore(r PublicRetention) *MemoryPublicHistoryStore {
	return &MemoryPublicHistoryStore{states: map[string]PublicHistoryState{}, payloads: map[string]map[string]json.RawMessage{}, retention: r.Defaults()}
}
func (s *MemoryPublicHistoryStore) Retention() PublicRetention { return s.retention }
func ClonePublicHistory(v PublicHistoryState) (PublicHistoryState, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return PublicHistoryState{}, e
	}
	var out PublicHistoryState
	e = json.Unmarshal(b, &out)
	return out, e
}
func (s *MemoryPublicHistoryStore) Read(ctx context.Context, key string) (PublicHistoryState, error) {
	if e := ctx.Err(); e != nil {
		return PublicHistoryState{}, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return ClonePublicHistory(s.states[key])
}
func publicPageKey(q BlockRange) string { return fmt.Sprintf("%d:%d", q.FromBlock, q.ToBlock) }
func (s *MemoryPublicHistoryStore) ReadPayload(ctx context.Context, key string, ref PublicPageReference) (json.RawMessage, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	raw := s.payloads[key][publicPageKey(ref.Range)]
	if sha256.Sum256(raw) != ref.Digest {
		return nil, errors.New("public page missing or corrupt")
	}
	return append(json.RawMessage(nil), raw...), nil
}
func (s *MemoryPublicHistoryStore) Update(ctx context.Context, key string, fn func(*PublicHistoryState) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, e := ClonePublicHistory(s.states[key])
	if e != nil {
		return e
	}
	if e = fn(&v); e != nil {
		return e
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	v.Revision++
	payloads := map[string]json.RawMessage{}
	for i := range v.Pages {
		ref := &v.Pages[i]
		pkey := publicPageKey(ref.Range)
		if len(ref.Data) > 0 {
			payloads[pkey] = append(json.RawMessage(nil), ref.Data...)
		} else {
			payloads[pkey] = s.payloads[key][pkey]
		}
		ref.Data = nil
	}
	s.states[key], e = ClonePublicHistory(v)
	if e == nil {
		s.payloads[key] = payloads
	}
	return e
}

func PutPublicPage(ctx context.Context, s PublicHistoryStore, key string, p PublicPage, generation uint64) error {
	raw, e := json.Marshal(p)
	if e != nil {
		return e
	}
	r := s.Retention().Defaults()
	if len(raw) > r.MaxPageBytes {
		return errors.New("public page exceeds byte limit; use a smaller range")
	}
	return s.Update(ctx, key, func(v *PublicHistoryState) error {
		if v.Generation != generation {
			return errors.New("public history generation changed")
		}
		var e error
		v.Coverage, e = MergeCoverage(v.Coverage, CoverageRange{p.Range.FromBlock, p.Block})
		if e != nil {
			return e
		}
		refs := v.Pages[:0]
		for _, ref := range v.Pages {
			if ref.Range != p.Range {
				refs = append(refs, ref)
			}
		}
		v.Pages = append(refs, PublicPageReference{p.Range, sha256.Sum256(raw), len(raw), raw})
		sort.Slice(v.Pages, func(i, j int) bool { return v.Pages[i].Range.ToBlock > v.Pages[j].Range.ToBlock })
		bytes := 0
		keep := 0
		for _, ref := range v.Pages {
			if keep == r.MaxPages || bytes+ref.Bytes > r.MaxBytes {
				break
			}
			bytes += ref.Bytes
			keep++
		}
		v.Pages = v.Pages[:keep]
		return nil
	})
}
func retainedPublicPage(v PublicHistoryState, q BlockRange) (PublicPage, bool) {
	result := PublicPage{Range: q}
	cursor := q.FromBlock
	for cursor <= q.ToBlock {
		found := false
		for _, r := range v.Pages {
			if r.Range.FromBlock > cursor || r.Range.ToBlock < cursor || r.Digest != sha256.Sum256(r.Data) {
				continue
			}
			var p PublicPage
			if json.Unmarshal(r.Data, &p) != nil {
				continue
			}
			end := r.Range.ToBlock
			if end > q.ToBlock {
				end = q.ToBlock
			}
			for n := cursor; ; n++ {
				index := n - r.Range.FromBlock
				if index >= uint64(len(p.Headers)) {
					break
				}
				result.Headers = append(result.Headers, p.Headers[index])
				result.Block = p.Headers[index]
				if n == end {
					break
				}
			}
			if result.Block.Number != end {
				continue
			}
			for _, tx := range p.Transactions {
				if tx.Receipt != nil && tx.Receipt.BlockNumber != nil && tx.Receipt.BlockNumber.Uint64() >= cursor && tx.Receipt.BlockNumber.Uint64() <= end {
					result.Transactions = append(result.Transactions, tx)
				}
			}
			if end == q.ToBlock {
				return result, true
			}
			cursor = end + 1
			found = true
			break
		}
		if !found {
			return PublicPage{}, false
		}
	}
	return PublicPage{}, false
}

type PublicHistoryOptions struct {
	Store       PublicHistoryStore
	Protocol    *chain.Protocol
	Provider    sdk.Provider
	ProtocolRPC *protocolrpc.ProtocolRpcClient
}

var publicHistoryGates sync.Map

func publicGate(store PublicHistoryStore) chan struct{} {
	gate := make(chan struct{}, 1)
	actual, _ := publicHistoryGates.LoadOrStore(store, gate)
	return actual.(chan struct{})
}

type PublicHistory struct {
	options PublicHistoryOptions
	key     string
	gate    chan struct{}
}

func NewPublicHistory(o PublicHistoryOptions) (*PublicHistory, error) {
	if o.Protocol == nil || o.Provider == nil || o.ProtocolRPC == nil {
		return nil, errors.New("public history connections required")
	}
	if o.Store == nil {
		o.Store = NewMemoryPublicHistoryStore(PublicRetention{})
	}
	d := o.Protocol.Deployment()
	key := fmt.Sprintf("%d:%s:%s:%d", o.Protocol.Context().ChainID, d.Protocol.Address.Hex(), d.Protocol.RuntimeCodeHash.Hex(), d.Protocol.FromBlock)
	return &PublicHistory{o, key, publicGate(o.Store)}, nil
}
func (h *PublicHistory) Store() PublicHistoryStore { return h.options.Store }
func (h *PublicHistory) Key() string               { return h.key }
func (h *PublicHistory) Snapshot(ctx context.Context) (PublicHistoryState, error) {
	return h.options.Store.Read(ctx, h.key)
}
func (h *PublicHistory) PruneBefore(ctx context.Context, block uint64) error {
	return h.options.Store.Update(ctx, h.key, func(v *PublicHistoryState) error {
		refs := v.Pages[:0]
		for _, r := range v.Pages {
			if r.Range.ToBlock >= block {
				refs = append(refs, r)
			}
		}
		v.Pages = refs
		return nil
	})
}
func (h *PublicHistory) checkCanonical(ctx context.Context, v PublicHistoryState) error {
	for _, r := range v.Coverage {
		head, e := h.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(r.Block.Number))
		if e != nil {
			return e
		}
		if head == nil || head.Hash() != r.Block.Hash {
			e = h.options.Store.Update(ctx, h.key, func(s *PublicHistoryState) error {
				if s.Generation != v.Generation {
					return errors.New("public history generation changed")
				}
				s.Generation++
				s.Coverage = nil
				s.Pages = nil
				return nil
			})
			if e != nil {
				return e
			}
			return ErrPublicHistoryReorg
		}
	}
	return nil
}
func (h *PublicHistory) Page(ctx context.Context, q BlockRange, offline bool) (PublicPage, error) {
	// Downloads continue for other waiters after an individual subscriber cancels.
	done := make(chan struct{})
	var result PublicPage
	var err error
	go func() {
		defer close(done)
		h.gate <- struct{}{}
		defer func() { <-h.gate }()
		result, err = h.page(context.WithoutCancel(ctx), q, offline)
	}()
	select {
	case <-ctx.Done():
		return PublicPage{}, ctx.Err()
	case <-done:
		return result, err
	}
}
func (h *PublicHistory) page(ctx context.Context, q BlockRange, offline bool) (PublicPage, error) {
	v, e := h.Snapshot(ctx)
	if e != nil {
		return PublicPage{}, e
	}
	if !offline {
		if e = h.checkCanonical(ctx, v); e != nil {
			return PublicPage{}, e
		}
	}
	loaded := v
	if payloadStore, ok := h.options.Store.(PublicHistoryPayloadStore); ok {
		loaded.Pages = append([]PublicPageReference(nil), v.Pages...)
		for i := range loaded.Pages {
			ref := &loaded.Pages[i]
			if ref.Range.ToBlock >= q.FromBlock && ref.Range.FromBlock <= q.ToBlock {
				raw, err := payloadStore.ReadPayload(ctx, h.key, *ref)
				if err == nil {
					ref.Data = raw
				}
			}
		}
	}
	if p, ok := retainedPublicPage(loaded, q); ok {
		return p, nil
	}
	if offline {
		return PublicPage{}, errors.New("public history page not retained")
	}
	if q.FromBlock > q.ToBlock || q.ToBlock-q.FromBlock >= 256 {
		return PublicPage{}, errors.New("public page must contain 1..256 blocks")
	}
	p := PublicPage{Range: q}
	for n := q.FromBlock; ; n++ {
		head, e := h.options.Provider.HeaderByNumber(ctx, new(big.Int).SetUint64(n))
		if e != nil {
			return p, e
		}
		if head == nil || head.Number == nil || head.Number.Uint64() != n {
			return p, errors.New("invalid public header")
		}
		b := protocolrpc.CanonicalBlock{Number: n, Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
		if len(p.Headers) > 0 && b.ParentHash != p.Headers[len(p.Headers)-1].Hash {
			return p, errors.New("public header chain mismatch")
		}
		p.Headers = append(p.Headers, b)
		p.Block = b
		if n == q.ToBlock {
			break
		}
	}
	query := protocolrpc.ProtocolUpdateRange{FromBlock: q.FromBlock, ToBlock: q.ToBlock}
	ns, e := h.options.ProtocolRPC.GetNoteUpdates(ctx, h.options.Protocol.Context(), query)
	if e != nil {
		return p, e
	}
	cs, e := h.options.ProtocolRPC.GetCreditTransitions(ctx, h.options.Protocol.Context(), query)
	if e != nil {
		return p, e
	}
	if ns.Block != p.Block || cs.Block != p.Block {
		return p, errors.New("public discovery anchor mismatch")
	}
	hashes := map[common.Hash]bool{}
	for _, n := range ns.Notes {
		hashes[n.TransactionHash] = true
	}
	for _, c := range cs.Transitions {
		hashes[c.TransactionHash] = true
	}
	decoder, e := receipts.NewDecoder()
	if e != nil {
		return p, e
	}
	pageBytes := 0
	for hash := range hashes {
		r, e := h.options.Provider.TransactionReceipt(ctx, hash)
		if e != nil {
			return p, e
		}
		if r == nil || r.Status != types.ReceiptStatusSuccessful || r.TxHash != hash || r.BlockNumber == nil || !r.BlockNumber.IsUint64() || r.BlockNumber.Uint64() < q.FromBlock || r.BlockNumber.Uint64() > q.ToBlock {
			return p, errors.New("invalid public receipt")
		}
		b := p.Headers[r.BlockNumber.Uint64()-q.FromBlock]
		if r.BlockHash != b.Hash {
			return p, errors.New("public receipt not canonical")
		}
		head, e := h.options.Provider.HeaderByNumber(ctx, r.BlockNumber)
		if e != nil {
			return p, e
		}
		logs := []types.Log{}
		for _, l := range r.Logs {
			if l == nil {
				return p, errors.New("nil public log")
			}
			if l.Address != h.options.Protocol.Context().ProtocolAddress {
				continue
			}
			if l.Removed || l.TxHash != hash || l.BlockHash != b.Hash || l.BlockNumber != b.Number {
				return p, errors.New("invalid public log binding")
			}
			logs = append(logs, *l)
		}
		batch, e := decoder.Decode(receipts.Block{Number: b.Number, Hash: b.Hash, ParentHash: b.ParentHash, Timestamp: b.Timestamp}, logs)
		if e != nil {
			return p, e
		}
		for _, o := range batch.Outputs {
			if _, e = h.options.ProtocolRPC.GetMembership(ctx, h.options.Protocol.Context(), protocolrpc.MembershipInput{TreeID: o.TreeID, LeafIndex: o.LeafIndex, Commitment: o.Commitment}, p.Block); e != nil {
				return p, e
			}
		}
		for _, o := range batch.CreditStates {
			if _, e = h.options.ProtocolRPC.GetMembership(ctx, h.options.Protocol.Context(), protocolrpc.MembershipInput{TreeID: o.TreeID, LeafIndex: o.LeafIndex, Commitment: o.Commitment}, p.Block); e != nil {
				return p, e
			}
		}
		tx := PublicTransaction{PositionValidated: true, Receipt: r, Header: head, Batch: batch}
		encoded, err := json.Marshal(tx)
		if err != nil {
			return p, err
		}
		pageBytes += len(encoded)
		if pageBytes > h.options.Store.Retention().Defaults().MaxPageBytes {
			return p, errors.New("public page exceeds byte limit; use a smaller range")
		}
		p.Transactions = append(p.Transactions, PublicTransaction{PositionValidated: true, Receipt: r, Header: head, Batch: batch})
	}
	sort.Slice(p.Transactions, func(i, j int) bool {
		a, b := p.Transactions[i].Receipt, p.Transactions[j].Receipt
		if a.BlockNumber.Cmp(b.BlockNumber) != 0 {
			return a.BlockNumber.Cmp(b.BlockNumber) < 0
		}
		return a.TransactionIndex < b.TransactionIndex
	})
	if e = h.checkCanonical(ctx, PublicHistoryState{Generation: v.Generation, Coverage: []CoverageRange{{p.Range.FromBlock, p.Block}}}); e != nil {
		return p, e
	}
	e = PutPublicPage(ctx, h.options.Store, h.key, p, v.Generation)
	return p, e
}
