// Package protocol decodes events using the current built Protocol ABI.
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
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/protocol/contracts"
)

type Decoder struct {
	abi     abi.ABI
	context *protocol.TransactionContext
}

func NewDecoder(context ...protocol.TransactionContext) (*Decoder, error) {
	a, e := contracts.InvisiCoreMetaData.GetAbi()
	if e != nil {
		return nil, e
	}
	d := &Decoder{abi: *a}
	if len(context) > 0 {
		c := context[0]
		d.context = &c
	}
	return d, nil
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
		core common.Address
		tx   common.Hash
		tree uint64
		leaf uint32
	}
	inserted := map[location]common.Hash{}
	insertionLengths := map[location]int{}
	seenLogs := map[string]bool{}
	for _, l := range ordered {
		key := fmt.Sprintf("%s:%s:%d", l.Address, l.TxHash, l.Index)
		if seenLogs[key] {
			return Batch{}, fmt.Errorf("duplicate log")
		}
		seenLogs[key] = true
		if block.Hash != (common.Hash{}) && (l.BlockHash != block.Hash || l.BlockNumber != block.Number) {
			return Batch{}, fmt.Errorf("log canonical block mismatch")
		}
		if d.context != nil && l.Address != d.context.ProtocolAddress {
			return Batch{}, fmt.Errorf("log Core mismatch")
		}
		if l.Removed {
			return Batch{}, fmt.Errorf("removed log %d", l.Index)
		}
		if len(l.Topics) == 0 {
			continue
		}
		e, eerr := d.abi.EventByID(l.Topics[0])
		if eerr != nil {
			raw, _ := json.Marshal(eventJSON(map[string]any{"topics": l.Topics, "data": l.Data}))
			batch.Events = append(batch.Events, Event{Block: block, Name: "Unknown", TransactionHash: l.TxHash, LogIndex: l.Index, Data: raw})
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
		batch.Events = append(batch.Events, Event{Block: block, Name: e.Name, TransactionHash: l.TxHash, LogIndex: l.Index, Data: raw})
		events = append(events, decoded{l, e.Name, m})
		if e.Name == "TreeCreated" {
			id, x := tree(m["treeId"])
			if x != nil {
				return Batch{}, x
			}
			batch.CreatedTrees = append(batch.CreatedTrees, id)
		}
		if e.Name == "CommitmentsInserted" {
			id, x := tree(m["treeId"])
			if x != nil || id == 0 {
				return Batch{}, fmt.Errorf("invalid output tree")
			}
			start := m["startLeafIndex"].(uint32)
			commitmentField := "commitments"

			cs := hashes(m[commitmentField])
			if len(cs) == 0 || uint64(start)+uint64(len(cs)) > 65536 {
				return Batch{}, fmt.Errorf("%s: invalid insertion range", e.Name)
			}
			batch.Insertions = append(batch.Insertions, Insertion{TreeID: id, Start: start, Commitments: cs, LogIndex: l.Index})
			insertionLengths[location{l.Address, l.TxHash, id, start}] = len(cs)
			for i, c := range cs {
				key := location{l.Address, l.TxHash, id, start + uint32(i)}
				if _, ok := inserted[key]; ok {
					return Batch{}, fmt.Errorf("duplicate inserted leaf")
				}
				inserted[key] = c
			}
		}
	}
	type operationKey struct {
		core   common.Address
		tx, id common.Hash
	}
	operations := map[operationKey]map[string]any{}
	claimed := map[location]bool{}
	outputSeen := map[operationKey]bool{}
	operationOutputs := map[operationKey][]Output{}
	facts := map[operationKey]map[string]map[string]any{}
	positions := map[operationKey]uint{}
	for _, e := range events {
		if e.name == "NoteOperation" {
			positions[operationKey{e.log.Address, e.log.TxHash, common.Hash(e.values["operationId"].([32]byte))}] = e.log.Index
		}
	}

	for _, e := range events {
		if e.name == "NoteOperation" {
			key := operationKey{e.log.Address, e.log.TxHash, common.Hash(e.values["operationId"].([32]byte))}
			if _, ok := operations[key]; ok {
				return Batch{}, fmt.Errorf("duplicate NoteOperation")
			}
			if e.values["kind"].(uint8) > 1 || e.values["outputCount"].(*big.Int).Cmp(big.NewInt(190)) > 0 {
				return Batch{}, fmt.Errorf("invalid Note operation kind")
			}
			operations[key] = e.values
		}
	}
	for _, e := range events {
		m, l := e.values, e.log
		if e.name == "NoteOutputs" || e.name == "NoteStateTransition" || e.name == "PublicDeposit" || e.name == "PublicWithdrawal" {
			key := operationKey{l.Address, l.TxHash, common.Hash(m["operationId"].([32]byte))}
			if operations[key] == nil {
				return Batch{}, fmt.Errorf("orphan Note operation fact")
			}
			if facts[key] == nil {
				facts[key] = map[string]map[string]any{}
			}
			if facts[key][e.name] != nil {
				return Batch{}, fmt.Errorf("duplicate Note operation fact")
			}
			facts[key][e.name] = m
		}
		addNull := func(v any) { batch.Nullifiers = append(batch.Nullifiers, hashes(v)...) }
		add := func(id uint64, leaf uint32, c common.Hash, token common.Address, v delivery) error {
			if actual, ok := inserted[location{l.Address, l.TxHash, id, leaf}]; !ok || actual != c {
				return fmt.Errorf("%s: output does not match transaction insertion", e.name)
			}
			batch.Outputs = append(batch.Outputs, Output{Block: block, TreeID: id, LeafIndex: leaf, Commitment: c, Token: token, Ciphertext: append([]byte(nil), v.Ciphertext...), Clue: append([]byte(nil), v.Clue...), TransactionHash: l.TxHash, LogIndex: l.Index, Kind: e.name, OperationID: func() common.Hash {
				if v, ok := m["operationId"]; ok {
					return common.Hash(v.([32]byte))
				}
				return common.Hash{}
			}()})
			if e.name == "NoteOutputs" {
				key := operationKey{l.Address, l.TxHash, common.Hash(m["operationId"].([32]byte))}
				operationOutputs[key] = append(operationOutputs[key], batch.Outputs[len(batch.Outputs)-1])
			}
			return nil
		}
		switch e.name {
		case "NoteStateTransition":
			key := operationKey{l.Address, l.TxHash, common.Hash(m["operationId"].([32]byte))}
			if operations[key] == nil {
				return Batch{}, fmt.Errorf("orphan NoteStateTransition")
			}
			ns := hashes(m["nullifiers"])
			sources := tuple[[]protocol.InputSource](m["sources"])
			seen := map[common.Hash]bool{}
			if operations[key]["kind"].(uint8) != 1 || len(ns) == 0 || len(sources) == 0 || len(sources) > len(ns) {
				return Batch{}, fmt.Errorf("invalid Note state transition")
			}
			var last uint64
			for _, source := range sources {
				if !source.TreeId.IsUint64() || source.TreeId.Uint64() <= last || source.MerkleRoot == ([32]byte{}) {
					return Batch{}, fmt.Errorf("invalid Note source order")
				}
				last = source.TreeId.Uint64()
			}
			for _, nf := range ns {
				if nf == (common.Hash{}) || seen[nf] {
					return Batch{}, fmt.Errorf("invalid Note nullifier")
				}
				seen[nf] = true
			}
			addNull(m["nullifiers"])
		case "PublicDeposit", "PublicWithdrawal":
			key := operationKey{l.Address, l.TxHash, common.Hash(m["operationId"].([32]byte))}
			amount := m["amount"].(*big.Int)
			if m["token"] != operations[key]["token"] || amount.Sign() <= 0 || amount.BitLen() > 128 {
				return Batch{}, fmt.Errorf("invalid public flow")
			}
		case "NoteOutputs":
			key := operationKey{l.Address, l.TxHash, common.Hash(m["operationId"].([32]byte))}
			op := operations[key]
			if op == nil || outputSeen[key] {
				return Batch{}, fmt.Errorf("missing or duplicate Note operation output")
			}
			outputSeen[key] = true
			id, x := tree(m["treeId"])
			if x != nil || id == 0 {
				return Batch{}, fmt.Errorf("invalid output tree")
			}
			start := m["startLeafIndex"].(uint32)
			countBig := op["outputCount"].(*big.Int)
			if !countBig.IsUint64() || countBig.Uint64() == 0 || countBig.Uint64() > 65536 || uint64(start)+countBig.Uint64() > 65536 {
				return Batch{}, fmt.Errorf("invalid Note output count")
			}
			count := int(countBig.Uint64())
			ranges := 0
			for loc, n := range insertionLengths {
				if loc.core == l.Address && loc.tx == l.TxHash && loc.tree == id && loc.leaf <= start && uint64(start)+uint64(count) <= uint64(loc.leaf)+uint64(n) {
					ranges++
				}
			}
			if ranges != 1 {
				return Batch{}, fmt.Errorf("Note outputs require one insertion subrange")
			}
			packed := m["packedDeliveries"].([]byte)
			cursor := 0
			for i := 0; i < count; i++ {
				if len(packed)-cursor < 2 {
					return Batch{}, fmt.Errorf("truncated Note envelope length")
				}
				size := int(binary.BigEndian.Uint16(packed[cursor : cursor+2]))
				cursor += 2
				v := delivery{}
				if size > 0 {
					if size > 4096 || len(packed)-cursor < size+68 {
						return Batch{}, fmt.Errorf("invalid Note envelope length")
					}
					v = delivery{packed[cursor : cursor+size], packed[cursor+size : cursor+size+68]}
					if v.Clue[64] != 24 {
						return Batch{}, fmt.Errorf("invalid clue precision")
					}
					cursor += size + 68
				}
				loc := location{l.Address, l.TxHash, id, start + uint32(i)}
				if claimed[loc] {
					return Batch{}, fmt.Errorf("duplicate Note output position")
				}
				claimed[loc] = true
				c, ok := inserted[loc]
				if !ok {
					return Batch{}, fmt.Errorf("missing Note insertion")
				}
				if err = add(id, loc.leaf, c, op["token"].(common.Address), v); err != nil {
					return Batch{}, err
				}
			}
			if cursor != len(packed) {
				return Batch{}, fmt.Errorf("trailing Note delivery bytes")
			}
		case "CreditCreated", "CreditUpdated":
			p := m
			if e.name == "CreditUpdated" {
				p = fields(m["params"])
				batch.CreditNullifiers = append(batch.CreditNullifiers, p["nullifier"].([32]byte))
			}
			id, x := tree(m["outputTreeId"])
			if x != nil {
				return Batch{}, x
			}
			c := common.Hash(p["creditCommitment"].([32]byte))
			leaf := m["leafIndex"].(uint32)
			vs := tuple[[]delivery](m["deliveries"])
			if len(vs) != 2 {
				return Batch{}, fmt.Errorf("%s: expected owner and spender state deliveries", e.name)
			}
			for _, v := range vs {
				if err = add(id, leaf, c, p["token"].(common.Address), v); err != nil {
					return Batch{}, err
				}
				last := len(batch.Outputs) - 1
				batch.CreditStates = append(batch.CreditStates, StateFromOutput(batch.Outputs[last], []Delivery{{Ciphertext: v.Ciphertext, Clue: v.Clue}}))
				batch.Outputs = batch.Outputs[:last]
			}
		case "CreditStateTransition":
			p := fields(m["state"])
			ns := hashes(p["nullifiers"])
			cs := hashes(p["creditCommitments"])
			ls := m["creditLeafIndices"].([]uint32)
			vs := tuple[[]delivery](m["creditDeliveries"])
			n := len(cs)
			if n == 0 || len(ns) != n || len(ls) != n || len(vs) != 2*n || len(hashes(p["recoveryAmounts"])) != n {
				return Batch{}, fmt.Errorf("CreditStateTransition: inconsistent dimensions")
			}
			batch.CreditNullifiers = append(batch.CreditNullifiers, ns...)
			id, x := tree(m["outputTreeId"])
			if x != nil {
				return Batch{}, x
			}
			var token common.Address
			matches := 0
			for _, draw := range events {
				if draw.name == "CreditDraw" && draw.log.Address == l.Address && draw.log.TxHash == l.TxHash && draw.values["noteCommitment"] == m["noteCommitment"] && draw.values["outputTreeId"].(*big.Int).Cmp(m["outputTreeId"].(*big.Int)) == 0 {
					token = draw.values["token"].(common.Address)
					matches++
				}
			}
			if matches != 1 {
				return Batch{}, fmt.Errorf("Credit state transition requires exactly one draw")
			}
			for i, c := range cs {
				if c == (common.Hash{}) {
					return Batch{}, fmt.Errorf("zero successor state")
				}
				for _, v := range vs[2*i : 2*i+2] {
					if err = add(id, ls[i], c, token, v); err != nil {
						return Batch{}, err
					}
					last := len(batch.Outputs) - 1
					batch.CreditStates = append(batch.CreditStates, StateFromOutput(batch.Outputs[last], []Delivery{{Ciphertext: v.Ciphertext, Clue: v.Clue}}))
					batch.Outputs = batch.Outputs[:last]
				}
			}
		case "CreditDraw":
			id, x := tree(m["outputTreeId"])
			if x != nil {
				return Batch{}, x
			}
			v := tuple[delivery](m["delivery"])
			if err = add(id, m["noteLeafIndex"].(uint32), common.Hash(m["noteCommitment"].([32]byte)), m["token"].(common.Address), v); err != nil {
				return Batch{}, err
			}

		}
	}
	for key, op := range operations {
		count := op["outputCount"].(*big.Int)
		if count.Sign() > 0 && !outputSeen[key] {
			return Batch{}, fmt.Errorf("missing NoteOutputs")
		}
		p := protocol.CrossTreeTransactParams{Token: op["token"].(common.Address), Sources: []protocol.InputSource{}, Nullifiers: [][32]byte{}, NoteCommitments: [][32]byte{}, WithdrawalAmount: new(big.Int), Deadline: op["deadline"].(*big.Int), FeeRecipient: op["feeRecipient"].(common.Address), FeeAmount: op["feeAmount"].(*big.Int)}
		funding := protocol.FundingParams{Amount: new(big.Int)}
		deliveries := []protocol.Delivery{}
		if f := facts[key]["NoteStateTransition"]; f != nil {
			p.Sources = tuple[[]protocol.InputSource](f["sources"])
			p.Nullifiers = f["nullifiers"].([][32]byte)
		}
		if f := facts[key]["PublicDeposit"]; f != nil {
			funding.Payer = f["payer"].(common.Address)
			funding.Amount = f["amount"].(*big.Int)
		}
		if f := facts[key]["PublicWithdrawal"]; f != nil {
			p.WithdrawalAddress = f["recipient"].(common.Address)
			p.WithdrawalAmount = f["amount"].(*big.Int)
		}
		for _, o := range operationOutputs[key] {
			p.NoteCommitments = append(p.NoteCommitments, o.Commitment)
			deliveries = append(deliveries, protocol.Delivery{Ciphertext: o.Ciphertext, Clue: o.Clue})
		}
		kind := op["kind"].(uint8)
		if kind == 0 {
			if count.Sign() <= 0 || count.Cmp(big.NewInt(10)) > 0 || len(p.Nullifiers) > 0 || funding.Amount.Sign() <= 0 || p.WithdrawalAmount.Sign() != 0 {
				return Batch{}, fmt.Errorf("invalid Shield operation")
			}
		} else if len(p.Nullifiers) == 0 && funding.Amount.Sign() == 0 {
			return Batch{}, fmt.Errorf("unfunded Note operation")
		}
		if d.context != nil {
			actual, e := protocol.NoteOperationID(*d.context, kind, p, deliveries, funding)
			if e != nil || actual != key.id {
				return Batch{}, fmt.Errorf("Note operation hash mismatch: %v", e)
			}
		}
	}
	members := map[operationKey]bool{}
	for _, e := range events {
		if e.name == "TransactAggregated" {
			ids := hashes(e.values["operationIds"])
			if len(ids) == 0 || len(ids) > 5 {
				return Batch{}, fmt.Errorf("invalid aggregate membership")
			}
			for _, id := range ids {
				key := operationKey{e.log.Address, e.log.TxHash, id}
				op := operations[key]
				if op == nil || op["kind"].(uint8) != 1 || positions[key] >= e.log.Index || members[key] {
					return Batch{}, fmt.Errorf("invalid aggregate operation association")
				}
				members[key] = true
			}
		}
	}
	combined := []CreditState{}
	for _, state := range batch.CreditStates {
		if len(combined) > 0 {
			last := &combined[len(combined)-1]
			if last.TransactionHash == state.TransactionHash && last.TreeID == state.TreeID && last.LeafIndex == state.LeafIndex {
				last.Deliveries = append(last.Deliveries, state.Deliveries...)
				continue
			}
		}
		combined = append(combined, state)
	}
	for _, state := range combined {
		if len(state.Deliveries) != 2 {
			return Batch{}, fmt.Errorf("Credit state must have exactly two deliveries")
		}
	}
	for i := range batch.Outputs {
		batch.Outputs[i].Role = "funds"
	}
	batch.CreditStates = combined
	return batch, nil
}
