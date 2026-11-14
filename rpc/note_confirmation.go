package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"reflect"
)

// TrustedNoteEvidence is evidence authenticated by the trusted RPC server.
// It does not provide the independent chain authentication of NoteLookup.
type TrustedNoteEvidence struct {
	Deployment   ProtocolDeployment
	IndexedHead  *CanonicalBlock
	Status       string
	Reason       *string
	Output       *ProtocolNoteOutput
	Membership   *TrustedNoteMembership
	Confirmation *TrustedNoteConfirmation
}
type TrustedNoteMembership struct {
	TreeID           uint64
	LeafIndex        uint32
	Commitment, Root common.Hash
	LeafCount        uint32
	Siblings         []common.Hash
	Block            CanonicalBlock
}
type TrustedNoteConfirmation struct {
	Finality                 string
	VerifiedAt, FinalityHead CanonicalBlock
}

// GetNoteFromTrustedRPC validates pinned deployment, public bindings and proof locally.
// Receipt, canonical chain and finality authentication remain trusted server claims.
func (c *ProtocolRpcClient) GetNoteFromTrustedRPC(ctx context.Context, scope DeploymentContext, commitment common.Hash, finality string) (*TrustedNoteEvidence, error) {
	d, e := c.context(scope)
	if e != nil {
		return nil, e
	}
	if !c.pinned {
		return nil, errors.New("explicit deployment pin required")
	}
	if finality != "safe" && finality != "finalized" {
		return nil, errors.New("safe or finalized required")
	}
	if commitment == (common.Hash{}) {
		return nil, errors.New("nonzero commitment required")
	}
	if _, e = primitives.HexToElement(commitment.Hex()); e != nil {
		return nil, e
	}
	var raw json.RawMessage
	if e = c.call(ctx, scope.ChainID, "invisi_getNote", []any{commitment, map[string]string{"finality": finality}}, &raw); e != nil {
		return nil, e
	}
	return parseTrustedNote(raw, d, commitment, finality)
}
func trustedNull(r json.RawMessage) bool { return bytes.Equal(bytes.TrimSpace(r), []byte("null")) }
func parseTrustedNote(raw json.RawMessage, d ProtocolDeployment, commitment common.Hash, finality string) (*TrustedNoteEvidence, error) {
	fail := func() (*TrustedNoteEvidence, error) { return nil, errors.New("invalid trusted note evidence") }
	m, e := wireObject(raw)
	if e != nil {
		return nil, e
	}
	for _, k := range []string{"apiVersion", "protocolVersion", "deployment", "indexedHead", "status", "reason", "output", "membership", "confirmation"} {
		if len(m[k]) == 0 {
			return fail()
		}
	}
	if string(m["apiVersion"]) != `"v1"` || string(m["protocolVersion"]) != `"1"` {
		return fail()
	}
	var remote ProtocolDeployment
	if json.Unmarshal(m["deployment"], &remote) != nil || remote.Validate() != nil || !reflect.DeepEqual(remote, d) {
		return fail()
	}
	result := &TrustedNoteEvidence{Deployment: cloneDeployment(d)}
	blocks := map[uint64]CanonicalBlock{}
	block := func(raw json.RawMessage) (CanonicalBlock, error) {
		b, e := wireBlock(raw)
		if e != nil {
			return b, e
		}
		if b.Number > 1<<53-1 || b.Timestamp > 1<<53-1 || b.Number < d.Protocol.FromBlock || b.Hash == (common.Hash{}) {
			return b, errors.New("invalid trusted block bounds")
		}
		if old, ok := blocks[b.Number]; ok && old != b {
			return b, errors.New("conflicting trusted blocks")
		}
		for n, old := range blocks {
			if n != b.Number && old.Hash == b.Hash {
				return b, errors.New("repeated trusted block hash")
			}
			if n < b.Number && old.Timestamp > b.Timestamp || n > b.Number && old.Timestamp < b.Timestamp {
				return b, errors.New("invalid block time order")
			}
			if n+1 == b.Number && b.ParentHash != old.Hash || b.Number+1 == n && old.ParentHash != b.Hash {
				return b, errors.New("invalid block ancestry")
			}
		}
		blocks[b.Number] = b
		return b, nil
	}
	if !trustedNull(m["indexedHead"]) {
		b, e := block(m["indexedHead"])
		if e != nil {
			return nil, e
		}
		result.IndexedHead = &b
	}
	if json.Unmarshal(m["status"], &result.Status) != nil {
		return fail()
	}
	if result.Status == "pending" {
		var reason string
		if json.Unmarshal(m["reason"], &reason) != nil {
			return fail()
		}
		switch reason {
		case "not_observed", "pending_index", "confirmation_pending", "finality_unsupported", "finality_unavailable", "finality_inconsistent":
		default:
			return fail()
		}
		if !trustedNull(m["output"]) || !trustedNull(m["membership"]) || !trustedNull(m["confirmation"]) {
			return fail()
		}
		result.Reason = &reason
		return result, nil
	}
	if result.Status != "confirmed" || !trustedNull(m["reason"]) || result.IndexedHead == nil {
		return fail()
	}
	out, e := parseNoteOutput(m["output"])
	if e != nil {
		return nil, e
	}
	if out.Commitment != commitment || out.TreeID > 1<<53-1 || out.TransactionHash == (common.Hash{}) {
		return fail()
	}
	om, _ := wireObject(m["output"])
	if _, e = block(om["block"]); e != nil {
		return nil, e
	}
	proof, e := wireObject(m["membership"])
	if e != nil {
		return nil, e
	}
	anchor, e := wireObject(proof["anchor"])
	if e != nil {
		return nil, e
	}
	tree, e := wireUint(anchor["treeId"], 64)
	if e != nil || tree != out.TreeID {
		return fail()
	}
	index, e := wireUint(proof["leafIndex"], 32)
	if e != nil || index != uint64(out.LeafIndex) {
		return fail()
	}
	leaf, e := wireHash(proof["commitment"])
	if e != nil || leaf != commitment {
		return fail()
	}
	count, e := wireUint(anchor["leafCount"], 32)
	if e != nil || count <= index || count > 1<<primitives.TreeDepth {
		return fail()
	}
	root, e := wireHash(anchor["root"])
	if e != nil {
		return nil, e
	}
	if _, e = primitives.HexToElement(root.Hex()); e != nil {
		return nil, e
	}
	rootBlock, e := block(anchor["block"])
	if e != nil {
		return nil, e
	}
	var siblings []json.RawMessage
	if json.Unmarshal(proof["siblings"], &siblings) != nil || len(siblings) != primitives.TreeDepth {
		return fail()
	}
	path := make([]common.Hash, len(siblings))
	current, _ := primitives.HexToElement(commitment.Hex())
	for level, raw := range siblings {
		h, e := wireHash(raw)
		if e != nil {
			return nil, e
		}
		s, e := primitives.HexToElement(h.Hex())
		if e != nil {
			return nil, e
		}
		path[level] = h
		if index>>level&1 == 0 {
			current, e = primitives.Poseidon2Hash(current, s)
		} else {
			current, e = primitives.Poseidon2Hash(s, current)
		}
		if e != nil {
			return nil, e
		}
	}
	if common.Hash(current.Bytes()) != root {
		return fail()
	}
	conf, e := wireObject(m["confirmation"])
	if e != nil {
		return nil, e
	}
	var tag string
	if json.Unmarshal(conf["finality"], &tag) != nil || tag != finality {
		return fail()
	}
	verified, e := block(conf["verifiedAt"])
	if e != nil {
		return nil, e
	}
	head, e := block(conf["finalityHead"])
	if e != nil {
		return nil, e
	}
	expected := result.IndexedHead.Number
	if head.Number < expected {
		expected = head.Number
	}
	if verified.Number != expected || rootBlock.Number > verified.Number || out.Block.Number > rootBlock.Number {
		return fail()
	}
	result.Output = &out
	result.Membership = &TrustedNoteMembership{tree, uint32(index), leaf, root, uint32(count), path, rootBlock}
	result.Confirmation = &TrustedNoteConfirmation{tag, verified, head}
	return result, nil
}
