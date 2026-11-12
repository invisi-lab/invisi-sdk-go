package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"math/big"
	"testing"
)

func TestCachedCreditJournalKeepsLocalWorkOfflineAndGatesSubmission(t *testing.T) {
	ctx := context.Background()
	prepared, grant := testPlan(t)
	store := NewMemoryStore()
	live := 0
	s, e := New(Options{Store: store, Prepare: func(context.Context, credit.Intent) (*PreparationEvidence, error) { return &PreparationEvidence{}, nil }, Validate: func(context.Context, credit.Intent) error { live++; return errors.New("live input consumed") }})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Track(ctx, grant); e != nil {
		t.Fatal(e)
	}
	op, e := s.Prepare(ctx, "offline", prepared)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = op.ExportSigningRequest(ctx); e != nil {
		t.Fatal(e)
	}
	state, _ := store.ReadSnapshot(ctx)
	if state.Operations["offline"].SubmissionValidation != "required" {
		t.Fatal("missing pending live validation")
	}
	resumed, e := s.Resume(ctx, "offline", nil)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = resumed.ExportSigningRequest(ctx); e != nil {
		t.Fatal(e)
	}
	if live != 0 {
		t.Fatal("local signing used live reader")
	}
	if e = resumed.ValidateSubmission(ctx); e == nil {
		t.Fatal("submission bypassed current state")
	}
	if e = resumed.Externalize(ctx, common.Hash{}); e == nil {
		t.Fatal("externalization bypassed current state")
	}
	if live != 2 {
		t.Fatalf("expected two live gates, got %d", live)
	}
}
func TestCachedRevokeCalldataRemainsOffline(t *testing.T) {
	ctx := context.Background()
	_, grant := testPlan(t)
	state := grant.Delivery.State
	c := credit.Context{ProtocolContext: state.ProtocolContext, TokenAddress: common.BigToAddress(primitives.ElementToBigInt(state.TokenAddress)), Timestamp: 2000000000}
	p, e := credit.PrepareRevoke(c, credit.RevokeParams{Token: c.TokenAddress, Owner: common.BigToAddress(primitives.ElementToBigInt(state.OwnerWalletIdentity)), Shard: 1, Deadline: big.NewInt(2000000300)})
	if e != nil {
		t.Fatal(e)
	}
	s, _ := New(Options{Store: NewMemoryStore(), Prepare: func(context.Context, credit.Intent) (*PreparationEvidence, error) { return &PreparationEvidence{}, nil }, Validate: func(context.Context, credit.Intent) error { return errors.New("offline") }})
	op, e := s.Prepare(ctx, "revoke", p)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = op.Calldata(ctx, common.BigToAddress(primitives.ElementToBigInt(state.OwnerWalletIdentity))); e != nil {
		t.Fatal(e)
	}
	if e = op.Externalize(ctx, common.Hash{}); e == nil {
		t.Fatal("offline externalization passed")
	}
}

func TestCachedCreditMissingAnchorFailsClosed(t *testing.T) {
	ctx := context.Background()
	prepared, grant := testPlan(t)
	store := NewMemoryStore()
	s, _ := New(Options{Store: store, Prepare: func(context.Context, credit.Intent) (*PreparationEvidence, error) { return &PreparationEvidence{}, nil }, Validate: func(context.Context, credit.Intent) error { return nil }})
	if err := s.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	op, err := s.Prepare(ctx, "missing-anchor", prepared)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Update(ctx, func(v *State) error {
		record := v.Operations["missing-anchor"]
		record.Preparation = nil
		v.Operations["missing-anchor"] = record
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = op.SigningRequest(ctx); err == nil {
		t.Fatal("signed unanchored cached Credit")
	}
	if err = op.ValidateSubmission(ctx); err == nil {
		t.Fatal("submitted unanchored cached Credit")
	}
}
