package tree

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DefaultPollInterval is the fallback polling interval when WebSocket is unavailable.
const DefaultPollInterval = 2 * time.Second

// MerkleTreeSyncOptions configures a MerkleTreeSync instance.
type MerkleTreeSyncOptions struct {
	Tree            *InMemoryMerkleTree
	Client          EthClient
	InstanceAddress common.Address
	FromBlock       uint64
	TreeID          *big.Int
	// PollInterval overrides the default polling interval (for testing).
	// Zero means use DefaultPollInterval.
	PollInterval time.Duration
}

// MerkleTreeSync syncs an InMemoryMerkleTree with on-chain events for a single treeID.
type MerkleTreeSync struct {
	tree            *InMemoryMerkleTree
	client          EthClient
	instanceAddress common.Address
	fromBlock       uint64
	treeID          *big.Int
	pollInterval    time.Duration

	synced          atomic.Bool
	lastSyncedBlock atomic.Uint64
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.Mutex
}

// NewMerkleTreeSync creates a new sync instance. Call Start to begin synchronization.
func NewMerkleTreeSync(opts MerkleTreeSyncOptions) *MerkleTreeSync {
	interval := opts.PollInterval
	if interval == 0 {
		interval = DefaultPollInterval
	}
	return &MerkleTreeSync{
		tree:            opts.Tree,
		client:          opts.Client,
		instanceAddress: opts.InstanceAddress,
		fromBlock:       opts.FromBlock,
		treeID:          opts.TreeID,
		pollInterval:    interval,
	}
}

// Start syncs historical events, then subscribes to real-time events.
// It blocks until historical sync is complete, then starts background subscription.
// Cancel the context or call Stop to stop the sync.
func (s *MerkleTreeSync) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Phase 1: Historical sync
	if err := s.scanHistory(ctx); err != nil {
		cancel()
		return fmt.Errorf("historical sync: %w", err)
	}

	s.synced.Store(true)

	// Phase 2: Real-time subscription (background)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.subscribe(ctx)
	}()

	return nil
}

// Stop stops the sync and waits for background goroutines to finish.
func (s *MerkleTreeSync) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

// Synced returns true once historical sync is complete.
func (s *MerkleTreeSync) Synced() bool {
	return s.synced.Load()
}

// LastSyncedBlock returns the last block number that was synced.
func (s *MerkleTreeSync) LastSyncedBlock() uint64 {
	return s.lastSyncedBlock.Load()
}

// buildFilterQuery creates an ethereum.FilterQuery for all tree events
// filtered by instance address and treeID.
func (s *MerkleTreeSync) buildFilterQuery(from, to uint64) ethereum.FilterQuery {
	// treeId is the 4th topic (index 3) — it's the 3rd indexed param in all events.
	// Topic[0] = event sig (we pass all four sigs to match any of them)
	// Topic[3] = treeId
	treeIDHash := common.BigToHash(s.treeID)

	return ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Addresses: []common.Address{s.instanceAddress},
		Topics: [][]common.Hash{
			{payOpenEventSig, payBlindEventSig, creditDepositEventSig, creditDrawEventSig},
			{},           // topic[1]: any (commitment/address)
			{},           // topic[2]: any (payer/depositor/tokenAddress)
			{treeIDHash}, // topic[3]: treeId
		},
	}
}

// scanHistory fetches all historical events and inserts leaves into the tree.
func (s *MerkleTreeSync) scanHistory(ctx context.Context) error {
	latestBlock, err := s.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("getting block number: %w", err)
	}

	if s.fromBlock > latestBlock {
		s.lastSyncedBlock.Store(latestBlock)
		return nil
	}

	query := s.buildFilterQuery(s.fromBlock, latestBlock)
	logs, err := s.client.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("filter logs: %w", err)
	}

	if err := s.applyLogs(logs); err != nil {
		return err
	}

	s.lastSyncedBlock.Store(latestBlock)
	return nil
}

// subscribe starts real-time event watching.
// Tries WebSocket subscription first; falls back to polling if that fails.
func (s *MerkleTreeSync) subscribe(ctx context.Context) {
	// Try WebSocket subscription
	ch := make(chan types.Log, 256)
	query := s.buildFilterQuery(0, 0) // from/to not used in subscriptions
	query.FromBlock = nil
	query.ToBlock = nil

	sub, err := s.client.SubscribeFilterLogs(ctx, query, ch)
	if err != nil {
		// WebSocket unavailable — fall back to polling
		s.poll(ctx)
		return
	}

	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-sub.Err():
			if err != nil && ctx.Err() == nil {
				// Subscription error — fall back to polling
				s.poll(ctx)
			}
			return
		case log := <-ch:
			s.mu.Lock()
			err := s.applyLogs([]types.Log{log})
			s.mu.Unlock()
			if err != nil {
				// Log error but continue — production code should add proper error handling
				continue
			}
			s.lastSyncedBlock.Store(log.BlockNumber)
		}
	}
}

// poll periodically fetches new events using FilterLogs.
func (s *MerkleTreeSync) poll(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lastBlock := s.lastSyncedBlock.Load()
			latestBlock, err := s.client.BlockNumber(ctx)
			if err != nil {
				continue
			}
			if latestBlock <= lastBlock {
				continue
			}

			query := s.buildFilterQuery(lastBlock+1, latestBlock)
			logs, err := s.client.FilterLogs(ctx, query)
			if err != nil {
				continue
			}

			s.mu.Lock()
			err = s.applyLogs(logs)
			s.mu.Unlock()
			if err != nil {
				continue
			}

			s.lastSyncedBlock.Store(latestBlock)
		}
	}
}

// applyLogs parses events into leaves and inserts them into the tree.
func (s *MerkleTreeSync) applyLogs(logs []types.Log) error {
	leaves, err := parseEventsToLeaves(logs)
	if err != nil {
		return err
	}

	for _, leaf := range leaves {
		// Convert [32]byte commitment to fr.Element
		var elem fr.Element
		if err := elem.SetBytesCanonical(leaf.commitment[:]); err != nil {
			return fmt.Errorf("invalid commitment at leafIndex %d: %w", leaf.leafIndex, err)
		}

		insertedIndex, err := s.tree.Insert(elem)
		if err != nil {
			return fmt.Errorf("insert at leafIndex %d: %w", leaf.leafIndex, err)
		}

		if insertedIndex != leaf.leafIndex {
			return fmt.Errorf("Merkle tree sync mismatch: expected leafIndex %d, got %d", leaf.leafIndex, insertedIndex)
		}
	}

	return nil
}
