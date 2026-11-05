package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

// ProtocolOutput is an indexed opaque delivery. Authenticate decryption and membership before spending.
type ProtocolOutput struct {
	Block            CanonicalBlock
	TreeID           uint64
	LeafIndex        uint32
	Commitment       common.Hash
	TokenAddress     common.Address
	Ciphertext, Clue []byte
	TransactionHash  common.Hash
	LogIndex         uint32
	Kind             string
}
type ProtocolOutputs struct {
	Block   CanonicalBlock
	Outputs []ProtocolOutput
}
type RegisteredReceivingAddress struct {
	ReceivingAddress string
	Nonce            uint64
	Block            CanonicalBlock
}

func wireObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	e := json.Unmarshal(raw, &m)
	if e != nil || m == nil {
		return nil, errors.New("expected RPC object")
	}
	return m, nil
}
func wireUint(raw json.RawMessage, bits int) (uint64, error) {
	s := string(raw)
	quoted := strings.HasPrefix(s, "\"")
	if quoted {
		if json.Unmarshal(raw, &s) != nil {
			return 0, errors.New("invalid RPC integer")
		}
	}
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return 0, errors.New("invalid RPC integer")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid RPC integer")
		}
	}
	n, e := strconv.ParseUint(s, 10, bits)
	if e == nil && !quoted && n > 1<<53-1 {
		e = errors.New("unsafe RPC integer")
	}
	return n, e
}
func wireBytes(raw json.RawMessage, min, max int) ([]byte, error) {
	var s string
	if json.Unmarshal(raw, &s) != nil || !strings.HasPrefix(s, "0x") {
		return nil, errors.New("invalid RPC hex")
	}
	b, e := hex.DecodeString(s[2:])
	if e != nil || len(b) < min || len(b) > max {
		return nil, errors.New("invalid RPC bytes")
	}
	return b, nil
}
func wireHash(raw json.RawMessage) (common.Hash, error) {
	b, e := wireBytes(raw, 32, 32)
	return common.BytesToHash(b), e
}
func wireBlock(raw json.RawMessage) (CanonicalBlock, error) {
	m, e := wireObject(raw)
	if e != nil {
		return CanonicalBlock{}, e
	}
	var b CanonicalBlock
	if b.Number, e = wireUint(m["number"], 64); e != nil {
		return b, e
	}
	if b.Timestamp, e = wireUint(m["timestamp"], 64); e != nil {
		return b, e
	}
	if b.Hash, e = wireHash(m["hash"]); e != nil {
		return b, e
	}
	b.ParentHash, e = wireHash(m["parentHash"])
	return b, e
}
func parseOutput(raw json.RawMessage) (ProtocolOutput, error) {
	var o ProtocolOutput
	m, e := wireObject(raw)
	if e != nil {
		return o, e
	}
	if o.Block, e = wireBlock(m["block"]); e != nil {
		return o, e
	}
	if o.TreeID, e = wireUint(m["treeId"], 64); e != nil || o.TreeID == 0 {
		return o, errors.New("invalid output tree")
	}
	n, e := wireUint(m["leafIndex"], 16)
	if e != nil {
		return o, e
	}
	o.LeafIndex = uint32(n)
	if o.Commitment, e = wireHash(m["commitment"]); e != nil {
		return o, e
	}
	if _, e = core.HexToElement(o.Commitment.Hex()); e != nil {
		return o, e
	}
	b, e := wireBytes(m["token"], 20, 20)
	if e != nil {
		return o, e
	}
	o.TokenAddress = common.BytesToAddress(b)
	if o.TokenAddress == (common.Address{}) {
		return o, errors.New("invalid output token")
	}
	if o.Ciphertext, e = wireBytes(m["ciphertext"], 0, 4096); e != nil {
		return o, e
	}
	clueSize := 68
	if len(o.Ciphertext) == 0 {
		clueSize = 0
	}
	if o.Clue, e = wireBytes(m["clue"], clueSize, clueSize); e != nil {
		return o, e
	}
	if o.TransactionHash, e = wireHash(m["transactionHash"]); e != nil {
		return o, e
	}
	n, e = wireUint(m["logIndex"], 32)
	if e != nil {
		return o, e
	}
	o.LogIndex = uint32(n)
	if json.Unmarshal(m["kind"], &o.Kind) != nil || o.Kind == "" {
		return o, errors.New("invalid output kind")
	}
	return o, nil
}
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
	address, e := core.EncodeLongAddress(core.ReceivingAddress{OwnerAddress: core.BigIntToElement(owner.Big()), ViewingPublicKey: view, ClueKey: clue})
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
