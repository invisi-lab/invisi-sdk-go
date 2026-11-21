package session

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
	creditsession "github.com/invisi-labs/invisi-sdk-go/session/credit"
)

func multiNoteRefreshFixture(t testing.TB, count int) (*WalletSession, common.Address, []common.Hash, *atomic.Int32) {
	t.Helper()
	s, token, _ := testSession(t, 2000000000)
	base := s.options.Provider.(sessionProvider)
	state, _ := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	var sample StoredNote
	for _, n := range state.Notes {
		sample = n
	}
	state.Notes = map[common.Hash]StoredNote{}
	keys := []common.Hash{}
	for i := 0; i < count; i++ {
		n := sample
		n.Note.Nonce = primitives.ElementFromUint64(uint64(1000 + i))
		n.TreeID = uint64(i + 1)
		identity, _ := primitives.ComputeNoteIdentity(n.Note.OwnerAddress, n.Note.Nonce)
		leaf, _ := primitives.ComputeNoteCommitment(n.Note.ChainID, n.Note.ProtocolAddress, identity, n.Note.TokenAddress, n.Note.Amount)
		n.Commitment = common.Hash(leaf.Bytes())
		n.Nullifier, _ = s.wallet.ExportNullifier().ForNote(n.Note)
		state.Notes[n.Commitment] = n
		keys = append(keys, n.Commitment)
		base.leaves.Store(n.TreeID, n.Commitment)
	}
	if err := s.options.Store.Update(context.Background(), s.scope, func(v *WalletState) error { *v = state; return nil }); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	d := s.protocol.Deployment()
	b := *state.ObservedThrough
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		_ = json.NewDecoder(r.Body).Decode(&q)
		var out any
		if q.Method == "invisi_getStatus" {
			calls.Add(1)
			out = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": d, "indexedHead": b}
		} else {
			out = retainedNoteFixture(base, b, q.Method, q.Params)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": out})
	}))
	t.Cleanup(server.Close)
	c, err := protocolrpc.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &d})
	if err != nil {
		t.Fatal(err)
	}
	s.rpc = c.WithProvider(base)
	return s, token, keys, &calls
}

func TestRefreshAndLazyCollectionBatchOnlyMissingPaths(t *testing.T) {
	for _, inventoryOnly := range []bool{false, true} {
		t.Run(map[bool]string{false: "default-prewarm", true: "inventory-only"}[inventoryOnly], func(t *testing.T) {
			s, token, keys, calls := multiNoteRefreshFixture(t, 4)
			if err := s.RefreshStateWithOptions(context.Background(), StateRefreshOptions{InventoryOnly: inventoryOnly}); err != nil {
				t.Fatal(err)
			}
			want := int32(1)
			if inventoryOnly {
				want = 0
			}
			if calls.Load() != want {
				t.Fatalf("refresh repeated full membership evidence: got %d want %d", calls.Load(), want)
			}
			op, err := s.PrepareCollection(context.Background(), token, common.HexToAddress("0x123"), keys)
			if err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 1 || op.Preview().InputCount != 4 || op.Preview().WithdrawalAmount.Cmp(big.NewInt(400)) != 0 {
				t.Fatal("collection did not batch selected inputs or changed exact payout", calls.Load())
			}
		})
	}
}

func TestInventoryOnlyRefreshDiscardsPrewarmAndRequiresLiveLazyPreparation(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, 2000000000)
	transport := &offlineProvider{Provider: s.options.Provider}
	s.options.Provider = transport
	s.rpc = s.rpc.WithProvider(transport)
	if err := s.RefreshState(ctx); err != nil {
		t.Fatal(err)
	}
	originalRPC := s.rpc
	dead, _ := protocolrpc.NewProtocolRpcClient("http://127.0.0.1:1/v1")
	s.rpc = dead.WithProvider(transport)
	if err := s.RefreshStateWithOptions(ctx, StateRefreshOptions{InventoryOnly: true}); err != nil {
		t.Fatal("inventory refresh requested transaction material", err)
	}
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil || state.PreparationSnapshot != nil || state.CheckedThrough == nil {
		t.Fatal("inventory-only retained stale preparation or lost checks", err)
	}
	transport.disabled.Store(true)
	address, _ := s.wallet.GetAddress()
	if _, err := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})}); err == nil {
		t.Fatal("lazy preparation reused discarded snapshot")
	}
	transport.disabled.Store(false)
	s.rpc = originalRPC
	var inputs []common.Hash
	for key := range state.Notes {
		inputs = append(inputs, key)
	}
	if _, err := s.PrepareCollection(ctx, token, common.HexToAddress("0x123"), inputs); err != nil {
		t.Fatal("lazy collection failed", err)
	}
}

type independentStateProvider struct {
	sdk.Provider
	active, peak, calls         atomic.Int32
	gate                        chan struct{}
	once                        sync.Once
	failAt                      int32
	spent, revoked, rootUnknown atomic.Bool
}

func (p *independentStateProvider) CallContract(ctx context.Context, m ethereum.CallMsg, b *big.Int) ([]byte, error) {
	n := p.calls.Add(1)
	a := p.active.Add(1)
	defer p.active.Add(-1)
	for {
		old := p.peak.Load()
		if a <= old || p.peak.CompareAndSwap(old, a) {
			break
		}
	}
	if p.gate != nil {
		if a >= 2 {
			p.once.Do(func() { close(p.gate) })
		}
		select {
		case <-p.gate:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if p.failAt != 0 && n == p.failAt {
		return nil, errors.New("state read failed")
	}
	abi, _ := protocol.ProtocolABI()
	if method, e := abi.MethodById(m.Data[:4]); e == nil {
		switch method.RawName {
		case "rootHistory":
			return method.Outputs.Pack(!p.rootUnknown.Load())
		case "isNullifierSpent":
			return method.Outputs.Pack(p.spent.Load())
		case "creditStatus":
			v := uint8(1)
			if p.revoked.Load() {
				v = 2
			}
			return method.Outputs.Pack(big.NewInt(int64(v)))
		}
	}
	return p.Provider.CallContract(ctx, m, b)
}

type changingIdentityProvider struct {
	sdk.Provider
	reads atomic.Int32
}

func (p *changingIdentityProvider) ChainID(ctx context.Context) (*big.Int, error) {
	if p.reads.Add(1) > 1 {
		return big.NewInt(1), nil
	}
	return p.Provider.ChainID(ctx)
}

func TestInventoryOnlyRejectsIdentityChangeAtFinalFence(t *testing.T) {
	s, _, _ := testSession(t)
	ctx := context.Background()
	before, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	s.options.Provider = &changingIdentityProvider{Provider: s.options.Provider}
	if err := s.RefreshStateWithOptions(ctx, StateRefreshOptions{InventoryOnly: true}); err == nil {
		t.Fatal("changed chain identity committed")
	}
	after, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if after.Revision != before.Revision {
		t.Fatal("failed identity fence changed inventory")
	}
}

func TestInventoryOnlyRetainsPrivateObservationFenceOutsidePublicCoverage(t *testing.T) {
	s, _, _ := testSession(t)
	ctx := context.Background()
	base := s.options.Provider.(sessionProvider)
	orphan := fixtureHeader(9, base.head.Time)
	orphan.Extra = []byte{99}
	point := protocolrpc.CanonicalBlock{Number: 9, Hash: orphan.Hash(), ParentHash: orphan.ParentHash, Timestamp: orphan.Time}
	if err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error { v.ObservedThrough = &point; return nil }); err != nil {
		t.Fatal(err)
	}
	before, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err := s.RefreshStateWithOptions(ctx, StateRefreshOptions{InventoryOnly: true}); err == nil {
		t.Fatal("orphaned private observation outside public coverage accepted")
	}
	after, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if after.Revision != before.Revision {
		t.Fatal("failed observation fence committed partial state")
	}
}

type privateAnchorReorgProvider struct {
	sdk.Provider
	changed atomic.Bool
}

func (p *privateAnchorReorgProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	abi, _ := protocol.ProtocolABI()
	if method, err := abi.MethodById(msg.Data[:4]); err == nil && method.RawName == "isNullifierSpent" {
		p.changed.Store(true)
	}
	return p.Provider.CallContract(ctx, msg, block)
}

func (p *privateAnchorReorgProvider) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	h, err := p.Provider.HeaderByNumber(ctx, number)
	if err == nil && h != nil && h.Number.Uint64() == 9 && p.changed.Load() {
		copy := *h
		copy.Extra = []byte{99}
		return &copy, nil
	}
	return h, err
}

func TestInventoryOnlyFencesPrivateAnchorAfterConsumptionReads(t *testing.T) {
	s, _, _ := testSession(t)
	ctx := context.Background()
	base := s.options.Provider.(sessionProvider)
	h := fixtureHeader(9, base.head.Time)
	point := protocolrpc.CanonicalBlock{Number: 9, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
	if err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		v.ObservedThrough = &point
		v.CheckedThrough = &point
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	before, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	provider := &privateAnchorReorgProvider{Provider: base}
	s.options.Provider = provider
	s.rpc = s.rpc.WithProvider(provider)
	if err := s.RefreshStateWithOptions(ctx, StateRefreshOptions{InventoryOnly: true}); err == nil {
		t.Fatal("private anchor reorg during state reads accepted")
	}
	after, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if !provider.changed.Load() || after.Revision != before.Revision {
		t.Fatal("test did not reach state reads or committed partial inventory")
	}
}

func TestOnlineRefreshDoesNotTrustSameBlockCachedRoot(t *testing.T) {
	s, _, _ := testSession(t)
	ctx := context.Background()
	p := &independentStateProvider{Provider: s.options.Provider}
	s.options.Provider = p
	s.rpc = s.rpc.WithProvider(p)
	if err := s.RefreshState(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		for key, m := range v.PreparationSnapshot.Memberships {
			m.Path[0] = common.HexToHash("0x01")
			root, _ := primitives.HexToElement(m.Leaf.Hex())
			for i, sibling := range m.Path {
				other, _ := primitives.HexToElement(sibling.Hex())
				if m.LeafIndex>>i&1 == 0 {
					root, _ = primitives.Poseidon2Hash(root, other)
				} else {
					root, _ = primitives.Poseidon2Hash(other, root)
				}
			}
			m.Root = common.Hash(root.Bytes())
			v.PreparationSnapshot.Memberships[key] = m
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	p.rootUnknown.Store(true)
	if err := s.RefreshState(ctx); err == nil {
		t.Fatal("same-block metadata authenticated an unknown cached root")
	}
}

func TestKnownStateReadsAreBoundedConcurrentAndErrorsDoNotPartiallyApply(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "parallel", true: "atomic failure"}[fail], func(t *testing.T) {
			s, _, _ := testSession(t)
			state, _ := s.options.Store.ReadSnapshot(context.Background(), s.scope)
			var sample StoredNote
			for _, n := range state.Notes {
				sample = n
			}
			state.Notes = map[common.Hash]StoredNote{}
			for i := 1; i <= 9; i++ {
				n := sample
				n.Commitment = common.BigToHash(big.NewInt(int64(i)))
				n.Nullifier = n.Commitment
				n.CheckedThrough = nil
				state.Notes[n.Commitment] = n
			}
			p := &independentStateProvider{Provider: s.options.Provider, gate: make(chan struct{})}
			if fail {
				p.failAt = 3
			}
			s.options.Provider = p
			h := s.options.Provider.(*independentStateProvider).Provider.(sessionProvider).head
			b := protocolrpc.CanonicalBlock{Number: h.Number.Uint64(), Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := s.checkKnownState(ctx, &state, b)
			if fail {
				if err == nil {
					t.Fatal("failure swallowed")
				}
				for _, n := range state.Notes {
					if n.CheckedThrough != nil {
						t.Fatal("partial state applied")
					}
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if p.peak.Load() < 2 || p.peak.Load() > 4 {
					t.Fatalf("unbounded or serial reads: peak=%d", p.peak.Load())
				}
				if p.calls.Load() != 9 {
					t.Fatal("consumption checks skipped")
				}
			}
		})
	}
}

func TestInventoryOnlyRefreshRechecksSpentAndRevocationAtEveryInvocation(t *testing.T) {
	s, _, _ := testSession(t)
	p := &independentStateProvider{Provider: s.options.Provider}
	s.options.Provider = p
	ctx := context.Background()
	err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		v.PostpaidCredit.Grants["grant"] = creditsession.StoredState{Nullifier: common.HexToHash("0x42"), Role: credit.Owner, Delivery: credit.DeliveryState{State: syncGrantState()}}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RefreshStateWithOptions(ctx, StateRefreshOptions{InventoryOnly: true}); err != nil {
		t.Fatal(err)
	}
	p.spent.Store(true)
	p.revoked.Store(true)
	if err := s.RefreshStateWithOptions(ctx, StateRefreshOptions{InventoryOnly: true}); err != nil {
		t.Fatal(err)
	}
	v, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range v.Notes {
		if !n.Spent {
			t.Fatal("stale spent state reused")
		}
	}
	g := v.PostpaidCredit.Grants["grant"]
	if !g.Consumed || !g.Revoked {
		t.Fatal("consumption/revocation not refreshed")
	}
}

func BenchmarkStateRefreshInventory(b *testing.B) {
	for _, inventoryOnly := range []bool{false, true} {
		name := "full-prewarm"
		if inventoryOnly {
			name = "inventory-only"
		}
		b.Run(name, func(b *testing.B) {
			s, _, _, calls := multiNoteRefreshFixture(b, 32)
			ctx := context.Background()
			b.ResetTimer()
			for range b.N {
				if err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error { v.PreparationSnapshot = nil; return nil }); err != nil {
					b.Fatal(err)
				}
				if err := s.RefreshStateWithOptions(ctx, StateRefreshOptions{InventoryOnly: inventoryOnly}); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(calls.Load())/float64(b.N), "status/op")
		})
	}
}
