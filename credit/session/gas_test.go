package session

import (
	"bytes"
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"math/big"
	"testing"
)

type managedEstimator struct{ call ethereum.CallMsg }

func (p *managedEstimator) EstimateGas(_ context.Context, call ethereum.CallMsg) (uint64, error) {
	p.call = call
	return 12345, nil
}
func TestManagedCreditEstimateUsesCalldataJournal(t *testing.T) {
	ctx := context.Background()
	original, _ := testPlan(t)
	intent, _ := original.ExportIntent()
	owner := common.HexToAddress("0x2c")
	prepared, err := credit.PrepareRevoke(intent.Context, credit.RevokeParams{Token: intent.Context.TokenAddress, Owner: owner, Shard: 1, Deadline: big.NewInt(2000000300)})
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
	if _, err = op.EstimateGas(ctx, nil, owner); err == nil {
		t.Fatal("missing estimator accepted")
	}
	before, _ := store.ReadSnapshot(ctx)
	if before.Operations[op.id].Status == "ready" {
		t.Fatal("missing estimator changed journal")
	}
	provider := &managedEstimator{}
	result, err := op.EstimateGas(ctx, provider, owner)
	if err != nil {
		t.Fatal(err)
	}
	state, err := store.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	record := state.Operations[op.id]
	if result.Source != "rpc" || result.GasUnits != 12345 || record.Status != "ready" || !bytes.Equal(record.Calldata, provider.call.Data) || provider.call.From != owner || provider.call.To == nil || *provider.call.To != common.BigToAddress(core.ElementToBigInt(intent.Context.ProtocolAddress)) {
		t.Fatal("managed estimate bypassed final calldata", result)
	}
}
