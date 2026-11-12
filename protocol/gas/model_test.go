package gas

import "testing"

func TestReferenceCopiesProvenance(t *testing.T) {
	query := Shape{"kind": "transact", "noteInputs": 1, "noteOutputs": 1}
	result := EstimateGas(query)
	if !result.Available {
		t.Fatal(result)
	}
	result.Reference.Shape["kind"] = "mutated"
	result.Reference.Source.Conditions = "mutated"
	result.Assumptions[0] = "mutated"
	again := EstimateGas(query)
	if !again.Available || again.Reference.Shape["kind"] != "transact" || again.Conditions == "mutated" || again.Assumptions[0] == "mutated" {
		t.Fatal("reference aliases catalog")
	}
}
