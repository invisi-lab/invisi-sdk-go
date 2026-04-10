package tree

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// mockClient implements EthClient for testing.
type mockClient struct {
	mu           sync.Mutex
	logs         []types.Log
	blockNumber  uint64
	subscribeCh  chan<- types.Log
	subscribeErr error
	callResults  map[string][]byte // keyed by hex-encoded calldata prefix
	filterLogsFn func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

func newMockClient() *mockClient {
	return &mockClient{
		blockNumber: 100,
		callResults: make(map[string][]byte),
	}
}

func (m *mockClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.filterLogsFn != nil {
		return m.filterLogsFn(ctx, q)
	}

	// Filter logs by address and topics
	var result []types.Log
	for _, l := range m.logs {
		if !matchesQuery(l, q) {
			continue
		}
		result = append(result, l)
	}
	return result, nil
}

func (m *mockClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	if m.subscribeErr != nil {
		return nil, m.subscribeErr
	}
	m.mu.Lock()
	m.subscribeCh = ch
	m.mu.Unlock()
	return &mockSubscription{ctx: ctx}, nil
}

func (m *mockClient) BlockNumber(ctx context.Context) (uint64, error) {
	return m.blockNumber, nil
}

func (m *mockClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	// Return pre-configured result based on first 4 bytes of data
	if len(msg.Data) >= 4 {
		key := common.Bytes2Hex(msg.Data[:4])
		if result, ok := m.callResults[key]; ok {
			return result, nil
		}
	}
	return nil, nil
}

// sendLog simulates a real-time event arriving via subscription.
func (m *mockClient) sendLog(l types.Log) {
	m.mu.Lock()
	ch := m.subscribeCh
	m.mu.Unlock()
	if ch != nil {
		ch <- l
	}
}

// matchesQuery checks if a log matches a filter query.
func matchesQuery(l types.Log, q ethereum.FilterQuery) bool {
	// Check address
	if len(q.Addresses) > 0 {
		found := false
		for _, addr := range q.Addresses {
			if l.Address == addr {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check topics (topic filters)
	for i, topicFilter := range q.Topics {
		if len(topicFilter) == 0 {
			continue // no filter for this topic position
		}
		if i >= len(l.Topics) {
			return false
		}
		found := false
		for _, t := range topicFilter {
			if l.Topics[i] == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// mockSubscription implements ethereum.Subscription.
type mockSubscription struct {
	ctx context.Context
}

func (s *mockSubscription) Err() <-chan error {
	ch := make(chan error)
	go func() {
		<-s.ctx.Done()
		ch <- s.ctx.Err()
	}()
	return ch
}

func (s *mockSubscription) Unsubscribe() {}

// buildPayOpenLog builds a mock PayOpen event log.
func buildPayOpenLog(
	addr common.Address,
	treeID *big.Int,
	paymentCommitment [32]byte,
	leafIndex uint32,
) types.Log {
	// Pack non-indexed data
	data, _ := protocolABI.Events["PayOpen"].Inputs.NonIndexed().Pack(
		leafIndex,        // leafIndex
		common.Address{}, // tokenAddress
		big.NewInt(0),    // amount
		[32]byte{},       // identityCommitment
		[32]byte{},       // nonce
		[32]byte{},       // refundLink
	)

	return types.Log{
		Address: addr,
		Topics: []common.Hash{
			payOpenEventSig,
			paymentCommitment,            // indexed: paymentCommitment
			common.BytesToHash([]byte{}), // indexed: payer
			common.BigToHash(treeID),     // indexed: treeId
		},
		Data: data,
	}
}

// buildPayBlindLog builds a mock PayBlind event log.
func buildPayBlindLog(
	addr common.Address,
	treeID *big.Int,
	paymentCommitment [32]byte,
	changeCommitment [32]byte,
	paymentLeafIndex uint32,
	changeLeafIndex uint32,
) types.Log {
	data, _ := protocolABI.Events["PayBlind"].Inputs.NonIndexed().Pack(
		changeCommitment, // changeCommitment
		paymentLeafIndex, // paymentLeafIndex
		changeLeafIndex,  // changeLeafIndex
		common.Address{}, // tokenAddress
		big.NewInt(0),    // inputAmount
		[32]byte{},       // refundLink
	)

	return types.Log{
		Address: addr,
		Topics: []common.Hash{
			payBlindEventSig,
			paymentCommitment,            // indexed: paymentCommitment
			common.BytesToHash([]byte{}), // indexed: payer
			common.BigToHash(treeID),     // indexed: treeId
		},
		Data: data,
	}
}

// buildCreditDepositLog builds a mock CreditDeposit event log.
func buildCreditDepositLog(
	addr common.Address,
	treeID *big.Int,
	commitment [32]byte,
	leafIndex uint32,
) types.Log {
	data, _ := protocolABI.Events["CreditDeposit"].Inputs.NonIndexed().Pack(
		leafIndex,        // leafIndex
		[32]byte{},       // identityCommitment
		[32]byte{},       // payerCommitment
		common.Address{}, // tokenAddress
		big.NewInt(0),    // amount
		[32]byte{},       // nonce
		big.NewInt(0),    // expiry
		[32]byte{},       // nullifierKeyHash
	)

	return types.Log{
		Address: addr,
		Topics: []common.Hash{
			creditDepositEventSig,
			commitment,                   // indexed: commitment
			common.BytesToHash([]byte{}), // indexed: depositor
			common.BigToHash(treeID),     // indexed: treeId
		},
		Data: data,
	}
}

// buildCreditDrawLog builds a mock CreditDraw event log.
func buildCreditDrawLog(
	addr common.Address,
	treeID *big.Int,
	remainderCommitments [][32]byte,
	remainderLeafIndices []uint32,
) types.Log {
	data, _ := protocolABI.Events["CreditDraw"].Inputs.NonIndexed().Pack(
		[32]byte{},              // merkleRoot
		[][32]byte{},            // nullifiers
		big.NewInt(0),           // totalAmount
		remainderCommitments,    // remainderCommitments
		remainderLeafIndices,    // remainderLeafIndices
		big.NewInt(0),           // expiry
	)

	return types.Log{
		Address: addr,
		Topics: []common.Hash{
			creditDrawEventSig,
			common.BytesToHash([]byte{}), // indexed: receivingAddress
			common.BytesToHash([]byte{}), // indexed: tokenAddress
			common.BigToHash(treeID),     // indexed: treeId
		},
		Data: data,
	}
}
