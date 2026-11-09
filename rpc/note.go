package rpc

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// NoteLookup contains public indexed metadata and independently authenticated membership.
type NoteLookup struct {
	Output     ProtocolNoteOutput
	Membership Membership
}

func (c *ProtocolRpcClient) GetNote(ctx context.Context, scope DeploymentContext, commitment common.Hash, block CanonicalBlock) (*NoteLookup, error) {
	deployment, e := c.context(scope)
	if e != nil {
		return nil, e
	}
	if commitment == (common.Hash{}) {
		return nil, errors.New("nonzero commitment required")
	}
	if _, e = core.HexToElement(commitment.Hex()); e != nil {
		return nil, e
	}
	if block.Number > 1<<53-1 {
		return nil, errors.New("block exceeds safe wire range")
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return nil, e
	}
	var raw json.RawMessage
	if e = c.call(ctx, scope.ChainID, "invisi_getNote", []any{commitment, block.Number}, &raw); e != nil {
		return nil, e
	}
	if string(raw) == "null" {
		return nil, nil
	}
	fields, e := wireObject(raw)
	if e != nil {
		return nil, e
	}
	output, e := parseNoteOutput(fields["output"])
	if e != nil {
		return nil, e
	}
	if output.Commitment != commitment || output.Block.Number < deployment.Protocol.FromBlock || output.Block.Number > block.Number {
		return nil, errors.New("RPC note output binding mismatch")
	}
	if _, e = c.canonical(ctx, output.Block); e != nil {
		return nil, e
	}
	proof, e := wireObject(fields["membership"])
	if e != nil {
		return nil, e
	}
	anchor, e := wireObject(proof["anchor"])
	if e != nil {
		return nil, e
	}
	root, e := wireHash(anchor["root"])
	if e != nil {
		return nil, e
	}
	membership, e := c.GetMembership(ctx, scope, MembershipInput{TreeID: output.TreeID, LeafIndex: output.LeafIndex, Commitment: commitment, Root: &root}, block)
	if e != nil {
		return nil, e
	}
	tree, e := wireUint(anchor["treeId"], 64)
	if e != nil || tree != membership.TreeID {
		return nil, errors.New("RPC note tree mismatch")
	}
	index, e := wireUint(proof["leafIndex"], 32)
	if e != nil || index != uint64(membership.LeafIndex) {
		return nil, errors.New("RPC note position mismatch")
	}
	leaf, e := wireHash(proof["commitment"])
	if e != nil || leaf != commitment {
		return nil, errors.New("RPC note commitment mismatch")
	}
	count, e := wireUint(anchor["leafCount"], 32)
	if e != nil || count <= uint64(output.LeafIndex) || count > 1<<core.TreeDepth {
		return nil, errors.New("invalid note leaf count")
	}
	var siblings []json.RawMessage
	if e = json.Unmarshal(proof["siblings"], &siblings); e != nil || len(siblings) != core.TreeDepth {
		return nil, errors.New("invalid note path")
	}
	for i, sibling := range siblings {
		h, e := wireHash(sibling)
		if e != nil || h != membership.Path[i] {
			return nil, errors.New("RPC note path mismatch")
		}
	}
	anchorBlock, e := wireBlock(anchor["block"])
	if e != nil || anchorBlock.Number > block.Number {
		return nil, errors.New("invalid note anchor")
	}
	if _, e = c.canonical(ctx, anchorBlock); e != nil {
		return nil, e
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return nil, e
	}
	return &NoteLookup{output, membership}, nil
}

// CreditStateLookup contains public indexed metadata and independently authenticated membership.
type CreditStateLookup struct {
	State      IndexedCreditState
	Membership Membership
}

func (c *ProtocolRpcClient) GetCreditState(ctx context.Context, scope DeploymentContext, commitment common.Hash, block CanonicalBlock) (*CreditStateLookup, error) {
	deployment, e := c.context(scope)
	if e != nil {
		return nil, e
	}
	if commitment == (common.Hash{}) {
		return nil, errors.New("nonzero commitment required")
	}
	if _, e = core.HexToElement(commitment.Hex()); e != nil {
		return nil, e
	}
	if block.Number > 1<<53-1 {
		return nil, errors.New("block exceeds safe wire range")
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return nil, e
	}
	var raw json.RawMessage
	if e = c.call(ctx, scope.ChainID, "invisi_getCreditState", []any{commitment, block.Number}, &raw); e != nil {
		return nil, e
	}
	if string(raw) == "null" {
		return nil, nil
	}
	fields, e := wireObject(raw)
	if e != nil {
		return nil, e
	}
	output, e := parseCreditState(fields["state"])
	if e != nil {
		return nil, e
	}
	if output.Commitment != commitment || output.Block.Number < deployment.Protocol.FromBlock || output.Block.Number > block.Number {
		return nil, errors.New("RPC Credit state output binding mismatch")
	}
	if _, e = c.canonical(ctx, output.Block); e != nil {
		return nil, e
	}
	proof, e := wireObject(fields["membership"])
	if e != nil {
		return nil, e
	}
	anchor, e := wireObject(proof["anchor"])
	if e != nil {
		return nil, e
	}
	root, e := wireHash(anchor["root"])
	if e != nil {
		return nil, e
	}
	membership, e := c.GetMembership(ctx, scope, MembershipInput{TreeID: output.TreeID, LeafIndex: output.LeafIndex, Commitment: commitment, Root: &root}, block)
	if e != nil {
		return nil, e
	}
	tree, e := wireUint(anchor["treeId"], 64)
	if e != nil || tree != membership.TreeID {
		return nil, errors.New("RPC Credit state tree mismatch")
	}
	index, e := wireUint(proof["leafIndex"], 32)
	if e != nil || index != uint64(membership.LeafIndex) {
		return nil, errors.New("RPC Credit state position mismatch")
	}
	leaf, e := wireHash(proof["commitment"])
	if e != nil || leaf != commitment {
		return nil, errors.New("RPC Credit state commitment mismatch")
	}
	count, e := wireUint(anchor["leafCount"], 32)
	if e != nil || count <= uint64(output.LeafIndex) || count > 1<<core.TreeDepth {
		return nil, errors.New("invalid Credit state leaf count")
	}
	var siblings []json.RawMessage
	if e = json.Unmarshal(proof["siblings"], &siblings); e != nil || len(siblings) != core.TreeDepth {
		return nil, errors.New("invalid Credit state path")
	}
	for i, sibling := range siblings {
		h, e := wireHash(sibling)
		if e != nil || h != membership.Path[i] {
			return nil, errors.New("RPC Credit state path mismatch")
		}
	}
	anchorBlock, e := wireBlock(anchor["block"])
	if e != nil || anchorBlock.Number > block.Number {
		return nil, errors.New("invalid Credit state anchor")
	}
	if _, e = c.canonical(ctx, anchorBlock); e != nil {
		return nil, e
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return nil, e
	}
	return &CreditStateLookup{output, membership}, nil
}
