package gas

import "testing"

func TestReturnedAssumptionsAreIndependent(t *testing.T) {
	query := Shape{"kind": "transact", "noteInputs": 1, "noteOutputs": 1}
	result := EstimateGas(query)
	if !result.Available {
		t.Fatal(result)
	}
	result.Assumptions[0] = "mutated"
	again := EstimateGas(query)
	if !again.Available || again.Assumptions[0] == "mutated" {
		t.Fatal("result shares mutable model metadata")
	}
}
