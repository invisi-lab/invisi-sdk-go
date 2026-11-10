package gas

import (
	"encoding/json"
	"os"
	"strings"
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
		got, ok := Reference(v.Query)
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
	for _, row := range DocumentedCatalog().References {
		if row.Shape["kind"] == "transact" {
			query = row.Shape
			break
		}
	}
	result, ok := Reference(query)
	if !ok {
		t.Fatal("reference absent")
	}
	result.Shape["kind"] = "mutated"
	result.Source.Conditions = "mutated"
	again, ok := Reference(query)
	if !ok || again.Shape["kind"] != "transact" || again.Source.Conditions == "mutated" {
		t.Fatal("reference aliases catalog")
	}
}

func TestCreditReferenceRequiresFullManifestIdentity(t *testing.T) {
	originalCatalog, originalRelease := catalogJSON, releaseJSON
	t.Cleanup(func() { catalogJSON, releaseJSON = originalCatalog, originalRelease })
	identity := func(hash string, version int) map[string]any {
		out := map[string]any{"creditManifestVersion": version, "creditFormatVersion": 8}
		if hash != "missing" {
			out["creditManifestSha256"] = hash
		}
		return out
	}
	sha := strings.Repeat("a", 64)
	for _, tc := range []struct {
		name, rowHash, currentHash string
		rowVersion                 int
		available                  bool
	}{
		{"exact manifest", sha, sha, 7, true},
		{"missing measured hash", "missing", sha, 7, false},
		{"missing current hash", sha, "missing", 7, false},
		{"both missing", "missing", "missing", 7, false},
		{"different manifest same versions", strings.Repeat("b", 64), sha, 7, false},
		{"malformed matching hash", "not-a-sha256", "not-a-sha256", 7, false},
		{"different manifest version", sha, sha, 99, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			releaseJSON, err = json.Marshal(identity(tc.currentHash, 7))
			if err != nil {
				t.Fatal(err)
			}
			query := Shape{"kind": "create-credit", "scenario": "fixture"}
			catalogJSON, err = json.Marshal(map[string]any{"sources": []Source{{Path: "fixture"}}, "references": []any{map[string]any{"source": "fixture", "shape": query, "gasUnits": "42", "releaseIdentity": identity(tc.rowHash, tc.rowVersion)}}})
			if err != nil {
				t.Fatal(err)
			}
			result, ok := Reference(query)
			if ok != tc.available || (ok && result.GasUnits != "42") {
				t.Fatalf("manifest match: available %v, result %+v", ok, result)
			}
		})
	}
}
