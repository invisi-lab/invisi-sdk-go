package session

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func seedLimitedCredit(t *testing.T, s *WalletSession, token common.Address, states ...core.CreditSpendingState) common.Hash {
	t.Helper()
	old := seedCredit(t, s, token)
	var leaf common.Hash
	err := s.options.Store.Update(context.Background(), s.scope, func(snapshot *WalletState) error {
		for _, role := range []string{"owner", "spender"} {
			entry := snapshot.Credits[transaction.CreditKey(old, role)]
			delete(snapshot.Credits, transaction.CreditKey(old, role))
			entry.State.Note.CreditSpendingState = core.CreditSpendingState{SingleLimit: core.ElementFromUint64(20), DailyLimit: core.ElementFromUint64(40), MonthlyLimit: core.ElementFromUint64(60), LastDay: core.ElementFromUint64(1), DailySpent: core.ElementFromUint64(10), LastMonth: core.ElementFromUint64(1), MonthlySpent: core.ElementFromUint64(15)}
			if len(states) > 0 {
				entry.State.Note.CreditSpendingState = states[0]
			}
			entry.State.Note.Expiry = core.ElementFromUint64(200000)
			var e error
			leaf, e = transaction.CreditCommitment(entry.State.Note)
			if e != nil {
				return e
			}
			nf, e := core.ComputeCreditNullifier(entry.State.Note.ChainID, entry.State.Note.CreditSharedKey, core.BigIntToElement(leaf.Big()))
			if e != nil {
				return e
			}
			entry.Commitment = leaf
			entry.Nullifier = transaction.FieldHash(nf)
			snapshot.Credits[transaction.CreditKey(leaf, role)] = entry
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	s.options.Provider.(sessionProvider).leaves.Store(uint64(2), leaf)
	return leaf
}
func TestLimitedCreditPreparationPreservesStateAndChecksFullConsumption(t *testing.T) {
	ctx := context.Background()
	for _, kind := range []string{"partial", "full", "top-up"} {
		t.Run(kind, func(t *testing.T) {
			s, token, _ := testSession(t)
			leaf := seedLimitedCredit(t, s, token)
			if kind == "full" {
				if _, err := s.PrepareDrawCredit(ctx, sdk.DrawCreditPreparation{TokenAddress: token, Credits: []sdk.CreditDrawInput{{Credit: leaf, Amount: big.NewInt(100)}}}); err == nil || !strings.Contains(err.Error(), "single limit") {
					t.Fatal("full draw bypassed limit", err)
				}
				return
			}
			var op *Operation
			var err error
			if kind == "partial" {
				op, err = s.PrepareDrawCredit(ctx, sdk.DrawCreditPreparation{TokenAddress: token, Credits: []sdk.CreditDrawInput{{Credit: leaf, Amount: big.NewInt(15)}}})
			} else {
				op, err = s.PrepareTopUpCredit(ctx, sdk.TopUpCreditPreparation{TokenAddress: token, Credit: leaf, AdditionalAmount: big.NewInt(10)})
			}
			if err != nil {
				t.Fatal(err)
			}
			if kind == "top-up" {
				if op.plan.CreditOutput.Note.CreditSpendingState != op.plan.CreditInputs[0].State.Note.CreditSpendingState {
					t.Fatal("top-up reset state")
				}
			}
			if op, err = op.Authorize(ctx); err != nil {
				t.Fatal(err)
			}
			private, err := op.privateInputs()
			if err != nil {
				t.Fatal(err)
			}
			if kind == "partial" {
				words, e := op.words()
				if e != nil || len(words) != 20 || new(big.Int).SetBytes(words[16][:]).Uint64() != 1 || new(big.Int).SetBytes(words[17][:]).Uint64() != 1 || private["daily_spents"].([]string)[0] != "10" {
					t.Fatal("wrong period/private witness", e)
				}
			} else if private["daily_spent"] != "10" || private["monthly_spent"] != "15" {
				t.Fatal("top-up witness reset usage")
			}
			if _, err = op.Prove(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestPreparedDrawRejectsUtcPeriodChangeAtEveryStage(t *testing.T) {
	for _, stage := range []string{"authorize", "prove", "export"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t)
			leaf := seedLimitedCredit(t, s, token)
			op, err := s.PrepareDrawCredit(ctx, sdk.DrawCreditPreparation{TokenAddress: token, Deadline: 200000, Credits: []sdk.CreditDrawInput{{Credit: leaf, Amount: big.NewInt(10)}}})
			if err != nil {
				t.Fatal(err)
			}
			if stage != "authorize" {
				op, err = op.Authorize(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}
			if stage == "export" {
				op, err = op.Prove(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}
			s.options.Provider.(sessionProvider).head.Time = 86400
			switch stage {
			case "authorize":
				_, err = op.Authorize(ctx)
			case "prove":
				_, err = op.Prove(ctx)
			case "export":
				_, err = op.ExportTransaction(ctx)
			}
			if err == nil || (!strings.Contains(err.Error(), "UTC") && !strings.Contains(err.Error(), "expired")) {
				t.Fatal("stale period reused", err)
			}
		})
	}
}

func TestRealNativeLimitedCreditOperations(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set pinned BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine limited Credit proofs")
	}
	for _, kind := range []string{"create", "draw", "top-up", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t)
			backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
			if err != nil {
				t.Fatal(err)
			}
			s.options.Prover = backend
			var op *Operation
			if kind == "create" {
				address, _ := s.wallet.GetAddress()
				op, err = s.PrepareCreateCredit(ctx, sdk.CreateCreditPreparation{TokenAddress: token, Spender: address, Amount: big.NewInt(30), Expiry: big.NewInt(1000), SingleLimit: big.NewInt(20), DailyLimit: big.NewInt(40), MonthlyLimit: big.NewInt(60)})
			} else {
				leaf := seedLimitedCredit(t, s, token)
				switch kind {
				case "draw":
					op, err = s.PrepareDrawCredit(ctx, sdk.DrawCreditPreparation{TokenAddress: token, Credits: []sdk.CreditDrawInput{{Credit: leaf, Amount: big.NewInt(15)}}})
				case "top-up":
					op, err = s.PrepareTopUpCredit(ctx, sdk.TopUpCreditPreparation{TokenAddress: token, Credit: leaf, AdditionalAmount: big.NewInt(10), SingleLimit: big.NewInt(10), DailyLimit: big.NewInt(5), MonthlyLimit: big.NewInt(0)})
				case "cancel":
					op, err = s.PrepareCancelCredit(ctx, sdk.CancelCreditPreparation{TokenAddress: token, Credit: leaf})
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if op, err = op.Authorize(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err = op.Prove(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTopUpExposesOptionalPolicyChanges(t *testing.T) {
	shape := reflect.TypeFor[sdk.TopUpCreditPreparation]()
	for _, name := range []string{"SingleLimit", "DailyLimit", "MonthlyLimit"} {
		field, ok := shape.FieldByName(name)
		if !ok || field.Type != reflect.TypeFor[*big.Int]() {
			t.Fatalf("top-up needs optional policy %s", name)
		}
	}
}

func TestTopUpEditsOnlyExplicitPolicyAndPreservesAllUsage(t *testing.T) {
	for _, name := range []string{"inherit", "disable", "lower", "increase", "later-day"} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			timestamp := uint64(100)
			if name == "later-day" {
				timestamp = 90000
			}
			s, token, _ := testSession(t, timestamp)
			leaf := seedLimitedCredit(t, s, token)
			input := sdk.TopUpCreditPreparation{TokenAddress: token, Credit: leaf, AdditionalAmount: big.NewInt(10)}
			switch name {
			case "disable":
				input.SingleLimit = big.NewInt(0)
				input.DailyLimit = big.NewInt(0)
				input.MonthlyLimit = big.NewInt(0)
			case "lower":
				input.SingleLimit = big.NewInt(1)
				input.DailyLimit = big.NewInt(5)
				input.MonthlyLimit = big.NewInt(10)
			case "increase":
				input.SingleLimit = big.NewInt(30)
				input.MonthlyLimit = big.NewInt(90)
			}
			op, err := s.PrepareTopUpCredit(ctx, input)
			if err != nil {
				t.Fatal(err)
			}
			previous := op.plan.CreditInputs[0].State.Note
			next := op.plan.CreditOutput.Note
			if next.LastDay != previous.LastDay || next.DailySpent != previous.DailySpent || next.LastMonth != previous.LastMonth || next.MonthlySpent != previous.MonthlySpent {
				t.Fatal("top-up reset/advanced usage")
			}
			expected := []*big.Int{input.SingleLimit, input.DailyLimit, input.MonthlyLimit}
			for i, value := range expected {
				want := previous.CreditSpendingState.Fields()[i]
				if value != nil {
					want = core.BigIntToElement(value)
				}
				if next.CreditSpendingState.Fields()[i] != want {
					t.Fatalf("wrong limit%d", i)
				}
			}
			if name == "lower" {
				if _, err = core.ApplyCreditDrawUsage(next, core.ElementFromUint64(1), 100); err == nil {
					t.Fatal("lowered policy did not freeze current period")
				}
				if _, err = core.ApplyCreditDrawUsage(next, core.ElementFromUint64(1), 3000000); err != nil {
					t.Fatal("new month did not start fresh", err)
				}
			}
			if op, err = op.Authorize(ctx); err != nil {
				t.Fatal(err)
			}
			witness, err := op.privateInputs()
			if err != nil {
				t.Fatal(err)
			}
			if witness["previous_single_limit"] != "20" || witness["previous_daily_limit"] != "40" || witness["previous_monthly_limit"] != "60" || witness["single_limit"] != next.SingleLimit.String() || witness["daily_limit"] != next.DailyLimit.String() || witness["monthly_limit"] != next.MonthlyLimit.String() {
				t.Fatal("old/new policy witness not bound")
			}
		})
	}
}

func TestRealNativeTopUpPolicyRequiresNewOwnerAuthorization(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set pinned tools for genuine owner authorization rejection")
	}
	ctx := context.Background()
	s, token, _ := testSession(t)
	leaf := seedLimitedCredit(t, s, token)
	op, err := s.PrepareTopUpCredit(ctx, sdk.TopUpCreditPreparation{TokenAddress: token, Credit: leaf, AdditionalAmount: big.NewInt(10), SingleLimit: big.NewInt(10), DailyLimit: big.NewInt(5), MonthlyLimit: big.NewInt(0)})
	if err != nil {
		t.Fatal(err)
	}
	if op, err = op.Authorize(ctx); err != nil {
		t.Fatal(err)
	}
	private, err := op.privateInputs()
	if err != nil {
		t.Fatal(err)
	}
	changed := op.plan.CreditOutput.Note.Clone()
	changed.SingleLimit = core.ElementFromUint64(11)
	commitment, err := transaction.CreditCommitment(changed)
	if err != nil {
		t.Fatal(err)
	}
	params := *op.plan.TopUpCredit
	params.Creation.CreditCommitment = commitment
	words, err := protocol.TopUpCreditPublicInputs(op.plan.Context, params, op.deliveries(), op.plan.Funding)
	if err != nil {
		t.Fatal(err)
	}
	inputs := make([]common.Hash, len(words))
	for i, value := range words {
		inputs[i] = value
	}
	private["single_limit"] = changed.SingleLimit.String()
	request, err := prover.Request("credit_top_up", "evm", private, inputs)
	if err != nil {
		t.Fatal(err)
	}
	backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = backend.Prove(ctx, request, prover.Options{}); err == nil {
		t.Fatal("new policy accepted using signature for another output Credit")
	}
}

func TestRealNativeUnlimitedUint128Usage(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set pinned tools for genuine uint128 usage proofs")
	}
	for _, amount := range []int64{15, 100} {
		t.Run(fmt.Sprintf("draw%d", amount), func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t)
			wide := core.BigIntToElement(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(3)))
			state := core.CreditSpendingState{LastDay: core.ElementFromUint64(1), DailySpent: wide, LastMonth: core.ElementFromUint64(1), MonthlySpent: wide}
			leaf := seedLimitedCredit(t, s, token, state)
			backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
			if err != nil {
				t.Fatal(err)
			}
			s.options.Prover = backend
			op, err := s.PrepareDrawCredit(ctx, sdk.DrawCreditPreparation{TokenAddress: token, Credits: []sdk.CreditDrawInput{{Credit: leaf, Amount: big.NewInt(amount)}}})
			if err != nil {
				t.Fatal(err)
			}
			if op, err = op.Authorize(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err = op.Prove(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}
