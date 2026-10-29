package receipts

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

type Decoder struct{ abi abi.ABI }

func NewDecoder() (*Decoder, error) {
	a, e := contracts.InvisiProtocolMetaData.GetAbi()
	if e != nil {
		return nil, e
	}
	return &Decoder{*a}, nil
}

type delivery struct {
	Ciphertext []byte
	Clue       []byte
}

func tuple[T any](v any) T { return *abi.ConvertType(v, new(T)).(*T) }
func hashes(v any) []common.Hash {
	a := v.([][32]byte)
	r := make([]common.Hash, len(a))
	for i := range a {
		r[i] = a[i]
	}
	return r
}
func tree(v any) (uint64, error) {
	n, ok := v.(*big.Int)
	if !ok || !n.IsUint64() {
		return 0, fmt.Errorf("tree ID exceeds uint64")
	}
	return n.Uint64(), nil
}
func fields(v any) map[string]any {
	r := map[string]any{}
	x := reflect.ValueOf(v)
	t := x.Type()
	for i := 0; i < x.NumField(); i++ {
		n := t.Field(i).Name
		n = string(bytes.ToLower([]byte(n[:1]))) + n[1:]
		r[n] = x.Field(i).Interface()
	}
	return r
}

func (d *Decoder) Decode(block Block, logs []types.Log) (batch Batch, err error) {
	batch = Batch{Block: block, Logs: append([]types.Log(nil), logs...)}
	// ABI conversion failures are malformed logs, never process crashes.
	defer func() {
		if r := recover(); r != nil {
			batch = Batch{}
			err = fmt.Errorf("malformed protocol event: %v", r)
		}
	}()
	ordered := append([]types.Log(nil), logs...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Index < ordered[j].Index })
	type decoded struct {
		log    types.Log
		name   string
		values map[string]any
	}
	events := []decoded{}
	type location struct {
		tx   common.Hash
		tree uint64
		leaf uint32
	}
	inserted := map[location]common.Hash{}
	insertionLengths := map[location]int{}
	for _, l := range ordered {
		if l.Removed {
			return Batch{}, fmt.Errorf("removed log %d", l.Index)
		}
		if len(l.Topics) == 0 {
			continue
		}
		e, eerr := d.abi.EventByID(l.Topics[0])
		if eerr != nil {
			raw, _ := json.Marshal(eventJSON(map[string]any{"topics": l.Topics, "data": l.Data}))
			batch.Events = append(batch.Events, Event{Name: "Unknown", TransactionHash: l.TxHash, LogIndex: l.Index, Data: raw})
			continue
		}
		m := map[string]any{}
		if err = e.Inputs.NonIndexed().UnpackIntoMap(m, l.Data); err != nil {
			return Batch{}, fmt.Errorf("%s log %d: %w", e.Name, l.Index, err)
		}
		indexed := abi.Arguments{}
		for _, a := range e.Inputs {
			if a.Indexed {
				indexed = append(indexed, a)
			}
		}
		if len(l.Topics) != len(indexed)+1 {
			return Batch{}, fmt.Errorf("%s: incorrect topic count", e.Name)
		}
		if err = abi.ParseTopicsIntoMap(m, indexed, l.Topics[1:]); err != nil {
			return Batch{}, err
		}
		raw, eerr := json.Marshal(eventJSON(m))
		if eerr != nil {
			return Batch{}, eerr
		}
		batch.Events = append(batch.Events, Event{Name: e.Name, TransactionHash: l.TxHash, LogIndex: l.Index, Data: raw})
		events = append(events, decoded{l, e.Name, m})
		if e.Name == "TreeCreated" {
			id, x := tree(m["treeId"])
			if x != nil {
				return Batch{}, x
			}
			batch.CreatedTrees = append(batch.CreatedTrees, id)
		}
		if e.Name == "CommitmentsInserted" || e.Name == "BatchOutputs" {
			id, x := tree(m["treeId"])
			if x != nil {
				return Batch{}, x
			}
			start := m["startLeafIndex"].(uint32)
			cs := hashes(m["commitments"])
			if len(cs) == 0 || uint64(start)+uint64(len(cs)) > 65536 {
				return Batch{}, fmt.Errorf("%s: invalid insertion range", e.Name)
			}
			batch.Insertions = append(batch.Insertions, Insertion{TreeID: id, Start: start, Commitments: cs, LogIndex: l.Index})
			insertionLengths[location{l.TxHash, id, start}] = len(cs)
			for i, c := range cs {
				key := location{l.TxHash, id, start + uint32(i)}
				if _, ok := inserted[key]; ok {
					return Batch{}, fmt.Errorf("duplicate inserted leaf")
				}
				inserted[key] = c
			}
		}
	}
	for _, e := range events {
		m, l := e.values, e.log
		addNull := func(v any) { batch.Nullifiers = append(batch.Nullifiers, hashes(v)...) }
		add := func(id uint64, leaf uint32, c common.Hash, token common.Address, v delivery, credit bool) error {
			minimum, version := 160, byte(2)
			if credit {
				minimum, version = 402, 4
			}
			if len(v.Ciphertext) < minimum || len(v.Ciphertext) > minimum+256 || v.Ciphertext[0] != version || len(v.Clue) != 68 || v.Clue[64] != 24 {
				return fmt.Errorf("%s: invalid delivery envelope or clue", e.name)
			}
			if actual, ok := inserted[location{l.TxHash, id, leaf}]; !ok || actual != c {
				return fmt.Errorf("%s: output does not match transaction insertion", e.name)
			}
			batch.Outputs = append(batch.Outputs, Output{TreeID: id, LeafIndex: leaf, Commitment: c, Token: token, Ciphertext: append([]byte(nil), v.Ciphertext...), Clue: append([]byte(nil), v.Clue...), TransactionHash: l.TxHash, LogIndex: l.Index, Kind: e.name, Role: "funds"})
			return nil
		}
		switch e.name {
		case "BatchOutputs":
			id, x := tree(m["treeId"])
			if x != nil {
				return Batch{}, x
			}
			cs := hashes(m["commitments"])
			ts := m["tokens"].([]common.Address)
			p := m["packedDeliveries"].([]byte)
			if len(ts) != len(cs) {
				return Batch{}, fmt.Errorf("BatchOutputs: array lengths differ")
			}
			for i, c := range cs {
				if len(p) < 2 {
					return Batch{}, fmt.Errorf("BatchOutputs: missing ciphertext length")
				}
				length := int(binary.BigEndian.Uint16(p[:2]))
				p = p[2:]
				if length < 160 || length > 416 || len(p) < length+68 {
					return Batch{}, fmt.Errorf("BatchOutputs: invalid or truncated delivery length")
				}
				if err = add(id, m["startLeafIndex"].(uint32)+uint32(i), c, ts[i], delivery{p[:length], p[length : length+68]}, false); err != nil {
					return Batch{}, err
				}
				p = p[length+68:]
			}
			if len(p) != 0 {
				return Batch{}, fmt.Errorf("BatchOutputs: trailing delivery bytes")
			}

		case "Shield":
			id, x := tree(m["treeId"])
			if x != nil {
				return Batch{}, x
			}
			vs := tuple[[]delivery](m["deliveries"])
			start := m["startLeafIndex"].(uint32)
			if n, ok := insertionLengths[location{l.TxHash, id, start}]; !ok || n != len(vs) {
				return Batch{}, fmt.Errorf("Shield: deliveries do not match insertion length")
			}
			for i, v := range vs {
				leaf := start + uint32(i)
				c, ok := inserted[location{l.TxHash, id, leaf}]
				if !ok {
					return Batch{}, fmt.Errorf("Shield: missing transaction insertion")
				}
				if err = add(id, leaf, c, m["token"].(common.Address), v, false); err != nil {
					return Batch{}, err
				}
			}
		case "Transact", "CrossTreeTransact":
			addNull(m["nullifiers"])
			id, x := tree(m["outputTreeId"])
			if x != nil {
				return Batch{}, x
			}
			cs := hashes(m["outputCommitments"])
			ls := m["leafIndices"].([]uint32)
			vs := tuple[[]delivery](m["deliveries"])
			if len(cs) != len(ls) || len(cs) != len(vs) {
				return Batch{}, fmt.Errorf("%s: array lengths differ", e.name)
			}
			for i, c := range cs {
				if err = add(id, ls[i], c, m["token"].(common.Address), vs[i], false); err != nil {
					return Batch{}, err
				}
			}
		case "BatchRequest":
			addNull(m["nullifiers"])
		case "CreditCancelled":
			batch.Nullifiers = append(batch.Nullifiers, m["nullifier"].([32]byte))
		case "CreditCreated", "CreditToppedUp":
			p := m
			if e.name == "CreditToppedUp" {
				q := fields(m["params"])
				p = fields(q["creation"])
				batch.Nullifiers = append(batch.Nullifiers, q["creditNullifier"].([32]byte))
			}
			addNull(p["nullifiers"])
			id, x := tree(m["outputTreeId"])
			if x != nil {
				return Batch{}, x
			}
			ls := m["leafIndices"].([]uint32)
			vs := tuple[[]delivery](m["deliveries"])
			credit := common.Hash(p["creditCommitment"].([32]byte))
			change := common.Hash(p["changeCommitment"].([32]byte))
			want := 1
			if change != (common.Hash{}) {
				want = 2
			}
			if len(ls) != want || len(vs) != 3 {
				return Batch{}, fmt.Errorf("%s: array lengths differ", e.name)
			}
			for i := 0; i < 2; i++ {
				if err = add(id, ls[0], credit, p["token"].(common.Address), vs[i], true); err != nil {
					return Batch{}, err
				}
				batch.Outputs[len(batch.Outputs)-1].Role = []string{"owner", "spender"}[i]
			}
			if want == 2 {
				if err = add(id, ls[1], change, p["token"].(common.Address), vs[2], false); err != nil {
					return Batch{}, err
				}
			} else if len(vs[2].Ciphertext) != 0 || len(vs[2].Clue) != 0 {
				return Batch{}, fmt.Errorf("nonempty absent change delivery")
			}
		case "CreditDraw":
			p := fields(m["params"])
			addNull(p["nullifiers"])
			cs := hashes(p["remainderCommitments"])
			ls := m["remainderLeafIndices"].([]uint32)
			vs := tuple[[]delivery](m["deliveries"])
			n := len(cs)
			if len(ls) != n || len(vs) != 2*n || len(hashes(p["nullifiers"])) != n || reflect.ValueOf(p["expiries"]).Len() != n || len(hashes(p["recoveryAmounts"])) != n {
				return Batch{}, fmt.Errorf("CreditDraw: array lengths differ")
			}
			id, x := tree(m["outputTreeId"])
			if x != nil {
				return Batch{}, x
			}
			for i, c := range cs {
				for j := 0; j < 2; j++ {
					v := vs[2*i+j]
					if c == (common.Hash{}) {
						if ls[i] != 0 || len(v.Ciphertext) != 0 || len(v.Clue) != 0 {
							return Batch{}, fmt.Errorf("CreditDraw: nonempty absent remainder")
						}
						continue
					}
					if err = add(id, ls[i], c, p["token"].(common.Address), v, true); err != nil {
						return Batch{}, err
					}
					batch.Outputs[len(batch.Outputs)-1].Role = []string{"owner", "spender"}[j]
				}
			}
		}
	}
	return batch, nil
}
