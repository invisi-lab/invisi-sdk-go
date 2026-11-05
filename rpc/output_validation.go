package rpc

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

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
