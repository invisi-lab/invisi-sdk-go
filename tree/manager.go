package tree

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type InvisiTreeManagerOptions struct {
	Client          EthClient
	InstanceAddress common.Address
	FromBlock       uint64
	PollInterval    time.Duration
	OnError         func(error)
}
type treeEntry struct {
	tree *InMemoryMerkleTree
	sync *MerkleTreeSync
}

// InvisiTreeManager retains every historical note tree and discovers new active
// trees. IDs begin at one; each tree holds at most 65,536 commitments.
type InvisiTreeManager struct {
	client          EthClient
	instanceAddress common.Address
	fromBlock       uint64
	pollInterval    time.Duration
	onError         func(error)
	ledger          *commitmentLedger
	mu              sync.RWMutex
	discoveryMu     sync.Mutex
	lifecycleMu     sync.Mutex
	runCtx          context.Context
	trees           map[string]*treeEntry
	treeCount       int
	err             error
	cancel          context.CancelFunc
	running         bool
	wg              sync.WaitGroup
}

func NewInvisiTreeManager(opts InvisiTreeManagerOptions) *InvisiTreeManager {
	interval := opts.PollInterval
	if interval <= 0 {
		interval = DefaultPollInterval
	}
	return &InvisiTreeManager{client: opts.Client, instanceAddress: opts.InstanceAddress, fromBlock: opts.FromBlock, pollInterval: interval, onError: opts.OnError, ledger: newCommitmentLedger(), trees: make(map[string]*treeEntry)}
}
func (m *InvisiTreeManager) Start(ctx context.Context) error {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	m.mu.RLock()
	err := m.err
	active := m.running && m.runCtx != nil && m.runCtx.Err() == nil
	m.mu.RUnlock()
	if err != nil {
		return err
	}
	if active {
		return nil
	}
	m.wg.Wait()
	if err := m.Error(); err != nil {
		return err
	}
	if m.client == nil {
		return fmt.Errorf("client is nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	m.mu.Lock()
	m.cancel = cancel
	m.runCtx = ctx
	m.running = true
	m.mu.Unlock()
	launched := false
	defer func() {
		if !launched {
			m.finishRun(cancel)
		}
	}()
	ch := make(chan types.Log, 256)
	query := ethereum.FilterQuery{FromBlock: new(big.Int).SetUint64(m.fromBlock), Addresses: []common.Address{m.instanceAddress}, Topics: [][]common.Hash{{treeCreatedEventSig}}}
	sub, subErr := m.client.SubscribeFilterLogs(ctx, query, ch)
	count, err := m.fetchActiveTreeID(ctx)
	if err == nil && count < m.TreeCount() {
		err = fmt.Errorf("active tree ID regressed; rebuild required")
	}
	if err == nil {
		m.mu.RLock()
		existing := make([]*treeEntry, 0, len(m.trees))
		for _, entry := range m.trees {
			existing = append(existing, entry)
		}
		m.mu.RUnlock()
		for _, entry := range existing {
			if err = entry.sync.Start(ctx); err != nil {
				break
			}
		}
	}
	if err == nil {
		err = m.discoverThrough(ctx, count)
	}
	if err != nil {
		if sub != nil {
			sub.Unsubscribe()
		}
		if ctx.Err() == nil {
			m.fail(err)
		}
		return err
	}
	if err := ctx.Err(); err != nil {
		if sub != nil {
			sub.Unsubscribe()
		}
		return err
	}
	m.wg.Add(1)
	launched = true
	go func() {
		defer m.wg.Done()
		defer m.finishRun(cancel)
		if subErr == nil {
			defer sub.Unsubscribe()
		}
		m.watch(ctx, sub, ch)
	}()
	return nil
}
func (m *InvisiTreeManager) Stop() {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	m.mu.Lock()
	cancel := m.cancel
	m.running = false
	entries := make([]*treeEntry, 0, len(m.trees))
	for _, entry := range m.trees {
		entries = append(entries, entry)
	}
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	m.wg.Wait()
	for _, entry := range entries {
		entry.sync.Stop()
	}
}

// finishRun releases every child subscription before allowing a restart.
func (m *InvisiTreeManager) finishRun(cancel context.CancelFunc) {
	cancel()
	m.mu.RLock()
	entries := make([]*treeEntry, 0, len(m.trees))
	for _, entry := range m.trees {
		entries = append(entries, entry)
	}
	m.mu.RUnlock()
	for _, entry := range entries {
		entry.sync.Stop()
	}
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()
}

func (m *InvisiTreeManager) Error() error { m.mu.RLock(); defer m.mu.RUnlock(); return m.err }
func (m *InvisiTreeManager) fail(err error) {
	m.mu.Lock()
	if m.err != nil {
		m.mu.Unlock()
		return
	}
	m.err = err
	m.running = false
	cancel := m.cancel
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if m.onError != nil {
		go m.onError(err)
	}
}
func (m *InvisiTreeManager) TreeCount() int { m.mu.RLock(); defer m.mu.RUnlock(); return m.treeCount }
func (m *InvisiTreeManager) GetTree(id *big.Int) (*InMemoryMerkleTree, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	if id == nil || id.Sign() <= 0 {
		return nil, fmt.Errorf("tree ID must be positive")
	}
	entry, ok := m.trees[id.String()]
	if !ok {
		return nil, fmt.Errorf("tree %s not found", id)
	}
	if err := entry.sync.Error(); err != nil {
		return nil, err
	}
	return entry.tree, nil
}
func (m *InvisiTreeManager) GetNoteProof(id *big.Int, index int) (MerkleProof, error) {
	tree, err := m.GetTree(id)
	if err != nil {
		return MerkleProof{}, err
	}
	return tree.GetProof(index)
}
func (m *InvisiTreeManager) discoverThrough(ctx context.Context, count int) error {
	m.discoveryMu.Lock()
	defer m.discoveryMu.Unlock()
	for id := m.TreeCount() + 1; id <= count; id++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		tree, err := NewInMemoryMerkleTree()
		if err != nil {
			return err
		}
		s := NewMerkleTreeSync(MerkleTreeSyncOptions{Tree: tree, Client: m.client, InstanceAddress: m.instanceAddress, FromBlock: m.fromBlock, TreeID: big.NewInt(int64(id)), PollInterval: m.pollInterval, OnError: m.fail, ledger: m.ledger})
		if err = s.Start(ctx); err != nil {
			return fmt.Errorf("sync tree %d: %w", id, err)
		}
		m.mu.Lock()
		if !m.running {
			m.mu.Unlock()
			s.Stop()
			return ctx.Err()
		}
		m.trees[big.NewInt(int64(id)).String()] = &treeEntry{tree, s}
		m.treeCount = id
		m.mu.Unlock()
	}
	return nil
}
func (m *InvisiTreeManager) fetchActiveTreeID(ctx context.Context) (int, error) {
	data, err := protocolABI.Pack("activeNoteTreeId")
	if err != nil {
		return 0, err
	}
	result, err := m.client.CallContract(ctx, ethereum.CallMsg{To: &m.instanceAddress, Data: data}, nil)
	if err != nil {
		return 0, err
	}
	values, err := protocolABI.Unpack("activeNoteTreeId", result)
	if err != nil {
		return 0, err
	}
	if len(values) != 1 {
		return 0, fmt.Errorf("activeNoteTreeId returned invalid result")
	}
	id := values[0].(*big.Int)
	if id.Sign() <= 0 || !id.IsInt64() || uint64(id.Int64()) > uint64(^uint(0)>>1) {
		return 0, fmt.Errorf("invalid active note tree ID %s", id)
	}
	return int(id.Int64()), nil
}
func (m *InvisiTreeManager) watch(ctx context.Context, sub ethereum.Subscription, ch <-chan types.Log) {
	var errors <-chan error
	if sub != nil {
		errors = sub.Err()
	}
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-errors:
			if ok && err != nil && ctx.Err() == nil {
				m.fail(err)
			}
			return
		case log, ok := <-ch:
			if !ok {
				m.fail(fmt.Errorf("tree discovery subscription closed"))
				return
			}
			if log.Address != m.instanceAddress || len(log.Topics) == 0 || log.Topics[0] != treeCreatedEventSig {
				continue
			}
			if log.Removed {
				m.fail(fmt.Errorf("TreeCreated removed by reorg; rebuild required"))
				return
			}
			if len(log.Topics) != 2 {
				m.fail(fmt.Errorf("invalid TreeCreated topics"))
				return
			}
			id := new(big.Int).SetBytes(log.Topics[1][:])
			if !id.IsInt64() || id.Sign() <= 0 || uint64(id.Int64()) > uint64(^uint(0)>>1) {
				m.fail(fmt.Errorf("invalid TreeCreated ID"))
				return
			}
			if err := m.discoverThrough(ctx, int(id.Int64())); err != nil {
				if ctx.Err() == nil {
					m.fail(err)
				}
				return
			}
		case <-ticker.C:
			count, err := m.fetchActiveTreeID(ctx)
			if err != nil {
				continue
			}
			if count < m.TreeCount() {
				m.fail(fmt.Errorf("active tree ID regressed; rebuild required"))
				return
			}
			if err = m.discoverThrough(ctx, count); err != nil {
				if ctx.Err() == nil {
					m.fail(err)
				}
				return
			}
		}
	}
}
