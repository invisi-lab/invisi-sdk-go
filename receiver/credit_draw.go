package receiver

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/tree"
)

// creditDrawABI is the ABI fragment for InvisiProtocol.creditDraw.
// creditDraw(bytes proof, tuple(bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount, address receivingAddress, uint256 treeId, bytes32[] remainderCommitments, uint256 expiry) params)
var creditDrawABI abi.ABI

func init() {
	var err error
	creditDrawABI, err = abi.JSON(strings.NewReader(`[{
		"type": "function",
		"name": "creditDraw",
		"inputs": [
			{"name": "proof", "type": "bytes"},
			{
				"name": "params",
				"type": "tuple",
				"components": [
					{"name": "merkleRoot", "type": "bytes32"},
					{"name": "nullifiers", "type": "bytes32[]"},
					{"name": "totalAmount", "type": "uint256"},
					{"name": "receivingAddress", "type": "address"},
					{"name": "treeId", "type": "uint256"},
					{"name": "remainderCommitments", "type": "bytes32[]"},
					{"name": "expiry", "type": "uint256"}
				]
			}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	}]`))
	if err != nil {
		panic(fmt.Sprintf("receiver: failed to parse creditDraw ABI: %v", err))
	}
}

// BuildCreditDrawWitness constructs the ZK circuit witness for a credit draw proof.
//
// Similar to BuildWithdrawWitness but includes credit-specific fields:
// remainder commitments, expiries, payer commitments, and nullifier key hashes.
//
// RemainderCommitments are filled with zero-value placeholders for now (phase 2
// will add proper computation from original credit amounts vs drawn amounts).
//
// All Merkle proofs must share the same root (same tree). The function
// validates input consistency before building the witness.
func BuildCreditDrawWitness(params BuildCreditDrawWitnessParams) (*CreditDrawWitness, error) {
	count := len(params.NullifierKeys)

	if count == 0 {
		return nil, fmt.Errorf("receiver: at least one credit required")
	}

	// Validate consistent lengths
	if len(params.Amounts) != count ||
		len(params.Nonces) != count ||
		len(params.Expiries) != count ||
		len(params.PayerCommitments) != count ||
		len(params.NullifierKeyHashes) != count ||
		len(params.MerkleProofs) != count {
		return nil, fmt.Errorf(
			"receiver: length mismatch: nullifierKeys=%d, amounts=%d, nonces=%d, expiries=%d, payerCommitments=%d, nullifierKeyHashes=%d, merkleProofs=%d",
			count, len(params.Amounts), len(params.Nonces), len(params.Expiries),
			len(params.PayerCommitments), len(params.NullifierKeyHashes), len(params.MerkleProofs),
		)
	}

	// Validate all merkle proofs share the same root
	firstRoot := params.MerkleProofs[0].Root
	for i := 1; i < count; i++ {
		if !params.MerkleProofs[i].Root.Equal(&firstRoot) {
			return nil, fmt.Errorf("receiver: merkle root mismatch at index %d", i)
		}
	}

	// Compute total amount
	totalAmount := new(big.Int)
	for _, amt := range params.Amounts {
		if amt == nil {
			return nil, fmt.Errorf("receiver: nil amount in credit draw batch")
		}
		totalAmount.Add(totalAmount, amt)
	}

	// Remainder commitments: zero-value placeholders for phase 1.
	// Full logic requires knowledge of original credit amounts vs drawn amounts.
	remainderCommitments := make([]fr.Element, count)

	// Deep copy all slices to avoid aliasing
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

	expiries := make([]*big.Int, count)
	for i, e := range params.Expiries {
		expiries[i] = new(big.Int).Set(e)
	}

	payerComms := make([]fr.Element, count)
	copy(payerComms, params.PayerCommitments)

	nkHashes := make([]fr.Element, count)
	copy(nkHashes, params.NullifierKeyHashes)

	return &CreditDrawWitness{
		NullifierKeys:        nks,
		Amounts:              amounts,
		Nonces:               nonces,
		MerkleProofs:         proofs,
		TreeRoot:             firstRoot,
		TotalAmount:          totalAmount,
		Receiver:             params.ReceivingAddress,
		RemainderCommitments: remainderCommitments,
		Expiries:             expiries,
		PayerCommitments:     payerComms,
		NullifierKeyHashes:   nkHashes,
	}, nil
}

// creditDrawParamsTuple matches the Solidity struct for creditDraw params.
type creditDrawParamsTuple struct {
	MerkleRoot           [32]byte
	Nullifiers           [][32]byte
	TotalAmount          *big.Int
	ReceivingAddress     common.Address
	TreeId               *big.Int
	RemainderCommitments [][32]byte
	Expiry               *big.Int
}

// BuildCreditDrawCalldata encodes a creditDraw call using the proof from a
// remote prover and the credit draw parameters.
//
// Call flow: BuildCreditDrawWitness -> remote prover -> BuildCreditDrawCalldata.
func BuildCreditDrawCalldata(params CreditDrawParams) (*CreditDrawCalldata, error) {
	if len(params.Nullifiers) == 0 {
		return nil, fmt.Errorf("receiver: at least one nullifier required for credit draw")
	}

	nullifiers := make([][32]byte, len(params.Nullifiers))
	for i, n := range params.Nullifiers {
		nullifiers[i] = n
	}

	remainderComms := make([][32]byte, len(params.RemainderCommitments))
	for i, rc := range params.RemainderCommitments {
		remainderComms[i] = rc
	}

	tupleParams := creditDrawParamsTuple{
		MerkleRoot:           [32]byte(params.MerkleRoot),
		Nullifiers:           nullifiers,
		TotalAmount:          params.TotalAmount,
		ReceivingAddress:     params.ReceivingAddress,
		TreeId:               params.TreeID,
		RemainderCommitments: remainderComms,
		Expiry:               params.Expiry,
	}

	data, err := creditDrawABI.Pack("creditDraw", params.Proof, tupleParams)
	if err != nil {
		return nil, fmt.Errorf("receiver: failed to encode creditDraw calldata: %w", err)
	}

	retNullifiers := make([]common.Hash, len(params.Nullifiers))
	copy(retNullifiers, params.Nullifiers)

	return &CreditDrawCalldata{
		Data:        data,
		Nullifiers:  retNullifiers,
		TotalAmount: new(big.Int).Set(params.TotalAmount),
	}, nil
}
