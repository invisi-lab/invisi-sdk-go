package tree

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// setupMockClientWithTreeCount configures the mock client to return a tree count.
func setupMockClientWithTreeCount(client *mockClient, count int) {
	const getTreeCountABI = `[{"type":"function","name":"getTreeCount","inputs":[],"outputs":[{"name":"count","type":"uint256"}]}]`
	parsed, _ := abi.JSON(strings.NewReader(getTreeCountABI))
	calldata, _ := parsed.Pack("getTreeCount")

	// Key by first 4 bytes (method selector)
	key := common.Bytes2Hex(calldata[:4])

	// ABI-encode the return value
	result, _ := parsed.Methods["getTreeCount"].Outputs.Pack(big.NewInt(int64(count)))
	client.callResults[key] = result
}

func TestManagerStartsWithZeroTrees(t *testing.T) {
	client := newMockClient()
	setupMockClientWithTreeCount(client, 0)

	mgr := NewInvisiTreeManager(InvisiTreeManagerOptions{
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Stop()

	if mgr.TreeCount() != 0 {
		t.Errorf("TreeCount() = %d, want 0", mgr.TreeCount())
	}
}

func TestManagerDiscoversTrees(t *testing.T) {
	client := newMockClient()
	setupMockClientWithTreeCount(client, 2)

	// Add events for tree 1 and tree 2
	c1 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	c2 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002")
	c3 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003")

	client.logs = append(client.logs,
		buildPayOpenLog(testAddr, big.NewInt(1), c1, 0),
		buildPayOpenLog(testAddr, big.NewInt(1), c2, 1),
		buildPayOpenLog(testAddr, big.NewInt(2), c3, 0),
	)

	mgr := NewInvisiTreeManager(InvisiTreeManagerOptions{
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Stop()

	if mgr.TreeCount() != 2 {
		t.Errorf("TreeCount() = %d, want 2", mgr.TreeCount())
	}

	tree1, err := mgr.GetTree(big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	if tree1.Size() != 2 {
		t.Errorf("tree1.Size() = %d, want 2", tree1.Size())
	}

	tree2, err := mgr.GetTree(big.NewInt(2))
	if err != nil {
		t.Fatal(err)
	}
	if tree2.Size() != 1 {
		t.Errorf("tree2.Size() = %d, want 1", tree2.Size())
	}
}

func TestManagerGetPaymentProof(t *testing.T) {
	client := newMockClient()
	setupMockClientWithTreeCount(client, 1)

	c := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000042")
	client.logs = append(client.logs,
		buildPayOpenLog(testAddr, big.NewInt(1), c, 0),
	)

	mgr := NewInvisiTreeManager(InvisiTreeManagerOptions{
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Stop()

	proof, err := mgr.GetPaymentProof(big.NewInt(1), 0)
	if err != nil {
		t.Fatal(err)
	}
	if proof.Index != 0 {
		t.Errorf("proof.Index = %d, want 0", proof.Index)
	}

	// Verify the proof
	valid, err := proof.Verify()
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("proof should verify")
	}
}

func TestManagerGetTreeNotFound(t *testing.T) {
	client := newMockClient()
	setupMockClientWithTreeCount(client, 0)

	mgr := NewInvisiTreeManager(InvisiTreeManagerOptions{
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Stop()

	_, err = mgr.GetTree(big.NewInt(99))
	if err == nil {
		t.Error("expected error for non-existent tree")
	}
}

func TestManagerGetPaymentProofNotFound(t *testing.T) {
	client := newMockClient()
	setupMockClientWithTreeCount(client, 0)

	mgr := NewInvisiTreeManager(InvisiTreeManagerOptions{
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Stop()

	_, err = mgr.GetPaymentProof(big.NewInt(99), 0)
	if err == nil {
		t.Error("expected error for non-existent tree")
	}
}

func TestManagerStopIsIdempotent(t *testing.T) {
	client := newMockClient()
	setupMockClientWithTreeCount(client, 1)

	c := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	client.logs = append(client.logs,
		buildPayOpenLog(testAddr, big.NewInt(1), c, 0),
	)

	mgr := NewInvisiTreeManager(InvisiTreeManagerOptions{
		Client:          client,
		InstanceAddress: testAddr,
		FromBlock:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Multiple stops should not panic
	mgr.Stop()
	mgr.Stop()
}
