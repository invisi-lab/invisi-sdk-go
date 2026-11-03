package rpc

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

// TransactionUpdateStatus describes index availability, not EVM execution success.
type TransactionUpdateStatus string

const (
	TransactionNotFound     TransactionUpdateStatus = "not_found"
	TransactionPendingIndex TransactionUpdateStatus = "pending_index"
	TransactionIndexed      TransactionUpdateStatus = "indexed"
)

// ProtocolCreditTransition contains public event data; authenticate its receipt before applying financial state.
type ProtocolCreditTransition struct {
	Block           CanonicalBlock  `json:"block"`
	Name            string          `json:"name"`
	TransactionHash common.Hash     `json:"transactionHash"`
	LogIndex        uint32          `json:"logIndex"`
	Data            json.RawMessage `json:"data"`
}

// ProtocolTransactionUpdates preserves nil collections until the transaction is indexed.
// IndexedHead is only an observed checkpoint in availability states and can be on a stale fork.
type ProtocolTransactionUpdates struct {
	TransactionHash   common.Hash
	Status            TransactionUpdateStatus
	IndexedHead       *CanonicalBlock
	Block             *CanonicalBlock
	ReceiptStatus     *uint64
	Outputs           []ProtocolOutput
	CreditTransitions []ProtocolCreditTransition
}

// GetTransactionUpdates queries RPC 1.1.4+ without scanning ranges or changing wallet state.
// Context controls cancellation. ReceiptStatus is 0 for reverted or 1 for successful execution.
func (c *ProtocolRpcClient) GetTransactionUpdates(ctx context.Context, scope DeploymentContext, transactionHash common.Hash) (ProtocolTransactionUpdates, error) {
	var result ProtocolTransactionUpdates
	if transactionHash == (common.Hash{}) {
		return result, errors.New("nonzero transaction hash required")
	}
	deployment, err := c.context(scope)
	if err != nil {
		return result, err
	}
	if _, err = c.readStatus(ctx, scope, false); err != nil {
		return result, err
	}
	var raw json.RawMessage
	if err = c.call(ctx, scope.ChainID, "invisi_getTransactionUpdates", []any{transactionHash.Hex()}, &raw); err != nil {
		return result, err
	}
	wire, err := wireObject(raw)
	if err != nil {
		return result, err
	}
	responseHash, err := wireHash(wire["transactionHash"])
	if err != nil || responseHash != transactionHash {
		return result, errors.New("RPC transaction hash mismatch")
	}
	if err = json.Unmarshal(wire["status"], &result.Status); err != nil {
		return ProtocolTransactionUpdates{}, err
	}
	switch result.Status {
	case TransactionNotFound, TransactionPendingIndex, TransactionIndexed:
	default:
		return ProtocolTransactionUpdates{}, errors.New("invalid transaction update status")
	}
	result.TransactionHash = transactionHash
	headRaw, exists := wire["indexedHead"]
	if !exists {
		return ProtocolTransactionUpdates{}, errors.New("missing indexed head")
	}
	if string(headRaw) != "null" {
		head, err := wireBlock(headRaw)
		if err != nil {
			return ProtocolTransactionUpdates{}, err
		}
		if result.Status == TransactionIndexed {
			head, err = c.canonical(ctx, head)
			if err != nil {
				return ProtocolTransactionUpdates{}, err
			}
		}
		result.IndexedHead = &head
	}
	if result.Status == TransactionNotFound {
		if len(wire["blockNumber"]) != 0 || len(wire["blockHash"]) != 0 || len(wire["receiptStatus"]) != 0 || string(wire["outputs"]) != "null" || string(wire["creditTransitions"]) != "null" {
			return ProtocolTransactionUpdates{}, errors.New("invalid not_found transaction updates")
		}
		return result, nil
	}
	number, err := wireUint(wire["blockNumber"], 64)
	if err != nil {
		return ProtocolTransactionUpdates{}, err
	}
	start := deployment.Protocol.FromBlock
	if deployment.Registry != nil && deployment.Registry.FromBlock < start {
		start = deployment.Registry.FromBlock
	}
	if number < start {
		return ProtocolTransactionUpdates{}, errors.New("transaction predates indexed deployment")
	}
	blockHash, err := wireHash(wire["blockHash"])
	if err != nil {
		return ProtocolTransactionUpdates{}, err
	}
	block, err := c.block(ctx, number)
	if err != nil {
		return ProtocolTransactionUpdates{}, err
	}
	if block.Hash != blockHash {
		return ProtocolTransactionUpdates{}, errors.New("RPC transaction block mismatch")
	}
	receiptStatus, err := wireUint(wire["receiptStatus"], 1)
	if err != nil {
		return ProtocolTransactionUpdates{}, err
	}
	result.Block = &block
	result.ReceiptStatus = &receiptStatus
	if result.Status == TransactionPendingIndex {
		if string(wire["outputs"]) != "null" || string(wire["creditTransitions"]) != "null" {
			return ProtocolTransactionUpdates{}, errors.New("unindexed transaction must not contain updates")
		}
		return result, nil
	}
	if result.IndexedHead == nil || result.IndexedHead.Number < number || (result.IndexedHead.Number == number && result.IndexedHead.Hash != block.Hash) {
		return ProtocolTransactionUpdates{}, errors.New("transaction exceeds indexed head")
	}
	var outputs, transitions []json.RawMessage
	if json.Unmarshal(wire["outputs"], &outputs) != nil || outputs == nil || json.Unmarshal(wire["creditTransitions"], &transitions) != nil || transitions == nil {
		return ProtocolTransactionUpdates{}, errors.New("indexed transaction requires arrays")
	}
	if receiptStatus == 0 && (len(outputs) > 0 || len(transitions) > 0) {
		return ProtocolTransactionUpdates{}, errors.New("reverted transaction cannot contain updates")
	}
	for _, row := range append(append([]json.RawMessage{}, outputs...), transitions...) {
		item, err := wireObject(row)
		if err != nil {
			return ProtocolTransactionUpdates{}, err
		}
		hash, err := wireHash(item["transactionHash"])
		if err != nil || hash != transactionHash {
			return ProtocolTransactionUpdates{}, errors.New("updates contain a different transaction")
		}
	}
	parsed, err := c.parseOutputRows(ctx, block, outputs, ProtocolUpdateRange{FromBlock: number, ToBlock: number}, 2)
	if err != nil {
		return ProtocolTransactionUpdates{}, err
	}
	result.Outputs = parsed.Outputs
	result.CreditTransitions = []ProtocolCreditTransition{}
	seen := map[uint32]string{}
	for _, row := range transitions {
		item, err := wireObject(row)
		if err != nil {
			return ProtocolTransactionUpdates{}, err
		}
		reference, err := wireBlock(item["block"])
		if err != nil {
			return ProtocolTransactionUpdates{}, err
		}
		if reference != block {
			return ProtocolTransactionUpdates{}, errors.New("Credit transition block mismatch")
		}
		var name string
		if json.Unmarshal(item["name"], &name) != nil {
			return ProtocolTransactionUpdates{}, errors.New("invalid Credit transition name")
		}
		switch name {
		case "CreditCreated", "CreditDraw", "CreditCancelled", "CreditToppedUp":
		default:
			return ProtocolTransactionUpdates{}, errors.New("invalid Credit transition name")
		}
		index, err := wireUint(item["logIndex"], 32)
		if err != nil {
			return ProtocolTransactionUpdates{}, err
		}
		data, err := wireObject(item["data"])
		if err != nil {
			return ProtocolTransactionUpdates{}, err
		}
		normalized, err := json.Marshal(data)
		if err != nil {
			return ProtocolTransactionUpdates{}, err
		}
		event := ProtocolCreditTransition{reference, name, transactionHash, uint32(index), normalized}
		encoded, _ := json.Marshal(event)
		previous, exists := seen[event.LogIndex]
		if exists && previous != string(encoded) {
			return ProtocolTransactionUpdates{}, errors.New("conflicting Credit transition")
		}
		if !exists {
			result.CreditTransitions = append(result.CreditTransitions, event)
		}
		seen[event.LogIndex] = string(encoded)
	}
	if _, err = c.canonical(ctx, block); err != nil {
		return ProtocolTransactionUpdates{}, err
	}
	return result, nil
}
