package provenance

import (
	"bytes"
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestEncryptedRecoverySQLiteRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "journal.db")
	db, e := sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	records, e := NewSQLiteRecoveryRecords(ctx, db)
	if e != nil {
		t.Fatal(e)
	}
	encrypted, _ := NewEncryptedRecoveryStore(records, bytes.Repeat([]byte{7}, 32))
	record := RecoveryRecord{Revision: 1, ID: "one", Intent: FrozenProvenanceIntent{Calldata: "0xfeed"}}
	if ok, e := encrypted.CompareAndSwap(ctx, "one", 0, record); e != nil || !ok {
		t.Fatal(e)
	}
	db.Close()
	db, e = sql.Open("sqlite", path)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	records, e = NewSQLiteRecoveryRecords(ctx, db)
	if e != nil {
		t.Fatal(e)
	}
	encrypted, _ = NewEncryptedRecoveryStore(records, bytes.Repeat([]byte{7}, 32))
	loaded, found, e := encrypted.Load(ctx, "one")
	if e != nil || !found || loaded.Intent.Calldata != "0xfeed" {
		t.Fatal(e)
	}
	if ok, _ := encrypted.CompareAndSwap(ctx, "one", 0, record); ok {
		t.Fatal("stale revision committed after restart")
	}
	var data []byte
	if e = db.QueryRow("SELECT data FROM invisi_provenance_recovery").Scan(&data); e != nil || bytes.Contains(data, []byte("0xfeed")) {
		t.Fatal("plaintext or invalid ciphertext", e)
	}
}
