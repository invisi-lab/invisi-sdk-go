package broadcaster

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
	"github.com/invisi-lab/invisi-sdk-go/protocol/settlement"
)

// MemberProof contains only public recursive proof evidence.
type MemberProof struct {
	FormatVersion  int           `json:"formatVersion"`
	CircuitName    string        `json:"circuitName"`
	Target         string        `json:"target"`
	BBVersion      string        `json:"bbVersion"`
	ArtifactSHA256 string        `json:"artifactSha256"`
	VKSHA256       string        `json:"vkSha256"`
	ProofHex       string        `json:"proofHex"`
	ProofSHA256    string        `json:"proofSha256"`
	PublicInputs   []common.Hash `json:"publicInputs"`
}

func aggregateCapacity(data []byte) (int, error) {
	r, e := settlement.DecodeSingle(data)
	if e != nil {
		return 0, e
	}
	c := len(r.Params.Nullifiers) + len(r.Params.NoteCommitments)
	if r.Method != "transact" || len(r.Params.Nullifiers) == 0 || c < 1 || c > 10 || r.Funding.Amount == nil || r.Funding.Amount.Sign() != 0 {
		return 0, errors.New("unfunded private C1 through C10 member required")
	}
	return c, nil
}
func validateMember(data []byte, p *MemberProof) error {
	c, e := aggregateCapacity(data)
	if e != nil {
		return e
	}
	r, e := settlement.DecodeSingle(data)
	if e != nil {
		return e
	}
	digest := regexp.MustCompile(`^[0-9a-f]{64}$`)
	if p == nil || p.FormatVersion != 1 || p.Target != "noir-recursive" || p.BBVersion != "6.0.0" || p.CircuitName != fmt.Sprintf("transact_user_slots%d", c) || !digest.MatchString(p.ArtifactSHA256) || !digest.MatchString(p.VKSHA256) || !digest.MatchString(p.ProofSHA256) || (len(p.ProofHex) != 2+458*64 || !regexp.MustCompile(`^0x[0-9a-fA-F]+$`).MatchString(p.ProofHex)) || !bytes.Equal(common.FromHex(p.ProofHex), r.Proof) || len(p.PublicInputs) != 17+3*c {
		return errors.New("invalid broadcaster member proof")
	}
	if p.PublicInputs[3].Big().Cmp(new(big.Int).SetUint64(uint64(len(r.Params.Nullifiers)))) != 0 || p.PublicInputs[4].Big().Cmp(new(big.Int).SetUint64(uint64(len(r.Params.NoteCommitments)))) != 0 {
		return errors.New("member proof public shape mismatch")
	}
	return nil
}

// Submission is the durable public JSON-RPC parameter object.
type Submission struct {
	Data          string       `json:"data"`
	Quote         Quote        `json:"quote"`
	Signature     string       `json:"signature"`
	ProofSidecars any          `json:"proofSidecars,omitempty"`
	MemberProof   *MemberProof `json:"memberProof,omitempty"`
}

func (c *Client) PrepareSubmission(data []byte, s SignedQuote, sidecars []provenance.ProofSidecar, member *MemberProof) (Submission, error) {
	var out Submission
	if s.Context == nil {
		return out, errors.New("missing verified quote context")
	}
	t := *s.Context
	t.Signer = c.configuration.QuoteSigner
	if e := VerifyQuote(s, t, data, 0); e != nil {
		return out, e
	}
	if s.Quote.FeeRecipient != c.configuration.FeeRecipient {
		return out, errors.New("unregistered fee recipient")
	}
	if (s.Quote.Execution == "aggregate") != (member != nil) {
		return out, errors.New("quote execution and member proof mismatch")
	}
	if member != nil {
		if e := validateMember(data, member); e != nil {
			return out, e
		}
	}
	params, e := sidecarParams(data, s.ProvenanceRequired, [][]provenance.ProofSidecar{sidecars})
	if e != nil {
		return out, e
	}
	out = Submission{Data: params["data"].(string), Quote: s.Quote, Signature: s.Signature, ProofSidecars: params["proofSidecars"], MemberProof: member}
	// Detach all evidence before the caller persists or sends it.
	b, e := json.Marshal(out)
	if e != nil {
		return out, e
	}
	e = json.Unmarshal(b, &out)
	return out, e
}
func (c *Client) SubmitPrepared(ctx context.Context, p Submission) (Status, error) {
	var out Status
	e := c.call(ctx, "broadcaster_submitTransaction", p, &out)
	if e != nil {
		return out, e
	}
	return out, checkStatus(out, "")
}
func (c *Client) SubmitMember(ctx context.Context, data []byte, s SignedQuote, sidecars []provenance.ProofSidecar, member *MemberProof) (Status, error) {
	p, e := c.PrepareSubmission(data, s, sidecars, member)
	if e != nil {
		return Status{}, e
	}
	return c.SubmitPrepared(ctx, p)
}
