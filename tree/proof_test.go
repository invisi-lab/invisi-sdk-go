package tree

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func TestGetProofReturns16Siblings(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var l1, l2, l3 fr.Element
	l1.SetUint64(111)
	l2.SetUint64(222)
	l3.SetUint64(333)

	tree.Insert(l1)
	tree.Insert(l2)
	tree.Insert(l3)

	proof, err := tree.GetProof(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(proof.Siblings) != Depth {
		t.Errorf("siblings length = %d, want %d", len(proof.Siblings), Depth)
	}
	if proof.Index != 1 {
		t.Errorf("proof.Index = %d, want 1", proof.Index)
	}

	root := tree.Root()
	if !proof.Root.Equal(&root) {
		t.Error("proof.Root should match tree.Root()")
	}
}

func TestProofVerification_Index0(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var leaf fr.Element
	leaf.SetUint64(42)
	tree.Insert(leaf)

	proof, err := tree.GetProof(0)
	if err != nil {
		t.Fatal(err)
	}

	valid, err := proof.Verify()
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("proof for index 0 should verify")
	}
}

func TestProofVerification_Index2(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var l1, l2, l3 fr.Element
	l1.SetUint64(100)
	l2.SetUint64(200)
	l3.SetUint64(300)

	tree.Insert(l1) // index 0
	tree.Insert(l2) // index 1
	tree.Insert(l3) // index 2

	proof, err := tree.GetProof(2)
	if err != nil {
		t.Fatal(err)
	}

	valid, err := proof.Verify()
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("proof for index 2 should verify")
	}
}

func TestProofVerification_AllIndices(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	// Insert 8 leaves
	for i := uint64(0); i < 8; i++ {
		var leaf fr.Element
		leaf.SetUint64((i + 1) * 1000)
		_, err := tree.Insert(leaf)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify proof for every leaf
	for i := 0; i < 8; i++ {
		proof, err := tree.GetProof(i)
		if err != nil {
			t.Fatalf("GetProof(%d): %v", i, err)
		}
		valid, err := proof.Verify()
		if err != nil {
			t.Fatalf("Verify proof for index %d: %v", i, err)
		}
		if !valid {
			t.Errorf("proof for index %d should verify", i)
		}
	}
}

func TestProofVerification_TamperedLeafFails(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var leaf fr.Element
	leaf.SetUint64(42)
	tree.Insert(leaf)

	proof, err := tree.GetProof(0)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with the leaf
	proof.Leaf.SetUint64(999)

	valid, err := proof.Verify()
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("tampered proof should NOT verify")
	}
}

func TestProofVerification_TamperedSiblingFails(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var leaf fr.Element
	leaf.SetUint64(42)
	tree.Insert(leaf)

	proof, err := tree.GetProof(0)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with a sibling
	proof.Siblings[0].SetUint64(999)

	valid, err := proof.Verify()
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("tampered sibling should cause verification failure")
	}
}

func TestGetProofOutOfRange(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	var leaf fr.Element
	leaf.SetUint64(1)
	tree.Insert(leaf)

	_, err = tree.GetProof(5)
	if err == nil {
		t.Error("expected error for out of range leafIndex")
	}

	_, err = tree.GetProof(-1)
	if err == nil {
		t.Error("expected error for negative leafIndex")
	}
}

func TestGetProofEmptyTree(t *testing.T) {
	tree, err := NewInMemoryMerkleTree()
	if err != nil {
		t.Fatal(err)
	}

	_, err = tree.GetProof(0)
	if err == nil {
		t.Error("expected error for empty tree")
	}
}
