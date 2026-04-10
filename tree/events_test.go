package tree

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestParsePayOpenEvent(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	commitment := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	log := buildPayOpenLog(addr, big.NewInt(1), commitment, 0)

	leaves, err := parseEventsToLeaves([]types.Log{log})
	if err != nil {
		t.Fatal(err)
	}
	if len(leaves) != 1 {
		t.Fatalf("expected 1 leaf, got %d", len(leaves))
	}
	if leaves[0].leafIndex != 0 {
		t.Errorf("leafIndex = %d, want 0", leaves[0].leafIndex)
	}
	if leaves[0].commitment != commitment {
		t.Errorf("commitment mismatch")
	}
}

func TestParsePayBlindEvent(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	pc := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	cc := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ff")

	log := buildPayBlindLog(addr, big.NewInt(1), pc, cc, 0, 1)

	leaves, err := parseEventsToLeaves([]types.Log{log})
	if err != nil {
		t.Fatal(err)
	}
	if len(leaves) != 2 {
		t.Fatalf("expected 2 leaves, got %d", len(leaves))
	}
	// Should be sorted by leafIndex
	if leaves[0].leafIndex != 0 || leaves[1].leafIndex != 1 {
		t.Errorf("leaf indices = (%d, %d), want (0, 1)", leaves[0].leafIndex, leaves[1].leafIndex)
	}
}

func TestParseCreditDepositEvent(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	commitment := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000005")
	log := buildCreditDepositLog(addr, big.NewInt(1), commitment, 3)

	leaves, err := parseEventsToLeaves([]types.Log{log})
	if err != nil {
		t.Fatal(err)
	}
	if len(leaves) != 1 {
		t.Fatalf("expected 1 leaf, got %d", len(leaves))
	}
	if leaves[0].leafIndex != 3 {
		t.Errorf("leafIndex = %d, want 3", leaves[0].leafIndex)
	}
}

func TestParseCreditDrawEvent(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	rc1 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000011")
	rc2 := [32]byte{} // zero — should be skipped
	rc3 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000033")

	log := buildCreditDrawLog(
		addr, big.NewInt(1),
		[][32]byte{rc1, rc2, rc3},
		[]uint32{4, 5, 6},
	)

	leaves, err := parseEventsToLeaves([]types.Log{log})
	if err != nil {
		t.Fatal(err)
	}
	// rc2 is zero, so only 2 leaves
	if len(leaves) != 2 {
		t.Fatalf("expected 2 leaves (skipping zero), got %d", len(leaves))
	}
	if leaves[0].leafIndex != 4 {
		t.Errorf("first leaf index = %d, want 4", leaves[0].leafIndex)
	}
	if leaves[1].leafIndex != 6 {
		t.Errorf("second leaf index = %d, want 6", leaves[1].leafIndex)
	}
}

func TestParseEventsAreSortedByLeafIndex(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	c1 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	c2 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002")

	// Insert in reverse order
	log1 := buildPayOpenLog(addr, big.NewInt(1), c2, 1)
	log2 := buildPayOpenLog(addr, big.NewInt(1), c1, 0)

	leaves, err := parseEventsToLeaves([]types.Log{log1, log2})
	if err != nil {
		t.Fatal(err)
	}
	if leaves[0].leafIndex != 0 || leaves[1].leafIndex != 1 {
		t.Errorf("leaves not sorted: (%d, %d)", leaves[0].leafIndex, leaves[1].leafIndex)
	}
}

func TestParseMixedEvents(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	treeID := big.NewInt(1)

	c1 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	c2 := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002")
	c3 := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000aa")
	c4 := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000bb")

	// PayOpen at index 0,1, then PayBlind at 2,3
	logs := []types.Log{
		buildPayOpenLog(addr, treeID, c1, 0),
		buildPayOpenLog(addr, treeID, c2, 1),
		buildPayBlindLog(addr, treeID, c3, c4, 2, 3),
	}

	leaves, err := parseEventsToLeaves(logs)
	if err != nil {
		t.Fatal(err)
	}
	if len(leaves) != 4 {
		t.Fatalf("expected 4 leaves, got %d", len(leaves))
	}
	for i := 0; i < 4; i++ {
		if leaves[i].leafIndex != i {
			t.Errorf("leaf[%d].leafIndex = %d, want %d", i, leaves[i].leafIndex, i)
		}
	}
}
