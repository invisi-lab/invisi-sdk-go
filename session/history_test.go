package session

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestHistoryDetachedBusinessRecordsAndIndependentDiscovery(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryWalletStore()
	scope := WalletScope{ChainID: 97}
	token := common.HexToAddress("0x1111111111111111111111111111111111111111")
	commitment := common.HexToHash("0x1234")
	block := protocolrpc.CanonicalBlock{Number: 12, Hash: common.HexToHash("0x12")}
	plan := transaction.Plan{ID: "operation-1", Kind: "shield", Context: protocol.TransactionContext{TokenAddress: token}, Block: block, FundingMode: "caller",
		Funding: protocol.FundingParams{Amount: big.NewInt(7)}, Outputs: []transaction.PreparedOutput{{Address: "recipient", Amount: big.NewInt(7), Note: &core.Note{Amount: core.ElementFromUint64(7), Payload: []byte{1, 2, 3}}}}}
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Update(ctx, scope, func(state *WalletState) error {
		state.DiscoveryThrough = &block
		state.ObservedThrough = &protocolrpc.CanonicalBlock{Number: 999}
		state.Operations[plan.ID] = OperationRecord{ID: plan.ID, Kind: "shield", Status: "uncertain", Plan: raw, Proof: json.RawMessage(`{"privateWitness":"do-not-expose"}`)}
		state.Notes[commitment] = StoredNote{Note: core.Note{TokenAddress: core.BigIntToElement(token.Big()), Amount: core.ElementFromUint64(7), Payload: []byte{4, 5}}, Commitment: commitment, Spent: true, Block: &block}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	session := &WalletSession{scope: scope, options: sessionOptions{Store: store}}
	history, err := session.GetHistory(ctx, HistoryOptions{TokenAddress: &token})
	if err != nil {
		t.Fatal(err)
	}
	if history.Revision != 1 || history.CheckedThrough == nil || history.CheckedThrough.Number != 12 || len(history.Outgoing) != 1 || len(history.Incoming) != 1 || !history.Incoming[0].Spent {
		t.Fatalf("invalid summary: %+v", history)
	}
	encoded, _ := json.Marshal(history)
	if strings.Contains(string(encoded), "privateWitness") || strings.Contains(string(encoded), "do-not-expose") {
		t.Fatal("private proof escaped history")
	}
	history.Outgoing[0].Outputs[0].Payload[0] = 99
	history.Outgoing[0].Outputs[0].Amount.SetInt64(999)
	history.Outgoing[0].Funding.Amount.SetInt64(999)
	history.Incoming[0].Payload[0] = 99
	history.Incoming[0].Amount.SetInt64(999)
	history.Incoming[0].Block.Number = 999
	history.CheckedThrough.Number = 999
	again, err := session.GetHistory(ctx, HistoryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if again.CheckedThrough.Number != 12 || again.Outgoing[0].Outputs[0].Payload[0] != 1 || again.Outgoing[0].Outputs[0].Amount.Int64() != 7 || again.Outgoing[0].Funding.Amount.Int64() != 7 || again.Incoming[0].Block.Number != 12 || again.Incoming[0].Amount.Int64() != 7 || !reflect.DeepEqual(again.Incoming[0].Payload, []byte{4, 5}) {
		t.Fatal("returned history aliases stored state")
	}
	other := common.HexToAddress("0x2222222222222222222222222222222222222222")
	filtered, err := session.GetHistory(ctx, HistoryOptions{TokenAddress: &other})
	if err != nil || len(filtered.Outgoing) != 0 || len(filtered.Incoming) != 0 {
		t.Fatal("asset filter failed", err)
	}
}

func TestHistoryDoesNotClaimDiscoveryAfterOnlyReconciliation(t *testing.T) {
	store := NewMemoryWalletStore()
	scope := WalletScope{ChainID: 97}
	if err := store.Update(context.Background(), scope, func(state *WalletState) error {
		state.ObservedThrough = &protocolrpc.CanonicalBlock{Number: 100}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	session := &WalletSession{scope: scope, options: sessionOptions{Store: store}}
	history, err := session.GetHistory(context.Background(), HistoryOptions{})
	if err != nil || history.CheckedThrough != nil {
		t.Fatalf("reconciliation became discovery: %+v, %v", history, err)
	}
}

func TestSubscriptionsFollowSuccessfulCommitsAndAreSessionLocal(t *testing.T) {
	ctx := context.Background()
	scope := WalletScope{ChainID: 97}
	base := NewMemoryWalletStore()
	notifications := newSessionNotifications()
	store := &notifyingWalletStore{base: base, notifications: notifications}
	session := &WalletSession{history: notifications}
	other := &WalletSession{history: newSessionNotifications()}
	called, otherCalled := make(chan struct{}, 8), make(chan struct{}, 8)
	stop := session.Subscribe(func() {
		if _, err := base.ReadSnapshot(ctx, scope); err != nil {
			t.Error(err)
		}
		called <- struct{}{}
	})
	other.Subscribe(func() { otherCalled <- struct{}{} })
	session.Subscribe(func() { panic("a UI callback cannot undo committed wallet state") })
	if err := store.Update(ctx, scope, func(state *WalletState) error { return errors.New("rollback") }); err == nil {
		t.Fatal("expected rollback")
	}
	select {
	case <-called:
		t.Fatal("failed update notified")
	default:
	}
	if err := store.Update(ctx, scope, func(state *WalletState) error { return nil }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("successful update did not notify")
	}
	select {
	case <-otherCalled:
		t.Fatal("another session was notified")
	default:
	}
	stop()
	stop()
	if err := store.Update(ctx, scope, func(state *WalletState) error { return nil }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-called:
		t.Fatal("unsubscribe failed")
	case <-time.After(20 * time.Millisecond):
	}
}

type historyReadHookStore struct {
	WalletStore
	beforeRead func()
}

func (s historyReadHookStore) ReadSnapshot(ctx context.Context, scope WalletScope) (WalletState, error) {
	s.beforeRead()
	return s.WalletStore.ReadSnapshot(ctx, scope)
}

func TestHistorySnapshotsAssetFilterBeforeReadingStore(t *testing.T) {
	ctx := context.Background()
	base := NewMemoryWalletStore()
	scope := WalletScope{ChainID: 97}
	token := common.HexToAddress("0x1111111111111111111111111111111111111111")
	original := token
	if err := base.Update(ctx, scope, func(state *WalletState) error {
		state.Notes[common.HexToHash("0x01")] = StoredNote{Note: core.Note{TokenAddress: core.BigIntToElement(token.Big()), Amount: core.ElementFromUint64(1)}}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	hook := historyReadHookStore{WalletStore: base, beforeRead: func() { token = common.HexToAddress("0x2222222222222222222222222222222222222222") }}
	session := &WalletSession{scope: scope, options: sessionOptions{Store: hook}}
	result, err := session.GetHistory(ctx, HistoryOptions{TokenAddress: &token})
	if err != nil || len(result.Incoming) != 1 || result.Incoming[0].TokenAddress != original {
		t.Fatalf("filter changed during read: %+v, %v", result, err)
	}
}

func TestSubscriptionMayReadTheOperationBeingCommitted(t *testing.T) {
	session, token, _ := testSession(t)
	prepared, err := session.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: walletAddress(session.wallet), Amount: big.NewInt(1)}})})
	if err != nil {
		t.Fatal(err)
	}
	observed := make(chan struct{}, 8)
	stop := session.Subscribe(func() {
		_ = prepared.Status()
		select {
		case observed <- struct{}{}:
		default:
		}
	})
	defer stop()
	done := make(chan error, 1)
	go func() { _, err := authorizeForTest(prepared, context.Background()); done <- err }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("subscription blocked an operation while its mutex was held")
	}
	select {
	case <-observed:
	case <-time.After(time.Second):
		t.Fatal("commit notification did not arrive")
	}
}
