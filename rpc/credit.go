package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func (c *ProtocolRpcClient) parseTransitions(ctx context.Context, rows []json.RawMessage, r ProtocolUpdateRange) ([]ProtocolCreditTransition, error) {
	out := []ProtocolCreditTransition{}
	seen := map[string]string{}
	for _, row := range rows {
		m, e := wireObject(row)
		if e != nil {
			return nil, e
		}
		b, e := wireBlock(m["block"])
		if e != nil {
			return nil, e
		}
		if b.Number < r.FromBlock || b.Number > r.ToBlock {
			return nil, errors.New("Credit transition outside update range")
		}
		if _, e = c.canonical(ctx, b); e != nil {
			return nil, e
		}
		var name string
		if json.Unmarshal(m["name"], &name) != nil {
			return nil, errors.New("invalid Credit transition name")
		}
		switch name {
		case "CreditCreated", "CreditDraw", "CreditUpdated", "CreditRevoked":
		default:
			return nil, errors.New("invalid Credit transition name")
		}
		h, e := wireHash(m["transactionHash"])
		if e != nil {
			return nil, e
		}
		index, e := wireUint(m["logIndex"], 32)
		if e != nil {
			return nil, e
		}
		data, e := wireObject(m["data"])
		if e != nil {
			return nil, e
		}
		normalized, e := json.Marshal(data)
		if e != nil {
			return nil, e
		}
		event := ProtocolCreditTransition{b, name, h, uint32(index), normalized}
		encoded, _ := json.Marshal(event)
		key := fmt.Sprintf("%d:%d", b.Number, index)
		previous, ok := seen[key]
		if ok && previous != string(encoded) {
			return nil, errors.New("conflicting Credit transition")
		}
		if !ok {
			out = append(out, event)
		}
		seen[key] = string(encoded)
	}
	return out, nil
}
func (c *ProtocolRpcClient) GetCreditTransitions(ctx context.Context, scope DeploymentContext, r ProtocolUpdateRange) (ProtocolCreditUpdates, error) {
	var out ProtocolCreditUpdates
	b, raw, e := c.updatePage(ctx, scope, r, "invisi_getCreditTransitions")
	if e != nil {
		return out, e
	}
	var rows []json.RawMessage
	if e = json.Unmarshal(raw, &rows); e != nil {
		return out, e
	}
	events, e := c.parseTransitions(ctx, rows, r)
	if e != nil {
		return out, e
	}
	if _, e = c.canonical(ctx, b); e != nil {
		return out, e
	}
	return ProtocolCreditUpdates{b, events}, nil
}

type IndexedCreditAllocation struct {
	Block CanonicalBlock
	Shard uint64
	Slot  uint16
}
type IndexedCreditRevocationBitmap struct {
	Block  CanonicalBlock
	Bitmap *big.Int
}

func (c *ProtocolRpcClient) queryBlock(ctx context.Context, scope DeploymentContext, block CanonicalBlock) error {
	if block.Number > 1<<53-1 {
		return errors.New("block exceeds safe wire range")
	}
	if _, e := c.readStatus(ctx, scope, false); e != nil {
		return e
	}
	_, e := c.canonical(ctx, block)
	return e
}
func (c *ProtocolRpcClient) GetCreditAllocation(ctx context.Context, scope DeploymentContext, owner, token common.Address, block CanonicalBlock) (IndexedCreditAllocation, error) {
	var out IndexedCreditAllocation
	if owner == (common.Address{}) || token == (common.Address{}) {
		return out, errors.New("nonzero owner and token required")
	}
	if e := c.queryBlock(ctx, scope, block); e != nil {
		return out, e
	}
	var raw json.RawMessage
	if e := c.call(ctx, scope.ChainID, "invisi_getCreditAllocation", []any{owner, token, block.Number}, &raw); e != nil {
		return out, e
	}
	m, e := wireObject(raw)
	if e != nil {
		return out, e
	}
	b, e := wireBlock(m["block"])
	if e != nil || b != block {
		return out, errors.New("RPC grant block mismatch")
	}
	shard, e := wireUint(m["shard"], 64)
	if e != nil || shard == 0 {
		return out, errors.New("invalid grant shard")
	}
	slot, e := wireUint(m["slot"], 16)
	if e != nil {
		return out, e
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return out, e
	}
	return IndexedCreditAllocation{b, shard, uint16(slot)}, nil
}
func (c *ProtocolRpcClient) GetCreditRevocationBitmap(ctx context.Context, scope DeploymentContext, owner, token common.Address, shard uint64, word uint8, block CanonicalBlock) (IndexedCreditRevocationBitmap, error) {
	var out IndexedCreditRevocationBitmap
	if owner == (common.Address{}) || token == (common.Address{}) || shard == 0 {
		return out, errors.New("invalid grant location")
	}
	if e := c.queryBlock(ctx, scope, block); e != nil {
		return out, e
	}
	var raw json.RawMessage
	if e := c.call(ctx, scope.ChainID, "invisi_getCreditRevocationBitmap", []any{owner, token, strconv.FormatUint(shard, 10), word, block.Number}, &raw); e != nil {
		return out, e
	}
	m, e := wireObject(raw)
	if e != nil {
		return out, e
	}
	b, e := wireBlock(m["block"])
	if e != nil || b != block {
		return out, errors.New("RPC grant block mismatch")
	}
	var decimal string
	if json.Unmarshal(m["bitmap"], &decimal) != nil {
		return out, errors.New("invalid bitmap")
	}
	bits, ok := new(big.Int).SetString(decimal, 10)
	if !ok || bits.Sign() < 0 || bits.BitLen() > 256 || bits.String() != decimal {
		return out, errors.New("invalid bitmap")
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return out, e
	}
	return IndexedCreditRevocationBitmap{b, bits}, nil
}
func (c *ProtocolRpcClient) consumed(ctx context.Context, scope DeploymentContext, nf common.Hash, block CanonicalBlock, credit bool) (bool, error) {
	if nf == (common.Hash{}) {
		return false, errors.New("nonzero nullifier required")
	}
	if _, e := primitives.HexToElement(nf.Hex()); e != nil {
		return false, e
	}
	if e := c.queryBlock(ctx, scope, block); e != nil {
		return false, e
	}
	method := "invisi_isNoteSpent"
	if credit {
		method = "invisi_isCreditConsumed"
	}
	var raw json.RawMessage
	if e := c.call(ctx, scope.ChainID, method, []any{nf, block.Number}, &raw); e != nil {
		return false, e
	}
	if string(raw) != "true" && string(raw) != "false" {
		return false, errors.New("invalid consumption observation")
	}
	if _, e := c.canonical(ctx, block); e != nil {
		return false, e
	}
	return string(raw) == "true", nil
}
func (c *ProtocolRpcClient) IsNoteSpent(ctx context.Context, scope DeploymentContext, nf common.Hash, block CanonicalBlock) (bool, error) {
	return c.consumed(ctx, scope, nf, block, false)
}
func (c *ProtocolRpcClient) IsCreditConsumed(ctx context.Context, scope DeploymentContext, nf common.Hash, block CanonicalBlock) (bool, error) {
	return c.consumed(ctx, scope, nf, block, true)
}
