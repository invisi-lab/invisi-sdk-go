// Package gas exposes measured benchmark references independently of transaction estimation.
package gas

import (
	_ "embed"
	"encoding/json"
	"reflect"
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

// documentedCatalog returns a detached copy of the exact documented reference catalog.
func documentedCatalog() Catalog {
	var c Catalog
	if err := json.Unmarshal(catalogJSON, &c); err != nil {
		panic("invalid embedded gas catalog")
	}
	return c
}

// Reference matches a complete benchmark shape. Unsupported shapes have no reference.
// The result describes the documented benchmark environment.
func reference(shape Shape) (Result, bool) {
	raw, err := json.Marshal(shape)
	if err != nil {
		return Result{}, false
	}
	var query Shape
	if json.Unmarshal(raw, &query) != nil {
		return Result{}, false
	}
	c := documentedCatalog()
	for _, row := range c.References {
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
