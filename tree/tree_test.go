package tree

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func TestNewInMemoryMerkleTree(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}
	if tree.Depth() != 16 {
		t.Errorf("Depth() = %d, want 16", tree.Depth())
	}
	if tree.Size() != 0 {
		t.Errorf("Size() = %d, want 0", tree.Size())
	}
}

func TestEmptyTreeRootIsNonZero(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}
	root := tree.Root()
	var zero fr.Element
	if root.Equal(&zero) {
		t.Error("empty tree root should not be zero (it is Poseidon2(zero, zero) recursively)")
	}
}

func TestZeroHashesConsistency(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}
	// zeros[0] must be the zero element
	var zero fr.Element
	if !tree.zeros[0].Equal(&zero) {
		t.Error("zeros[0] should be the zero field element")
	}
	// Empty tree root must equal zeros[Depth]
	emptyRoot := tree.zeros[Depth]
	root := tree.Root()
	if !root.Equal(&emptyRoot) {
		t.Error("empty tree Root() should equal zeros[Depth]")
	}
}

func TestInsertReturnsSequentialIndices(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var leaf1, leaf2, leaf3 fr.Element
	leaf1.SetUint64(1)
	leaf2.SetUint64(2)
	leaf3.SetUint64(3)

	i0, err := tree.Insert(leaf1)
	if err != nil {
		t.Fatal(err)
	}
	i1, err := tree.Insert(leaf2)
	if err != nil {
		t.Fatal(err)
	}
	i2, err := tree.Insert(leaf3)
	if err != nil {
		t.Fatal(err)
	}

	if i0 != 0 || i1 != 1 || i2 != 2 {
		t.Errorf("indices = (%d, %d, %d), want (0, 1, 2)", i0, i1, i2)
	}
	if tree.Size() != 3 {
		t.Errorf("Size() = %d, want 3", tree.Size())
	}
}

func TestRootChangesAfterInsert(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	rootBefore := tree.Root()

	var leaf fr.Element
	leaf.SetUint64(42)
	_, err = tree.Insert(leaf)
	if err != nil {
		t.Fatal(err)
	}

	rootAfter := tree.Root()
	if rootBefore.Equal(&rootAfter) {
		t.Error("root should change after insert")
	}
}

func TestDeterministicRoots(t *testing.T) {
	t1, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}
	t2, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var leaf fr.Element
	leaf.SetUint64(100)

	_, err = t1.Insert(leaf)
	if err != nil {
		t.Fatal(err)
	}
	_, err = t2.Insert(leaf)
	if err != nil {
		t.Fatal(err)
	}

	r1 := t1.Root()
	r2 := t2.Root()
	if !r1.Equal(&r2) {
		t.Error("same inserts should produce same root")
	}
}

func TestInsertFullTree(t *testing.T) {
	// We won't insert 65536 leaves (too slow), but verify the error at capacity.
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}
	// Manually set size to max to test overflow
	tree.size = MaxLeaves
	var leaf fr.Element
	leaf.SetUint64(999)
	_, err = tree.Insert(leaf)
	if err == nil {
		t.Error("expected error when tree is full")
	}
}

func TestMultipleInsertsProduceDifferentRoots(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var leaf1, leaf2 fr.Element
	leaf1.SetUint64(111)
	leaf2.SetUint64(222)

	_, err = tree.Insert(leaf1)
	if err != nil {
		t.Fatal(err)
	}
	rootAfterOne := tree.Root()

	_, err = tree.Insert(leaf2)
	if err != nil {
		t.Fatal(err)
	}
	rootAfterTwo := tree.Root()

	if rootAfterOne.Equal(&rootAfterTwo) {
		t.Error("root should differ after second insert")
	}
}

// --- Task 8: Integration tests ---

func TestIntegration_InsertAndProveAllLeaves(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	// Insert 16 leaves
	for i := uint64(0); i < 16; i++ {
		var leaf fr.Element
		leaf.SetUint64(i * 100)
		idx, err := tree.Insert(leaf)
		if err != nil {
			t.Fatalf("Insert(%d): %v", i, err)
		}
		if idx != int(i) {
			t.Fatalf("Insert(%d) = %d, want %d", i, idx, i)
		}
	}

	if tree.Size() != 16 {
		t.Fatalf("Size() = %d, want 16", tree.Size())
	}

	// Prove and verify every leaf
	for i := 0; i < 16; i++ {
		proof, err := tree.GetProof(i)
		if err != nil {
			t.Fatalf("GetProof(%d): %v", i, err)
		}
		valid, err := proof.Verify()
		if err != nil {
			t.Fatalf("Verify(%d): %v", i, err)
		}
		if !valid {
			t.Errorf("proof for leaf %d does not verify", i)
		}
	}
}

func TestIntegration_RootStability(t *testing.T) {
	// Two trees with identical inserts must produce identical roots.
	t1, _ := NewInMemoryMerkleTree()
	t2, _ := NewInMemoryMerkleTree()

	for i := uint64(0); i < 32; i++ {
		var leaf fr.Element
		leaf.SetUint64(i + 1)
		t1.Insert(leaf)
		t2.Insert(leaf)
	}

	r1 := t1.Root()
	r2 := t2.Root()
	if !r1.Equal(&r2) {
		t.Error("identical trees should have identical roots")
	}
}

func TestIntegration_ProofRootMatchesTreeRoot(t *testing.T) {
	tree, _ := NewInMemoryMerkleTree()

	var l1, l2 fr.Element
	l1.SetUint64(1)
	l2.SetUint64(2)

	tree.Insert(l1)
	tree.Insert(l2)

	proof0, _ := tree.GetProof(0)
	proof1, _ := tree.GetProof(1)

	root := tree.Root()
	if !proof0.Root.Equal(&root) {
		t.Error("proof0.Root != tree.Root()")
	}
	if !proof1.Root.Equal(&root) {
		t.Error("proof1.Root != tree.Root()")
	}
}

func TestIntegration_LargeTree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large tree test in short mode")
	}

	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	// Insert 256 leaves
	for i := uint64(0); i < 256; i++ {
		var leaf fr.Element
		leaf.SetUint64(i)
		_, err := tree.Insert(leaf)
		if err != nil {
			t.Fatalf("Insert(%d): %v", i, err)
		}
	}

	if tree.Size() != 256 {
		t.Fatalf("Size() = %d, want 256", tree.Size())
	}

	// Spot-check proofs at various indices
	for _, idx := range []int{0, 1, 127, 128, 255} {
		proof, err := tree.GetProof(idx)
		if err != nil {
			t.Fatalf("GetProof(%d): %v", idx, err)
		}
		valid, err := proof.Verify()
		if err != nil {
			t.Fatalf("Verify(%d): %v", idx, err)
		}
		if !valid {
			t.Errorf("proof for leaf %d does not verify", idx)
		}
	}
}
