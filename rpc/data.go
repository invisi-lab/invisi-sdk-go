package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

const MaxUpdateBlocks uint64 = 10000

func (c *ProtocolRpcClient) block(ctx context.Context, number uint64) (CanonicalBlock, error) {
	if c.provider == nil {
		return CanonicalBlock{}, errors.New("Chain RPC provider required")
	}
	h, e := c.provider.HeaderByNumber(ctx, new(big.Int).SetUint64(number))
	if e != nil {
		return CanonicalBlock{}, e
	}
	if h == nil || h.Number == nil || !h.Number.IsUint64() || h.Number.Uint64() != number {
		return CanonicalBlock{}, errors.New("incorrect canonical block")
	}
	return CanonicalBlock{number, h.Hash(), h.ParentHash, h.Time}, nil
}
func (c *ProtocolRpcClient) canonical(ctx context.Context, remote CanonicalBlock) (CanonicalBlock, error) {
	b, e := c.block(ctx, remote.Number)
	if e != nil {
		return b, e
	}
	if b.Hash != remote.Hash || b.ParentHash != remote.ParentHash || b.Timestamp != remote.Timestamp {
		return b, errors.New("RPC canonical block mismatch")
	}
	return b, nil
}
func (c *ProtocolRpcClient) GetStatus(ctx context.Context, context DeploymentContext) (ProtocolDataStatus, error) {
	return c.readStatus(ctx, context, true)
}
func (c *ProtocolRpcClient) readStatus(ctx context.Context, context DeploymentContext, checkCanonical bool) (ProtocolDataStatus, error) {
	if c.provider == nil {
		return ProtocolDataStatus{}, errors.New("Chain RPC provider required")
	}
	chain, e := c.provider.ChainID(ctx)
	if e != nil {
		return ProtocolDataStatus{}, e
	}
	if !chain.IsUint64() || chain.Uint64() != context.ChainID {
		return ProtocolDataStatus{}, errors.New("provider chain mismatch")
	}

	d, e := c.context(context)
	if e != nil {
		return ProtocolDataStatus{}, e
	}
	var response struct {
		APIVersion, ProtocolVersion string
		Deployment                  ProtocolDeployment
		IndexedHead                 *CanonicalBlock
	}
	if e = c.call(ctx, context.ChainID, "invisi_getStatus", []any{}, &response); e != nil {
		return ProtocolDataStatus{}, e
	}
	a, _ := json.Marshal(d)
	b, _ := json.Marshal(response.Deployment)
	if response.APIVersion != "v1" || response.ProtocolVersion != "1" || string(a) != string(b) {
		return ProtocolDataStatus{}, errors.New("invalid indexed status")
	}
	if response.IndexedHead != nil && (response.IndexedHead.Number < d.Protocol.FromBlock || response.IndexedHead.Number > 1<<53-1 || response.IndexedHead.Timestamp > 1<<53-1) {
		return ProtocolDataStatus{}, errors.New("invalid indexed head bounds")
	}
	if checkCanonical && response.IndexedHead != nil {
		canonical, e := c.canonical(ctx, *response.IndexedHead)
		if e != nil {
			return ProtocolDataStatus{}, e
		}
		if canonical.Number < d.Protocol.FromBlock {
			return ProtocolDataStatus{}, errors.New("indexed head predates deployment")
		}
		response.IndexedHead = &canonical
	}
	return ProtocolDataStatus{IndexedHead: response.IndexedHead}, nil
}
func (c *ProtocolRpcClient) knownRoot(ctx context.Context, context DeploymentContext, tree uint64, root common.Hash, block CanonicalBlock) (bool, error) {
	data := append([]byte{}, crypto.Keccak256([]byte("rootHistory(uint256,bytes32)"))[:4]...)
	data = append(data, common.LeftPadBytes(new(big.Int).SetUint64(tree).Bytes(), 32)...)
	data = append(data, root[:]...)
	out, e := c.provider.CallContract(ctx, ethereum.CallMsg{To: &context.ProtocolAddress, Data: data}, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return false, e
	}
	if len(out) != 32 || new(big.Int).SetBytes(out).Cmp(big.NewInt(1)) > 0 {
		return false, errors.New("invalid rootHistory response")
	}
	return out[31] == 1, nil
}
func (c *ProtocolRpcClient) GetMembership(ctx context.Context, context DeploymentContext, input MembershipInput, block CanonicalBlock) (Membership, error) {
	var result Membership
	d, e := c.context(context)
	if e != nil {
		return result, e
	}
	if input.TreeID == 0 || input.TreeID > 1<<53-1 || input.LeafIndex >= 1<<primitives.TreeDepth {
		return result, errors.New("invalid membership position")
	}
	leaf, e := primitives.HexToElement(input.Commitment.Hex())
	if e != nil {
		return result, e
	}
	if _, e = c.GetStatus(ctx, context); e != nil {
		return result, e
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return result, e
	}
	var roots []rootAnchor
	if e = c.call(ctx, context.ChainID, "invisi_getRoots", []any{input.TreeID}, &roots); e != nil {
		return result, e
	}
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].Block.Number == roots[j].Block.Number {
			return roots[i].LeafCount > roots[j].LeafCount
		}
		return roots[i].Block.Number > roots[j].Block.Number
	})
	var selected *rootAnchor
	for i := range roots {
		r := &roots[i]
		if r.TreeID != input.TreeID || r.LeafCount > 1<<primitives.TreeDepth {
			return result, errors.New("invalid root metadata")
		}
		if r.Block.Number > block.Number || r.Block.Number < d.Protocol.FromBlock || r.LeafCount <= input.LeafIndex || (input.Root != nil && r.Root != *input.Root) {
			continue
		}
		if _, e = primitives.HexToElement(r.Root.Hex()); e != nil {
			return result, e
		}
		if _, e = c.canonical(ctx, r.Block); e != nil {
			return result, e
		}
		known, e := c.knownRoot(ctx, context, input.TreeID, r.Root, block)
		if e != nil {
			return result, e
		}
		if known {
			selected = r
			break
		}
	}
	if selected == nil {
		return result, errors.New("no chain-known indexed membership root")
	}
	var response struct {
		Anchor     rootAnchor    `json:"anchor"`
		LeafIndex  uint32        `json:"leafIndex"`
		Commitment common.Hash   `json:"commitment"`
		Path       []common.Hash `json:"siblings"`
	}
	if e = c.call(ctx, context.ChainID, "invisi_getMembership", []any{input.TreeID, input.LeafIndex, selected.Root}, &response); e != nil {
		return result, e
	}
	if response.Anchor != *selected || response.LeafIndex != input.LeafIndex || response.Commitment != input.Commitment || len(response.Path) != primitives.TreeDepth {
		return result, errors.New("RPC membership binding mismatch")
	}
	current := leaf
	for level, sibling := range response.Path {
		s, e := primitives.HexToElement(sibling.Hex())
		if e != nil {
			return result, e
		}
		if input.LeafIndex>>level&1 == 0 {
			current, e = primitives.Poseidon2Hash(current, s)
		} else {
			current, e = primitives.Poseidon2Hash(s, current)
		}
		if e != nil {
			return result, e
		}
	}
	if common.Hash(current.Bytes()) != selected.Root {
		return result, fmt.Errorf("invalid Merkle path")
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return result, e
	}
	return Membership{input.TreeID, input.LeafIndex, input.Commitment, selected.Root, append([]common.Hash(nil), response.Path...), block}, nil
}
