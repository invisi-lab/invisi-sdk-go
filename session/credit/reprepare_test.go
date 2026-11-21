package session

import (
	"context"
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/prover"
	"github.com/invisi-labs/invisi-sdk-go/rpc/broadcaster"
)

type drawRepreparer interface {
	ReprepareDraw(context.Context, credit.DrawPreparation) error
}

func reprepare(t *testing.T, op *Operation, request credit.DrawPreparation) error {
	t.Helper()
	capability, ok := any(op).(drawRepreparer)
	if !ok {
		t.Fatal("managed draw has no safe re-preparation capability")
	}
	return capability.ReprepareDraw(context.Background(), request)
}
func reprepareFixture(t *testing.T) (*Session, *Operation, *MemoryStore, StoredState) {
	t.Helper()
	ctx := context.Background()
	p, g := testPlan(t)
	store := NewMemoryStore()
	s, err := New(Options{Store: store, ChainTimestamp: func(context.Context) (uint64, error) { return 2000073600, nil }, Build: func(context.Context, any) (*credit.Prepared, error) {
		p, _ := testPlanAtTimestamp(t, 2000073600)
		return p, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Track(ctx, g); err != nil {
		t.Fatal(err)
	}
	op, err := s.Prepare(ctx, "draw", p)
	if err != nil {
		t.Fatal(err)
	}
	return s, op, store, g
}
func TestReprepareDrawReplacesStaleReadyJournalWithoutReleasingInputs(t *testing.T) {
	s, op, store, g := reprepareFixture(t)
	ctx := context.Background()
	before, _ := store.ReadSnapshot(ctx)
	oldworker := op.worker
	if err := store.Update(ctx, func(st *State) error {
		r := st.Operations[op.id]
		r.Status = "ready"
		r.Proof = &prover.ProofBundle{}
		r.Calldata = []byte{1}
		r.RevocationSignature = []byte{2}
		r.CancellationConfirmed = true
		st.Operations[op.id] = r
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	preparedBefore := op.prepared
	checks := 0
	s.prepare = func(context.Context, credit.Intent) (*PreparationEvidence, error) {
		return &PreparationEvidence{Generation: 9}, nil
	}
	s.checkPreparation = func(context.Context, PreparationEvidence) error { checks++; return nil }
	s.build = func(ctx context.Context, _ any) (*credit.Prepared, error) {
		st, _ := store.ReadSnapshot(ctx)
		if st.Grants[g.key()].ReservedBy != op.id {
			t.Fatal("reservation released during rebuild")
		}
		p, _ := testPlanAtTimestamp(t, 2000073600)
		return p, nil
	}
	if err := reprepare(t, op, credit.DrawPreparation{}); err != nil {
		t.Fatal(err)
	}
	after, _ := store.ReadSnapshot(ctx)
	r := after.Operations[op.id]
	if r.Status != "prepared" || r.WorkerID == oldworker || r.WorkerID != op.worker || r.Intent.Context.Timestamp != 2000073600 || r.Proof != nil || len(r.Calldata) != 0 || len(r.RevocationSignature) != 0 || r.CancellationConfirmed || r.Externalized || r.BroadcasterRequest != nil || r.BroadcasterRequestID != "" || r.TransactionHash != (common.Hash{}) {
		t.Fatalf("stale artifacts retained: %+v", r)
	}
	if r.Preparation == nil || r.Preparation.Generation != 9 || r.SubmissionValidation != "required" || checks != 2 {
		t.Fatalf("fresh preparation gates missing: %+v checks=%d", r.Preparation, checks)
	}
	if op.prepared == preparedBefore || after.Grants[g.key()].ReservedBy != "draw" || !reflect.DeepEqual(before.Grants[g.key()].Delivery, after.Grants[g.key()].Delivery) {
		t.Fatal("grant reservation or original delivery changed")
	}
	if err := (&Operation{session: s, id: op.id, worker: oldworker, prepared: preparedBefore}).check(ctx); err == nil {
		t.Fatal("old operation handle retained lease")
	}
	if _, err := op.prepared.Proof(); err == nil {
		t.Fatal("old proof survived replacement")
	}
}
func TestReprepareDrawRefusesUnsafeOrUnchangedOperations(t *testing.T) {
	for _, name := range []string{"same period", "missing clock", "clock failure", "externalized no hash", "externalized hash", "hash only", "retained request id", "retained request", "terminal", "lost lease"} {
		t.Run(name, func(t *testing.T) {
			s, op, store, _ := reprepareFixture(t)
			ctx := context.Background()
			calls := 0
			s.build = func(context.Context, any) (*credit.Prepared, error) {
				calls++
				p, _ := testPlanAtTimestamp(t, 2000073600)
				return p, nil
			}
			switch name {
			case "same period":
				s.chainTimestamp = func(context.Context) (uint64, error) { return 2000000001, nil }
			case "missing clock":
				s.chainTimestamp = nil
			case "clock failure":
				s.chainTimestamp = func(context.Context) (uint64, error) { return 0, errors.New("offline") }
			default:
				if err := store.Update(ctx, func(st *State) error {
					r := st.Operations[op.id]
					switch name {
					case "externalized no hash":
						r.Externalized = true
					case "externalized hash":
						r.Externalized = true
						r.TransactionHash = common.HexToHash("1")
					case "hash only":
						r.TransactionHash = common.HexToHash("1")
					case "retained request id":
						r.BroadcasterRequestID = "sent"
					case "retained request":
						r.BroadcasterRequest = &broadcaster.Submission{Data: "0x010203", Signature: "retained", Quote: broadcaster.Quote{Execution: "direct", Deadline: "1"}}
					case "terminal":
						r.Status = "failed"
					case "lost lease":
						r.WorkerID = "other"
					}
					st.Operations[op.id] = r
					return nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := store.ReadSnapshot(ctx)
			old := op.prepared
			if err := reprepare(t, op, credit.DrawPreparation{}); err == nil {
				t.Fatal("accepted unsafe replacement")
			}
			after, _ := store.ReadSnapshot(ctx)
			if calls != 0 || old != op.prepared || !reflect.DeepEqual(before, after) {
				t.Fatal("refusal mutated journal or built replacement")
			}
		})
	}
}

type reprepareFailStore struct {
	Store
	fail bool
}

func (s *reprepareFailStore) Update(ctx context.Context, change func(*State) error) error {
	if s.fail {
		return errors.New("storage unavailable")
	}
	return s.Store.Update(ctx, change)
}

func TestReprepareDrawFailuresPreserveFrozenJournalAndReservations(t *testing.T) {
	for _, name := range []string{"build", "prepare evidence", "first check", "second check", "store", "input mismatch", "lease race", "journal race", "consumed input", "revoked input", "reservation race"} {
		t.Run(name, func(t *testing.T) {
			s, op, store, g := reprepareFixture(t)
			ctx := context.Background()
			before, _ := store.ReadSnapshot(ctx)
			oldworker, oldprepared := op.worker, op.prepared
			s.build = func(ctx context.Context, _ any) (*credit.Prepared, error) {
				if PreparationReservation(ctx) != op.id {
					t.Fatal("missing reservation capability")
				}
				if name == "build" {
					return nil, errors.New("build unavailable")
				}
				if name == "lease race" || name == "journal race" || name == "consumed input" || name == "revoked input" || name == "reservation race" {
					err := store.Update(ctx, func(st *State) error {
						r := st.Operations[op.id]
						grant := st.Grants[g.key()]
						switch name {
						case "lease race":
							r.WorkerID = "resumed"
						case "journal race":
							r.Calldata = []byte{9}
						case "consumed input":
							grant.Consumed = true
						case "revoked input":
							grant.GrantStatus = 2
							grant.Revoked = true
						case "reservation race":
							grant.ReservedBy = "other"
						}
						st.Operations[op.id] = r
						st.Grants[g.key()] = grant
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
				}
				p, _ := testPlanAtTimestamp(t, 2000073600)
				if name == "input mismatch" {
					i, err := p.ExportIntent()
					if err != nil {
						t.Fatal(err)
					}
					i.Inputs[0].State.Nonce = primitives.ElementFromUint64(2)
					tr, err := i.Inputs[0].State.Draw(i.Inputs[0].Debit, i.Context.Timestamp)
					if err != nil {
						t.Fatal(err)
					}
					i.Draw.Nullifiers[0] = common.Hash(tr.Nullifier.Bytes())
					i.Draw.CreditCommitments[0] = common.Hash(tr.StateCommitment.Bytes())
					i.Draw.RecoveryAmounts[0] = common.Hash(tr.RecoveryAmount.Bytes())
					p, err = credit.PrepareDraw(i.Context, *i.Draw, i.Deliveries, i.Inputs, i.OutputIdentity)
					if err != nil {
						t.Fatal(err)
					}
				}
				return p, nil
			}
			if name == "prepare evidence" {
				s.prepare = func(context.Context, credit.Intent) (*PreparationEvidence, error) {
					return nil, errors.New("evidence unavailable")
				}
			}
			if name == "first check" || name == "second check" {
				s.prepare = func(context.Context, credit.Intent) (*PreparationEvidence, error) { return &PreparationEvidence{}, nil }
				calls := 0
				s.checkPreparation = func(context.Context, PreparationEvidence) error {
					calls++
					if name == "first check" || calls == 2 {
						return errors.New("fork changed")
					}
					return nil
				}
			}
			if name == "store" {
				s.store = &reprepareFailStore{Store: store, fail: true}
			}
			if err := reprepare(t, op, credit.DrawPreparation{}); err == nil {
				t.Fatal("replacement accepted failing dependency or race")
			}
			after, _ := store.ReadSnapshot(ctx)
			if op.worker != oldworker || op.prepared != oldprepared {
				t.Fatal("failed transaction changed live handle")
			}
			switch name {
			case "lease race", "journal race", "consumed input", "revoked input", "reservation race":
				if after.Operations[op.id].Intent.Context.Timestamp != before.Operations[op.id].Intent.Context.Timestamp {
					t.Fatal("concurrent state replaced")
				}
			default:
				if !reflect.DeepEqual(before, after) {
					t.Fatal("failed replacement changed journal or reservations")
				}
			}
		})
	}
}

func TestReprepareDrawFreezesRequestBeforeChainCallbacks(t *testing.T) {
	s, op, _, _ := reprepareFixture(t)
	request := credit.DrawPreparation{Credits: []credit.DrawMember{{Amount: big.NewInt(7)}}}
	s.chainTimestamp = func(context.Context) (uint64, error) { request.Credits[0].Amount.SetInt64(99); return 2000073600, nil }
	s.build = func(_ context.Context, value any) (*credit.Prepared, error) {
		frozen := value.(credit.DrawPreparation)
		if frozen.Credits[0].Amount.Int64() != 7 {
			t.Fatalf("callback changed frozen request amount to %s", frozen.Credits[0].Amount)
		}
		p, _ := testPlanAtTimestamp(t, 2000073600)
		return p, nil
	}
	if err := op.ReprepareDraw(context.Background(), request); err != nil {
		t.Fatal(err)
	}
}

func TestReprepareDrawPermitsSamePeriodExpiredDeadlineAndRefusal(t *testing.T) {
	for _, reason := range []string{"deadline", "quote", "refusal"} {
		t.Run(reason, func(t *testing.T) {
			s, op, store, g := reprepareFixture(t)
			now := uint64(2000000001)
			if reason == "deadline" {
				now = 2000000301
			}
			s.chainTimestamp = func(context.Context) (uint64, error) { return now, nil }
			s.build = func(context.Context, any) (*credit.Prepared, error) {
				p, _ := testPlanAtTimestamp(t, now)
				return p, nil
			}
			if err := store.Update(context.Background(), func(st *State) error {
				r := st.Operations[op.id]
				if reason == "refusal" {
					r.SubmissionRefusal = "CREDIT_PERIOD_ENDING"
				}
				if reason == "quote" {
					r.Intent.Broadcaster = &credit.BroadcasterBinding{Quote: broadcaster.SignedQuote{Quote: broadcaster.Quote{Deadline: "1", Execution: "direct"}}}
				}
				st.Operations[op.id] = r
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if err := op.ReprepareDraw(context.Background(), credit.DrawPreparation{}); err != nil {
				t.Fatal(err)
			}
			state, _ := store.ReadSnapshot(context.Background())
			r := state.Operations[op.id]
			if r.Intent.Context.Timestamp != now || r.Intent.Broadcaster != nil || r.SubmissionRefusal != "" || state.Grants[g.key()].ReservedBy != op.id {
				t.Fatal("refresh lost reservation or stale evidence survived")
			}
		})
	}
}

func TestAbandonExpiredKnownUnsentReadyOperation(t *testing.T) {
	for _, mode := range []string{"expired", "live", "unknown", "offline", "quote-only"} {
		t.Run(mode, func(t *testing.T) {
			s, op, store, g := reprepareFixture(t)
			now := uint64(2000000301)
			if mode == "live" || mode == "quote-only" {
				now = 2000000001
			}
			s.chainTimestamp = func(context.Context) (uint64, error) {
				if mode == "offline" {
					return 0, errors.New("offline")
				}
				return now, nil
			}
			if err := store.Update(context.Background(), func(st *State) error {
				r := st.Operations[op.id]
				r.Status = "ready"
				r.Calldata = []byte{1}
				if mode == "quote-only" {
					r.Intent.Broadcaster = &credit.BroadcasterBinding{Quote: broadcaster.SignedQuote{Quote: broadcaster.Quote{Deadline: "1", Execution: "direct"}}}
				}
				if mode == "unknown" {
					r.Externalized = true
				}
				st.Operations[op.id] = r
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			err := op.Abandon(context.Background())
			if (err == nil) != (mode == "expired") {
				t.Fatalf("abandon %s: %v", mode, err)
			}
			state, _ := store.ReadSnapshot(context.Background())
			if (state.Grants[g.key()].ReservedBy == "") != (mode == "expired") {
				t.Fatal("unsafe reservation release")
			}
		})
	}
}
