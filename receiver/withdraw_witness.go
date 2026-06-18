package receiver

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/invisi-lab/invisi-sdk-go/tree"
)

// BuildWithdrawWitness constructs the complete ZK circuit witness for a
// batch settlement proof. The witness contains:
//   - Per-payment: nullifierKey, amount, nonce, merkle proof
//   - Global: treeRoot, totalAmount, receivingAddress
//
// All Merkle proofs must share the same root (same tree). The function
// validates input consistency before building the witness.
//
// After building, send the witness to a remote prover service for proof
// generation, then use BuildWithdrawCalldata to encode the on-chain call.
func BuildWithdrawWitness(params BuildWithdrawWitnessParams) (*WithdrawWitness, error) {
	count := len(params.NullifierKeys)

	// Validate non-empty
	if count == 0 {
		return nil, fmt.Errorf("receiver: at least one payment required")
	}

	// Validate consistent lengths
	if len(params.Amounts) != count ||
		len(params.Nonces) != count ||
		len(params.MerkleProofs) != count {
		return nil, fmt.Errorf(
			"receiver: length mismatch: nullifierKeys=%d, amounts=%d, nonces=%d, merkleProofs=%d",
			count, len(params.Amounts), len(params.Nonces), len(params.MerkleProofs),
		)
	}

	// Validate all merkle proofs share the same root
	firstRoot := params.MerkleProofs[0].Root
	for i := 1; i < count; i++ {
		if !params.MerkleProofs[i].Root.Equal(&firstRoot) {
			return nil, fmt.Errorf(
				"receiver: merkle root mismatch at index %d: all proofs must share the same tree root",
				i,
			)
		}
	}

	// Compute total amount
	totalAmount := new(big.Int)
	for _, amt := range params.Amounts {
		if amt == nil {
			return nil, fmt.Errorf("receiver: nil amount in batch")
		}
		totalAmount.Add(totalAmount, amt)
	}

	// Deep copy slices to avoid aliasing
	nks := make([]fr.Element, count)
	copy(nks, params.NullifierKeys)

	amounts := make([]*big.Int, count)
	for i, a := range params.Amounts {
		amounts[i] = new(big.Int).Set(a)
	}

	nonces := make([]fr.Element, count)
	copy(nonces, params.Nonces)

	proofs := make([]tree.MerkleProof, count)
	copy(proofs, params.MerkleProofs)

	return &WithdrawWitness{
		NullifierKeys: nks,
		Amounts:       amounts,
		Nonces:        nonces,
		MerkleProofs:  proofs,
		TreeRoot:      firstRoot,
		TotalAmount:   totalAmount,
		Receiver:      params.ReceivingAddress,
	}, nil
}
