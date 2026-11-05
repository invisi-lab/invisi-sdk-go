package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type RegisteredBroadcaster struct {
	Operator     common.Address `json:"operator"`
	Endpoint     string         `json:"endpoint"`
	MetadataURI  string         `json:"metadataURI"`
	QuoteSigner  common.Address `json:"quoteSigner"`
	FeeRecipient common.Address `json:"feeRecipient"`
	Enabled      bool           `json:"enabled"`
	Block        CanonicalBlock `json:"block"`
}
type BroadcasterQuery struct {
	Enabled *bool
	After   *common.Address
	Limit   int
	Block   *CanonicalBlock
}
type BroadcasterPage struct {
	Block        CanonicalBlock          `json:"block"`
	Broadcasters []RegisteredBroadcaster `json:"broadcasters"`
	Next         *common.Address         `json:"next,omitempty"`
}

func (c *ProtocolRpcClient) GetBroadcasters(ctx context.Context, d DeploymentContext, q BroadcasterQuery) (BroadcasterPage, error) {
	var out BroadcasterPage
	deployment, e := c.context(d)
	if e != nil {
		return out, e
	}
	registry := deployment.BroadcasterRegistry
	if registry == nil {
		return out, errors.New("deployment has no broadcaster registry")
	}
	if q.Limit == 0 {
		q.Limit = 100
	}
	if q.Limit < 1 || q.Limit > 100 || q.After != nil && (q.Block == nil || *q.After == (common.Address{})) {
		return out, errors.New("invalid broadcaster query")
	}
	if q.Block != nil && q.Block.Number < registry.FromBlock {
		return out, errors.New("broadcaster anchor predates deployment")
	}
	p := map[string]any{"limit": q.Limit}
	if q.Enabled != nil {
		p["enabled"] = *q.Enabled
	}
	if q.After != nil {
		p["after"] = q.After.Hex()
	}
	if q.Block != nil {
		p["block"] = q.Block
	}
	var raw json.RawMessage
	if e = c.call(ctx, d.ChainID, "invisi_getBroadcasters", []any{p}, &raw); e != nil {
		return out, e
	}
	m, e := wireObject(raw)
	if e != nil {
		return out, e
	}
	for k := range m {
		if k != "block" && k != "broadcasters" && k != "next" {
			return out, errors.New("unsupported broadcaster page fields")
		}
	}
	out.Block, e = wireBlock(m["block"])
	if e != nil {
		return out, e
	}
	var entries []map[string]json.RawMessage
	if json.Unmarshal(m["broadcasters"], &entries) != nil || entries == nil {
		return out, errors.New("invalid broadcaster entries")
	}
	out.Broadcasters = make([]RegisteredBroadcaster, 0, len(entries))
	for _, entry := range entries {
		for _, key := range []string{"operator", "endpoint", "metadataURI", "quoteSigner", "feeRecipient", "enabled", "block"} {
			if _, ok := entry[key]; !ok {
				return out, errors.New("missing broadcaster field")
			}
		}
		if string(entry["enabled"]) != "true" && string(entry["enabled"]) != "false" {
			return out, errors.New("invalid broadcaster enabled flag")
		}
		for _, key := range []string{"endpoint", "metadataURI"} {
			var value *string
			if json.Unmarshal(entry[key], &value) != nil || value == nil {
				return out, errors.New("invalid broadcaster URI")
			}
		}
		for key := range entry {
			switch key {
			case "operator", "endpoint", "metadataURI", "quoteSigner", "feeRecipient", "enabled", "block":
			default:
				return out, errors.New("unsupported broadcaster fields")
			}
		}
		block, err := wireBlock(entry["block"])
		if err != nil {
			return out, err
		}
		delete(entry, "block")
		encoded, err := json.Marshal(entry)
		if err != nil {
			return out, err
		}
		var b RegisteredBroadcaster
		if json.Unmarshal(encoded, &b) != nil {
			return out, errors.New("invalid broadcaster registration")
		}
		b.Block = block
		out.Broadcasters = append(out.Broadcasters, b)
	}
	if next, ok := m["next"]; ok {
		if string(next) == "null" || json.Unmarshal(next, &out.Next) != nil || out.Next == nil || *out.Next == (common.Address{}) {
			return out, errors.New("invalid broadcaster cursor")
		}
	}
	if out.Block.Number < registry.FromBlock || q.Block != nil && out.Block != *q.Block || out.Broadcasters == nil || len(out.Broadcasters) > q.Limit {
		return out, errors.New("invalid broadcaster page")
	}
	previous := common.Address{}
	if q.After != nil {
		previous = *q.After
	}
	for _, b := range out.Broadcasters {
		if b.Operator.Big().Cmp(previous.Big()) <= 0 || b.QuoteSigner == (common.Address{}) || b.FeeRecipient == (common.Address{}) || q.Enabled != nil && b.Enabled != *q.Enabled {
			return out, errors.New("invalid broadcaster registration")
		}
		previous = b.Operator
		for _, s := range []string{b.Endpoint, b.MetadataURI} {
			if len(s) > 2048 {
				return out, errors.New("invalid broadcaster URI")
			}
			for _, r := range s {
				if r < 32 || r > 126 {
					return out, errors.New("invalid broadcaster URI")
				}
			}
		}
		if !strings.HasPrefix(b.Endpoint, "https://") || len(b.Endpoint) <= 8 || b.Block.Number < registry.FromBlock || b.Block.Number > out.Block.Number || b.Block.Number == out.Block.Number && b.Block != out.Block {
			return out, errors.New("invalid broadcaster registration anchor")
		}
	}
	if out.Next != nil && (len(out.Broadcasters) != q.Limit || *out.Next != previous) {
		return out, errors.New("invalid broadcaster cursor")
	}
	return out, nil
}
