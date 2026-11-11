package session

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

type historyChain struct {
	sessionProvider
	headers map[uint64]*types.Header
}

func (p historyChain) HeaderByNumber(_ context.Context, n *big.Int) (*types.Header, error) {
	if n == nil {
		return p.headers[10], nil
	}
	h := p.headers[n.Uint64()]
	if h == nil {
		return nil, errors.New("missing header")
	}
	return h, nil
}
func sharedEmptyFixture(t *testing.T) (*WalletSession, *atomic.Int32) {
	t.Helper()
	s, _, _ := testSession(t)
	base := s.options.Provider.(sessionProvider)
	headers := map[uint64]*types.Header{}
	for n := uint64(1); n <= 10; n++ {
		h := &types.Header{Number: new(big.Int).SetUint64(n), Time: n}
		if n > 1 {
			h.ParentHash = headers[n-1].Hash()
		}
		headers[n] = h
	}
	provider := historyChain{base, headers}
	deployment := s.protocol.Deployment()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		var result any
		switch request.Method {
		case "invisi_getStatus":
			h := headers[10]
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": protocolrpc.CanonicalBlock{Number: 10, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}, "checkedBits": 8, "detectionEnabled": true}
		case "invisi_getNoteUpdates":
			calls.Add(1)
			var q struct{ ToBlock uint64 }
			_ = json.Unmarshal(request.Params[0], &q)
			result = map[string]any{"blockHash": headers[q.ToBlock].Hash(), "notes": []any{}}
		case "invisi_getCreditTransitions":
			result = []any{}
		default:
			t.Errorf("unexpected %s", request.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	rpc, e := sdk.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &deployment})
	if e != nil {
		t.Fatal(e)
	}
	s.rpc = rpc.WithProvider(provider)
	s.options.Provider = provider
	s.publicHistory, e = NewPublicHistory(PublicHistoryOptions{Protocol: s.protocol, Provider: provider, ProtocolRPC: s.rpc})
	if e != nil {
		t.Fatal(e)
	}
	s.options.Store = NewMemoryWalletStore()
	return s, &calls
}
func TestSharedNormalRefreshPersistsEmptyCoverageAndOfflineSecondAccount(t *testing.T) {
	ctx := context.Background()
	a, calls := sharedEmptyFixture(t)
	var decrypts atomic.Int32
	d := wallet.DeliveryDecryptorFuncs{Note: func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error) {
		decrypts.Add(1)
		return nil, nil
	}}
	pending, e := a.Refresh(ctx, RefreshOptions{PageSize: 3})
	if e != nil {
		t.Fatal(e)
	}
	before, e := a.GetScanStatus(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if before.FullyScannedHeight != nil {
		t.Fatal("staging advanced private cursor")
	}
	if _, e = pending.Decrypt(ctx, d); e != nil {
		t.Fatal("A decrypt", e)
	}
	status, e := a.GetScanStatus(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if !status.ScanComplete || !status.IsSynced || status.FullyScannedHeight == nil || *status.FullyScannedHeight != 10 {
		t.Fatalf("%+v", status)
	}
	if decrypts.Load() != 0 {
		t.Fatal("empty pages decrypted")
	}
	count := calls.Load()
	b := *a
	b.options = a.options
	b.options.Store = NewMemoryWalletStore()
	b.scope.Namespace = "second-account"
	b.refreshGate = make(chan struct{}, 1)
	if _, e = b.ScanHistory(ctx, d, ScanOptions{Offline: true, PageSize: 3}); e != nil {
		v, _ := a.publicHistory.Snapshot(ctx)
		t.Fatal("B scan", e, v.Pages)
	}
	if calls.Load() != count {
		t.Fatal("B redownloaded public candidates")
	}
}
func TestScanCancellationPreservesCommittedGaps(t *testing.T) {
	s, _ := sharedEmptyFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	_, e := s.ScanHistory(ctx, wallet.DeliveryDecryptorFuncs{}, ScanOptions{PageSize: 3, OnProgress: func(p HistoryProgress) {
		if p.Phase == "committed" {
			cancel()
		}
	}})
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	v, e := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	if e != nil {
		t.Fatal(e)
	}
	if len(v.Scan.Coverage) != 1 || v.Scan.Coverage[0].FromBlock != 8 || v.Scan.Coverage[0].Block.Number != 10 || v.DiscoveryThrough != nil {
		t.Fatalf("cursor lost/overclaimed %+v", v.Scan)
	}
}

func TestSameHeightReorgWithNoGapInvalidatesAndResetDropsOrphanCandidates(t *testing.T) {
	s, _ := sharedEmptyFixture(t)
	ctx := context.Background()
	if _, e := s.ScanHistory(ctx, wallet.DeliveryDecryptorFuncs{}, ScanOptions{}); e != nil {
		t.Fatal(e)
	}
	if e := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		v.Scan.Unresolved = map[string]UnresolvedDelivery{"orphan": {Method: wallet.DecryptNote}}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	provider := s.options.Provider.(historyChain)
	replacement := types.CopyHeader(provider.headers[10])
	replacement.Extra = []byte{99}
	provider.headers[10] = replacement
	if e := s.RefreshState(ctx); !errors.Is(e, ErrPublicHistoryReorg) {
		t.Fatalf("fully-covered reorg escaped: %v", e)
	}
	status, e := s.GetScanStatus(ctx)
	if e != nil || !status.Invalidated || status.IsSynced {
		t.Fatal("orphan history remained usable", e)
	}
	if _, e = s.ScanHistory(ctx, wallet.DeliveryDecryptorFuncs{}, ScanOptions{Reset: true}); e != nil {
		t.Fatal(e)
	}
	v, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil || v.Invalidated || len(v.Scan.Unresolved) != 0 {
		t.Fatal("reset retained orphan ambiguity", e)
	}
}

func TestMemoryResetRestoresArchivedExternalizedJournalAndInputLock(t *testing.T) {
	s, token, _ := testSession(t)
	ctx := context.Background()
	address, _ := s.wallet.GetAddress()
	operation, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	snapshot, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	var note StoredNote
	for _, n := range snapshot.Notes {
		note = n
	}
	store := NewMemoryWalletStore(MemoryWalletStoreOptions{RecentRecords: 0})
	if e = store.Update(ctx, s.scope, func(v *WalletState) error {
		*v = snapshot
		record := v.Operations[operation.plan.ID]
		record.Status = "confirmed"
		record.Externalized = true
		v.Operations[record.ID] = record
		for key, n := range v.Notes {
			n.Spent = true
			n.ReservedBy = ""
			v.Notes[key] = n
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	before, e := store.ReadSnapshot(ctx, s.scope)
	if e != nil || len(before.Operations) != 0 || len(before.Notes) != 0 {
		t.Fatal("fixture failed to archive", e)
	}
	_, _, page := publicNoteFixture(t)
	page.Range = BlockRange{1, 10}
	page.Headers = nil
	for n := uint64(1); n <= 10; n++ {
		h := fixtureHeader(n, 100)
		page.Headers = append(page.Headers, protocolrpc.CanonicalBlock{Number: n, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time})
	}
	page.Block = page.Headers[9]
	page.Transactions[0].Batch.Outputs[0].Commitment = note.Commitment
	if e = PutPublicPage(ctx, s.publicHistory.Store(), s.publicHistory.Key(), page, 0); e != nil {
		t.Fatal(e)
	}
	s.options.Store = store
	if _, e = s.ScanHistory(ctx, testDeliveryDecryptor(s), ScanOptions{Reset: true, Offline: true}); e != nil {
		t.Fatal(e)
	}
	after, e := store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	if after.Operations[operation.plan.ID].Status != "reorged" || after.Notes[note.Commitment].ReservedBy != operation.plan.ID {
		t.Fatal("reset lost externalized authority or input lock")
	}
	history, e := store.ReadRecords(ctx, s.scope, RecordQuery{Kind: "notes", Limit: 50})
	if e != nil || len(history.Items) != 1 {
		t.Fatal("reset duplicated active/archive Note", len(history.Items), e)
	}
}

func TestPendingRefreshHandsUnretainedPageDirectlyToScan(t *testing.T) {
	s, calls := sharedEmptyFixture(t)
	history, e := NewPublicHistory(PublicHistoryOptions{Store: NewMemoryPublicHistoryStore(PublicRetention{MaxBytes: 1}), Protocol: s.protocol, Provider: s.options.Provider, ProtocolRPC: s.rpc})
	if e != nil {
		t.Fatal(e)
	}
	s.publicHistory = history
	pending, e := s.Refresh(context.Background(), RefreshOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = pending.Decrypt(context.Background(), wallet.DeliveryDecryptorFuncs{}); e != nil {
		t.Fatal(e)
	}
	if calls.Load() != 1 {
		t.Fatalf("staged page fetched %d times", calls.Load())
	}
}

type reorgAfterPageStore struct {
	*MemoryPublicHistoryStore
	reads atomic.Int32
}

func (s *reorgAfterPageStore) Read(ctx context.Context, key string) (PublicHistoryState, error) {
	if s.reads.Add(1) == 3 {
		if e := s.MemoryPublicHistoryStore.Update(ctx, key, func(v *PublicHistoryState) error { v.Generation++; v.Coverage = nil; v.Pages = nil; return nil }); e != nil {
			return PublicHistoryState{}, e
		}
	}
	return s.MemoryPublicHistoryStore.Read(ctx, key)
}
func TestStagingCannotRelabelOldPageAfterGenerationChange(t *testing.T) {
	s, _ := sharedEmptyFixture(t)
	store := &reorgAfterPageStore{MemoryPublicHistoryStore: NewMemoryPublicHistoryStore(PublicRetention{})}
	h, e := NewPublicHistory(PublicHistoryOptions{Store: store, Protocol: s.protocol, Provider: s.options.Provider, ProtocolRPC: s.rpc})
	if e != nil {
		t.Fatal(e)
	}
	s.publicHistory = h
	if _, e = s.Refresh(context.Background(), RefreshOptions{}); e == nil {
		t.Fatal("staging relabeled old page with new generation")
	}
}
