package rpc

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
)

// MaxMembershipBatch bounds per-leaf work without extending the RPC wire API.
const MaxMembershipBatch = 256

// GetMemberships authenticates an ordered batch against one canonical block.
// Inputs from the same tree share one root covering their highest leaf. All
// evidence is private to this call; no mutable status survives into another call.
func (c *ProtocolRpcClient) GetMemberships(ctx context.Context, scope DeploymentContext, inputs []MembershipInput, block CanonicalBlock) ([]Membership, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(inputs) > MaxMembershipBatch {
		return nil, errors.New("membership batch exceeds 256 inputs")
	}
	if len(inputs) == 0 {
		return []Membership{}, nil
	}
	d, err := c.context(scope)
	if err != nil {
		return nil, err
	}
	type treeGroup struct {
		tree     uint64
		highest  uint32
		root     *common.Hash
		selected rootAnchor
	}
	groups := []*treeGroup{}
	byTree := map[uint64]*treeGroup{}
	for _, input := range inputs {
		if input.TreeID == 0 || input.TreeID > 1<<53-1 || input.LeafIndex >= 1<<primitives.TreeDepth {
			return nil, errors.New("invalid membership position")
		}
		if _, err := primitives.HexToElement(input.Commitment.Hex()); err != nil {
			return nil, err
		}
		g := byTree[input.TreeID]
		if g == nil {
			g = &treeGroup{tree: input.TreeID}
			byTree[input.TreeID] = g
			groups = append(groups, g)
		}
		if input.LeafIndex > g.highest {
			g.highest = input.LeafIndex
		}
		if input.Root != nil {
			if g.root != nil && *g.root != *input.Root {
				return nil, errors.New("conflicting roots for one membership tree")
			}
			root := *input.Root
			g.root = &root
		}
	}
	if _, err = c.GetStatus(ctx, scope); err != nil {
		return nil, err
	}
	if _, err = c.canonical(ctx, block); err != nil {
		return nil, err
	}
	for _, g := range groups {
		var roots []rootAnchor
		if err = c.call(ctx, scope.ChainID, "invisi_getRoots", []any{g.tree}, &roots); err != nil {
			return nil, err
		}
		sort.Slice(roots, func(i, j int) bool {
			if roots[i].Block.Number == roots[j].Block.Number {
				return roots[i].LeafCount > roots[j].LeafCount
			}
			return roots[i].Block.Number > roots[j].Block.Number
		})
		found := false
		for _, root := range roots {
			if root.TreeID != g.tree || root.LeafCount > 1<<primitives.TreeDepth {
				return nil, errors.New("invalid root metadata")
			}
			if root.Block.Number > block.Number || root.Block.Number < d.Protocol.FromBlock || root.LeafCount <= g.highest || (g.root != nil && root.Root != *g.root) {
				continue
			}
			if _, err = primitives.HexToElement(root.Root.Hex()); err != nil {
				return nil, err
			}
			if _, err = c.canonical(ctx, root.Block); err != nil {
				return nil, err
			}
			known, e := c.knownRoot(ctx, scope, g.tree, root.Root, block)
			if e != nil {
				return nil, e
			}
			if known {
				g.selected = root
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("no chain-known indexed membership root")
		}
	}
	results := make([]Membership, len(inputs))
	failures := make([]error, len(inputs))
	var mu sync.Mutex
	next := 0
	var workers sync.WaitGroup
	for range min(4, len(inputs)) {
		workers.Go(func() {
			for {
				mu.Lock()
				i := next
				next++
				mu.Unlock()
				if i >= len(inputs) {
					return
				}
				if e := ctx.Err(); e != nil {
					failures[i] = e
					continue
				}
				input := inputs[i]
				selected := byTree[input.TreeID].selected
				results[i], failures[i] = c.membershipPath(ctx, scope, input, selected, block)
			}
		})
	}
	workers.Wait()
	for _, e := range failures {
		if e != nil {
			return nil, e
		}
	}
	// A batch shares root evidence across paths. Recheck every distinct numeric
	// root anchor before publishing that shared evidence, not just the target.
	anchors := map[uint64]CanonicalBlock{block.Number: block}
	for _, group := range groups {
		anchor := group.selected.Block
		if prior, exists := anchors[anchor.Number]; exists && prior != anchor {
			return nil, errors.New("conflicting membership canonical anchors")
		}
		anchors[anchor.Number] = anchor
	}
	numbers := make([]uint64, 0, len(anchors))
	for number := range anchors {
		if number != block.Number {
			numbers = append(numbers, number)
		}
	}
	sort.Slice(numbers, func(i, j int) bool { return numbers[i] < numbers[j] })
	for _, number := range numbers {
		if _, err = c.canonical(ctx, anchors[number]); err != nil {
			return nil, err
		}
	}
	if _, err = c.canonical(ctx, block); err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *ProtocolRpcClient) membershipPath(ctx context.Context, scope DeploymentContext, input MembershipInput, selected rootAnchor, block CanonicalBlock) (Membership, error) {
	var response struct {
		Anchor     rootAnchor    `json:"anchor"`
		LeafIndex  uint32        `json:"leafIndex"`
		Commitment common.Hash   `json:"commitment"`
		Path       []common.Hash `json:"siblings"`
	}
	if err := c.call(ctx, scope.ChainID, "invisi_getMembership", []any{input.TreeID, input.LeafIndex, selected.Root}, &response); err != nil {
		return Membership{}, err
	}
	if response.Anchor != selected || response.LeafIndex != input.LeafIndex || response.Commitment != input.Commitment || len(response.Path) != primitives.TreeDepth {
		return Membership{}, errors.New("RPC membership binding mismatch")
	}
	current, err := primitives.HexToElement(input.Commitment.Hex())
	if err != nil {
		return Membership{}, err
	}
	for level, sibling := range response.Path {
		s, err := primitives.HexToElement(sibling.Hex())
		if err != nil {
			return Membership{}, err
		}
		if input.LeafIndex>>level&1 == 0 {
			current, err = primitives.Poseidon2Hash(current, s)
		} else {
			current, err = primitives.Poseidon2Hash(s, current)
		}
		if err != nil {
			return Membership{}, err
		}
	}
	if common.Hash(current.Bytes()) != selected.Root {
		return Membership{}, errors.New("invalid Merkle path")
	}
	return Membership{input.TreeID, input.LeafIndex, input.Commitment, selected.Root, append([]common.Hash(nil), response.Path...), block}, nil
}
