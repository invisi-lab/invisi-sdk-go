package tree

import (
	"fmt"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

// MaxLeaves is the maximum number of leaves (2^Depth).
const MaxLeaves = 1 << Depth // 65536

// InMemoryMerkleTree is a depth-16 Poseidon2 Merkle tree.
// It stores all layers for arbitrary proof generation.
type InMemoryMerkleTree struct {
	mu sync.RWMutex
	// zeros[i] is the default hash at level i for an empty subtree.
	zeros [Depth + 1]fr.Element
	// layers[0] = leaves, layers[Depth] = [root].
	// Each layer is a slice that grows as leaves are inserted.
	layers [Depth + 1][]fr.Element
	// size is the number of inserted leaves.
	size int
}

// hash2 hashes two sibling field elements using Poseidon2.
// This is the tree's internal hash: Poseidon2Hash(left, right) with no domain tag.
func hash2(a, b fr.Element) (fr.Element, error) {
	return core.Poseidon2Hash(a, b)
}

// NewInMemoryMerkleTree creates a new empty Merkle tree with precomputed zero hashes.
func NewInMemoryMerkleTree() (*InMemoryMerkleTree, error) {
	t := &InMemoryMerkleTree{}

	// Precompute zero hashes: zeros[0] = 0, zeros[i] = hash(zeros[i-1], zeros[i-1])
	// zeros[0] is the default fr.Element zero value
	for i := 1; i <= Depth; i++ {
		h, err := hash2(t.zeros[i-1], t.zeros[i-1])
		if err != nil {
			return nil, fmt.Errorf("precomputing zero hash level %d: %w", i, err)
		}
		t.zeros[i] = h
	}

	// Initialize empty layers
	for i := 0; i <= Depth; i++ {
		t.layers[i] = []fr.Element{}
	}

	return t, nil
}

// Depth returns the tree depth (16).
func (t *InMemoryMerkleTree) Depth() int {
	return Depth
}

// Size returns the number of inserted leaves.
func (t *InMemoryMerkleTree) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

// Root returns the current Merkle root.
// For an empty tree, returns zeros[Depth].
func (t *InMemoryMerkleTree) Root() fr.Element {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.rootLocked()
}

func (t *InMemoryMerkleTree) rootLocked() fr.Element {
	topLayer := t.layers[Depth]
	if len(topLayer) == 0 {
		return t.zeros[Depth]
	}
	return topLayer[0]
}

// Insert adds a leaf to the tree and returns its index.
// Returns an error if the tree is full (>= 2^16 leaves).
func (t *InMemoryMerkleTree) Insert(leaf fr.Element) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.insertLocked(leaf)
}

func (t *InMemoryMerkleTree) insertLocked(leaf fr.Element) (int, error) {
	index := t.size
	if index >= MaxLeaves {
		return 0, fmt.Errorf("tree full: max %d leaves", MaxLeaves)
	}

	// Set leaf at level 0
	if index < len(t.layers[0]) {
		t.layers[0][index] = leaf
	} else {
		t.layers[0] = append(t.layers[0], leaf)
	}

	// Recompute path from leaf to root
	currentIndex := index
	for level := 0; level < Depth; level++ {
		parentIndex := currentIndex >> 1
		isRight := currentIndex & 1

		var left, right fr.Element
		if isRight == 1 {
			left = t.layers[level][currentIndex-1]
			right = t.layers[level][currentIndex]
		} else {
			left = t.layers[level][currentIndex]
			if currentIndex+1 < len(t.layers[level]) {
				right = t.layers[level][currentIndex+1]
			} else {
				right = t.zeros[level]
			}
		}

		hash, err := hash2(left, right)
		if err != nil {
			return 0, fmt.Errorf("hashing at level %d: %w", level, err)
		}

		// Set parent
		nextLayer := t.layers[level+1]
		if parentIndex < len(nextLayer) {
			nextLayer[parentIndex] = hash
		} else {
			// Grow the layer — parentIndex should equal len(nextLayer) for append
			t.layers[level+1] = append(nextLayer, hash)
		}

		currentIndex = parentIndex
	}

	t.size++
	return index, nil
}

// GetProof generates a Merkle inclusion proof for the leaf at the given index.
func (t *InMemoryMerkleTree) GetProof(leafIndex int) (MerkleProof, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if leafIndex < 0 || leafIndex >= t.size {
		return MerkleProof{}, fmt.Errorf("leafIndex %d out of range [0, %d)", leafIndex, t.size)
	}

	var proof MerkleProof
	proof.Leaf = t.layers[0][leafIndex]
	proof.Index = leafIndex
	proof.Root = t.rootLocked()

	currentIndex := leafIndex
	for level := 0; level < Depth; level++ {
		isRight := currentIndex & 1
		var siblingIndex int
		if isRight == 1 {
			siblingIndex = currentIndex - 1
		} else {
			siblingIndex = currentIndex + 1
		}

		if siblingIndex < len(t.layers[level]) {
			proof.Siblings[level] = t.layers[level][siblingIndex]
		} else {
			proof.Siblings[level] = t.zeros[level]
		}

		currentIndex >>= 1
	}

	return proof, nil
}

// GetLeaf returns a canonical field commitment at an inserted leaf index.
func (t *InMemoryMerkleTree) GetLeaf(index int) (fr.Element, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if index < 0 || index >= t.size {
		return fr.Element{}, fmt.Errorf("leaf index %d out of range", index)
	}
	return t.layers[0][index], nil
}

func (t *InMemoryMerkleTree) cloneLocked() *InMemoryMerkleTree {
	copyTree := &InMemoryMerkleTree{zeros: t.zeros, size: t.size}
	for i := range t.layers {
		copyTree.layers[i] = append([]fr.Element(nil), t.layers[i]...)
	}
	return copyTree
}
