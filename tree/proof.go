package tree

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/invisi-labs/invisi-sdk-go/core"
)

const Depth = core.TreeDepth // 16

// MerkleProof contains a Merkle inclusion proof for a leaf.
type MerkleProof struct {
	Root     fr.Element
	Leaf     fr.Element
	Index    int
	Siblings [Depth]fr.Element
}

// Verify checks the Merkle proof by recomputing the root from the leaf and siblings.
// Returns true if the recomputed root matches the proof root.
func (p *MerkleProof) Verify() (bool, error) {
	current := p.Leaf
	idx := p.Index

	for i := 0; i < Depth; i++ {
		var left, right fr.Element
		if idx&1 == 0 {
			// current is left child
			left = current
			right = p.Siblings[i]
		} else {
			// current is right child
			left = p.Siblings[i]
			right = current
		}
		hash, err := core.Poseidon2Hash(left, right)
		if err != nil {
			return false, err
		}
		current = hash
		idx >>= 1
	}

	return current.Equal(&p.Root), nil
}
