package sqlite

import (
	"context"
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/session"
	"math/big"
	"testing"
)

func TestPersistentPublicPagesEvictPayloadsKeepCoverageAndReopen(t *testing.T) {
	ctx := context.Background()
	db, e := sql.Open("sqlite", ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	store, e := NewPublicHistoryStore(ctx, db, session.PublicRetention{MaxPages: 1})
	if e != nil {
		t.Fatal(e)
	}
	for _, n := range []uint64{200, 100} {
		b := protocolrpc.CanonicalBlock{Number: n, Hash: common.BigToHash(new(big.Int).SetUint64(n))}
		if e = session.PutPublicPage(ctx, store, "deployment", session.PublicPage{Range: session.BlockRange{FromBlock: n, ToBlock: n}, Block: b, Headers: []protocolrpc.CanonicalBlock{b}}, 0); e != nil {
			t.Fatal(e)
		}
	}
	reopened, e := NewPublicHistoryStore(ctx, db, session.PublicRetention{MaxPages: 1})
	if e != nil {
		t.Fatal(e)
	}
	v, e := reopened.Read(ctx, "deployment")
	if e != nil {
		t.Fatal(e)
	}
	if len(v.Coverage) != 2 || len(v.Pages) != 1 || v.Pages[0].Range.FromBlock != 200 || len(v.Pages[0].Data) != 0 {
		t.Fatal("metadata/payload retention incorrect")
	}
	if _, e = reopened.ReadPayload(ctx, "deployment", v.Pages[0]); e != nil {
		t.Fatal(e)
	}
	var count int
	if e = db.QueryRow("SELECT COUNT(*) FROM invisi_public_history_pages").Scan(&count); e != nil || count != 1 {
		t.Fatal("orphan payload row", count, e)
	}
	if e = reopened.Update(ctx, "deployment", func(v *session.PublicHistoryState) error { v.Pages = nil; return nil }); e != nil {
		t.Fatal(e)
	}
	if e = db.QueryRow("SELECT COUNT(*) FROM invisi_public_history_pages").Scan(&count); e != nil || count != 0 {
		t.Fatal("physical payload pruning failed")
	}
}
