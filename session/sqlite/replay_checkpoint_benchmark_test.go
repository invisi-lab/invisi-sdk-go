package sqlite

import (
	"context"
	"database/sql"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/session"
)

// Compare the previous full-record Update path and the optional cursor-only
// capability on the same encrypted SQLite adapter and retained inventory.
func BenchmarkReplayCheckpoint(b *testing.B) {
	for _, mode := range []string{"FullRecordUpdate", "MetadataOnly"} {
		b.Run(mode, func(b *testing.B) {
			ctx := context.Background()
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()
			codec, err := session.NewEncryptedWalletCodec(make([]byte, 32))
			if err != nil {
				b.Fatal(err)
			}
			store, err := NewWalletStore(ctx, db, codec)
			if err != nil {
				b.Fatal(err)
			}
			scope := session.WalletScope{ChainID: 97}
			if err = store.Update(ctx, scope, func(v *session.WalletState) error {
				v.Scan.CreditReplayTarget = 1 << 40
				v.Scan.CreditReplayRevision = 1
				for n := int64(1); n <= 512; n++ {
					hash := common.BigToHash(big.NewInt(n))
					v.Notes[hash] = session.StoredNote{Commitment: hash, ReservedBy: "externalized-pending"}
				}
				return nil
			}); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				checkpoint := session.CreditReplayCheckpoint{Revision: uint64(i + 1), ReplayRevision: 1, Target: 1 << 40, Cursor: uint64(i + 1)}
				if mode == "MetadataOnly" {
					err = store.CommitCreditReplayCheckpoint(ctx, scope, checkpoint)
				} else {
					err = store.Update(ctx, scope, checkpoint.Apply)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(512, "notes")
			b.ReportAllocs()
		})
	}
}
