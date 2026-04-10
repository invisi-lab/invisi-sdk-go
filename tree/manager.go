package tree

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// InvisiTreeManagerOptions configures an InvisiTreeManager.
type InvisiTreeManagerOptions struct {
	Client          EthClient
	InstanceAddress common.Address
	FromBlock       uint64
}

// treeEntry holds a tree and its sync.
type treeEntry struct {
	tree *InMemoryMerkleTree
	sync *MerkleTreeSync
}

// InvisiTreeManager manages multiple Payment Trees indexed by treeID.
// treeId starts at 1 (index 0 is reserved/invalid on-chain).
type InvisiTreeManager struct {
	client          EthClient
	instanceAddress common.Address
	fromBlock       uint64

	mu        sync.RWMutex
	trees     map[string]*treeEntry // keyed by treeID.String()
	treeCount int
	cancel    context.CancelFunc
}

// NewInvisiTreeManager creates a new multi-tree manager.
func NewInvisiTreeManager(opts InvisiTreeManagerOptions) *InvisiTreeManager {
	return &InvisiTreeManager{
		client:          opts.Client,
		instanceAddress: opts.InstanceAddress,
		fromBlock:       opts.FromBlock,
		trees:           make(map[string]*treeEntry),
	}
}

// Start discovers all trees from the contract and starts syncing each.
func (m *InvisiTreeManager) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	count, err := m.fetchTreeCount(ctx)
	if err != nil {
		cancel()
		return fmt.Errorf("fetching tree count: %w", err)
	}
	m.treeCount = count

	// Start all trees concurrently
	var wg sync.WaitGroup
	errCh := make(chan error, count)

	for i := 1; i <= count; i++ {
		wg.Add(1)
		treeID := big.NewInt(int64(i))
		go func(id *big.Int) {
			defer wg.Done()
			if err := m.initTree(ctx, id); err != nil {
				errCh <- fmt.Errorf("init tree %s: %w", id.String(), err)
			}
		}(treeID)
	}

	wg.Wait()
	close(errCh)

	// Return first error if any
	for err := range errCh {
		if err != nil {
			cancel()
			return err
		}
	}

	return nil
}

// Stop stops all tree syncs.
func (m *InvisiTreeManager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, entry := range m.trees {
		entry.sync.Stop()
	}
}

// TreeCount returns the number of trees discovered from the contract.
func (m *InvisiTreeManager) TreeCount() int {
	return m.treeCount
}

// GetTree returns the InMemoryMerkleTree for the given treeID.
func (m *InvisiTreeManager) GetTree(treeID *big.Int) (*InMemoryMerkleTree, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.trees[treeID.String()]
	if !ok {
		return nil, fmt.Errorf("tree %s not found", treeID.String())
	}
	return entry.tree, nil
}

// GetPaymentProof returns a MerkleProof for a leaf in the specified tree.
func (m *InvisiTreeManager) GetPaymentProof(treeID *big.Int, leafIndex int) (MerkleProof, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.trees[treeID.String()]
	if !ok {
		return MerkleProof{}, fmt.Errorf("tree %s not found", treeID.String())
	}
	return entry.tree.GetProof(leafIndex)
}

// initTree creates and starts a single tree sync.
func (m *InvisiTreeManager) initTree(ctx context.Context, treeID *big.Int) error {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		return err
	}

	s := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          m.client,
		InstanceAddress: m.instanceAddress,
		FromBlock:       m.fromBlock,
		TreeID:          treeID,
	})

	if err := s.Start(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	m.trees[treeID.String()] = &treeEntry{tree: tree, sync: s}
	m.mu.Unlock()

	return nil
}

// fetchTreeCount calls getTreeCount() on the contract.
func (m *InvisiTreeManager) fetchTreeCount(ctx context.Context) (int, error) {
	// ABI-encode getTreeCount() call.
	// getTreeCount() has selector 0x091f3d5e (keccak256("getTreeCount()")[:4]).
	// We construct it from ABI to be safe.
	const getTreeCountABI = `[{"type":"function","name":"getTreeCount","inputs":[],"outputs":[{"name":"count","type":"uint256"}]}]`
	parsed, err := abi.JSON(strings.NewReader(getTreeCountABI))
	if err != nil {
		return 0, err
	}

	calldata, err := parsed.Pack("getTreeCount")
	if err != nil {
		return 0, err
	}

	result, err := m.client.CallContract(ctx, ethereum.CallMsg{
		To:   &m.instanceAddress,
		Data: calldata,
	}, nil)
	if err != nil {
		return 0, err
	}

	outputs, err := parsed.Unpack("getTreeCount", result)
	if err != nil {
		return 0, err
	}
	if len(outputs) == 0 {
		return 0, fmt.Errorf("getTreeCount returned no outputs")
	}

	count, ok := outputs[0].(*big.Int)
	if !ok {
		return 0, fmt.Errorf("getTreeCount: unexpected type %T", outputs[0])
	}
	return int(count.Int64()), nil
}
