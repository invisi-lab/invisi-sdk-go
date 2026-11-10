package rpc

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

type leafMetadata struct {
	Block           CanonicalBlock
	TreeID          uint64
	LeafIndex       uint32
	Commitment      common.Hash
	TokenAddress    common.Address
	TransactionHash common.Hash
	LogIndex        uint32
	Kind            string
}

func parseLeaf(raw json.RawMessage) (leafMetadata, map[string]json.RawMessage, error) {
	var o leafMetadata
	m, e := wireObject(raw)
	if e != nil {
		return o, nil, e
	}
	if o.Block, e = wireBlock(m["block"]); e != nil {
		return o, nil, e
	}
	if o.TreeID, e = wireUint(m["treeId"], 64); e != nil || o.TreeID == 0 {
		return o, nil, errors.New("invalid output tree")
	}
	n, e := wireUint(m["leafIndex"], 16)
	if e != nil {
		return o, nil, e
	}
	o.LeafIndex = uint32(n)
	if o.Commitment, e = wireHash(m["commitment"]); e != nil {
		return o, nil, e
	}
	if _, e = primitives.HexToElement(o.Commitment.Hex()); e != nil {
		return o, nil, e
	}
	token, e := wireBytes(m["token"], 20, 20)
	if e != nil {
		return o, nil, e
	}
	o.TokenAddress = common.BytesToAddress(token)
	if o.TokenAddress == (common.Address{}) {
		return o, nil, errors.New("invalid output token")
	}
	if o.TransactionHash, e = wireHash(m["transactionHash"]); e != nil {
		return o, nil, e
	}
	n, e = wireUint(m["logIndex"], 32)
	if e != nil {
		return o, nil, e
	}
	o.LogIndex = uint32(n)
	if json.Unmarshal(m["kind"], &o.Kind) != nil || o.Kind == "" {
		return o, nil, errors.New("invalid output kind")
	}
	return o, m, nil
}
func parseDelivery(m map[string]json.RawMessage, omitted bool) (IndexedDelivery, error) {
	var d IndexedDelivery
	var e error
	if d.Ciphertext, e = wireBytes(m["ciphertext"], 0, 4096); e != nil {
		return d, e
	}
	if len(d.Ciphertext) == 0 && !omitted {
		return d, errors.New("Credit role delivery cannot be omitted")
	}
	size := 68
	if len(d.Ciphertext) == 0 {
		size = 0
	}
	d.Clue, e = wireBytes(m["clue"], size, size)
	return d, e
}
func parseNoteOutput(raw json.RawMessage) (ProtocolNoteOutput, error) {
	var o ProtocolNoteOutput
	m, fields, e := parseLeaf(raw)
	if e != nil {
		return o, e
	}
	switch m.Kind {
	case "Shield", "Transact", "CrossTreeTransact", "BatchOutputs", "CreditDraw":
	default:
		return o, errors.New("Credit state cannot appear in Note results")
	}
	d, e := parseDelivery(fields, true)
	if e != nil {
		return o, e
	}
	return ProtocolNoteOutput{m.Block, m.TreeID, m.LeafIndex, m.Commitment, m.TokenAddress, d.Ciphertext, d.Clue, m.TransactionHash, m.LogIndex, m.Kind}, nil
}
func parseCreditState(raw json.RawMessage) (IndexedCreditState, error) {
	var o IndexedCreditState
	m, fields, e := parseLeaf(raw)
	if e != nil {
		return o, e
	}
	switch m.Kind {
	case "CreditCreated", "CreditDraw", "CreditUpdated":
	default:
		return o, errors.New("invalid Credit state event")
	}
	var rows []json.RawMessage
	if json.Unmarshal(fields["deliveries"], &rows) != nil || len(rows) != 2 {
		return o, errors.New("Credit state requires both role deliveries")
	}
	ds := make([]IndexedDelivery, 0, 2)
	for _, row := range rows {
		fields, e := wireObject(row)
		if e != nil {
			return o, e
		}
		d, e := parseDelivery(fields, false)
		if e != nil {
			return o, e
		}
		ds = append(ds, d)
	}
	return IndexedCreditState{m.Block, m.TreeID, m.LeafIndex, m.Commitment, m.TokenAddress, ds, m.TransactionHash, m.LogIndex, m.Kind}, nil
}
func (o ProtocolNoteOutput) leafPosition() leafMetadata {
	return leafMetadata{o.Block, o.TreeID, o.LeafIndex, o.Commitment, o.TokenAddress, o.TransactionHash, o.LogIndex, o.Kind}
}
func (o IndexedCreditState) leafPosition() leafMetadata {
	return leafMetadata{o.Block, o.TreeID, o.LeafIndex, o.Commitment, o.TokenAddress, o.TransactionHash, o.LogIndex, o.Kind}
}
