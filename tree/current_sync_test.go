package tree

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

var testAddress = common.HexToAddress("0x1234")

func commitment(n int64) [32]byte { return common.BigToHash(big.NewInt(n)) }
func insertion(t *testing.T, name string, id int64, start uint32, values ...int64) types.Log {
	t.Helper()
	leaves := make([][32]byte, len(values))
	for i, n := range values {
		leaves[i] = commitment(n)
	}
	args := []interface{}{start, leaves}
	if name == "BatchOutputs" {
		args = []interface{}{start, big.NewInt(0), leaves, []byte{}}
	}
	data, err := protocolABI.Events[name].Inputs.NonIndexed().Pack(args...)
	if err != nil {
		t.Fatal(err)
	}
	return types.Log{Address: testAddress, Topics: []common.Hash{protocolABI.Events[name].ID, common.BigToHash(big.NewInt(id))}, Data: data, BlockNumber: 10}
}
func testSync(t *testing.T) *MerkleTreeSync {
	t.Helper()
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}
	return NewMerkleTreeSync(MerkleTreeSyncOptions{Tree: tree, InstanceAddress: testAddress, TreeID: big.NewInt(1)})
}
func TestCanonicalInsertionsAndOverlap(t *testing.T) {
	s := testSync(t)
	a := insertion(t, "CommitmentsInserted", 1, 0, 11, 12)
	b := insertion(t, "BatchOutputs", 1, 2, 13, 14)
	business := types.Log{Address: testAddress, Topics: []common.Hash{protocolABI.Events["Shield"].ID}, Data: []byte("ignored")}
	other := insertion(t, "CommitmentsInserted", 2, 0, 99)
	if err := s.applyLogs([]types.Log{b, business, other, a, a}); err != nil {
		t.Fatal(err)
	}
	if s.tree.Size() != 4 {
		t.Fatalf("size=%d", s.tree.Size())
	}
	root := s.tree.Root()
	if err := s.applyLogs([]types.Log{a, b}); err != nil {
		t.Fatal(err)
	}
	if s.tree.Root() != root || s.tree.Size() != 4 {
		t.Fatal("overlap changed tree")
	}
	proof, err := s.tree.GetProof(3)
	if err != nil {
		t.Fatal(err)
	}
	if valid, err := proof.Verify(); err != nil || !valid {
		t.Fatalf("proof invalid %v %v", valid, err)
	}
	query := s.buildFilterQuery(0, 10)
	if len(query.Topics) != 2 || query.Topics[1][0] != common.BigToHash(big.NewInt(1)) {
		t.Fatal("wrong canonical topic filter")
	}
}
func TestInsertionFailuresAreAtomic(t *testing.T) {
	for _, scenario := range []string{"gap", "conflict", "same batch conflict", "duplicate commitment", "invalid field", "zero", "capacity", "reorg", "malformed"} {
		t.Run(scenario, func(t *testing.T) {
			s := testSync(t)
			if err := s.applyLogs([]types.Log{insertion(t, "CommitmentsInserted", 1, 0, 11)}); err != nil {
				t.Fatal(err)
			}
			root := s.tree.Root()
			logs := []types.Log{insertion(t, "CommitmentsInserted", 1, 1, 12)}
			switch scenario {
			case "gap":
				logs = append(logs, insertion(t, "BatchOutputs", 1, 3, 14))
			case "conflict":
				logs = append(logs, insertion(t, "CommitmentsInserted", 1, 0, 15))
			case "same batch conflict":
				logs = append(logs, insertion(t, "BatchOutputs", 1, 1, 15))
			case "duplicate commitment":
				logs = append(logs, insertion(t, "CommitmentsInserted", 1, 2, 11))
			case "invalid field":
				bad := insertion(t, "CommitmentsInserted", 1, 2, 13)
				for i := 96; i < 128; i++ {
					bad.Data[i] = 255
				}
				logs = append(logs, bad)
			case "zero":
				logs = append(logs, insertion(t, "CommitmentsInserted", 1, 2, 0))
			case "capacity":
				logs = append(logs, insertion(t, "CommitmentsInserted", 1, MaxLeaves, 13))
			case "reorg":
				bad := insertion(t, "CommitmentsInserted", 1, 0, 11)
				bad.Removed = true
				logs = append(logs, bad)
			case "malformed":
				bad := insertion(t, "CommitmentsInserted", 1, 2, 13)
				bad.Data = []byte{1}
				logs = append(logs, bad)
			}
			if err := s.applyLogs(logs); err == nil {
				t.Fatal("expected error")
			}
			if s.tree.Size() != 1 || s.tree.Root() != root || len(s.ledger.leaves) != 1 {
				t.Fatal("failed batch partially mutated state")
			}
		})
	}
}
func TestGlobalCommitmentUniqueness(t *testing.T) {
	one := testSync(t)
	two := testSync(t)
	two.treeID = big.NewInt(2)
	two.ledger = one.ledger
	if err := one.applyLogs([]types.Log{insertion(t, "CommitmentsInserted", 1, 0, 11)}); err != nil {
		t.Fatal(err)
	}
	if err := two.applyLogs([]types.Log{insertion(t, "BatchOutputs", 2, 0, 12, 11)}); err == nil {
		t.Fatal("accepted cross-tree duplicate")
	}
	if two.tree.Size() != 0 || len(one.ledger.leaves) != 1 {
		t.Fatal("duplicate changed global state")
	}
}

type recordedSubscription struct {
	query ethereum.FilterQuery
	ch    chan<- types.Log
}
type currentClient struct {
	mu                  sync.Mutex
	logs                []types.Log
	active              int64
	latest              uint64
	subscriptions       []recordedSubscription
	activeSubscriptions int
	subscribeErr        error
	historyHook         func()
}

func (c *currentClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	c.mu.Lock()
	hook := c.historyHook
	c.historyHook = nil
	logs := append([]types.Log(nil), c.logs...)
	c.mu.Unlock()
	if hook != nil {
		hook()
	}
	var result []types.Log
	for _, log := range logs {
		if matches(log, q) {
			result = append(result, log)
		}
	}
	return result, nil
}
func (c *currentClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.subscribeErr != nil {
		return nil, c.subscribeErr
	}
	c.subscriptions = append(c.subscriptions, recordedSubscription{q, ch})
	c.activeSubscriptions++
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer func() { c.mu.Lock(); c.activeSubscriptions--; c.mu.Unlock() }()
		select {
		case <-ctx.Done():
			return nil
		case <-quit:
			return nil
		}
	}), nil
}
func (c *currentClient) BlockNumber(context.Context) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.latest, nil
}
func (c *currentClient) CallContract(_ context.Context, msg ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	if string(msg.Data) != string(protocolABI.Methods["activeNoteTreeId"].ID) {
		return nil, errors.New("must query activeNoteTreeId")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return protocolABI.Methods["activeNoteTreeId"].Outputs.Pack(big.NewInt(c.active))
}
func (c *currentClient) send(log types.Log) {
	c.mu.Lock()
	subscriptions := append([]recordedSubscription(nil), c.subscriptions...)
	c.mu.Unlock()
	for _, sub := range subscriptions {
		if matches(log, sub.query) {
			sub.ch <- log
		}
	}
}
func matches(log types.Log, q ethereum.FilterQuery) bool {
	if q.FromBlock != nil && log.BlockNumber < q.FromBlock.Uint64() {
		return false
	}
	if q.ToBlock != nil && log.BlockNumber > q.ToBlock.Uint64() {
		return false
	}
	if len(q.Addresses) > 0 && log.Address != q.Addresses[0] {
		return false
	}
	for i, alternatives := range q.Topics {
		if len(alternatives) == 0 {
			continue
		}
		if len(log.Topics) <= i {
			return false
		}
		found := false
		for _, topic := range alternatives {
			if log.Topics[i] == topic {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func await(t *testing.T, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out")
}
func TestSyncSubscribesBeforeHistoryAndReportsLiveFailure(t *testing.T) {
	s := testSync(t)
	c := &currentClient{latest: 10}
	s.client = c
	c.historyHook = func() { c.send(insertion(t, "CommitmentsInserted", 1, 0, 11)) }
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	await(t, func() bool { return s.tree.Size() == 1 })
	c.send(insertion(t, "BatchOutputs", 1, 2, 13))
	await(t, func() bool { return s.Error() != nil })
	if s.Synced() || s.tree.Size() != 1 {
		t.Fatal("failed live sync remained valid or mutated")
	}
}
func TestManagerActiveDiscoveryRolloverAndHistoricalProofs(t *testing.T) {
	c := &currentClient{active: 1, latest: 10, logs: []types.Log{insertion(t, "CommitmentsInserted", 1, 0, 11)}}
	m := NewInvisiTreeManager(InvisiTreeManagerOptions{Client: c, InstanceAddress: testAddress, PollInterval: 5 * time.Millisecond})
	if err := m.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer m.Stop()
	before, err := m.GetNoteProof(big.NewInt(1), 0)
	if err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	c.active = 2
	c.logs = append(c.logs, insertion(t, "BatchOutputs", 2, 0, 22))
	c.mu.Unlock()
	c.send(types.Log{Address: testAddress, Topics: []common.Hash{treeCreatedEventSig, common.BigToHash(big.NewInt(2))}, BlockNumber: 10})
	await(t, func() bool { return m.TreeCount() == 2 })
	if _, err := m.GetNoteProof(big.NewInt(2), 0); err != nil {
		t.Fatal(err)
	}
	after, err := m.GetNoteProof(big.NewInt(1), 0)
	if err != nil || after.Root != before.Root {
		t.Fatal("historical proof lost during rollover")
	}
	if _, err = m.GetTree(big.NewInt(0)); err == nil {
		t.Fatal("accepted reserved tree zero")
	}
	removed := types.Log{Address: testAddress, Topics: []common.Hash{treeCreatedEventSig, common.BigToHash(big.NewInt(2))}, Removed: true, BlockNumber: 10}
	c.send(removed)
	await(t, func() bool { return m.Error() != nil })
	if _, err = m.GetTree(big.NewInt(1)); err == nil {
		t.Fatal("manager served proofs after reorg")
	}
}
func TestManagerPollingDiscoversNewTree(t *testing.T) {
	c := &currentClient{active: 1, latest: 10, subscribeErr: errors.New("websocket unavailable")}
	m := NewInvisiTreeManager(InvisiTreeManagerOptions{Client: c, InstanceAddress: testAddress, PollInterval: 5 * time.Millisecond})
	if err := m.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer m.Stop()
	c.mu.Lock()
	c.active = 2
	c.mu.Unlock()
	await(t, func() bool { return m.TreeCount() == 2 })
}
func TestSyncRejectsReservedTree(t *testing.T) {
	s := testSync(t)
	s.treeID = big.NewInt(0)
	s.client = &currentClient{}
	if err := s.Start(context.Background()); err == nil {
		t.Fatal("accepted reserved tree zero")
	}
}

func TestManagerRestartRescansExistingTrees(t *testing.T) {
	c := &currentClient{active: 1, latest: 10, logs: []types.Log{insertion(t, "CommitmentsInserted", 1, 0, 11)}}
	m := NewInvisiTreeManager(InvisiTreeManagerOptions{Client: c, InstanceAddress: testAddress})
	if err := m.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	m.Stop()
	c.mu.Lock()
	c.logs = append(c.logs, insertion(t, "BatchOutputs", 1, 1, 12))
	c.mu.Unlock()
	if err := m.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer m.Stop()
	tree, err := m.GetTree(big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	if tree.Size() != 2 {
		t.Fatal("restart failed to rescan existing tree")
	}
}

func TestCanonicalReplayRejectsOfflineTailRemoval(t *testing.T) {
	for _, mode := range []string{"restart", "poll"} {
		for _, change := range []string{"missing tail", "replacement tail", "missing all"} {
			t.Run(mode+"/"+change, func(t *testing.T) {
				first := insertion(t, "CommitmentsInserted", 1, 0, 11)
				tail := insertion(t, "BatchOutputs", 1, 1, 12)
				c := &currentClient{latest: 10, logs: []types.Log{first, tail}}
				if mode == "poll" {
					c.subscribeErr = errors.New("no websocket")
				}
				s := testSync(t)
				s.client = c
				s.pollInterval = 5 * time.Millisecond
				if err := s.Start(context.Background()); err != nil {
					t.Fatal(err)
				}
				defer s.Stop()
				root := s.tree.Root()
				if mode == "restart" {
					s.Stop()
				}
				c.mu.Lock()
				c.latest = 11
				switch change {
				case "missing tail":
					c.logs = []types.Log{first}
				case "replacement tail":
					c.logs = []types.Log{first, insertion(t, "BatchOutputs", 1, 1, 99, 100)}
				case "missing all":
					c.logs = nil
				}
				c.mu.Unlock()
				if mode == "restart" {
					if err := s.Start(context.Background()); err == nil {
						t.Fatal("restart accepted changed canonical history")
					}
				} else {
					await(t, func() bool { return s.Error() != nil })
				}
				if s.Error() == nil || s.Synced() {
					t.Fatal("reorg did not invalidate sync")
				}
				if s.tree.Size() != 2 || s.tree.Root() != root || len(s.ledger.leaves) != 2 || s.LastSyncedBlock() != 10 {
					t.Fatal("rejected canonical snapshot mutated tree, ledger, or checkpoint")
				}
			})
		}
	}
}

func TestManagerRestartsAfterContextCancellationWithoutLeakingSubscriptions(t *testing.T) {
	c := &currentClient{active: 1, latest: 10, logs: []types.Log{insertion(t, "CommitmentsInserted", 1, 0, 11)}}
	m := NewInvisiTreeManager(InvisiTreeManagerOptions{Client: c, InstanceAddress: testAddress})
	ctx, cancel := context.WithCancel(context.Background())
	if err := m.Start(ctx); err != nil {
		t.Fatal(err)
	}
	cancel()
	c.mu.Lock()
	c.active = 2
	c.logs = append(c.logs, insertion(t, "BatchOutputs", 2, 0, 22))
	c.mu.Unlock()
	// No explicit Stop or waiting: Start must retire the cancelled run itself.
	if err := m.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer m.Stop()
	if m.TreeCount() != 2 {
		t.Fatalf("fresh context did not rediscover active tree: %d", m.TreeCount())
	}
	if _, err := m.GetNoteProof(big.NewInt(2), 0); err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	active := c.activeSubscriptions
	c.mu.Unlock()
	if active != 3 {
		t.Fatalf("expected manager plus two live tree subscriptions, got %d", active)
	}
	m.Stop()
	c.mu.Lock()
	active = c.activeSubscriptions
	c.mu.Unlock()
	if active != 0 {
		t.Fatalf("Stop leaked %d subscriptions", active)
	}
	if m.Error() != nil {
		t.Fatalf("normal cancellation became terminal: %v", m.Error())
	}
}

func TestPollingCanonicalReplayAppendsValidNewLeaves(t *testing.T) {
	first := insertion(t, "CommitmentsInserted", 1, 0, 11)
	c := &currentClient{latest: 10, logs: []types.Log{first}, subscribeErr: errors.New("no websocket")}
	s := testSync(t)
	s.client = c
	s.pollInterval = 5 * time.Millisecond
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	next := insertion(t, "BatchOutputs", 1, 1, 12, 13)
	next.BlockNumber = 11
	c.mu.Lock()
	c.latest = 11
	c.logs = append(c.logs, next)
	c.mu.Unlock()
	await(t, func() bool { return s.tree.Size() == 3 })
	if s.Error() != nil || !s.Synced() {
		t.Fatalf("valid canonical history rejected: %v", s.Error())
	}
}
