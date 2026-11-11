package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"math/big"
	"reflect"
	"testing"
)

func TestManagedCreditEstimateLeavesJournalUnchanged(t *testing.T) {
	ctx := context.Background()
	original, _ := testPlan(t)
	intent, _ := original.ExportIntent()
	prepared, err := credit.PrepareRevoke(intent.Context, credit.RevokeParams{Token: intent.Context.TokenAddress, Owner: common.HexToAddress("0x2c"), Shard: 1, Deadline: big.NewInt(2000000300)})
	if err != nil {
		t.Fatal(err)
	}
	store := NewMemoryStore()
	s, err := New(Options{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	op, err := s.Prepare(ctx, "estimate", prepared)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := store.ReadSnapshot(ctx)
	got, err := op.EstimateGas()
	after, _ := store.ReadSnapshot(ctx)
	if err != nil || got.Source != "benchmark" || got.Available || got.Reason == "" || !reflect.DeepEqual(before, after) {
		t.Fatal(got, err, "estimate changed journal")
	}
}
