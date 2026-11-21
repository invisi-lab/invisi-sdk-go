package session

import (
	"context"
	"errors"
	"testing"

	"github.com/invisi-labs/invisi-sdk-go/credit"
)

func TestManagedCreditSubmissionUsesFreshChainPeriod(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := uint64(2000000000)
	reads := 0
	managed, err := New(Options{Store: store, ChainTimestamp: func(context.Context) (uint64, error) { reads++; return now, nil }})
	if err != nil {
		t.Fatal(err)
	}
	p, grant := testPlan(t)
	if err = managed.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	op, err := managed.Prepare(ctx, "period", p)
	if err != nil {
		t.Fatal(err)
	}
	if err = op.ValidateSubmission(ctx); err != nil {
		t.Fatal(err)
	}
	now = 2000073600 // 2033-05-19T00:00:00Z, next UTC day after preparation.
	if err = op.ValidateSubmission(ctx); !errors.Is(err, credit.ErrCreditPeriodChanged) {
		t.Fatalf("accepted changed chain day: %v", err)
	}
	if reads != 2 {
		t.Fatalf("did not reread latest chain time: %d", reads)
	}
	// Recovery drops process-local capabilities; the session must supply its fresh reader.
	resumed, err := managed.Resume(ctx, "period", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = resumed.ValidateSubmission(ctx); !errors.Is(err, credit.ErrCreditPeriodChanged) {
		t.Fatalf("resume bypassed fresh period: %v", err)
	}
}

func TestManagedCreditSubmissionFailsClosedWithoutChainTimestamp(t *testing.T) {
	ctx := context.Background()
	managed, err := New(Options{Store: NewMemoryStore()})
	if err != nil {
		t.Fatal(err)
	}
	p, grant := testPlan(t)
	if err = managed.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	op, err := managed.Prepare(ctx, "no-reader", p)
	if err != nil {
		t.Fatal(err)
	}
	if err = op.ValidateSubmission(ctx); err == nil {
		t.Fatal("accepted payment without current chain timestamp")
	}
}

func TestExternalizedCreditValidationPreservesReconciliation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	managed, err := New(Options{Store: store, ChainTimestamp: func(context.Context) (uint64, error) { return 0, errors.New("chain unavailable after send") }})
	if err != nil {
		t.Fatal(err)
	}
	p, grant := testPlan(t)
	if err = managed.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	op, err := managed.Prepare(ctx, "sent", p)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Update(ctx, func(s *State) error {
		record := s.Operations[op.id]
		record.Externalized = true
		record.Status = "uncertain"
		s.Operations[op.id] = record
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	managed.validate = func(context.Context, credit.Intent) error { return credit.ErrCreditPeriodChanged }
	if err = op.ValidateSubmission(ctx); err != nil {
		t.Fatalf("existing uncertain operation blocked: %v", err)
	}
}

func TestCachedCreditLocalWorkRetainsFrozenPeriodUntilSubmission(t *testing.T) {
	ctx := context.Background()
	reads := 0
	unavailable := errors.New("chain offline")
	managed, err := New(Options{Store: NewMemoryStore(), ChainTimestamp: func(context.Context) (uint64, error) { reads++; return 0, unavailable }, Prepare: func(context.Context, credit.Intent) (*PreparationEvidence, error) { return &PreparationEvidence{}, nil }, Validate: func(context.Context, credit.Intent) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	// A valid preparation from 2023 remains available for local work independently
	// of the machine's date, while its eventual submission still fails closed.
	p, grant := testPlanAtTimestamp(t, 1700000000)
	if err = managed.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	op, err := managed.Prepare(ctx, "chain-local", p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = op.ExportSigningRequest(ctx); err != nil {
		t.Fatalf("local clock replaced frozen period: %v", err)
	}
	if reads != 0 {
		t.Fatalf("cached local operation read network %d times", reads)
	}
	if err = op.ValidateSubmission(ctx); !errors.Is(err, unavailable) {
		t.Fatalf("offline fallback authorized submission: %v", err)
	}
	if reads != 1 {
		t.Fatalf("submission did not require fresh reader: %d", reads)
	}
}
