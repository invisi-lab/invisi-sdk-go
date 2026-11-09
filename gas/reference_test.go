package gas

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/internal/release"
	"os"
	"testing"
)

func TestDocumentedReferenceVectors(t *testing.T) {
	data, err := os.ReadFile("testdata/documented-reference-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		CatalogSHA256          string
		PackagedCreditIdentity ReleaseIdentity
		Unsupported            []Shape
		Vectors                []struct {
			Query     Shape
			Available bool
			GasUnits  string
		}
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, catalogJSON); err != nil {
		t.Fatal(err)
	}
	if fixtures.CatalogSHA256 != fmt.Sprintf("%x", sha256.Sum256(compact.Bytes())) {
		t.Fatal("catalog vectors do not authenticate the exact catalog")
	}
	if fixtures.PackagedCreditIdentity.CreditManifestVersion != release.CreditManifestVersion || fixtures.PackagedCreditIdentity.CreditFormatVersion != release.CreditFormatVersion || fixtures.PackagedCreditIdentity.CreditManifestSHA256 != release.CreditManifestSHA256 {
		t.Fatal("vector release identity differs from packaged circuits")
	}
	for _, shape := range fixtures.Unsupported {
		if _, ok := Reference(shape); ok {
			t.Fatalf("unsupported shape has a reference: %v", shape)
		}
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
	query := Shape{}
	for _, row := range DocumentedCatalog().References {
		if _, ok := Reference(row.Shape); ok {
			query = row.Shape
			break
		}
	}
	expectedKind := query["kind"]
	result, ok := Reference(query)
	if !ok {
		t.Fatal("reference absent")
	}
	result.Shape["kind"] = "mutated"
	result.Source.Conditions = "mutated"
	again, ok := Reference(query)
	if !ok || again.Shape["kind"] != expectedKind || again.Source.Conditions == "mutated" {
		t.Fatal("reference aliases catalog")
	}
}

func TestCreditReferencesRequireExactManifestIdentity(t *testing.T) {
	current := ReleaseIdentity{CreditManifestVersion: release.CreditManifestVersion, CreditFormatVersion: release.CreditFormatVersion, CreditManifestSHA256: release.CreditManifestSHA256}
	if !currentCreditRelease(&current) {
		t.Fatal("current manifest identity rejected")
	}
	for _, sha := range []string{"", "00", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"} {
		bad := current
		bad.CreditManifestSHA256 = sha
		if currentCreditRelease(&bad) {
			t.Fatal("reference accepted missing or mismatched manifest", sha)
		}
	}
	if currentCreditRelease(nil) {
		t.Fatal("missing release identity accepted")
	}
}
