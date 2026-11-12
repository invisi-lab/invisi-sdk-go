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
