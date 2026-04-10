package tree

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

var testAddr = common.HexToAddress("0x0000000000000000000000000000000000000001")

func TestSyncStartsWithEmptyTree(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	client := newMockClient()
	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Stop()

	if tree.Size() != 0 {
		t.Errorf("Size() = %d, want 0", tree.Size())
	}
	if !sync.Synced() {
		t.Error("sync should be marked as synced")
	}
}

func TestSyncInsertsHistoricalPayOpenEvents(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	c1 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	c2 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002")

	client := newMockClient()
	client.logs = append(client.logs,
		buildPayOpenLog(testAddr, big.NewInt(1), c1, 0),
		buildPayOpenLog(testAddr, big.NewInt(1), c2, 1),
	)

	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Stop()

	if tree.Size() != 2 {
		t.Errorf("Size() = %d, want 2", tree.Size())
	}
}

func TestSyncInsertsPayBlindEvents(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	pc := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	cc := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ff")

	client := newMockClient()
	client.logs = append(client.logs,
		buildPayBlindLog(testAddr, big.NewInt(1), pc, cc, 0, 1),
	)

	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Stop()

	if tree.Size() != 2 {
		t.Errorf("Size() = %d, want 2 (payment + change)", tree.Size())
	}
}

func TestSyncFiltersByTreeID(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	c1 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	c2 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002")
	c3 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003")

	client := newMockClient()
	client.logs = append(client.logs,
		buildPayOpenLog(testAddr, big.NewInt(1), c1, 0), // treeID=1
		buildPayOpenLog(testAddr, big.NewInt(2), c2, 0), // treeID=2 — should be filtered out
		buildPayOpenLog(testAddr, big.NewInt(1), c3, 1), // treeID=1
	)

	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Stop()

	if tree.Size() != 2 {
		t.Errorf("Size() = %d, want 2 (only treeID=1 events)", tree.Size())
	}
}

func TestSyncMixedEvents(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	c1 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	c2 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002")
	c3 := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000aa")
	c4 := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000bb")

	client := newMockClient()
	treeID := big.NewInt(1)
	client.logs = append(client.logs,
		buildPayOpenLog(testAddr, treeID, c1, 0),
		buildPayOpenLog(testAddr, treeID, c2, 1),
		buildPayBlindLog(testAddr, treeID, c3, c4, 2, 3),
	)

	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Stop()

	if tree.Size() != 4 {
		t.Errorf("Size() = %d, want 4", tree.Size())
	}
}

func TestSyncMismatchError(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	// Gap: leafIndex jumps from 0 to 2, skipping 1
	c1 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	c2 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003")

	client := newMockClient()
	client.logs = append(client.logs,
		buildPayOpenLog(testAddr, big.NewInt(1), c1, 0),
		buildPayOpenLog(testAddr, big.NewInt(1), c2, 2), // gap: missing index 1
	)

	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err == nil {
		t.Error("expected mismatch error for non-contiguous leaf indices")
	}
}

func TestSyncLastSyncedBlock(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	client := newMockClient()
	client.blockNumber = 42

	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Stop()

	if sync.LastSyncedBlock() != 42 {
		t.Errorf("LastSyncedBlock() = %d, want 42", sync.LastSyncedBlock())
	}
}

func TestSyncStopIsIdempotent(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	client := newMockClient()
	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Stop multiple times should not panic
	sync.Stop()
	sync.Stop()
}

// --- Task 6: Real-time subscription + polling fallback tests ---

func TestSyncReceivesRealTimeEvents(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	client := newMockClient()
	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Stop()

	// Wait briefly for subscription goroutine to be ready
	time.Sleep(50 * time.Millisecond)

	// Simulate a real-time PayOpen event
	c := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000042")
	log := buildPayOpenLog(testAddr, big.NewInt(1), c, 0)
	log.BlockNumber = 101
	client.sendLog(log)

	// Wait until LastSyncedBlock reaches 101 — it's updated atomically after applyLogs completes.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sync.LastSyncedBlock() >= 101 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if sync.LastSyncedBlock() != 101 {
		t.Errorf("LastSyncedBlock() = %d, want 101", sync.LastSyncedBlock())
	}

	// Acquire sync's lock to safely read tree state after the event is fully processed.
	sync.mu.Lock()
	size := tree.Size()
	sync.mu.Unlock()

	if size != 1 {
		t.Errorf("Size() = %d, want 1 after real-time event", size)
	}
}

func TestSyncFallsBackToPolling(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	client := newMockClient()
	// Make subscription fail — triggers polling fallback
	client.subscribeErr = fmt.Errorf("WebSocket not available")

	c := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")

	// No initial logs
	sync := NewMerkleTreeSync(MerkleTreeSyncOptions{
		Tree:            tree,
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
		TreeID:          big.NewInt(1),
		PollInterval:    50 * time.Millisecond, // fast polling for test
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sync.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sync.Stop()

	// Now add a log that will be picked up by polling
	client.mu.Lock()
	client.logs = append(client.logs,
		buildPayOpenLog(testAddr, big.NewInt(1), c, 0),
	)
	client.blockNumber = 150
	client.mu.Unlock()

	// Wait until the sync has processed block 150, indicating the poll cycle completed.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sync.LastSyncedBlock() >= 150 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Acquire the sync's lock to safely read tree state after polling is done.
	sync.mu.Lock()
	size := tree.Size()
	sync.mu.Unlock()

	if size != 1 {
		t.Errorf("Size() = %d, want 1 after polling", size)
	}
}
