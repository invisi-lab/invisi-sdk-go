package gas

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUnavailableEstimateOmitsGasUnits(t *testing.T) {
	got := EstimateGas(Shape{"kind": "unknown"})
	data, err := json.Marshal(got)
	if err != nil || got.Available || got.Source != "benchmark" || got.Reason != "shape-not-covered" || strings.Contains(string(data), "gasUnits") {
		t.Fatal(string(data), err)
	}
}
func TestCreditBenchmarkIgnoresPackagedRelease(t *testing.T) {
	got := EstimateGas(Shape{"kind": "create-credit", "scenario": "first-allocation-empty-tree", "feeEnabled": false, "feeMechanism": "none"})
	if !got.Available || got.GasUnits == nil || got.ProtocolCommit == "" {
		t.Fatal(got)
	}
}

func TestEstimateUsesSinglePublicResultLayout(t *testing.T) {
	result := EstimateGas(Shape{"kind": "transact", "noteInputs": 1, "noteOutputs": 1})
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["reference"]; ok {
		t.Fatal("obsolete nested gas reference remains in public result")
	}
}
