package tree

import (
	"context"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/contracts"
)

// EthClient is the RPC surface required for note-tree synchronization.
type EthClient interface {
	FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(context.Context, ethereum.FilterQuery, chan<- types.Log) (ethereum.Subscription, error)
	BlockNumber(context.Context) (uint64, error)
	CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)
}

type leafEntry struct {
	commitment [32]byte
	leafIndex  int
}

var protocolABI = mustProtocolABI()
var commitmentsInsertedEventSig = protocolABI.Events["CommitmentsInserted"].ID
var batchOutputsEventSig = protocolABI.Events["BatchOutputs"].ID
var treeCreatedEventSig = protocolABI.Events["TreeCreated"].ID

func mustProtocolABI() abi.ABI {
	parsed, err := contracts.InvisiProtocolMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	return *parsed
}

// parseEventsToLeaves reads only canonical insertions for the requested tree.
// Business events may repeat commitments and must never drive Merkle insertion.
func parseEventsToLeaves(logs []types.Log, treeID *big.Int) ([]leafEntry, error) {
	leaves := make([]leafEntry, 0)
	for _, log := range logs {
		if len(log.Topics) == 0 || (log.Topics[0] != commitmentsInsertedEventSig && log.Topics[0] != batchOutputsEventSig) {
			continue
		}
		if len(log.Topics) != 2 {
			return nil, fmt.Errorf("canonical insertion: expected two topics")
		}
		if log.Topics[1] != common.BigToHash(treeID) {
			continue
		}
		if log.Removed {
			return nil, fmt.Errorf("canonical insertion removed by reorg; rebuild required")
		}
		name := "CommitmentsInserted"
		commitmentOffset := 1
		if log.Topics[0] == batchOutputsEventSig {
			name = "BatchOutputs"
			commitmentOffset = 2
		}
		values, err := protocolABI.Events[name].Inputs.NonIndexed().Unpack(log.Data)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", name, err)
		}
		start := values[0].(uint32)
		commitments := values[commitmentOffset].([][32]byte)
		if uint64(start)+uint64(len(commitments)) > MaxLeaves {
			return nil, fmt.Errorf("canonical insertion exceeds tree capacity")
		}
		for i, c := range commitments {
			leaves = append(leaves, leafEntry{c, int(start) + i})
		}
	}
	sort.SliceStable(leaves, func(i, j int) bool { return leaves[i].leafIndex < leaves[j].leafIndex })
	return leaves, nil
}
