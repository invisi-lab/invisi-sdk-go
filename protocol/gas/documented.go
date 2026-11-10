// Package gas exposes measured benchmark references independently of transaction estimation.
package gas

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
)

// Shape identifies every published measurement condition. Omitted dimensions do not match.
type Shape map[string]any

type Source struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Conditions string `json:"conditions"`
}

type ReleaseIdentity struct {
	CreditManifestSHA256  string `json:"creditManifestSha256,omitempty"`
	CreditManifestVersion int    `json:"creditManifestVersion"`
	CreditFormatVersion   int    `json:"creditFormatVersion"`
}
type Measurement struct {
	ReleaseIdentity *ReleaseIdentity `json:"releaseIdentity,omitempty"`
	ID              string           `json:"id"`
	Source          string           `json:"source"`
	Shape           Shape            `json:"shape"`
	GasUnits        string           `json:"gasUnits"`
	Statistic       string           `json:"statistic"`
	SampleCount     int              `json:"sampleCount"`
}

type Catalog struct {
	FormatVersion int           `json:"formatVersion"`
	Kind          string        `json:"kind"`
	DocsCommit    string        `json:"docsCommit"`
	Sources       []Source      `json:"sources"`
	Assumptions   []string      `json:"assumptions"`
	References    []Measurement `json:"references"`
}

type Result struct {
	Measurement
	Source      Source   `json:"source"`
	DocsCommit  string   `json:"docsCommit"`
	Assumptions []string `json:"assumptions"`
}

//go:embed data/documented-references.json
var catalogJSON []byte

//go:embed data/credit-release.json
var releaseJSON []byte

// DocumentedCatalog returns a detached copy of the exact documented reference catalog.
func DocumentedCatalog() Catalog {
	var c Catalog
	if err := json.Unmarshal(catalogJSON, &c); err != nil {
		panic("invalid embedded gas catalog")
	}
	return c
}

// Reference matches a complete benchmark shape. Unsupported shapes have no reference.
// The result describes the documented environment; use eth_estimateGas for a deployment.
func Reference(shape Shape) (Result, bool) {
	raw, err := json.Marshal(shape)
	if err != nil {
		return Result{}, false
	}
	var query Shape
	if json.Unmarshal(raw, &query) != nil {
		return Result{}, false
	}
	var current ReleaseIdentity
	if json.Unmarshal(releaseJSON, &current) != nil {
		panic("invalid current Credit release")
	}
	c := DocumentedCatalog()
	for _, row := range c.References {
		if kind, ok := row.Shape["kind"].(string); ok && strings.Contains(kind, "credit") {
			if !currentCreditRelease(row.ReleaseIdentity) {
				continue
			}
		}
		if !reflect.DeepEqual(row.Shape, query) {
			continue
		}
		for _, source := range c.Sources {
			if source.Path == row.Source {
				return Result{row, source, c.DocsCommit, c.Assumptions}, true
			}
		}
	}
	return Result{}, false
}

func currentCreditRelease(identity *ReleaseIdentity) bool {
	var current ReleaseIdentity
	if json.Unmarshal(releaseJSON, &current) != nil {
		panic("invalid current Credit release")
	}
	return identity != nil && gasDigest.MatchString(current.CreditManifestSHA256) && gasDigest.MatchString(identity.CreditManifestSHA256) && *identity == current
}
