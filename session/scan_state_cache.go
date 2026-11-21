package session

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
)

type scanStateCacheKey struct{}
type spentCheckKey struct {
	Block     protocolrpc.CanonicalBlock
	Nullifier common.Hash
}
type statusCheckKey struct {
	Block     protocolrpc.CanonicalBlock
	Namespace string
}
type scanStateCache struct {
	spent    map[spentCheckKey]bool
	statuses map[statusCheckKey]uint8
}

// A scan is serialized by refreshGate. Cache only successful immutable state at
// an exact numeric block/hash, never latest, errors, or cross-invocation results.
// Canonicality is still checked on every checkKnownState call and commit fence.
func withScanStateCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, scanStateCacheKey{}, &scanStateCache{spent: map[spentCheckKey]bool{}, statuses: map[statusCheckKey]uint8{}})
}
func scanCache(ctx context.Context) *scanStateCache {
	cache, _ := ctx.Value(scanStateCacheKey{}).(*scanStateCache)
	return cache
}

func (c *scanStateCache) trim() {
	if len(c.spent)+len(c.statuses) >= 16384 {
		clear(c.spent)
		clear(c.statuses)
	}
}
