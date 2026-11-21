package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	return CanonicalBlock{Number: number, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}, nil
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
		Finality                    json.RawMessage
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
	var finality *ChainFinality
	if response.Finality != nil {
		finality, e = checkedFinality(ctx, c.provider, response.Finality)
		if e != nil {
			return ProtocolDataStatus{}, e
		}
	}
	return ProtocolDataStatus{IndexedHead: response.IndexedHead, Finality: finality}, nil
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

// GetMembership verifies one path with the same fences as a membership batch.
func (c *ProtocolRpcClient) GetMembership(ctx context.Context, scope DeploymentContext, input MembershipInput, block CanonicalBlock) (Membership, error) {
	values, err := c.GetMemberships(ctx, scope, []MembershipInput{input}, block)
	if err != nil {
		return Membership{}, err
	}
	return values[0], nil
}
