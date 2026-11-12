package transaction

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/gas"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

func gasPlan(inputs, outputs int) Plan {
	p := Plan{Kind: "transact", Target: "evm", Inputs: make([]NoteRecord, inputs), Outputs: make([]PreparedOutput, outputs)}
	for _, c := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190} {
		if c >= inputs+outputs {
			p.Capacity = c
			break
		}
	}
	if inputs > 0 {
		p.Params.Sources = make([]protocol.InputSource, 1)
	} else {
		p.Funding.Amount = big.NewInt(1)
	}
	if outputs == 0 {
		p.Params.WithdrawalAmount = big.NewInt(1)
	}
	for i := range p.Outputs {
		p.Outputs[i].Delivery = protocol.Delivery{Ciphertext: make([]byte, 160), Clue: make([]byte, 68)}
	}
	return p
}
func TestPreparedBenchmarkIsStableAcrossProofAndSupportsPayload(t *testing.T) {
	p := gasPlan(0, 2)
	prepared, _ := Detached(&Execution{Plan: p, Status: "prepared"}, nil)
	got, err := prepared.EstimateGas()
	want := gas.EstimateGas(gas.Shape{"kind": "transact", "capacity": 2, "noteInputs": 0, "noteOutputs": 2, "deposit": true, "withdrawal": false, "feeEnabled": false})
	if err != nil || !got.Available || !reflect.DeepEqual(got, want) {
		t.Fatal(got, want, err)
	}
	prepared.operation.Status = "ready"
	prepared.operation.Data = []byte{1}
	after, _ := prepared.EstimateGas()
	if !reflect.DeepEqual(got, after) {
		t.Fatal("proof stage changed estimate")
	}
	p.Outputs[0].Delivery.Ciphertext = make([]byte, 260)
	larger := EstimatePlanGas(p)
	if !larger.Available || larger.Statistic != "estimate" || *larger.GasUnits <= *got.GasUnits {
		t.Fatal(larger)
	}
}
func TestPreparedAllTransactShapesAndCashFlowsStayPositive(t *testing.T) {
	for _, pair := range [][2]int{{1, 0}, {0, 2}, {1, 1}, {1, 2}, {3, 3}, {7, 4}, {95, 95}, {190, 0}, {0, 190}} {
		p := gasPlan(pair[0], pair[1])
		prepared, _ := Detached(&Execution{Plan: p, Status: "prepared"}, nil)
		got, err := prepared.EstimateGas()
		if err != nil || !got.Available || *got.GasUnits == 0 {
			t.Fatal(pair, got, err)
		}
		p.Funding.Amount = big.NewInt(1)
		p.Params.WithdrawalAmount = big.NewInt(1)
		p.Params.FeeAmount = big.NewInt(1)
		got = EstimatePlanGas(p)
		if !got.Available || *got.GasUnits == 0 {
			t.Fatal(pair, got)
		}
	}
}
func TestPreparedShieldEstimatesBeforeSigning(t *testing.T) {
	for n := 1; n <= 10; n++ {
		p := gasPlan(0, n)
		p.Kind = "shield"
		p.Target = ""
		p.Capacity = 0
		prepared, _ := Detached(&Execution{Plan: p, Status: "prepared"}, nil)
		got, err := prepared.EstimateGas()
		want := gas.EstimateGas(gas.Shape{"kind": "shield", "noteOutputs": n})
		if err != nil || !got.Available || !reflect.DeepEqual(got, want) {
			t.Fatal(n, got, want, err)
		}
		p.Shield.FeeAmount = big.NewInt(1)
		p.Funding.Permit.Deadline = big.NewInt(1)
		p.FundingMode = "signature"
		paid := EstimatePlanGas(p)
		if !paid.Available || paid.Statistic != "estimate" || *paid.GasUnits <= *got.GasUnits {
			t.Fatal(paid)
		}
	}
}
func TestQuotedAggregatePreviewIsNotDirectBenchmark(t *testing.T) {
	p := gasPlan(1, 1)
	p.Broadcaster = &BroadcasterPlan{Signed: broadcaster.SignedQuote{Quote: broadcaster.Quote{Execution: "aggregate"}}}
	if EstimatePlanGas(p).Available {
		t.Fatal("aggregate quote mislabeled as direct")
	}
}
