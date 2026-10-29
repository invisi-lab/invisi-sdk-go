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
	"github.com/invisi-lab/invisi-sdk-go/core"
)

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
		DetectionEnabled            *bool
		CheckedBits                 int
	}
	if e = c.call(ctx, context.ChainID, "invisi_getStatus", []any{}, &response); e != nil {
		return ProtocolDataStatus{}, e
	}
	a, _ := json.Marshal(d)
	b, _ := json.Marshal(response.Deployment)
	if response.APIVersion != "v1" || response.ProtocolVersion != "1" || string(a) != string(b) || response.CheckedBits < 1 || response.CheckedBits > 24 || response.DetectionEnabled == nil {
		return ProtocolDataStatus{}, errors.New("invalid indexed status")
	}
	if response.IndexedHead != nil {
		canonical, e := c.canonical(ctx, *response.IndexedHead)
		if e != nil {
			return ProtocolDataStatus{}, e
		}
		if canonical.Number < d.Protocol.FromBlock {
			return ProtocolDataStatus{}, errors.New("indexed head predates deployment")
		}
		response.IndexedHead = &canonical
	}
	return ProtocolDataStatus{response.IndexedHead, *response.DetectionEnabled, response.CheckedBits}, nil
}
func (c *ProtocolRpcClient) GetUpdates(ctx context.Context, context DeploymentContext, r ProtocolUpdateRange) (ProtocolUpdates, error) {
	d, e := c.context(context)
	if e != nil {
		return ProtocolUpdates{}, e
	}
	if r.FromBlock < d.Protocol.FromBlock || r.ToBlock < r.FromBlock || r.ToBlock-r.FromBlock > 99 || r.ToBlock > 1<<53-1 {
		return ProtocolUpdates{}, errors.New("updates require at most 100 deployment blocks")
	}
	status, e := c.GetStatus(ctx, context)
	if e != nil {
		return ProtocolUpdates{}, e
	}
	if status.IndexedHead == nil || status.IndexedHead.Number < r.ToBlock {
		return ProtocolUpdates{}, errors.New("Protocol RPC is behind the requested range")
	}
	block, e := c.block(ctx, r.ToBlock)
	if e != nil {
		return ProtocolUpdates{}, e
	}
	anchor := struct {
		FromBlock, ToBlock uint64
		BlockHash          common.Hash
	}{r.FromBlock, r.ToBlock, block.Hash}
	// Wire field names are explicit and shared with the TS client.
	wire := map[string]any{"fromBlock": anchor.FromBlock, "toBlock": anchor.ToBlock, "blockHash": anchor.BlockHash}
	type ref struct {
		TransactionHash common.Hash     `json:"transactionHash"`
		LogIndex        *uint64         `json:"logIndex"`
		Block           *CanonicalBlock `json:"block"`
	}
	var response struct {
		BlockHash  common.Hash `json:"blockHash"`
		Outputs    []ref       `json:"outputs"`
		Candidates []struct {
			Output   ref   `json:"output"`
			Fallback *bool `json:"fallback"`
		} `json:"candidates"`
		CreditTransitions []ref `json:"creditTransitions"`
		CheckedBits       int   `json:"checkedBits"`
	}
	method := "invisi_getUpdates"
	params := []any{wire}
	if len(r.DetectionKey) > 0 {
		if _, e := core.DeriveClueKey(r.DetectionKey); e != nil {
			return ProtocolUpdates{}, e
		}
		method = "invisi_detectOutputs"
		params = append(params, "0x"+common.Bytes2Hex(r.DetectionKey))
	}
	if e = c.call(ctx, context.ChainID, method, params, &response); e != nil {
		return ProtocolUpdates{}, e
	}
	if response.BlockHash != block.Hash {
		return ProtocolUpdates{}, errors.New("RPC update anchor mismatch")
	}
	refs := response.Outputs
	if len(r.DetectionKey) > 0 {
		if response.CheckedBits < 1 || response.CheckedBits > 24 {
			return ProtocolUpdates{}, errors.New("invalid detection precision")
		}
		refs = nil
		for _, candidate := range response.Candidates {
			if candidate.Fallback == nil {
				return ProtocolUpdates{}, errors.New("invalid detection candidate")
			}
			refs = append(refs, candidate.Output)
		}
	}
	seen := map[common.Hash]bool{}
	result := ProtocolUpdates{Block: block}
	verifiedBlocks := map[uint64]CanonicalBlock{block.Number: block}
	for _, ref := range append(refs, response.CreditTransitions...) {
		if ref.TransactionHash == (common.Hash{}) || ref.LogIndex == nil {
			return ProtocolUpdates{}, errors.New("invalid transaction reference")
		}
		if ref.Block == nil || ref.Block.Number < r.FromBlock || ref.Block.Number > r.ToBlock {
			return ProtocolUpdates{}, errors.New("missing or out-of-range transaction block")
		}
		canonical, ok := verifiedBlocks[ref.Block.Number]
		if !ok {
			canonical, e = c.canonical(ctx, *ref.Block)
			if e != nil {
				return ProtocolUpdates{}, e
			}
			verifiedBlocks[canonical.Number] = canonical
		}
		if canonical != *ref.Block {
			return ProtocolUpdates{}, errors.New("RPC transaction block context mismatch")
		}
		result.References = append(result.References, TransactionReference{TransactionHash: ref.TransactionHash, LogIndex: *ref.LogIndex, Block: canonical})
		if !seen[ref.TransactionHash] {
			seen[ref.TransactionHash] = true
			result.TransactionHashes = append(result.TransactionHashes, ref.TransactionHash)
		}
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return ProtocolUpdates{}, e
	}
	return result, nil
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
	if input.TreeID == 0 || input.TreeID > 1<<53-1 || input.LeafIndex >= 1<<core.TreeDepth {
		return result, errors.New("invalid membership position")
	}
	leaf, e := core.HexToElement(input.Commitment.Hex())
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
		if r.TreeID != input.TreeID || r.LeafCount > 1<<core.TreeDepth {
			return result, errors.New("invalid root metadata")
		}
		if r.Block.Number > block.Number || r.Block.Number < d.Protocol.FromBlock || r.LeafCount <= input.LeafIndex || (input.Root != nil && r.Root != *input.Root) {
			continue
		}
		if _, e = core.HexToElement(r.Root.Hex()); e != nil {
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
	if response.Anchor != *selected || response.LeafIndex != input.LeafIndex || response.Commitment != input.Commitment || len(response.Path) != core.TreeDepth {
		return result, errors.New("RPC membership binding mismatch")
	}
	current := leaf
	for level, sibling := range response.Path {
		s, e := core.HexToElement(sibling.Hex())
		if e != nil {
			return result, e
		}
		if input.LeafIndex>>level&1 == 0 {
			current, e = core.Poseidon2Hash(current, s)
		} else {
			current, e = core.Poseidon2Hash(s, current)
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
