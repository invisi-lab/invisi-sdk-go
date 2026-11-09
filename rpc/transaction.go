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

type TransactionAvailability struct {
	TransactionHash common.Hash
	Status          TransactionUpdateStatus
	IndexedHead     *CanonicalBlock
	Block           *CanonicalBlock
	ReceiptStatus   *uint64
}
type ProtocolTransactionNoteUpdates struct {
	TransactionAvailability
	Notes []ProtocolNoteOutput
}
type ProtocolTransactionCreditUpdates struct {
	TransactionAvailability
	States      []IndexedCreditState
	Transitions []ProtocolCreditTransition
}
type transactionUpdates struct {
	TransactionAvailability
	Notes       []ProtocolNoteOutput
	States      []IndexedCreditState
	Transitions []ProtocolCreditTransition
}

func (c *ProtocolRpcClient) GetTransactionNoteUpdates(ctx context.Context, scope DeploymentContext, hash common.Hash) (ProtocolTransactionNoteUpdates, error) {
	v, e := c.readTransaction(ctx, scope, hash, false)
	return ProtocolTransactionNoteUpdates{v.TransactionAvailability, v.Notes}, e
}
func (c *ProtocolRpcClient) GetTransactionCreditUpdates(ctx context.Context, scope DeploymentContext, hash common.Hash) (ProtocolTransactionCreditUpdates, error) {
	v, e := c.readTransaction(ctx, scope, hash, true)
	return ProtocolTransactionCreditUpdates{v.TransactionAvailability, v.States, v.Transitions}, e
}

// readTransaction validates availability before selecting a typed business collection.
// Context controls cancellation. ReceiptStatus is 0 for reverted or 1 for successful execution.
func (c *ProtocolRpcClient) readTransaction(ctx context.Context, scope DeploymentContext, transactionHash common.Hash, credit bool) (transactionUpdates, error) {
	var result transactionUpdates
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
	method := "invisi_getTransactionNoteUpdates"
	fields := []string{"notes"}
	if credit {
		method = "invisi_getTransactionCreditUpdates"
		fields = []string{"states", "transitions"}
	}
	var raw json.RawMessage
	if err = c.call(ctx, scope.ChainID, method, []any{transactionHash.Hex()}, &raw); err != nil {
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
		return transactionUpdates{}, err
	}
	switch result.Status {
	case TransactionNotFound, TransactionPendingIndex, TransactionIndexed:
	default:
		return transactionUpdates{}, errors.New("invalid transaction update status")
	}
	result.TransactionHash = transactionHash
	headRaw, exists := wire["indexedHead"]
	if !exists {
		return transactionUpdates{}, errors.New("missing indexed head")
	}
	if string(headRaw) != "null" {
		head, err := wireBlock(headRaw)
		if err != nil {
			return transactionUpdates{}, err
		}
		if result.Status == TransactionIndexed {
			head, err = c.canonical(ctx, head)
			if err != nil {
				return transactionUpdates{}, err
			}
		}
		result.IndexedHead = &head
	}
	if result.Status == TransactionNotFound {
		if len(wire["blockNumber"]) != 0 || len(wire["blockHash"]) != 0 || len(wire["receiptStatus"]) != 0 || !nullCollections(wire, fields) {
			return transactionUpdates{}, errors.New("invalid not_found transaction updates")
		}
		return result, nil
	}
	number, err := wireUint(wire["blockNumber"], 64)
	if err != nil {
		return transactionUpdates{}, err
	}
	start := deployment.Protocol.FromBlock
	if deployment.Registry != nil && deployment.Registry.FromBlock < start {
		start = deployment.Registry.FromBlock
	}
	if deployment.BroadcasterRegistry != nil && deployment.BroadcasterRegistry.FromBlock < start {
		start = deployment.BroadcasterRegistry.FromBlock
	}
	if number < start {
		return transactionUpdates{}, errors.New("transaction predates indexed deployment")
	}
	blockHash, err := wireHash(wire["blockHash"])
	if err != nil {
		return transactionUpdates{}, err
	}
	block, err := c.block(ctx, number)
	if err != nil {
		return transactionUpdates{}, err
	}
	if block.Hash != blockHash {
		return transactionUpdates{}, errors.New("RPC transaction block mismatch")
	}
	receiptStatus, err := wireUint(wire["receiptStatus"], 1)
	if err != nil {
		return transactionUpdates{}, err
	}
	result.Block = &block
	result.ReceiptStatus = &receiptStatus
	if result.Status == TransactionPendingIndex {
		if !nullCollections(wire, fields) {
			return transactionUpdates{}, errors.New("unindexed transaction must not contain updates")
		}
		return result, nil
	}
	if result.IndexedHead == nil || result.IndexedHead.Number < number || (result.IndexedHead.Number == number && result.IndexedHead.Hash != block.Hash) {
		return transactionUpdates{}, errors.New("transaction exceeds indexed head")
	}

	collections := map[string][]json.RawMessage{}
	for _, field := range fields {
		var rows []json.RawMessage
		if json.Unmarshal(wire[field], &rows) != nil || rows == nil {
			return transactionUpdates{}, errors.New("indexed transaction requires arrays")
		}
		if receiptStatus == 0 && len(rows) > 0 {
			return transactionUpdates{}, errors.New("reverted transaction cannot contain updates")
		}
		for _, row := range rows {
			item, e := wireObject(row)
			if e != nil {
				return transactionUpdates{}, e
			}
			h, e := wireHash(item["transactionHash"])
			if e != nil || h != transactionHash {
				return transactionUpdates{}, errors.New("updates contain a different transaction")
			}
		}
		collections[field] = rows
	}
	r := ProtocolUpdateRange{FromBlock: number, ToBlock: number}
	if credit {
		result.States, err = parseRows(ctx, c, block, collections["states"], r, parseCreditState)
		if err == nil {
			result.Transitions, err = c.parseTransitions(ctx, collections["transitions"], r)
		}
	} else {
		result.Notes, err = parseRows(ctx, c, block, collections["notes"], r, parseNoteOutput)
	}
	if err != nil {
		return transactionUpdates{}, err
	}
	if _, err = c.canonical(ctx, block); err != nil {
		return transactionUpdates{}, err
	}
	return result, nil
}
func nullCollections(wire map[string]json.RawMessage, fields []string) bool {
	for _, field := range fields {
		if string(wire[field]) != "null" {
			return false
		}
	}
	return true
}
