package transaction

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type preparationRead struct {
	ready chan struct{}
	at    time.Time
	value []byte
	err   error
}
type preparationReads struct {
	mu      sync.Mutex
	entries map[string]*preparationRead
	now     func() time.Time
}

func newPreparationReads() *preparationReads {
	return &preparationReads{entries: make(map[string]*preparationRead), now: time.Now}
}
func (r *preparationReads) read(ctx context.Context, key string, load func() ([]byte, error)) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	now := r.now()
	for k, e := range r.entries {
		if now.Sub(e.at) >= 30*time.Second {
			delete(r.entries, k)
		}
	}
	if e := r.entries[key]; e != nil {
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-e.ready:
			return append([]byte(nil), e.value...), e.err
		}
	}
	if len(r.entries) >= 1024 {
		r.mu.Unlock()
		value, err := load()
		return append([]byte(nil), value...), err
	}
	e := &preparationRead{ready: make(chan struct{}), at: now}
	r.entries[key] = e
	r.mu.Unlock()
	value, err := load()
	r.mu.Lock()
	e.value = append([]byte(nil), value...)
	e.err = err
	if err != nil {
		if r.entries[key] == e {
			delete(r.entries, key)
		}
	}
	close(e.ready)
	r.mu.Unlock()
	return append([]byte(nil), value...), err
}

type preparationProvider struct {
	Provider
	reads    *preparationReads
	protocol common.Address
	mu       sync.Mutex
	blocks   map[string]common.Hash
}

func (p *preparationProvider) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	h, err := p.Provider.HeaderByNumber(ctx, n)
	if err == nil && h != nil && h.Number != nil {
		p.mu.Lock()
		key := h.Number.String()
		if old, ok := p.blocks[key]; ok && old != h.Hash() {
			p.reads.mu.Lock()
			clear(p.reads.entries)
			p.reads.mu.Unlock()
		}
		if len(p.blocks) >= 1024 {
			clear(p.blocks)
		}
		p.blocks[key] = h.Hash()
		p.mu.Unlock()
	}
	return h, err
}
func (p *preparationProvider) CodeAt(ctx context.Context, a common.Address, n *big.Int) ([]byte, error) {
	if n == nil {
		return p.Provider.CodeAt(ctx, a, n)
	}
	return p.reads.read(ctx, "code:"+a.Hex()+":"+n.String(), func() ([]byte, error) { return p.Provider.CodeAt(ctx, a, n) })
}

var preparationSelectors = func() map[string]bool {
	result := map[string]bool{}
	for _, signature := range []string{"VERSION()", "addressRegistry()", "rootHistory(uint256,bytes32)", "isNullifierSpent(bytes32)", "nonces(address)", "creditAllocation(address,address)", "creditRevocationBitmap(address,address,uint64,uint8)"} {
		result[string(crypto.Keccak256([]byte(signature))[:4])] = true
	}
	return result
}()

func (p *preparationProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, n *big.Int) ([]byte, error) {
	// Latest nonce, balances, allowances and signature checks must remain live.
	if n == nil || msg.To == nil || *msg.To != p.protocol || len(msg.Data) < 4 || !preparationSelectors[string(msg.Data[:4])] {
		return p.Provider.CallContract(ctx, msg, n)
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(raw)
	return p.reads.read(ctx, "call:"+n.String()+":"+common.Hash(digest).Hex(), func() ([]byte, error) { return p.Provider.CallContract(ctx, msg, n) })
}
func (s *Runtime) preparationRuntime() *Runtime {
	next := *s
	if _, ok := next.Options.Provider.(*preparationProvider); !ok {
		next.Options.Provider = &preparationProvider{Provider: next.Options.Provider, reads: newPreparationReads(), protocol: s.Scope.ProtocolAddress, blocks: make(map[string]common.Hash)}
		next.RPC = s.RPC.WithProvider(next.Options.Provider)
	}
	return &next
}
