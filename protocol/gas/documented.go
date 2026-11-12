// Package gas 提供不依赖网络的实测与基准校准 gas 预估。
package gas

// Shape 描述本地收据查找或模型估算的操作形状。
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
