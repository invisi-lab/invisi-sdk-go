package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"testing"
	"time"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestClientQuoteRejectsCopiedAndNilPreparedWithoutNetwork(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := session.wallet.GetAddress()
	operation, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if err != nil {
		t.Fatal(err)
	}
	state, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	prepared := detachedForTest(t, operation, state)
	client, calls := testBroadcaster(t, operation, []int64{5}, false)
	copied := *prepared
	options := sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)}
	if err = client.Quote(ctx, &copied, options); err == nil {
		t.Fatal("copied handle accepted")
	}
	var absent *sdk.PreparedTransaction
	if err = client.Quote(ctx, absent, options); err == nil {
		t.Fatal("typed nil prepared accepted")
	}
	var absentSession *Operation
	if err = client.Quote(ctx, absentSession, options); err == nil {
		t.Fatal("typed nil session operation accepted")
	}
	if calls.Load() != 0 || prepared.Preview().FeeAmount.Sign() != 0 {
		t.Fatal("invalid target changed original or called the network")
	}
	if err = client.Quote(ctx, prepared, options); err != nil {
		t.Fatal(err)
	}
	if prepared.Preview().FeeAmount.Int64() != 5 {
		t.Fatal("original handle did not receive quote")
	}
	if transaction.PreparedIsCurrent(&copied) {
		t.Fatal("old prepared generation remained current")
	}
}

func TestClientQuoteResumeBindsCurrentSessionWorker(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := session.wallet.GetAddress()
	original, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := session.Resume(ctx, original.ID())
	if err != nil {
		t.Fatal(err)
	}
	client, calls := testBroadcaster(t, resumed, []int64{5}, false)
	options := sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)}
	if err = client.Quote(ctx, original, options); err == nil || calls.Load() != 0 {
		t.Fatal("stale worker reached quotation", err)
	}
	if err = client.Quote(ctx, resumed, options); err != nil {
		t.Fatal(err)
	}
	if resumed.plan.Broadcaster == nil || original.plan.Broadcaster != nil {
		t.Fatal("resumed quote changed wrong worker")
	}
	state, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	record := state.Operations[resumed.ID()]
	if record.WorkerID != resumed.workerID {
		t.Fatal("quote did not install current worker")
	}
	var saved transaction.Plan
	if err = json.Unmarshal(record.Plan, &saved); err != nil || saved.Broadcaster == nil {
		t.Fatal("quote was not durably saved", err)
	}
}

type rejectQuoteStore struct{ WalletStore }

func (s rejectQuoteStore) Update(ctx context.Context, scope WalletScope, update func(*WalletState) error) error {
	return s.WalletStore.Update(ctx, scope, func(state *WalletState) error {
		if err := update(state); err != nil {
			return err
		}
		for _, record := range state.Operations {
			var plan transaction.Plan
			if err := json.Unmarshal(record.Plan, &plan); err != nil {
				return err
			}
			if plan.Broadcaster != nil {
				return errors.New("quote persistence unavailable")
			}
		}
		return nil
	})
}
func TestClientQuoteStoreFailureRetainsOriginalAndReservations(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := session.wallet.GetAddress()
	original, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if err != nil {
		t.Fatal(err)
	}
	before, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	session.options.Store = rejectQuoteStore{session.options.Store}
	client, _ := testBroadcaster(t, original, []int64{5}, false)
	if err = client.Quote(ctx, original, broadcaster.QuoteOptions{MaxFeeAmount: big.NewInt(10)}); err == nil {
		t.Fatal("quote succeeded without persistence")
	}
	if original.plan.Broadcaster != nil || original.Preview().FeeAmount.Sign() != 0 {
		t.Fatal("failed quote changed original handle")
	}
	if err = original.checkLease(ctx); err != nil {
		t.Fatal("failed quote lost original worker", err)
	}
	after, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before.Operations[original.ID()].Plan, after.Operations[original.ID()].Plan) {
		t.Fatal("failed quote changed durable original plan")
	}
	for commitment, note := range before.Notes {
		if note.ReservedBy != after.Notes[commitment].ReservedBy {
			t.Fatal("failed quote changed reservations")
		}
	}
}
