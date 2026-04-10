package tree

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthClient abstracts the Ethereum client methods needed by tree sync.
// In production, *ethclient.Client satisfies this. Tests provide a mock.
type EthClient interface {
	// FilterLogs returns logs matching the filter query.
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	// SubscribeFilterLogs subscribes to new logs matching the filter.
	SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
	// BlockNumber returns the current block number.
	BlockNumber(ctx context.Context) (uint64, error)
	// CallContract executes a contract call (for getTreeCount).
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

// leafEntry is a parsed commitment with its tree leaf index.
type leafEntry struct {
	commitment [32]byte
	leafIndex  int
}

// Event topic hashes — computed from event signatures.
// These are keccak256 of the event signature string.
var (
	payOpenEventSig       = common.HexToHash("0x") // placeholder, set in init
	payBlindEventSig      = common.HexToHash("0x") // placeholder, set in init
	creditDepositEventSig = common.HexToHash("0x") // placeholder, set in init
	creditDrawEventSig    = common.HexToHash("0x") // placeholder, set in init

	protocolABI abi.ABI
)

func init() {
	// Minimal ABI with only the events we need to parse.
	// Full ABI is in the contracts package; we inline the event definitions here
	// to avoid a circular dependency (tree -> contracts).
	const abiJSON = `[
		{
			"type": "event",
			"name": "PayOpen",
			"inputs": [
				{"name": "paymentCommitment", "type": "bytes32", "indexed": true},
				{"name": "payer", "type": "address", "indexed": true},
				{"name": "treeId", "type": "uint256", "indexed": true},
				{"name": "leafIndex", "type": "uint32", "indexed": false},
				{"name": "tokenAddress", "type": "address", "indexed": false},
				{"name": "amount", "type": "uint256", "indexed": false},
				{"name": "identityCommitment", "type": "bytes32", "indexed": false},
				{"name": "nonce", "type": "bytes32", "indexed": false},
				{"name": "refundLink", "type": "bytes32", "indexed": false}
			]
		},
		{
			"type": "event",
			"name": "PayBlind",
			"inputs": [
				{"name": "paymentCommitment", "type": "bytes32", "indexed": true},
				{"name": "payer", "type": "address", "indexed": true},
				{"name": "treeId", "type": "uint256", "indexed": true},
				{"name": "changeCommitment", "type": "bytes32", "indexed": false},
				{"name": "paymentLeafIndex", "type": "uint32", "indexed": false},
				{"name": "changeLeafIndex", "type": "uint32", "indexed": false},
				{"name": "tokenAddress", "type": "address", "indexed": false},
				{"name": "inputAmount", "type": "uint256", "indexed": false},
				{"name": "refundLink", "type": "bytes32", "indexed": false}
			]
		},
		{
			"type": "event",
			"name": "CreditDeposit",
			"inputs": [
				{"name": "commitment", "type": "bytes32", "indexed": true},
				{"name": "depositor", "type": "address", "indexed": true},
				{"name": "treeId", "type": "uint256", "indexed": true},
				{"name": "leafIndex", "type": "uint32", "indexed": false},
				{"name": "identityCommitment", "type": "bytes32", "indexed": false},
				{"name": "payerCommitment", "type": "bytes32", "indexed": false},
				{"name": "tokenAddress", "type": "address", "indexed": false},
				{"name": "amount", "type": "uint256", "indexed": false},
				{"name": "nonce", "type": "bytes32", "indexed": false},
				{"name": "expiry", "type": "uint256", "indexed": false},
				{"name": "nullifierKeyHash", "type": "bytes32", "indexed": false}
			]
		},
		{
			"type": "event",
			"name": "CreditDraw",
			"inputs": [
				{"name": "receivingAddress", "type": "address", "indexed": true},
				{"name": "tokenAddress", "type": "address", "indexed": true},
				{"name": "treeId", "type": "uint256", "indexed": true},
				{"name": "merkleRoot", "type": "bytes32", "indexed": false},
				{"name": "nullifiers", "type": "bytes32[]", "indexed": false},
				{"name": "totalAmount", "type": "uint256", "indexed": false},
				{"name": "remainderCommitments", "type": "bytes32[]", "indexed": false},
				{"name": "remainderLeafIndices", "type": "uint32[]", "indexed": false},
				{"name": "expiry", "type": "uint256", "indexed": false}
			]
		}
	]`

	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		panic(fmt.Sprintf("failed to parse protocol ABI: %v", err))
	}
	protocolABI = parsed

	payOpenEventSig = protocolABI.Events["PayOpen"].ID
	payBlindEventSig = protocolABI.Events["PayBlind"].ID
	creditDepositEventSig = protocolABI.Events["CreditDeposit"].ID
	creditDrawEventSig = protocolABI.Events["CreditDraw"].ID
}

// parseEventsToLeaves extracts leaf entries from raw logs.
// It handles PayOpen, PayBlind, CreditDeposit, and CreditDraw events.
func parseEventsToLeaves(logs []types.Log) ([]leafEntry, error) {
	var leaves []leafEntry

	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}

		eventSig := log.Topics[0]

		switch eventSig {
		case payOpenEventSig:
			entry, err := parsePayOpen(log)
			if err != nil {
				return nil, fmt.Errorf("parsing PayOpen: %w", err)
			}
			leaves = append(leaves, entry)

		case payBlindEventSig:
			entries, err := parsePayBlind(log)
			if err != nil {
				return nil, fmt.Errorf("parsing PayBlind: %w", err)
			}
			leaves = append(leaves, entries...)

		case creditDepositEventSig:
			entry, err := parseCreditDeposit(log)
			if err != nil {
				return nil, fmt.Errorf("parsing CreditDeposit: %w", err)
			}
			leaves = append(leaves, entry)

		case creditDrawEventSig:
			entries, err := parseCreditDraw(log)
			if err != nil {
				return nil, fmt.Errorf("parsing CreditDraw: %w", err)
			}
			leaves = append(leaves, entries...)
		}
	}

	// Sort by leafIndex to ensure sequential insertion
	sort.Slice(leaves, func(i, j int) bool {
		return leaves[i].leafIndex < leaves[j].leafIndex
	})

	return leaves, nil
}

func parsePayOpen(log types.Log) (leafEntry, error) {
	// paymentCommitment is indexed (topic[1])
	if len(log.Topics) < 4 {
		return leafEntry{}, fmt.Errorf("PayOpen: expected 4 topics, got %d", len(log.Topics))
	}
	commitment := log.Topics[1] // bytes32, indexed

	// Unpack non-indexed fields
	values, err := protocolABI.Events["PayOpen"].Inputs.NonIndexed().Unpack(log.Data)
	if err != nil {
		return leafEntry{}, fmt.Errorf("PayOpen unpack: %w", err)
	}
	// values[0] = leafIndex (uint32)
	leafIndex := values[0].(uint32)

	return leafEntry{
		commitment: commitment,
		leafIndex:  int(leafIndex),
	}, nil
}

func parsePayBlind(log types.Log) ([]leafEntry, error) {
	// paymentCommitment is indexed (topic[1])
	if len(log.Topics) < 4 {
		return nil, fmt.Errorf("PayBlind: expected 4 topics, got %d", len(log.Topics))
	}
	paymentCommitment := log.Topics[1]

	// Unpack non-indexed fields
	values, err := protocolABI.Events["PayBlind"].Inputs.NonIndexed().Unpack(log.Data)
	if err != nil {
		return nil, fmt.Errorf("PayBlind unpack: %w", err)
	}
	// values[0] = changeCommitment (bytes32)
	// values[1] = paymentLeafIndex (uint32)
	// values[2] = changeLeafIndex (uint32)
	changeCommitment := values[0].([32]byte)
	paymentLeafIndex := values[1].(uint32)
	changeLeafIndex := values[2].(uint32)

	return []leafEntry{
		{commitment: paymentCommitment, leafIndex: int(paymentLeafIndex)},
		{commitment: changeCommitment, leafIndex: int(changeLeafIndex)},
	}, nil
}

func parseCreditDeposit(log types.Log) (leafEntry, error) {
	// commitment is indexed (topic[1])
	if len(log.Topics) < 4 {
		return leafEntry{}, fmt.Errorf("CreditDeposit: expected 4 topics, got %d", len(log.Topics))
	}
	commitment := log.Topics[1]

	values, err := protocolABI.Events["CreditDeposit"].Inputs.NonIndexed().Unpack(log.Data)
	if err != nil {
		return leafEntry{}, fmt.Errorf("CreditDeposit unpack: %w", err)
	}
	// values[0] = leafIndex (uint32)
	leafIndex := values[0].(uint32)

	return leafEntry{
		commitment: commitment,
		leafIndex:  int(leafIndex),
	}, nil
}

func parseCreditDraw(log types.Log) ([]leafEntry, error) {
	values, err := protocolABI.Events["CreditDraw"].Inputs.NonIndexed().Unpack(log.Data)
	if err != nil {
		return nil, fmt.Errorf("CreditDraw unpack: %w", err)
	}
	// values[0] = merkleRoot (bytes32)
	// values[1] = nullifiers ([]bytes32)
	// values[2] = totalAmount (uint256)
	// values[3] = remainderCommitments ([]bytes32)
	// values[4] = remainderLeafIndices ([]uint32)
	// values[5] = expiry (uint256)
	remainders := values[3].([][32]byte)
	indices := values[4].([]uint32)

	var entries []leafEntry
	for i := 0; i < len(remainders); i++ {
		// Skip zero commitments
		isZero := true
		for _, b := range remainders[i] {
			if b != 0 {
				isZero = false
				break
			}
		}
		if !isZero {
			entries = append(entries, leafEntry{
				commitment: remainders[i],
				leafIndex:  int(indices[i]),
			})
		}
	}

	return entries, nil
}
