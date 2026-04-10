package tree

// Package tree provides an in-memory depth-16 Poseidon2 Merkle tree
// with chain synchronization for the Invisible Protocol.
//
// Three main components:
//   - InMemoryMerkleTree: pure computational Merkle tree with insert/proof
//   - MerkleTreeSync: syncs a single tree from on-chain events
//   - InvisiTreeManager: manages multiple trees by treeId
