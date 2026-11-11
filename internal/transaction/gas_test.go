package transaction

import (
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/gas"
	"math/big"
	"reflect"
	"testing"
)

func TestPreparedExactBenchmark(t *testing.T) {
	p := Plan{Kind: "transact", Target: "evm", Capacity: 2, Outputs: make([]PreparedOutput, 2), Deliveries: []protocol.Delivery{{Ciphertext: make([]byte, 160), Clue: make([]byte, 68)}, {Ciphertext: make([]byte, 160), Clue: make([]byte, 68)}}}
	for i := range p.Outputs {
		p.Outputs[i].Delivery = p.Deliveries[i]
	}
	prepared, _ := Detached(&Execution{Plan: p, Status: "prepared"}, nil)
	got, err := prepared.EstimateGas()
	want := gas.EstimateGas(gas.Shape{"kind": "transact", "capacity": 2, "noteInputs": 0, "noteOutputs": 2, "sourceTrees": 0, "noteCiphertextBytes": 160, "clueBytes": 68, "scenario": "public-funding"})
	if err != nil || !got.Available || !reflect.DeepEqual(got, want) {
		t.Fatal(got, want, err)
	}
	prepared.operation.Status = "ready"
	prepared.operation.Data = []byte{1}
	after, _ := prepared.EstimateGas()
	if !reflect.DeepEqual(got, after) {
		t.Fatal("proof stage changed estimate")
	}
	p.Outputs[0].Delivery.Ciphertext = []byte{1}
	if EstimatePlanGas(p).Available {
		t.Fatal("nonstandard delivery estimated")
	}
}

func TestPreparedMixedCashFlowsSelectMeasuredScenario(t *testing.T) {
	p := Plan{Kind: "transact", Target: "evm", Capacity: 4, Inputs: make([]NoteRecord, 1), Outputs: make([]PreparedOutput, 3), Params: protocol.CrossTreeTransactParams{Sources: make([]protocol.InputSource, 1), WithdrawalAmount: big.NewInt(1), FeeAmount: big.NewInt(1)}, Funding: protocol.FundingParams{Amount: big.NewInt(1)}}
	for i := range p.Outputs {
		p.Outputs[i].Delivery = protocol.Delivery{Ciphertext: make([]byte, 160), Clue: make([]byte, 68)}
	}
	prepared, _ := Detached(&Execution{Plan: p, Status: "prepared"}, nil)
	got, err := prepared.EstimateGas()
	if err != nil || !got.Available || *got.GasUnits != 1768554 || got.Reference.ID != "5-62" {
		t.Fatal(got, err)
	}
	p.Funding.Amount = nil
	p.Params.WithdrawalAmount = nil
	p.Params.FeeAmount = nil
	if EstimatePlanGas(p).Available {
		t.Fatal("unmeasured private C4 used mixed row")
	}
	p.Capacity = 1
	p.Outputs = nil
	p.Params.WithdrawalAmount = big.NewInt(1)
	p.Params.FeeAmount = big.NewInt(1)
	got = EstimatePlanGas(p)
	if !got.Available || *got.GasUnits != 1083512 {
		t.Fatal(got)
	}
	p.Params.FeeAmount = nil
	if EstimatePlanGas(p).Available {
		t.Fatal("unmeasured withdrawal C1 used mixed row")
	}
}
