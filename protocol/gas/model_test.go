package gas

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDocumentedReferenceVectors(t *testing.T) {
	data, err := os.ReadFile("data/documented-reference-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Vectors []struct {
			Query     Shape
			Available bool
			GasUnits  string
		}
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures.Vectors) == 0 {
		t.Fatal("empty vectors")
	}
	for _, v := range fixtures.Vectors {
		got, ok := reference(v.Query)
		if ok != v.Available {
			t.Fatalf("query %v: availability %v", v.Query, ok)
		}
		if ok && got.GasUnits != v.GasUnits {
			t.Fatalf("query %v: gas %s != %s", v.Query, got.GasUnits, v.GasUnits)
		}
	}
}

func TestReferenceCopiesProvenance(t *testing.T) {
	var query Shape
	for _, row := range documentedCatalog().References {
		if row.Shape["kind"] == "transact" {
			query = row.Shape
			break
		}
	}
	result, ok := reference(query)
	if !ok {
		t.Fatal("reference absent")
	}
	result.Shape["kind"] = "mutated"
	result.Source.Conditions = "mutated"
	again, ok := reference(query)
	if !ok || again.Shape["kind"] != "transact" || again.Source.Conditions == "mutated" {
		t.Fatal("reference aliases catalog")
	}
}
