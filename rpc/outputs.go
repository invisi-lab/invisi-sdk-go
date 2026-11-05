package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

func (c *ProtocolRpcClient) GetOutputs(ctx context.Context, d DeploymentContext, r ProtocolUpdateRange) (ProtocolOutputs, error) {
	return c.readOutputs(ctx, d, r, 0)
}

func (c *ProtocolRpcClient) GetCreditOutputs(ctx context.Context, d DeploymentContext, r ProtocolUpdateRange) (ProtocolOutputs, error) {
	return c.readOutputs(ctx, d, r, 1)
}

func (c *ProtocolRpcClient) GetEncryptedOutputs(ctx context.Context, d DeploymentContext, r ProtocolUpdateRange) (ProtocolOutputs, error) {
	return c.readOutputs(ctx, d, r, 2)
}

func (c *ProtocolRpcClient) readOutputs(ctx context.Context, d DeploymentContext, r ProtocolUpdateRange, mode int) (ProtocolOutputs, error) {
	var result ProtocolOutputs
	deployment, e := c.context(d)
	if e != nil {
		return result, e
	}
	if r.FromBlock < deployment.Protocol.FromBlock || r.ToBlock < r.FromBlock || r.ToBlock-r.FromBlock >= MaxUpdateBlocks || r.ToBlock > 1<<53-1 {
		return result, errors.New("updates require at most 10000 deployment blocks")
	}
	status, e := c.readStatus(ctx, d, false)
	if e != nil {
		return result, e
	}
	if status.IndexedHead == nil || status.IndexedHead.Number < r.ToBlock {
		return result, errors.New("Protocol RPC is behind requested range")
	}
	block, e := c.block(ctx, r.ToBlock)
	if e != nil {
		return result, e
	}
	// Reuse the range anchor to reject a stale indexer without another chain read.
	if status.IndexedHead.Number == block.Number && *status.IndexedHead != block {
		return result, errors.New("RPC canonical block mismatch")
	}
	wire := map[string]any{"fromBlock": r.FromBlock, "toBlock": r.ToBlock, "blockHash": block.Hash}
	method := "invisi_getUpdates"
	params := []any{wire}
	if len(r.DetectionKey) > 0 {
		if _, e = core.DeriveClueKey(r.DetectionKey); e != nil {
			return result, e
		}
		method = "invisi_detectOutputs"
		params = append(params, "0x"+hex.EncodeToString(r.DetectionKey))
	}
	var raw json.RawMessage
	if e = c.call(ctx, d.ChainID, method, params, &raw); e != nil {
		return result, e
	}
	m, e := wireObject(raw)
	if e != nil {
		return result, e
	}
	anchor, e := wireHash(m["blockHash"])
	if e != nil || anchor != block.Hash {
		return result, errors.New("RPC update anchor mismatch")
	}
	var rows []json.RawMessage
	if len(r.DetectionKey) > 0 {
		bits, e := wireUint(m["checkedBits"], 8)
		if e != nil || bits < 1 || bits > 24 {
			return result, errors.New("invalid detection precision")
		}
		var candidates []json.RawMessage
		if json.Unmarshal(m["candidates"], &candidates) != nil || candidates == nil {
			return result, errors.New("invalid candidates")
		}
		for _, raw := range candidates {
			candidate, e := wireObject(raw)
			if e != nil {
				return result, e
			}
			if string(candidate["fallback"]) != "true" && string(candidate["fallback"]) != "false" {
				return result, errors.New("invalid detection candidate")
			}
			rows = append(rows, candidate["output"])
		}
	} else {
		if json.Unmarshal(m["outputs"], &rows) != nil || rows == nil {
			return result, errors.New("invalid outputs")
		}
	}
	var transitions []json.RawMessage
	if json.Unmarshal(m["creditTransitions"], &transitions) != nil || transitions == nil {
		return result, errors.New("invalid credit transitions")
	}
	return c.parseOutputRows(ctx, block, rows, r, mode)
}

func (c *ProtocolRpcClient) parseOutputRows(ctx context.Context, block CanonicalBlock, rows []json.RawMessage, r ProtocolUpdateRange, mode int) (ProtocolOutputs, error) {
	seen := map[string]bool{}
	positions := map[string]string{}
	commitments := map[common.Hash]string{}
	envelopes := map[common.Hash]map[string]string{}
	result := ProtocolOutputs{Block: block, Outputs: []ProtocolOutput{}}
	verified := map[uint64]CanonicalBlock{block.Number: block}
	for _, raw := range rows {
		o, e := parseOutput(raw)
		if e != nil {
			return ProtocolOutputs{}, e
		}
		known, ok := verified[o.Block.Number]
		if !ok {
			known, e = c.canonical(ctx, o.Block)
			if e != nil {
				return ProtocolOutputs{}, e
			}
			verified[o.Block.Number] = known
		}
		if known != o.Block {
			return ProtocolOutputs{}, errors.New("RPC output block mismatch")
		}
		if o.Block.Number < r.FromBlock || o.Block.Number > r.ToBlock {
			return ProtocolOutputs{}, errors.New("RPC output outside update range")
		}
		creditEvent := o.Kind == "CreditCreated" || o.Kind == "CreditToppedUp" || o.Kind == "CreditDraw"
		credit := creditEvent && len(o.Ciphertext) > 0 && o.Ciphertext[0]&0x80 != 0
		if mode != 2 && credit != (mode == 1) {
			continue
		}
		paired := credit
		if mode == 2 {
			paired = creditEvent && len(o.Ciphertext) > 0
		}
		bytes, _ := json.Marshal(o)
		fingerprint := string(bytes)
		identity := fingerprint
		if paired {
			copy := o
			copy.Ciphertext = nil
			copy.Clue = nil
			bytes, _ = json.Marshal(copy)
			identity = string(bytes)
		}
		position := fmt.Sprintf("%d:%d", o.TreeID, o.LeafIndex)
		if previous, ok := positions[position]; ok && previous != identity {
			return ProtocolOutputs{}, errors.New("conflicting RPC output position")
		}
		if previous, ok := commitments[o.Commitment]; ok && previous != identity {
			return ProtocolOutputs{}, errors.New("conflicting RPC commitment")
		}
		if seen[fingerprint] {
			continue
		}
		if paired {
			roles := envelopes[o.Commitment]
			if roles == nil {
				roles = map[string]string{}
				envelopes[o.Commitment] = roles
			}
			key := string(o.Ciphertext)
			if previous, ok := roles[key]; ok && previous != fingerprint {
				return ProtocolOutputs{}, errors.New("conflicting Credit role delivery")
			}
			if len(roles) >= 2 {
				return ProtocolOutputs{}, errors.New("too many Credit role envelopes")
			}
			roles[key] = fingerprint
		}
		seen[fingerprint] = true
		positions[position] = identity
		commitments[o.Commitment] = identity
		result.Outputs = append(result.Outputs, o)
	}
	return result, nil
}
