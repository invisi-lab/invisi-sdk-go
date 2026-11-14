package session

import (
	"context"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestImportAndCollectExactInventory(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t)
	before, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	var input sdk.NoteInput
	for _, n := range before.Notes {
		input = sdk.NoteInput{Note: n.Note, Commitment: n.Commitment, TreeID: n.TreeID, LeafIndex: n.LeafIndex}
	}
	if err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		v.Notes = map[common.Hash]StoredNote{}
		v.ObservedThrough = nil
		v.CheckedThrough = nil
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.ImportNotes(ctx, []sdk.NoteInput{input}); err != nil {
		t.Fatal(err)
	}
	b, err := s.Balance(ctx, token)
	if err != nil || b.AvailableKnown == nil || b.AvailableKnown.Int64() != 100 {
		t.Fatal(b, err)
	}
	recipient := common.HexToAddress("0xabcdef")
	op, err := s.PrepareCollection(ctx, token, recipient, []common.Hash{input.Commitment})
	if err != nil {
		t.Fatal(err)
	}
	base := op.plan
	execution := op.execution()
	for _, fee := range []int64{7, 11} {
		if err := execution.ApplyBroadcasterFee(ctx, base, common.HexToAddress("0x123"), big.NewInt(fee), sdk.BroadcasterQuoteOptions{FeeSource: sdk.BroadcasterFeePayout}); err != nil {
			t.Fatal(err)
		}
		if execution.Plan.Params.WithdrawalAmount.Int64() != 100-fee || len(execution.Plan.Outputs) != 0 || len(execution.Plan.Inputs) != 1 || !execution.Plan.NeedsAuthorization {
			t.Fatal("incorrect collection payout")
		}
	}
	if err := s.ImportNotes(ctx, []sdk.NoteInput{input}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.PrepareCollection(ctx, token, recipient, []common.Hash{input.Commitment}); err == nil {
		t.Fatal("reserved note collected twice")
	}
	bad := input
	bad.Commitment = common.HexToHash("0x123")
	if err := s.ImportNotes(ctx, []sdk.NoteInput{bad}); err == nil {
		t.Fatal("invalid opening imported")
	}

}

func TestCollectionSignedQuoteRepricesFromFrozenGross(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	state, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	var keys []common.Hash
	for k := range state.Notes {
		keys = append(keys, k)
	}
	recipient := common.HexToAddress("0xabcdef")
	op, err := s.PrepareCollection(ctx, token, recipient, keys)
	if err != nil {
		t.Fatal(err)
	}
	client, _ := testBroadcaster(t, op, []int64{5, 8, 8, 8}, false)
	if err = client.Quote(ctx, op, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10), FeeSource: sdk.BroadcasterFeePayout}); err != nil {
		t.Fatal(err)
	}
	if op.plan.Params.WithdrawalAmount.Int64() != 92 || op.plan.Params.WithdrawalAddress != recipient || !op.plan.NeedsAuthorization || len(op.plan.Outputs) != 0 {
		t.Fatal("quote changed collection shape")
	}
	resumed, resumeErr := s.Resume(ctx, op.ID())
	if resumeErr != nil {
		t.Fatal(resumeErr)
	}
	if resumed.plan.Params.WithdrawalAmount.Int64() != 92 || len(resumed.plan.Outputs) != 0 || !resumed.plan.NeedsAuthorization {
		t.Fatal("collection recovery changed frozen payout")
	}
	if _, err = resumed.Prove(ctx); err == nil {
		t.Fatal("unsigned cold payout was proved")
	}
}

func TestImportRejectsAdversarialOpeningsAndPreservesSpent(t *testing.T) {
	ctx := context.Background()
	s, _, _ := testSession(t)
	state, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	var input sdk.NoteInput
	for _, n := range state.Notes {
		input = sdk.NoteInput{Note: n.Note, Commitment: n.Commitment, TreeID: n.TreeID, LeafIndex: n.LeafIndex}
	}
	for _, change := range []func(*sdk.NoteInput){
		func(n *sdk.NoteInput) { n.Note.ChainID = primitives.ElementFromUint64(1) },
		func(n *sdk.NoteInput) { n.Note.ProtocolAddress = primitives.ElementFromUint64(1) },
		func(n *sdk.NoteInput) { n.Note.OwnerAddress = primitives.ElementFromUint64(1) },
		func(n *sdk.NoteInput) { n.Note.Amount = primitives.ElementFromUint64(99) },
		func(n *sdk.NoteInput) { n.LeafIndex = 1 },
	} {
		bad := input
		change(&bad)
		if err := s.ImportNotes(ctx, []sdk.NoteInput{bad}); err == nil {
			t.Fatal("adversarial opening accepted")
		}
	}
	if err := s.ImportNotes(ctx, []sdk.NoteInput{input, input}); err == nil {
		t.Fatal("duplicate opening accepted")
	}
	if err := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		n := v.Notes[input.Commitment]
		n.Spent = true
		v.Notes[input.Commitment] = n
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.ImportNotes(ctx, []sdk.NoteInput{input}); err != nil {
		t.Fatal(err)
	}
	after, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if !after.Notes[input.Commitment].Spent {
		t.Fatal("spent note revived")
	}
}

func TestPayoutFeeRejectsChangeFundingAndNonpositiveNet(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t)
	state, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	var keys []common.Hash
	for k := range state.Notes {
		keys = append(keys, k)
	}
	op, err := s.PrepareCollection(ctx, token, common.HexToAddress("0xabcd"), keys)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*transaction.Plan){
		func(p *transaction.Plan) { p.Outputs = append(p.Outputs, transaction.PreparedOutput{}) },
		func(p *transaction.Plan) { p.Funding.Amount = big.NewInt(1) },
		func(p *transaction.Plan) { p.Params.WithdrawalAmount = big.NewInt(5) },
		func(p *transaction.Plan) { p.Kind = "shield" },
		func(p *transaction.Plan) { p.Params.WithdrawalAmount = big.NewInt(99) },
	} {
		base, _ := transaction.ClonePlan(op.plan)
		change(&base)
		if err := op.execution().ApplyBroadcasterFee(ctx, base, common.HexToAddress("0x123"), big.NewInt(5), sdk.BroadcasterQuoteOptions{FeeSource: sdk.BroadcasterFeePayout}); err == nil {
			t.Fatal("invalid payout shape accepted")
		}
	}
}

func TestCollectionDeadlineFrozenAndPersisted(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t)
	state, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	var keys []common.Hash
	for k := range state.Notes {
		keys = append(keys, k)
	}
	option := CollectionOptions{Deadline: 3700}
	if _, err := s.PrepareCollection(ctx, token, common.HexToAddress("0xabcd"), keys, option, option); err == nil {
		t.Fatal("multiple collection options accepted")
	}
	op, err := s.PrepareCollection(ctx, token, common.HexToAddress("0xabcd"), keys, option)
	if err != nil {
		t.Fatal(err)
	}
	option.Deadline = 4000
	resumed, err := s.Resume(ctx, op.ID())
	if err != nil {
		t.Fatal(err)
	}
	if resumed.plan.Deadline != 3700 || resumed.plan.Params.Deadline.Uint64() != 3700 {
		t.Fatal("collection deadline changed")
	}
}
