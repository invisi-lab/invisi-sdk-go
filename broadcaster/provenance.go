package broadcaster

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
)

const ProvenanceRequestHashFormat = "invisi-broadcaster-provenance-intent-v1"

func VerifyProvenanceQuote(s SignedQuote, t Trust, data []byte, now uint64, requiredPolicy common.Hash) error {
	if s.ProvenancePolicyID == nil || *s.ProvenancePolicyID != requiredPolicy || s.RequestHashFormat != ProvenanceRequestHashFormat {
		return errors.New("broadcaster provenance policy mismatch")
	}
	original, e := RequestHash(data)
	if e != nil {
		return e
	}
	bound, e := provenance.ProvenanceIntentHash(original.Hex(), requiredPolicy.Hex())
	if e != nil {
		return e
	}
	return verifyQuoteWithHash(SignedQuote{Quote: s.Quote, Signature: s.Signature, Context: s.Context}, t, data, now, common.HexToHash(bound))
}

// NewProvenanceClient returns the same client used by normal prepared operations.
func NewProvenanceClient(config Configuration, policy common.Hash, options ...Options) (*Client, error) {
	base, e := NewClient(config, options...)
	if e != nil {
		return nil, e
	}
	return base.WithProvenancePolicy(policy)
}
func (c *Client) WithProvenancePolicy(policy common.Hash) (*Client, error) {
	if policy == (common.Hash{}) {
		return nil, errors.New("nonzero provenance policy required")
	}
	return &Client{configuration: c.configuration, http: c.http, requiredPolicy: policy}, nil
}
func (c *Client) RequiredProvenancePolicy() common.Hash { return c.requiredPolicy }
func VerifyRequiredQuote(s SignedQuote, t Trust, data []byte, now uint64, policy common.Hash) error {
	if policy == (common.Hash{}) {
		return VerifyQuote(s, t, data, now)
	}
	return VerifyProvenanceQuote(s, t, data, now, policy)
}
func (c *Client) sidecarParams(data []byte, sidecars [][]provenance.ProofSidecar) (map[string]any, error) {
	if len(sidecars) > 1 {
		return nil, errors.New("at most one sidecar list required")
	}
	p := map[string]any{"data": "0x" + common.Bytes2Hex(data)}
	if c.requiredPolicy == (common.Hash{}) {
		if len(sidecars) > 0 && len(sidecars[0]) > 0 {
			return nil, errors.New("proof sidecars require explicit provenance policy")
		}
		return p, nil
	}
	var selected []provenance.ProofSidecar
	if len(sidecars) == 1 {
		selected = sidecars[0]
	}
	normalized, e := provenance.ProofSidecarsToWire(selected)
	if e != nil {
		return nil, e
	}
	p["proofSidecars"] = normalized
	return p, nil
}
