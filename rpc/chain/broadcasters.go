package chain

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

// BroadcasterOptions filters registered operators by a live token capability probe.
type BroadcasterOptions struct {
	TokenAddress  common.Address
	ClientOptions broadcaster.Options
}

// GetBroadcasters returns enabled registered clients that advertise fees for this token.
// Unsupported and unreachable operators are skipped. Registry pagination remains pinned
// to the first page's indexed block; this is public publisher metadata, not endorsement.
func (p *Protocol) GetBroadcasters(ctx context.Context, options BroadcasterOptions) ([]*broadcaster.Client, error) {
	if p.deployment.BroadcasterRegistry == nil {
		return nil, errors.New("deployment has no broadcaster registry")
	}
	if p.rpc == nil {
		return nil, errors.New("broadcaster discovery requires ProtocolFromRPC")
	}
	if options.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address required")
	}
	enabled := true
	q := protocolrpc.BroadcasterQuery{Enabled: &enabled, Limit: 100}
	clients := []*broadcaster.Client{}
	seen := map[common.Address]bool{}
	for {
		page, e := p.rpc.GetBroadcasters(ctx, p.Context(), q)
		if e != nil {
			return nil, e
		}
		if q.Block == nil {
			block := page.Block
			q.Block = &block
		}
		for _, entry := range page.Broadcasters {
			if !entry.Enabled || seen[entry.Operator] {
				continue
			}
			seen[entry.Operator] = true
			c, e := broadcaster.NewClient(broadcaster.Configuration{Endpoint: entry.Endpoint, QuoteSigner: entry.QuoteSigner, FeeRecipient: entry.FeeRecipient}, options.ClientOptions)
			if e != nil {
				continue
			}
			supported, e := c.SupportsToken(ctx, options.TokenAddress, broadcaster.Trust{ChainID: new(big.Int).SetUint64(p.Context().ChainID), Protocol: p.Context().ProtocolAddress})
			if e != nil || !supported {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				continue
			}
			clients = append(clients, c)
		}
		if page.Next == nil {
			return clients, nil
		}
		if q.After != nil && page.Next.Big().Cmp(q.After.Big()) <= 0 {
			return nil, errors.New("nonadvancing broadcaster cursor")
		}
		cursor := *page.Next
		q.After = &cursor
	}
}
