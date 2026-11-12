package session

import (
	"context"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
	"testing"
	"time"
)

func TestHistoryReflectsSharedHistoryReset(t *testing.T) {
	ctx := context.Background()
	s, _, _ := testSession(t)
	if err := s.publicHistory.options.Store.Update(ctx, s.publicHistory.key, func(v *PublicHistoryState) error { v.Generation++; return nil }); err != nil {
		t.Fatal(err)
	}
	h, err := s.GetHistory(ctx, HistoryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !h.Invalidated {
		t.Fatalf("full history returned invalidated=false after public generation reset; entries=%d", len(h.Incoming))
	}
}
func TestHistoryPageMarksInvalidatedRecordsUnknown(t *testing.T) {
	ctx := context.Background()
	s, _, _ := testSession(t)
	if err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error { v.Invalidated = true; return nil }); err != nil {
		t.Fatal(err)
	}
	page, err := s.GetHistoryPage(ctx, HistoryPageOptions{Direction: "incoming", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range page.Items {
		if v.Canonical != "unknown" {
			t.Fatalf("invalidated page marked entry %q", v.Canonical)
		}
	}
}

func TestAbandonReleasesUnexternalizedAuthorizedAndReadyInputs(t *testing.T) {
	for _, stage := range []string{"authorized", "ready"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t, 2000000000)
			address, _ := s.wallet.GetAddress()
			op, err := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
			if err != nil {
				t.Fatal(err)
			}
			request, err := op.ExportWalletAuthorization(ctx)
			if err != nil {
				t.Fatal(err)
			}
			signature, err := signFixtureRequest(request)
			if err != nil {
				t.Fatal(err)
			}
			op, err = op.WithWalletSignature(ctx, signature)
			if err != nil {
				t.Fatal(err)
			}
			if stage == "ready" {
				op, err = op.Prove(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err = op.Abandon(ctx); err != nil {
				t.Fatal("unexternalized operation cannot release inputs:", err)
			}
			state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
			if err != nil {
				t.Fatal(err)
			}
			for _, note := range state.Notes {
				if note.ReservedBy != "" {
					t.Fatal("abandon retained an input reservation")
				}
			}
			if state.Operations[op.plan.ID].Status != "failed" {
				t.Fatal("abandon status differs from TS")
			}
			if _, err = op.ExportWalletAuthorization(ctx); err == nil {
				t.Fatal("abandoned handle retained signing authority")
			}

		})
	}
}

func TestAbandonedReadyOperationsCannotResumeRetainedSubmission(t *testing.T) {
	for _, relay := range []bool{false, true} {
		t.Run(fmt.Sprint(relay), func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t, uint64(time.Now().Unix()))
			address, _ := s.wallet.GetAddress()
			op, err := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
			if err != nil {
				t.Fatal(err)
			}
			if relay {
				client, _ := testBroadcaster(t, op, []int64{5}, false)
				op, err = op.WithBroadcaster(ctx, client, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
				if err != nil {
					t.Fatal(err)
				}
			}
			op, err = authorizeForTest(op, ctx)
			if err != nil {
				t.Fatal(err)
			}
			op, err = op.Prove(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if err = op.Abandon(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err = s.Resume(ctx, op.plan.ID); err == nil {
				t.Fatal("abandoned proof/calldata can be resumed for a submission retry")
			}
			state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
			if err != nil {
				t.Fatal(err)
			}
			record := state.Operations[op.plan.ID]
			if (len(record.Proof) != 0 && string(record.Proof) != "null") || len(record.CallData) != 0 {
				t.Fatal("abandoned journal retained executable proof/calldata")
			}
		})
	}
}
