package sdk

import (
	"context"

	"github.com/invisi-lab/invisi-sdk-go/provenance"
)

type ProvenanceProof = provenance.ProvenanceProof

func ValidateProvenanceProof(packet ProvenanceProof, trust provenance.TrustContext) error {
	return provenance.ValidateProvenanceProof(packet, trust)
}
func SubmitProvenanceProof(ctx context.Context, client *provenance.Client, packet ProvenanceProof) (provenance.Result, error) {
	return provenance.SubmitProvenanceProof(ctx, client, packet)
}
