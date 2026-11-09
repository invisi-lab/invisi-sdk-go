package session

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestStoreUpdatesAreAtomicAndDefensive(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryWalletStore()
	scope := WalletScope{ChainID: 1}
	hash := common.HexToHash("0x01")
	err := store.Update(ctx, scope, func(state *WalletState) error { state.Notes[hash] = StoredNote{}; return errors.New("rollback") })
	if err == nil {
		t.Fatal("expected callback error")
	}
	snapshot, _ := store.ReadSnapshot(ctx, scope)
	if len(snapshot.Notes) != 0 {
		t.Fatal("partial update leaked")
	}
	store.Update(ctx, scope, func(state *WalletState) error { state.Notes[hash] = StoredNote{}; return nil })
	snapshot, _ = store.ReadSnapshot(ctx, scope)
	delete(snapshot.Notes, hash)
	again, _ := store.ReadSnapshot(ctx, scope)
	if len(again.Notes) != 1 {
		t.Fatal("snapshot mutated stored state")
	}
}
func TestEncryptedStoreBindsScopeAndRejectsTampering(t *testing.T) {
	codec, e := NewEncryptedWalletCodec(make([]byte, 32))
	if e != nil {
		t.Fatal(e)
	}
	scope := WalletScope{ChainID: 1}
	state := emptyState()
	encoded, e := codec.Seal(scope, state)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = codec.Open(scope, encoded); e != nil {
		t.Fatal(e)
	}
	if _, e = codec.Open(WalletScope{ChainID: 2}, encoded); e == nil {
		t.Fatal("scope substitution accepted")
	}
	encoded[len(encoded)-1] ^= 1
	if _, e = codec.Open(scope, encoded); e == nil {
		t.Fatal("modified ciphertext accepted")
	}
}
func TestWalletPermissionsAndSecretRedaction(t *testing.T) {
	wallet, e := sdk.WalletFromPrivateKey("0x" + strings.Repeat("0", 63) + "1")
	if e != nil {
		t.Fatal(e)
	}
	address, e := wallet.GetAddress()
	if e != nil || len(address) != 138 {
		t.Fatal(address, e)
	}
	public, e := sdk.PublicWalletFromAddress(address)
	if e != nil {
		t.Fatal(e)
	}
	if _, exists := reflect.TypeOf(public).MethodByName("GetViewer"); exists {
		t.Fatal("public address granted viewing permission")
	}
	viewer, e := wallet.GetViewer()
	if e != nil {
		t.Fatal(e)
	}
	copy := viewer.ExportKey()
	copy[0] ^= 1
	if copy[0] == viewer.ExportKey()[0] {
		t.Fatal("private export aliases internal key")
	}
	if strings.Contains(wallet.GoString(), "80d9cd") {
		t.Fatal("secret logged")
	}
}

func TestRefreshRestoresPendingNoteReservations(t *testing.T) {
	state := emptyState()
	noteHash := common.HexToHash("0x01")
	plan := transaction.Plan{Inputs: []StoredNote{{Commitment: noteHash}}}
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	state.Operations["pending"] = OperationRecord{ID: "pending", Status: "failed", Externalized: true, Plan: raw}
	state.Notes[noteHash] = StoredNote{Commitment: noteHash}
	if err := restoreReservations(&state); err != nil {
		t.Fatal(err)
	}
	if state.Notes[noteHash].ReservedBy != "pending" {
		t.Fatal("pending spend was released during inventory rebuild")
	}
	state.Notes[noteHash] = StoredNote{Commitment: noteHash}
	record := state.Operations["pending"]
	record.Status = "expired"
	state.Operations["pending"] = record
	if err := restoreReservations(&state); err != nil {
		t.Fatal(err)
	}
	if state.Notes[noteHash].ReservedBy != "" {
		t.Fatal("expired operation reserved a rediscovered note")
	}
}

func TestReorgInvalidatesBalancesAndProtectsConfirmedInputs(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t)
	snapshot, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	var note StoredNote
	for _, current := range snapshot.Notes {
		note = current
	}
	raw, err := json.Marshal(transaction.Plan{Inputs: []StoredNote{note}})
	if err != nil {
		t.Fatal(err)
	}
	err = session.options.Store.Update(ctx, session.scope, func(state *WalletState) error {
		state.Operations["confirmed"] = OperationRecord{ID: "confirmed", Status: "confirmed", Externalized: true, WorkerID: "old", Plan: raw}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.invalidate(ctx); err != nil {
		t.Fatal(err)
	}
	balance, err := session.Balance(ctx, token)
	if err != nil || !balance.Invalidated || balance.AvailableKnown != nil || balance.CheckedThrough != nil {
		t.Fatal(balance, err)
	}
	state, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	if state.Operations["confirmed"].Status != "reorged" || state.Operations["confirmed"].WorkerID != "" {
		t.Fatal("old operation handle survived reorg")
	}
	state.Notes[note.Commitment] = note
	if err = restoreReservations(&state); err != nil {
		t.Fatal(err)
	}
	if state.Notes[note.Commitment].ReservedBy != "confirmed" {
		t.Fatal("reorg released externalized inputs")
	}
}
