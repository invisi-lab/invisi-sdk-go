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

const DefaultPollInterval = 2 * time.Second

type leafLocation struct {
	tree  string
	index int
}
type commitmentLedger struct {
	mu     sync.Mutex
	leaves map[[32]byte]leafLocation
}

func newCommitmentLedger() *commitmentLedger {
	return &commitmentLedger{leaves: make(map[[32]byte]leafLocation)}
}

type MerkleTreeSyncOptions struct {
	Tree            *InMemoryMerkleTree
	Client          EthClient
	InstanceAddress common.Address
	FromBlock       uint64
	TreeID          *big.Int
	PollInterval    time.Duration
	// OnError asynchronously reports a terminal live sync failure. Rebuild from canonical history.
	OnError func(error)
	ledger  *commitmentLedger
}

type MerkleTreeSync struct {
	tree            *InMemoryMerkleTree
	client          EthClient
	instanceAddress common.Address
	fromBlock       uint64
	treeID          *big.Int
	pollInterval    time.Duration
	onError         func(error)
	ledger          *commitmentLedger
	synced          atomic.Bool
	lastSyncedBlock atomic.Uint64
	mu              sync.Mutex
	statusMu        sync.Mutex
	lifecycleMu     sync.Mutex
	runCtx          context.Context
	err             error
	cancel          context.CancelFunc
	started         bool
	wg              sync.WaitGroup
}

func NewMerkleTreeSync(opts MerkleTreeSyncOptions) *MerkleTreeSync {
	interval := opts.PollInterval
	if interval <= 0 {
		interval = DefaultPollInterval
	}
	ledger := opts.ledger
	if ledger == nil {
		ledger = newCommitmentLedger()
	}
	id := new(big.Int)
	if opts.TreeID != nil {
		id.Set(opts.TreeID)
	}
	return &MerkleTreeSync{tree: opts.Tree, client: opts.Client, instanceAddress: opts.InstanceAddress, fromBlock: opts.FromBlock, treeID: id, pollInterval: interval, onError: opts.OnError, ledger: ledger}
}

// Start subscribes before taking the historical snapshot so rollover-era events
// cannot fall between history and live sync. Repeated overlapping leaves are idempotent.
func (s *MerkleTreeSync) Start(ctx context.Context) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	s.statusMu.Lock()
	err := s.err
	active := s.started && s.runCtx != nil && s.runCtx.Err() == nil
	s.statusMu.Unlock()
	if err != nil {
		return err
	}
	if active {
		return nil
	}
	// A cancelled run must release its subscription before its replacement starts.
	s.wg.Wait()
	if err := s.Error(); err != nil {
		return err
	}
	if s.treeID.Sign() <= 0 || s.tree == nil || s.client == nil {
		return fmt.Errorf("sync requires a positive tree ID, tree, and client")
	}
	ctx, cancel := context.WithCancel(ctx)
	s.statusMu.Lock()
	s.cancel = cancel
	s.runCtx = ctx
	s.started = true
	s.statusMu.Unlock()
	launched := false
	defer func() {
		if !launched {
			cancel()
			s.statusMu.Lock()
			s.started = false
			s.synced.Store(false)
			s.statusMu.Unlock()
		}
	}()
	ch := make(chan types.Log, 256)
	query := s.buildFilterQuery(s.fromBlock, 0)
	query.ToBlock = nil
	sub, subErr := s.client.SubscribeFilterLogs(ctx, query, ch)
	if err := s.scanHistory(ctx); err != nil {
		if sub != nil {
			sub.Unsubscribe()
		}
		if ctx.Err() == nil {
			s.fail(err)
		}
		return err
	}
	// Reconcile notifications buffered during the historical RPC before reporting
	// synchronization complete. They intentionally overlap the snapshot.
	buffered := make([]types.Log, 0, len(ch))
Drain:
	for {
		select {
		case log := <-ch:
			buffered = append(buffered, log)
		default:
			break Drain
		}
	}
	if err := s.applyLogs(buffered); err != nil {
		if sub != nil {
			sub.Unsubscribe()
		}
		if ctx.Err() == nil {
			s.fail(err)
		}
		return err
	}
	if err := ctx.Err(); err != nil {
		if sub != nil {
			sub.Unsubscribe()
		}
		return err
	}
	s.synced.Store(true)
	s.wg.Add(1)
	launched = true
	go func() {
		defer s.wg.Done()
		defer func() { s.statusMu.Lock(); s.started = false; s.synced.Store(false); s.statusMu.Unlock() }()
		if subErr != nil {
			s.poll(ctx)
			return
		}
		defer sub.Unsubscribe()
		s.subscribe(ctx, sub, ch)
	}()
	return nil
}
func (s *MerkleTreeSync) Stop() {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	s.statusMu.Lock()
	cancel := s.cancel
	s.started = false
	s.synced.Store(false)
	s.statusMu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.wg.Wait()
}
func (s *MerkleTreeSync) Synced() bool            { return s.synced.Load() }
func (s *MerkleTreeSync) LastSyncedBlock() uint64 { return s.lastSyncedBlock.Load() }
func (s *MerkleTreeSync) Error() error            { s.statusMu.Lock(); defer s.statusMu.Unlock(); return s.err }
func (s *MerkleTreeSync) fail(err error) {
	s.statusMu.Lock()
	if s.err != nil {
		s.statusMu.Unlock()
		return
	}
	s.err = err
	s.synced.Store(false)
	cancel := s.cancel
	s.statusMu.Unlock()
	if cancel != nil {
		cancel()
	}
	if s.onError != nil {
		go s.onError(err)
	}
}
func (s *MerkleTreeSync) buildFilterQuery(from, to uint64) ethereum.FilterQuery {
	return ethereum.FilterQuery{FromBlock: new(big.Int).SetUint64(from), ToBlock: new(big.Int).SetUint64(to), Addresses: []common.Address{s.instanceAddress}, Topics: [][]common.Hash{{commitmentsInsertedEventSig, batchOutputsEventSig}, {common.BigToHash(s.treeID)}}}
}
func (s *MerkleTreeSync) scanHistory(ctx context.Context) error {
	latest, err := s.client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	if latest < s.lastSyncedBlock.Load() {
		return fmt.Errorf("chain head regressed; rebuild required")
	}
	var logs []types.Log
	if s.fromBlock <= latest {
		logs, err = s.client.FilterLogs(ctx, s.buildFilterQuery(s.fromBlock, latest))
		if err != nil {
			return err
		}
	}
	if err = s.applyCanonicalHistory(logs); err != nil {
		return err
	}
	s.lastSyncedBlock.Store(latest)
	return nil
}
func (s *MerkleTreeSync) subscribe(ctx context.Context, sub ethereum.Subscription, ch <-chan types.Log) {
	errors := sub.Err()
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-errors:
			if ok && err != nil && ctx.Err() == nil {
				s.fail(fmt.Errorf("subscription: %w", err))
			}
			return
		case log, ok := <-ch:
			if !ok {
				s.fail(fmt.Errorf("insertion subscription closed"))
				return
			}
			if err := s.applyLogs([]types.Log{log}); err != nil {
				s.fail(err)
				return
			}
		}
	}
}
func (s *MerkleTreeSync) poll(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			latest, err := s.client.BlockNumber(ctx)
			if err != nil {
				continue
			}
			if latest < s.lastSyncedBlock.Load() {
				s.fail(fmt.Errorf("chain head regressed; rebuild required"))
				return
			}
			// EthClient has no block-header API. Replaying canonical history detects
			// removals even when a reorg has no replacement insertion or removed log.
			var logs []types.Log
			if s.fromBlock <= latest {
				logs, err = s.client.FilterLogs(ctx, s.buildFilterQuery(s.fromBlock, latest))
				if err != nil {
					continue
				}
			}
			if err = s.applyCanonicalHistory(logs); err != nil {
				s.fail(err)
				return
			}
			s.lastSyncedBlock.Store(latest)
		}
	}
}

// applyLogs validates and hashes the complete batch before publishing any leaves
// or global commitment reservations. A failure leaves the tree and ledger intact.
func (s *MerkleTreeSync) applyLogs(logs []types.Log) error {
	return s.applyLogBatch(logs, false)
}

func (s *MerkleTreeSync) applyCanonicalHistory(logs []types.Log) error {
	return s.applyLogBatch(logs, true)
}

func (s *MerkleTreeSync) applyLogBatch(logs []types.Log, completeHistory bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := make([]types.Log, 0, len(logs))
	for _, log := range logs {
		if log.Address == s.instanceAddress {
			filtered = append(filtered, log)
		}
	}
	leaves, err := parseEventsToLeaves(filtered, s.treeID)
	if err != nil {
		return err
	}
	s.ledger.mu.Lock()
	defer s.ledger.mu.Unlock()
	s.tree.mu.Lock()
	defer s.tree.mu.Unlock()
	if completeHistory {
		canonical := make(map[int][32]byte, len(leaves))
		for _, leaf := range leaves {
			canonical[leaf.leafIndex] = leaf.commitment
		}
		for i, leaf := range s.tree.layers[0] {
			commitment, exists := canonical[i]
			if !exists || commitment != leaf.Bytes() {
				return fmt.Errorf("canonical history changed at leaf index %d; reorg rebuild required", i)
			}
		}
	}
	pending := make(map[int]fr.Element)
	commitments := make(map[[32]byte]leafLocation)
	for i, leaf := range s.tree.layers[0] {
		commitments[leaf.Bytes()] = leafLocation{s.treeID.String(), i}
	}
	for _, leaf := range leaves {
		var value fr.Element
		if err = value.SetBytesCanonical(leaf.commitment[:]); err != nil {
			return fmt.Errorf("invalid commitment at index %d: %w", leaf.leafIndex, err)
		}
		if value.IsZero() {
			return fmt.Errorf("zero commitment at index %d", leaf.leafIndex)
		}
		location := leafLocation{s.treeID.String(), leaf.leafIndex}
		if previous, ok := s.ledger.leaves[leaf.commitment]; ok && previous != location {
			return fmt.Errorf("duplicate commitment across note trees")
		}
		if previous, ok := commitments[leaf.commitment]; ok && previous != location {
			return fmt.Errorf("duplicate commitment at leaf index %d", leaf.leafIndex)
		}
		commitments[leaf.commitment] = location
		if leaf.leafIndex < s.tree.size {
			if s.tree.layers[0][leaf.leafIndex] != value {
				return fmt.Errorf("commitment conflict at index %d", leaf.leafIndex)
			}
			continue
		}
		if previous, ok := pending[leaf.leafIndex]; ok && previous != value {
			return fmt.Errorf("commitment conflict at index %d", leaf.leafIndex)
		}
		pending[leaf.leafIndex] = value
	}
	for i := s.tree.size; i < s.tree.size+len(pending); i++ {
		if _, ok := pending[i]; !ok {
			return fmt.Errorf("gap at leaf index %d; rebuild required", i)
		}
	}
	if len(pending) > 0 {
		staged := s.tree.cloneLocked()
		for i := s.tree.size; i < s.tree.size+len(pending); i++ {
			if _, err := staged.insertLocked(pending[i]); err != nil {
				return err
			}
		}
		s.tree.layers = staged.layers
		s.tree.size = staged.size
	}
	for commitment, location := range commitments {
		s.ledger.leaves[commitment] = location
	}
	latest := s.lastSyncedBlock.Load()
	for _, log := range filtered {
		if len(log.Topics) == 2 && (log.Topics[0] == commitmentsInsertedEventSig || log.Topics[0] == batchOutputsEventSig) && log.Topics[1] == common.BigToHash(s.treeID) && log.BlockNumber > latest {
			latest = log.BlockNumber
		}
	}
	s.lastSyncedBlock.Store(latest)
	return nil
}
