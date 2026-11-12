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
	if client == nil {
		var err error
		client, err = (&provenance.OfficialClients{}).Get(ctx, provenance.Scope{ChainID: packet.Context.ChainID, ProtocolAddress: packet.Context.ProtocolAddress})
		if err != nil {
			return nil, err
		}
	}
	return provenance.SubmitProvenanceProof(ctx, client, packet)
}
