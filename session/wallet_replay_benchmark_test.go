package session

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-labs/invisi-sdk-go/protocol/receipts"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/wallet"
)

// This fixture deliberately contains another wallet's Credit transition. Both
// versions validate the retained canonical page; only the old version schedules
// deployment-wide private replay. Copy this file and the testing.TB fixture
// signature changes to the baseline revision for a directly comparable run.
func BenchmarkWalletRetainedForeignCredit(b *testing.B) {
	s, _ := sharedEmptyFixture(b)
	ctx := context.Background()
	head := s.options.Provider.(historyChain).headers[10]
	head.Difficulty = big.NewInt(0)
	block := protocolrpc.CanonicalBlock{Number: 10, Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	page := PublicPage{Range: BlockRange{10, 10}, Block: block, Headers: []protocolrpc.CanonicalBlock{block}, Transactions: []PublicTransaction{{
		PositionValidated: true, Header: head, Receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful, BlockNumber: head.Number, BlockHash: head.Hash(), Logs: []*types.Log{}},
		Batch: receipts.Batch{CreditNullifiers: []common.Hash{common.HexToHash("0x999")}},
	}}}
	if err := PutPublicPage(ctx, s.publicHistory.Store(), s.publicHistory.Key(), page, 0); err != nil {
		b.Fatal(err)
	}
	// Warm every historical page to measure repeated retained-history replay rather
	// than letting network download only the baseline's missing old pages.
	for from := uint64(1); from < 10; from += 3 {
		if _, err := s.publicHistory.Page(ctx, BlockRange{from, from + 2}, false); err != nil {
			b.Fatal(err)
		}
	}
	verify, err := s.publicHistory.Page(ctx, BlockRange{10, 10}, true)
	if err != nil {
		b.Fatal(err)
	}
	if len(verify.Transactions) != 1 || len(verify.Transactions[0].Batch.CreditNullifiers) != 1 {
		b.Fatalf("invalid retained foreign fixture %+v", verify)
	}
	from := uint64(10)
	replays := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		s.options.Store = NewMemoryWalletStore()
		b.StartTimer()
		_, err := s.ScanHistory(ctx, wallet.DeliveryDecryptorFuncs{}, ScanOptions{FromBlock: &from, PageSize: 3, SkipStateRefresh: true, OnProgress: func(p HistoryProgress) {
			if p.Direction == "credit-recovery" {
				replays++
			}
		}})
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(replays)/float64(b.N), "replay-pages/op")
	b.ReportAllocs()
}
