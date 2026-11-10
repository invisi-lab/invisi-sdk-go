package broadcaster

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ServiceStatus struct {
	APIVersion       string         `json:"apiVersion"`
	ProtocolVersion  string         `json:"protocolVersion"`
	ChainID          string         `json:"chainId"`
	Protocol         common.Address `json:"protocol"`
	Sender           common.Address `json:"sender"`
	FeeRecipient     common.Address `json:"feeRecipient"`
	SubmissionHalted bool           `json:"submissionHalted"`
	Provenance       struct {
		Enabled           bool   `json:"enabled"`
		RequestHashFormat string `json:"requestHashFormat"`
		TrackingRequired  bool   `json:"trackingRequired"`
		TrackingSupported bool   `json:"trackingSupported"`
	} `json:"provenance"`
}

func (c *Client) SubmissionRequirements(ctx context.Context, s SignedQuote) (bool, error) {
	if s.Context == nil {
		return false, errors.New("missing verified quote context")
	}
	var status ServiceStatus
	if e := c.call(ctx, "broadcaster_getStatus", nil, &status); e != nil {
		return false, e
	}
	chain, ok := new(big.Int).SetString(status.ChainID, 10)
	if !ok || status.APIVersion != "v1" || status.ProtocolVersion != "1" || chain.Cmp(s.Context.ChainID) != 0 || status.Protocol != s.Context.Protocol || status.Sender != c.configuration.QuoteSigner || status.FeeRecipient != c.configuration.FeeRecipient {
		return false, errors.New("invalid broadcaster service status")
	}
	if status.SubmissionHalted {
		return false, errors.New("broadcaster submission halted")
	}
	p := status.Provenance
	if p.Enabled != s.ProvenanceRequired || (p.Enabled && (p.RequestHashFormat != ProvenanceRequestHashFormat || !p.TrackingRequired || !p.TrackingSupported)) {
		return false, errors.New("broadcaster provenance requirement changed; obtain a new quote")
	}
	return p.Enabled, nil
}

func (c *Client) SupportsToken(ctx context.Context, token common.Address, trust Trust) (bool, error) {
	if token == (common.Address{}) || trust.ChainID == nil {
		return false, errors.New("token and deployment context required")
	}
	var raw struct {
		ServiceStatus
		TokenAddress common.Address `json:"tokenAddress"`
		Supported    bool           `json:"supported"`
	}
	if e := c.call(ctx, "broadcaster_getTokenSupport", map[string]any{"tokenAddress": token.Hex()}, &raw); e != nil {
		return false, e
	}
	n, ok := new(big.Int).SetString(raw.ChainID, 10)
	if !ok || raw.APIVersion != "v1" || raw.ProtocolVersion != "1" || n.Cmp(trust.ChainID) != 0 || raw.Protocol != trust.Protocol || raw.Sender != c.configuration.QuoteSigner || raw.FeeRecipient != c.configuration.FeeRecipient || raw.TokenAddress != token {
		return false, errors.New("invalid broadcaster token support")
	}
	return !raw.SubmissionHalted && raw.Supported, nil
}
