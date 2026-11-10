package rpc

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func (c *ProtocolRpcClient) GetReceivingAddress(ctx context.Context, owner common.Address) (*RegisteredReceivingAddress, error) {
	if owner == (common.Address{}) {
		return nil, errors.New("invalid receiving-address owner")
	}
	c.mu.Lock()
	var d ProtocolDeployment
	if c.deployment != nil {
		d = cloneDeployment(*c.deployment)
	}
	c.mu.Unlock()
	if d.Registry == nil {
		return nil, errors.New("deployment has no receiving-key registry")
	}
	chain, e := d.ChainNumber()
	if e != nil {
		return nil, e
	}
	var raw json.RawMessage
	if e = c.call(ctx, chain.Uint64(), "invisi_getReceivingAddress", []any{owner.Hex()}, &raw); e != nil {
		return nil, e
	}
	if string(raw) == "null" {
		return nil, nil
	}
	m, e := wireObject(raw)
	if e != nil {
		return nil, e
	}
	identity, e := wireBytes(m["ownerAddress"], 20, 20)
	if e != nil || common.BytesToAddress(identity) != owner {
		return nil, errors.New("RPC receiving-address identity mismatch")
	}
	view, e := wireBytes(m["viewingPublicKey"], 33, 33)
	if e != nil {
		return nil, e
	}
	clue, e := wireBytes(m["clueKey"], 32, 32)
	if e != nil {
		return nil, e
	}
	address, e := primitives.EncodeLongAddress(primitives.ReceivingAddress{OwnerAddress: primitives.BigIntToElement(owner.Big()), ViewingPublicKey: view, ClueKey: clue})
	if e != nil {
		return nil, e
	}
	block, e := wireBlock(m["block"])
	if e != nil {
		return nil, e
	}
	if block.Number < d.Registry.FromBlock {
		return nil, errors.New("registry index predates deployment")
	}
	nonce, e := wireUint(m["nonce"], 64)
	if e != nil {
		return nil, e
	}
	return &RegisteredReceivingAddress{address, nonce, block}, nil
}
