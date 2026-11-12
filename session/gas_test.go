package session

import (
	"context"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"

	"reflect"
	"testing"
)

func TestPreparedBenchmarkHasNoRuntimeDependency(t *testing.T) {
	invisi, input, _ := explicitFixture(t)

	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	execution := transaction.PreparedExecution(prepared)
	execution.Runtime = nil
	before := execution.Status
	result, err := prepared.EstimateGas()
	if err != nil || result.Source != "benchmark" {
		t.Fatal(result, err)
	}
	if execution.Status != before {
		t.Fatal("estimate changed lifecycle")
	}
	op := &Operation{plan: execution.Plan, status: before}
	managed, err := op.EstimateGas()
	if err != nil || !reflect.DeepEqual(result, managed) || op.status != before {
		t.Fatal(managed, err)
	}
}

func TestManagedPreparedExactBenchmarkBeforeProof(t *testing.T) {
	s, token, _ := testSession(t)
	address, _ := s.wallet.GetAddress()
	op, err := s.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(100)}})})
	if err != nil {
		t.Fatal(err)
	}
	s.options.Provider = nil
	before, _ := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	result, err := op.EstimateGas()
	after, _ := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	if err != nil || !result.Available || result.Reference == nil || *result.GasUnits == 0 || !reflect.DeepEqual(before, after) {
		t.Fatal(result, err)
	}
}

func TestRealPreparedMixedBenchmarkBeforeProof(t *testing.T) {
	s, token, _ := testSession(t)
	address, _ := s.wallet.GetAddress()
	op, err := s.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}, {To: address, Amount: big.NewInt(30)}, {To: address, Amount: big.NewInt(30)}}), Funding: &sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(10), Authorization: "caller"}, Withdrawal: &sdk.Withdrawal{To: s.wallet.OwnerAddress(), Amount: big.NewInt(20)}})
	if err != nil {
		t.Fatal(err)
	}
	s.options.Provider = nil
	before, _ := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	got, err := op.EstimateGas()
	after, _ := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	if err != nil || !got.Available || *got.GasUnits == 0 || got.Statistic != "estimate" || !reflect.DeepEqual(before, after) {
		t.Fatal(got, err)
	}
}
