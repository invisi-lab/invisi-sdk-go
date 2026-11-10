package provenance

import _ "embed"

//go:embed manifest.json
var reviewedManifest string

// ManifestBytes returns a copy of the reviewed public circuit manifest.
func ManifestBytes() []byte { return []byte(reviewedManifest) }
