package gas

import (
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
)

func TestMeasuredReceiptsAndHeldout(t *testing.T) {
	var model struct {
		Rows []struct {
			Shape    Shape
			GasUnits uint64
		}
	}
	raw, err := os.ReadFile("data/measured-model.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &model); err != nil {
		t.Fatal(err)
	}
	if len(model.Rows) != 241 {
		t.Fatalf("receipts: %d", len(model.Rows))
	}
	for i, row := range model.Rows {
		got := EstimateGas(row.Shape)
		if !got.Available || got.GasUnits == nil || *got.GasUnits != row.GasUnits || got.Reference == nil || got.Reference.Statistic != "median" || got.Reference.ID != "receipt-"+strconv.Itoa(i) {
			t.Fatalf("receipt %d: %+v", i, got)
		}
	}
	var heldout struct {
		Cases []struct {
			Shape       Shape
			ModeledGas  uint64
			MeasuredGas uint64
		}
	}
	raw, err = os.ReadFile("data/model-validation.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &heldout); err != nil {
		t.Fatal(err)
	}
	if len(heldout.Cases) != 12 {
		t.Fatalf("heldout: %d", len(heldout.Cases))
	}
	for i, row := range heldout.Cases {
		row.Shape["extraDeliveryBytes"] = 1
		got := EstimateGas(row.Shape)
		if !got.Available || got.GasUnits == nil || (*got.GasUnits < row.ModeledGas+31 || *got.GasUnits > row.ModeledGas+33) || got.Reference.Statistic != "estimate" {
			t.Fatalf("heldout %d: %+v want %d", i, got, row.ModeledGas)
		}
		if delta := math.Abs(float64(*got.GasUnits-32)-float64(row.MeasuredGas)) / float64(row.MeasuredGas); delta > 0.15 {
			t.Fatalf("heldout %d exceeds 15%% error: %f", i, delta)
		}
	}
}

func TestTypeScriptParity(t *testing.T) {
	raw, err := os.ReadFile("testdata/ts-parity-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		AvailableDefaults map[string]any
		Vectors           []struct {
			Query  Shape
			Result map[string]any
		}
	}
	if err = json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures.Vectors) != 1639 {
		t.Fatalf("parity cases %d", len(fixtures.Vectors))
	}
	for i, v := range fixtures.Vectors {
		got := EstimateGas(v.Query)
		raw, err = json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		var actual map[string]any
		if err = json.Unmarshal(raw, &actual); err != nil {
			t.Fatal(err)
		}
		expected := map[string]any{}
		if v.Result["available"] == true {
			for key, value := range fixtures.AvailableDefaults {
				expected[key] = value
			}
		}
		for key, value := range v.Result {
			expected[key] = value
		}
		for key, want := range expected {
			value := actual[key]
			if key == "gasUnits" {
				value = strconv.FormatUint(*got.GasUnits, 10)
			}
			if key == "sampleCount" && value == nil {
				value = float64(0)
			}
			a, _ := json.Marshal(value)
			b, _ := json.Marshal(want)
			if string(a) != string(b) {
				t.Fatalf("case %d query %v field %s: %s != %s", i, v.Query, key, a, b)
			}
		}
	}
}

func TestAllLegalDirectShapes(t *testing.T) {
	check := func(q Shape) {
		t.Helper()
		got, err := EstimateGasChecked(q)
		if err != nil || !got.Available || got.GasUnits == nil || *got.GasUnits == 0 {
			t.Fatalf("%v: %+v %v", q, got, err)
		}
	}
	for n := 1; n <= 10; n++ {
		for _, fee := range []bool{false, true} {
			check(Shape{"kind": "shield", "noteOutputs": n, "fee": fee})
		}
	}
	for total := 1; total <= 190; total++ {
		for inputs := 0; inputs <= total; inputs++ {
			outputs := total - inputs
			for flags := 0; flags < 8; flags++ {
				deposit, withdrawal, fee := flags&1 != 0, flags&2 != 0, flags&4 != 0
				if inputs == 0 && (outputs < 2 || !deposit) || outputs == 0 && !withdrawal && !fee {
					continue
				}
				for _, trees := range []int{min(inputs, 1), inputs} {
					check(Shape{"kind": "transact", "noteInputs": inputs, "noteOutputs": outputs, "sourceTrees": trees, "deposit": deposit, "withdrawal": withdrawal, "fee": fee})
				}
			}
		}
	}
	for n := 1; n <= 80; n++ {
		for payers := 1; payers <= n; payers++ {
			for trees := 1; trees <= n; trees++ {
				for _, fee := range []bool{false, true} {
					check(Shape{"kind": "draw-credit", "creditInputs": n, "distinctPayers": payers, "sourceTrees": trees, "fee": fee})
				}
			}
		}
	}
	for _, kind := range []string{"create-credit", "update-credit"} {
		for _, fee := range []bool{false, true} {
			check(Shape{"kind": kind, "fee": fee})
		}
	}
	for _, authorization := range []string{"direct", "eoa", "erc1271", "unknown"} {
		for _, bitmap := range []string{"empty", "nonempty"} {
			for _, fee := range []bool{false, true} {
				check(Shape{"kind": "revoke-credit", "authorization": authorization, "bitmapWord": bitmap, "fee": fee})
			}
		}
	}
}

func TestInvalidShapesReturnCheckedErrors(t *testing.T) {
	cases := []Shape{
		{"kind": "shield", "noteOutputs": 0}, {"kind": "shield", "noteOutputs": 11},
		{"kind": "transact", "noteInputs": 1, "noteOutputs": 1, "capacity": 3},
		{"kind": "transact", "noteInputs": 0, "noteOutputs": 2},
		{"kind": "transact", "noteInputs": 1, "noteOutputs": 0},
		{"kind": "transact", "noteInputs": 1, "noteOutputs": 1, "sourceTrees": 0},
		{"kind": "transact", "noteInputs": 1, "noteOutputs": 1, "sourceTrees": 2},
		{"kind": "transact", "noteInputs": 190, "noteOutputs": 1},
		{"kind": "draw-credit", "creditInputs": 81}, {"kind": "draw-credit", "creditInputs": 1, "distinctPayers": 0},
		{"kind": "draw-credit", "creditInputs": 1, "sourceTrees": 2},
		{"kind": "revoke-credit", "authorization": "invalid"},
		{"kind": "shield", "noteOutputs": 1, "extraDeliveryBytes": -1},
		{"kind": "shield", "noteOutputs": 1.5}, {"kind": "shield", "noteOutputs": "1"},
	}
	for _, q := range cases {
		if _, err := EstimateGasChecked(q); err == nil {
			t.Fatalf("accepted %v", q)
		}
		if EstimateGas(q).Available {
			t.Fatalf("unchecked accepted %v", q)
		}
	}
}

func TestModeledEstimateSerializesZeroSampleCount(t *testing.T) {
	got := EstimateGas(Shape{"kind": "revoke-credit"})
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err = json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["sampleCount"] != float64(0) {
		t.Fatalf("missing zero sample count: %s", raw)
	}
}
