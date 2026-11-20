package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/session"
)

func TestReplayCheckpointPreservesEncryptedRecordsAndFencesRevision(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	codec, err := session.NewEncryptedWalletCodec(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewWalletStore(ctx, db, codec)
	if err != nil {
		t.Fatal(err)
	}
	scope := session.WalletScope{ChainID: 97}
	hash := common.HexToHash("0x12")
	if err = store.Update(ctx, scope, func(v *session.WalletState) error {
		v.Notes[hash] = session.StoredNote{Commitment: hash, ReservedBy: "pending"}
		v.Scan.CreditReplayFrom = 10
		v.Scan.CreditReplayTarget = 100
		v.Scan.CreditReplayRevision = 3
		v.Scan.Generation = 2
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var recordBefore []byte
	if err = db.QueryRow("SELECT data FROM invisi_wallet_records").Scan(&recordBefore); err != nil {
		t.Fatal(err)
	}
	checkpoint := session.CreditReplayCheckpoint{Revision: 1, Generation: 2, ReplayRevision: 3, Target: 100, Cursor: 30}
	if err = store.CommitCreditReplayCheckpoint(ctx, scope, checkpoint); err != nil {
		t.Fatal(err)
	}
	var recordAfter []byte
	if err = db.QueryRow("SELECT data FROM invisi_wallet_records").Scan(&recordAfter); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(recordBefore, recordAfter) {
		t.Fatal("checkpoint rewrote encrypted private record")
	}
	reopened, err := NewWalletStore(ctx, db, codec)
	if err != nil {
		t.Fatal(err)
	}
	state, err := reopened.ReadSnapshot(ctx, scope)
	if err != nil {
		t.Fatal(err)
	}
	if state.Revision != 2 || state.Scan.CreditReplayCursor != 30 || state.Notes[hash].ReservedBy != "pending" {
		t.Fatal("durable checkpoint lost state")
	}
	if err = store.CommitCreditReplayCheckpoint(ctx, scope, checkpoint); err == nil {
		t.Fatal("stale revision accepted")
	}
	checkpoint.Revision = 2
	checkpoint.Generation = 1
	if err = store.CommitCreditReplayCheckpoint(ctx, scope, checkpoint); err == nil {
		t.Fatal("stale generation accepted")
	}
	checkpoint.Generation = 2
	checkpoint.ReplayRevision = 2
	if err = store.CommitCreditReplayCheckpoint(ctx, scope, checkpoint); err == nil {
		t.Fatal("stale replay revision accepted")
	}
	checkpoint.ReplayRevision = 3
	checkpoint.Cursor = 100
	if err = store.CommitCreditReplayCheckpoint(ctx, scope, checkpoint); err == nil {
		t.Fatal("metadata shortcut may not finish replay")
	}
	checkpoint.Cursor = 29
	if err = store.CommitCreditReplayCheckpoint(ctx, scope, checkpoint); err == nil {
		t.Fatal("checkpoint moved backwards")
	}
}

func TestReplayCheckpointNeverHidesCorruptRecordsAtFullRead(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	codec, err := session.NewEncryptedWalletCodec(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewWalletStore(ctx, db, codec)
	if err != nil {
		t.Fatal(err)
	}
	scope := session.WalletScope{ChainID: 97}
	hash := common.HexToHash("0x44")
	if err = store.Update(ctx, scope, func(v *session.WalletState) error {
		v.Notes[hash] = session.StoredNote{Commitment: hash}
		v.Scan.CreditReplayTarget = 100
		v.Scan.CreditReplayRevision = 1
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("DELETE FROM invisi_wallet_records"); err != nil {
		t.Fatal(err)
	}
	if err = store.CommitCreditReplayCheckpoint(ctx, scope, session.CreditReplayCheckpoint{Revision: 1, ReplayRevision: 1, Target: 100, Cursor: 10}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.ReadSnapshot(ctx, scope); err == nil {
		t.Fatal("full validation accepted missing authenticated record")
	}
	if err = store.CommitCreditReplayCheckpoint(ctx, scope, session.CreditReplayCheckpoint{Revision: 2, ReplayRevision: 1, Target: 100, Cursor: 100}); err == nil {
		t.Fatal("corrupt wallet completed via checkpoint")
	}
}
