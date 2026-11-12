package sqlite

import (
	"context"
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/session"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
	"math/big"
	"testing"
)

func TestTerminalRecordsArchiveAndPageWithoutFullLoad(t *testing.T) {
	ctx := context.Background()
	db, e := sql.Open("sqlite", ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	codec, _ := session.NewEncryptedWalletCodec(make([]byte, 32))
	store, e := NewWalletStore(ctx, db, codec, WalletStoreOptions{RecentRecords: 2})
	if e != nil {
		t.Fatal(e)
	}
	scope := session.WalletScope{ChainID: 1}
	if e = store.Update(ctx, scope, func(v *session.WalletState) error {
		for i := 1; i <= 10; i++ {
			h := common.BigToHash(big.NewInt(int64(i)))
			v.Notes[h] = session.StoredNote{Commitment: h, Spent: true}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	v, e := store.ReadSnapshot(ctx, scope)
	if e != nil {
		t.Fatal(e)
	}
	if len(v.Notes) != 2 {
		t.Fatalf("active count %d", len(v.Notes))
	}
	page, e := store.ReadRecords(ctx, scope, session.RecordQuery{Kind: "notes", Limit: 3})
	if e != nil {
		t.Fatal(e)
	}
	if len(page.Items) != 3 || page.NextCursor == "" {
		t.Fatalf("%+v", page)
	}
	count := len(page.Items)
	for page.NextCursor != "" {
		page, e = store.ReadRecords(ctx, scope, session.RecordQuery{Kind: "notes", Limit: 3, Cursor: page.NextCursor})
		if e != nil {
			t.Fatal(e)
		}
		count += len(page.Items)
	}
	if count != 10 {
		t.Fatalf("history count %d", count)
	}
	if _, e = db.Exec("DELETE FROM invisi_wallet_records WHERE rowid IN (SELECT rowid FROM invisi_wallet_records WHERE archived=0 LIMIT 1)"); e != nil {
		t.Fatal(e)
	}
	if _, e = store.ReadSnapshot(ctx, scope); e == nil {
		t.Fatal("missing active row accepted")
	}
}

func TestCreditReplayPinsParentsWithZeroRecentRecords(t *testing.T) {
	ctx := context.Background()
	db, e := sql.Open("sqlite", ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	codec, _ := session.NewEncryptedWalletCodec(make([]byte, 32))
	store, e := NewWalletStore(ctx, db, codec, WalletStoreOptions{RecentRecords: 0})
	if e != nil {
		t.Fatal(e)
	}
	scope := session.WalletScope{ChainID: 1}
	if e = store.Update(ctx, scope, func(v *session.WalletState) error {
		v.Scan.CreditReplayTarget = 16
		v.PostpaidCredit.Grants["parent"] = creditsession.StoredState{Consumed: true, Delivery: credit.DeliveryState{State: credit.State{SpenderWalletIdentity: primitives.ElementFromUint64(1), OwnerWalletIdentity: primitives.ElementFromUint64(2), RevocationShard: primitives.ElementFromUint64(1), Expiry: primitives.ElementFromUint64(100), CreditSharedKey: primitives.ElementFromUint64(3)}}}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	v, e := store.ReadSnapshot(ctx, scope)
	if e != nil || len(v.PostpaidCredit.Grants) != 1 {
		t.Fatal("required parent archived", e)
	}
	if e = store.Update(ctx, scope, func(v *session.WalletState) error { v.Scan.CreditReplayTarget = 0; return nil }); e != nil {
		t.Fatal(e)
	}
	v, e = store.ReadSnapshot(ctx, scope)
	if e != nil || len(v.PostpaidCredit.Grants) != 0 {
		t.Fatal("terminal parent retained after replay", e)
	}
}
func TestArchivedExternalizedJournalRestoresBeforeReset(t *testing.T) {
	ctx := context.Background()
	db, e := sql.Open("sqlite", ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	codec, _ := session.NewEncryptedWalletCodec(make([]byte, 32))
	store, e := NewWalletStore(ctx, db, codec, WalletStoreOptions{RecentRecords: 0})
	if e != nil {
		t.Fatal(e)
	}
	scope := session.WalletScope{ChainID: 1}
	if e = store.Update(ctx, scope, func(v *session.WalletState) error {
		v.Operations["exposed"] = session.OperationRecord{ID: "exposed", Status: "confirmed", Externalized: true, CallData: []byte{1, 2}}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if e = store.RestoreExternalizedOperations(ctx, scope); e != nil {
		t.Fatal(e)
	}
	v, e := store.ReadSnapshot(ctx, scope)
	if e != nil {
		t.Fatal(e)
	}
	if !v.Invalidated || v.Operations["exposed"].Status != "reorged" || len(v.Operations["exposed"].CallData) != 2 {
		t.Fatal("external authority lost")
	}
	page, e := store.ReadRecords(ctx, scope, session.RecordQuery{Kind: "operations"})
	if e != nil || len(page.Items) != 1 {
		t.Fatal("reactivation duplicated count", e)
	}
}

func TestDroppingEncryptedStorageMetadataRollsBack(t *testing.T) {
	ctx := context.Background()
	db, e := sql.Open("sqlite", ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	codec, _ := session.NewEncryptedWalletCodec(make([]byte, 32))
	store, e := NewWalletStore(ctx, db, codec)
	if e != nil {
		t.Fatal(e)
	}
	scope := session.WalletScope{ChainID: 1}
	if e = store.Update(ctx, scope, func(v *session.WalletState) error { return nil }); e != nil {
		t.Fatal(e)
	}
	before, _ := store.ReadSnapshot(ctx, scope)
	if e = store.Update(ctx, scope, func(v *session.WalletState) error { v.Records = session.RecordMetadata{}; return nil }); e == nil {
		t.Fatal("metadata omission accepted")
	}
	after, e := store.ReadSnapshot(ctx, scope)
	if e != nil || after.Revision != before.Revision {
		t.Fatal("failed metadata update committed")
	}
}
