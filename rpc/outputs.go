package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

func (c *ProtocolRpcClient) updatePage(ctx context.Context, d DeploymentContext, r ProtocolUpdateRange, method string) (CanonicalBlock, json.RawMessage, error) {
	deployment, e := c.context(d)
	if e != nil {
		return CanonicalBlock{}, nil, e
	}
	if r.FromBlock < deployment.Protocol.FromBlock || r.ToBlock < r.FromBlock || r.ToBlock-r.FromBlock >= MaxUpdateBlocks || r.ToBlock > 1<<53-1 {
		return CanonicalBlock{}, nil, errors.New("updates require at most 10000 deployment blocks")
	}
	status, e := c.readStatus(ctx, d, false)
	if e != nil {
		return CanonicalBlock{}, nil, e
	}
	if status.IndexedHead == nil || status.IndexedHead.Number < r.ToBlock {
		return CanonicalBlock{}, nil, errors.New("Protocol RPC is behind requested range")
	}
	block, e := c.block(ctx, r.ToBlock)
	if e != nil {
		return CanonicalBlock{}, nil, e
	}
	if status.IndexedHead.Number == block.Number && *status.IndexedHead != block {
		return CanonicalBlock{}, nil, errors.New("RPC canonical block mismatch")
	}
	params := []any{map[string]any{"fromBlock": r.FromBlock, "toBlock": r.ToBlock, "blockHash": block.Hash}}
	var raw json.RawMessage
	if e = c.call(ctx, d.ChainID, method, params, &raw); e != nil {
		return CanonicalBlock{}, nil, e
	}
	if method != "invisi_getCreditTransitions" {
		m, e := wireObject(raw)
		if e != nil {
			return CanonicalBlock{}, nil, e
		}
		hash, e := wireHash(m["blockHash"])
		if e != nil || hash != block.Hash {
			return CanonicalBlock{}, nil, errors.New("RPC update anchor mismatch")
		}
	}
	return block, raw, nil
}
func outputRows(raw json.RawMessage, field string) ([]json.RawMessage, error) {
	m, err := wireObject(raw)
	if err != nil {
		return nil, err
	}
	var rows []json.RawMessage
	if err = json.Unmarshal(m[field], &rows); err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []json.RawMessage{}
	}
	return rows, nil
}

type indexedLeaf interface{ leafPosition() leafMetadata }

func parseRows[T indexedLeaf](ctx context.Context, c *ProtocolRpcClient, block CanonicalBlock, rows []json.RawMessage, r ProtocolUpdateRange, parse func(json.RawMessage) (T, error)) ([]T, error) {
	seen := map[string]bool{}
	positions := map[string]string{}
	commitments := map[common.Hash]string{}
	out := []T{}
	verified := map[uint64]CanonicalBlock{block.Number: block}
	for _, row := range rows {
		value, e := parse(row)
		if e != nil {
			return nil, e
		}
		m := value.leafPosition()
		known, ok := verified[m.Block.Number]
		if !ok {
			known, e = c.canonical(ctx, m.Block)
			if e != nil {
				return nil, e
			}
			verified[m.Block.Number] = known
		}
		if known != m.Block {
			return nil, errors.New("RPC output block mismatch")
		}
		if m.Block.Number < r.FromBlock || m.Block.Number > r.ToBlock {
			return nil, errors.New("RPC output outside update range")
		}
		encoded, _ := json.Marshal(value)
		fingerprint := string(encoded)
		position := fmt.Sprintf("%d:%d", m.TreeID, m.LeafIndex)
		if previous, ok := positions[position]; ok && previous != fingerprint {
			return nil, errors.New("conflicting RPC output position")
		}
		if previous, ok := commitments[m.Commitment]; ok && previous != fingerprint {
			return nil, errors.New("conflicting RPC commitment")
		}
		if !seen[fingerprint] {
			out = append(out, value)
		}
		seen[fingerprint] = true
		positions[position] = fingerprint
		commitments[m.Commitment] = fingerprint
	}
	return out, nil
}
func (c *ProtocolRpcClient) GetNoteUpdates(ctx context.Context, d DeploymentContext, r ProtocolUpdateRange) (ProtocolNoteUpdates, error) {
	var out ProtocolNoteUpdates
	block, raw, e := c.updatePage(ctx, d, r, "invisi_getNoteUpdates")
	if e != nil {
		return out, e
	}
	rows, e := outputRows(raw, "notes")
	if e != nil {
		return out, e
	}
	notes, e := parseRows(ctx, c, block, rows, r, parseNoteOutput)
	if e != nil {
		return out, e
	}
	if _, e = c.canonical(ctx, block); e != nil {
		return out, e
	}
	return ProtocolNoteUpdates{block, notes}, nil
}
