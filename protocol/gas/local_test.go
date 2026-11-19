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

func TestInitialPaymentMeasuredGas(t *testing.T) {
	for _, tc := range []struct {
		kind  string
		fee   bool
		units uint64
	}{{"create-credit", false, 1569276}, {"create-credit", true, 1632766}, {"create-credit-initial-payment", false, 1677215}, {"create-credit-initial-payment", true, 1701560}} {
		q := Shape{"kind": tc.kind, "creationAbi": "initial-payment", "feeEnabled": tc.fee}
		r, e := EstimateGasChecked(q)
		if e != nil || !r.Available || *r.GasUnits != tc.units || r.SampleCount != 3 || r.ReleaseIdentity["creditManifestVersion"] != 7 {
			t.Fatal(q, r, e)
		}
		q["extraDeliveryBytes"] = 1
		r, e = EstimateGasChecked(q)
		if e != nil || *r.GasUnits != tc.units+32 || r.SampleCount != 0 {
			t.Fatal(q, r, e)
		}
	}
}
