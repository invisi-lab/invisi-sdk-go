package session

import (
	"context"
	"math/big"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

type preparationCountingStore struct {
	RecordWalletStore
	snapshots int
}

func (s *preparationCountingStore) ReadSnapshot(ctx context.Context, scope WalletScope) (WalletState, error) {
	s.snapshots++
	return s.RecordWalletStore.ReadSnapshot(ctx, scope)
}
func TestPreparationFrameReadsSummaryInsteadOfEntireInventory(t *testing.T) {
	s, token, _ := testSession(t)
	records, ok := s.options.Store.(*notifyingWalletStore).base.(RecordWalletStore)
	if !ok {
		t.Fatal("fixture requires paged store")
	}
	counted := &preparationCountingStore{RecordWalletStore: records}
	s.options.Store = counted
	address, _ := s.wallet.GetAddress()
	_, err := s.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if err != nil {
		t.Fatal(err)
	}
	if counted.snapshots != 1 {
		t.Fatalf("full inventory reads: %d, want 1", counted.snapshots)
	}
}
