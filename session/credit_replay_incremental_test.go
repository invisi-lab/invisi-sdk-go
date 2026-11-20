package session

import (
	"context"
	"errors"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/protocol/receipts"
	creditsession "github.com/invisi-labs/invisi-sdk-go/session/credit"
)

func TestForeignCreditDoesNotQueueWalletReplay(t *testing.T) {
	s, _, page := publicNoteFixture(t)
	page.Transactions[0].Batch.CreditNullifiers = []common.Hash{common.HexToHash("0x999")}
	if err := PutPublicPage(context.Background(), s.publicHistory.Store(), s.publicHistory.Key(), page, 0); err != nil {
		t.Fatal(err)
	}
	from := uint64(10)
	if _, err := s.ScanHistory(context.Background(), testDeliveryDecryptor(s), ScanOptions{FromBlock: &from, Offline: true}); err != nil {
		t.Fatal(err)
	}
	state, err := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	if err != nil {
		t.Fatal(err)
	}
	if state.Scan.CreditReplayTarget != 0 {
		t.Fatalf("foreign Credit queued replay to %d", state.Scan.CreditReplayTarget)
	}
}

func TestCreditReplaySchedulingPreservesProgressAndRewindsHistoricalParent(t *testing.T) {
	var state PrivateScanState
	scheduleCreditReplay(&state, 900, 1000)
	if state.CreditReplayFrom != 900 || state.CreditReplayTarget != 1000 {
		t.Fatalf("unexpected initial replay %+v", state)
	}
	state.CreditReplayCursor = 950
	scheduleCreditReplay(&state, 980, 1100)
	if state.CreditReplayCursor != 950 || state.CreditReplayFrom != 900 || state.CreditReplayTarget != 1100 {
		t.Fatalf("forward extension lost checkpoint %+v", state)
	}
	scheduleCreditReplay(&state, 700, 1100)
	if state.CreditReplayCursor != 0 || state.CreditReplayFrom != 700 {
		t.Fatalf("historical parent not replayed %+v", state)
	}
	state.CreditReplayCursor = 800
	scheduleCreditReplay(&state, 750, 1100)
	if state.CreditReplayCursor != 0 || state.CreditReplayFrom != 700 {
		t.Fatalf("new parent inside completed prefix not replayed %+v", state)
	}
}

func TestCreditReplayLegacyPendingObligationRemainsConservative(t *testing.T) {
	state := PrivateScanState{CreditReplayTarget: 1000, CreditReplayCursor: 950}
	scheduleCreditReplay(&state, 980, 1100)
	if state.CreditReplayFrom != 0 || state.CreditReplayCursor != 950 {
		t.Fatalf("legacy replay narrowed unsafely %+v", state)
	}
}

func TestCreditReplayRelevantIncludesCiphertextFreeOwnedNullifier(t *testing.T) {
	state := emptyState()
	nf := common.HexToHash("0x42")
	state.PostpaidCredit.Grants["owned"] = creditsession.StoredState{Nullifier: nf, Role: credit.Owner}
	page := PublicPage{Transactions: []PublicTransaction{{Batch: receipts.Batch{CreditNullifiers: []common.Hash{nf}}}}}
	if !creditReplayRelevant(state, page) {
		t.Fatal("owned transition without delivery omitted")
	}
	page.Transactions[0].Batch.CreditNullifiers[0] = common.HexToHash("0x99")
	if creditReplayRelevant(state, page) {
		t.Fatal("foreign nullifier selected")
	}
}

func TestCreditReplayStartsAtDirtyRangeAndResumesCommittedCursor(t *testing.T) {
	s, _ := sharedEmptyFixture(t)
	ctx := context.Background()
	if err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error { scheduleCreditReplay(&v.Scan, 8, 10); return nil }); err != nil {
		t.Fatal(err)
	}
	target := protocolrpc.CanonicalBlock{Number: 10}
	interrupted, cancel := context.WithCancel(ctx)
	var first []BlockRange
	err := s.replayCredits(interrupted, target, 1, ScanOptions{SkipStateRefresh: true, OnProgress: func(p HistoryProgress) { first = append(first, p.Range); cancel() }}, &RefreshResult{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want interrupted replay: %v", err)
	}
	if len(first) != 1 || first[0] != (BlockRange{8, 8}) {
		t.Fatalf("replayed before owned parent: %v", first)
	}
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	if state.Scan.CreditReplayCursor != 8 || state.Scan.CreditReplayFrom != 8 || state.Scan.CreditReplayTarget != 10 {
		t.Fatalf("checkpoint lost %+v", state.Scan)
	}
	var resumed []BlockRange
	if err = s.replayCredits(ctx, target, 1, ScanOptions{SkipStateRefresh: true, OnProgress: func(p HistoryProgress) { resumed = append(resumed, p.Range) }}, &RefreshResult{}); err != nil {
		t.Fatal(err)
	}
	if len(resumed) != 2 || resumed[0] != (BlockRange{9, 9}) || resumed[1] != (BlockRange{10, 10}) {
		t.Fatalf("resumed wrong range %v", resumed)
	}
	state, err = s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	if state.Scan.CreditReplayTarget != 0 || state.Scan.CreditReplayFrom != 0 || state.Scan.CreditReplayCursor != 0 {
		t.Fatalf("completed replay remains pending %+v", state.Scan)
	}
	if err = s.replayCredits(ctx, target, 1, ScanOptions{OnProgress: func(HistoryProgress) { t.Fatal("unchanged state replayed again") }}, &RefreshResult{}); err != nil {
		t.Fatal(err)
	}
}

type checkpointGenerationRaceStore struct {
	*MemoryWalletStore
	history *PublicHistory
}

func (m *checkpointGenerationRaceStore) CommitCreditReplayCheckpoint(ctx context.Context, scope WalletScope, c CreditReplayCheckpoint) error {
	if err := m.MemoryWalletStore.CommitCreditReplayCheckpoint(ctx, scope, c); err != nil {
		return err
	}
	return m.history.options.Store.Update(ctx, m.history.key, func(v *PublicHistoryState) error { v.Generation++; return nil })
}

func TestCreditReplayEmptyCheckpointInvalidatesOnGenerationRace(t *testing.T) {
	s, _ := sharedEmptyFixture(t)
	ctx := context.Background()
	store := &checkpointGenerationRaceStore{NewMemoryWalletStore(), s.publicHistory}
	s.options.Store = store
	if err := store.Update(ctx, s.scope, func(v *WalletState) error { scheduleCreditReplay(&v.Scan, 1, 10); return nil }); err != nil {
		t.Fatal(err)
	}
	_, err := s.checkpointEmptyCreditReplay(ctx, PublicPage{}, 10, 1, 2)
	if err == nil {
		t.Fatal("public generation race accepted")
	}
	state, err := store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Invalidated || state.Scan.CreditReplayTarget == 0 {
		t.Fatal("generation race lost recovery obligation")
	}
}

type countingReplayStore struct {
	*MemoryWalletStore
	snapshots, checkpoints int
}

func (m *countingReplayStore) ReadSnapshot(ctx context.Context, scope WalletScope) (WalletState, error) {
	m.snapshots++
	return m.MemoryWalletStore.ReadSnapshot(ctx, scope)
}
func (m *countingReplayStore) CommitCreditReplayCheckpoint(ctx context.Context, scope WalletScope, c CreditReplayCheckpoint) error {
	m.checkpoints++
	return m.MemoryWalletStore.CommitCreditReplayCheckpoint(ctx, scope, c)
}

func TestCreditReplayEmptyWindowsAvoidRepeatedRecordLoads(t *testing.T) {
	s, _ := sharedEmptyFixture(t)
	ctx := context.Background()
	store := &countingReplayStore{MemoryWalletStore: NewMemoryWalletStore()}
	s.options.Store = store
	if err := store.Update(ctx, s.scope, func(v *WalletState) error { scheduleCreditReplay(&v.Scan, 1, 10); return nil }); err != nil {
		t.Fatal(err)
	}
	if err := s.replayCredits(ctx, protocolrpc.CanonicalBlock{Number: 10}, 1, ScanOptions{SkipStateRefresh: true}, &RefreshResult{}); err != nil {
		t.Fatal(err)
	}
	if store.snapshots != 2 || store.checkpoints != 9 {
		t.Fatalf("empty-window full reads=%d checkpoints=%d", store.snapshots, store.checkpoints)
	}
	state, err := store.MemoryWalletStore.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	if state.Scan.CreditReplayTarget != 0 || state.CheckedThrough != nil || state.Invalidated {
		t.Fatal("checkpoint changed validation claims")
	}
}
