package broadcaster

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"math/big"
	"time"
)

const ProvenanceRequestHashFormat = "invisi-broadcaster-provenance-intent-v1"

type ProvenanceSignedQuote struct {
	Quote              Quote        `json:"quote"`
	Signature          string       `json:"signature"`
	ProvenancePolicyID *common.Hash `json:"provenancePolicyId"`
	RequestHashFormat  string       `json:"requestHashFormat"`
	Context            *Trust       `json:"-"`
}

func (s *SignedQuote) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if e := json.Unmarshal(data, &raw); e != nil {
		return e
	}
	if _, ok := raw["provenancePolicyId"]; ok {
		return errors.New("unexpected provenance policy on legacy broadcaster")
	}
	if _, ok := raw["requestHashFormat"]; ok {
		return errors.New("unexpected provenance request hash format")
	}
	type plain SignedQuote
	var decoded plain
	if e := json.Unmarshal(data, &decoded); e != nil {
		return e
	}
	*s = SignedQuote(decoded)
	return nil
}
func VerifyProvenanceQuote(s ProvenanceSignedQuote, t Trust, data []byte, now uint64, requiredPolicy common.Hash) error {
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

type ProvenanceClient struct {
	base           *Client
	requiredPolicy common.Hash
}

func NewProvenanceClient(config Configuration, requiredPolicy common.Hash, options ...Options) (*ProvenanceClient, error) {
	base, e := NewClient(config, options...)
	if e != nil {
		return nil, e
	}
	return &ProvenanceClient{base, requiredPolicy}, nil
}
func (c *ProvenanceClient) Fee(ctx context.Context, token common.Address, data []byte) (Fee, error) {
	return c.base.Fee(ctx, token, data)
}
func (c *ProvenanceClient) Quote(ctx context.Context, data []byte, trust Trust, sidecars []provenance.ProofSidecar) (ProvenanceSignedQuote, error) {
	var s ProvenanceSignedQuote
	normalized, e := provenance.ProofSidecarsToWire(sidecars)
	if e != nil {
		return s, e
	}
	if e := c.base.call(ctx, "broadcaster_getQuote", map[string]any{"data": "0x" + common.Bytes2Hex(data), "proofSidecars": normalized}, &s); e != nil {
		return s, e
	}
	trust.Signer = c.base.configuration.QuoteSigner
	if e := VerifyProvenanceQuote(s, trust, data, uint64(time.Now().Unix()), c.requiredPolicy); e != nil {
		return s, e
	}
	if s.Quote.FeeRecipient != c.base.configuration.FeeRecipient {
		return s, errors.New("unregistered fee recipient")
	}
	trust.ChainID = new(big.Int).Set(trust.ChainID)
	s.Context = &trust
	return s, nil
}
func (c *ProvenanceClient) Submit(ctx context.Context, data []byte, s ProvenanceSignedQuote, key string, sidecars []provenance.ProofSidecar) (Status, error) {
	var out Status
	if s.Context == nil {
		return out, errors.New("missing verified quote context")
	}
	trust := *s.Context
	trust.Signer = c.base.configuration.QuoteSigner
	if e := VerifyProvenanceQuote(s, trust, data, 0, c.requiredPolicy); e != nil {
		return out, e
	}
	if s.Quote.FeeRecipient != c.base.configuration.FeeRecipient {
		return out, errors.New("unregistered fee recipient")
	}
	normalized, e := provenance.ProofSidecarsToWire(sidecars)
	if e != nil {
		return out, e
	}
	if e := c.base.call(ctx, "broadcaster_submitTransaction", map[string]any{"data": "0x" + common.Bytes2Hex(data), "quote": s.Quote, "signature": s.Signature, "idempotencyKey": key, "proofSidecars": normalized}, &out); e != nil {
		return out, e
	}
	return out, checkStatus(out, key)
}
func (c *ProvenanceClient) Status(ctx context.Context, id string) (Status, error) {
	return c.base.Status(ctx, id)
}
