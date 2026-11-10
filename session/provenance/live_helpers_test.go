package provenance

import "encoding/json"

type liveTrust struct {
	TrackingID       string           `json:"trackingId"`
	Recovery         json.RawMessage  `json:"recovery,omitempty"`
	Format           string           `json:"format"`
	Endpoint         string           `json:"endpoint"`
	Scope            Scope            `json:"scope"`
	Generation       string           `json:"chainGeneration"`
	ProviderID       string           `json:"providerId"`
	InitialPublicKey string           `json:"initialSigningPublicKey"`
	InitialKeyID     string           `json:"initialSigningKeyId"`
	CurrentPublicKey string           `json:"currentSigningPublicKey"`
	CurrentKeyID     string           `json:"currentSigningKeyId"`
	Manifest         string           `json:"manifestSha256"`
	Rotations        []RotationRecord `json:"keyRotations"`
	CheckedAt        string           `json:"checkedAt"`
	RootID           string           `json:"rootId"`
	Commitment       string           `json:"commitment"`
	JobID            string           `json:"jobId"`
}
