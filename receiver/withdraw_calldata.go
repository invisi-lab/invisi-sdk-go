package receiver

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// withdrawABI is the ABI fragment for InvisiProtocol.withdraw.
// withdraw(bytes proof, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount, address receivingAddress, uint256 treeId)
var withdrawABI abi.ABI

func init() {
	var err error
	withdrawABI, err = abi.JSON(strings.NewReader(`[{
		"type": "function",
		"name": "withdraw",
		"inputs": [
			{"name": "proof", "type": "bytes"},
			{"name": "merkleRoot", "type": "bytes32"},
			{"name": "nullifiers", "type": "bytes32[]"},
			{"name": "totalAmount", "type": "uint256"},
			{"name": "receivingAddress", "type": "address"},
			{"name": "treeId", "type": "uint256"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	}]`))
	if err != nil {
		panic(fmt.Sprintf("receiver: failed to parse withdraw ABI: %v", err))
	}
}

// BuildWithdrawCalldata encodes a withdraw call using the proof from a
// remote prover and the settlement parameters.
//
// Call flow: BuildWithdrawWitness -> remote prover -> BuildWithdrawCalldata.
func BuildWithdrawCalldata(params WithdrawParams) (*WithdrawCalldata, error) {
	if len(params.Nullifiers) == 0 {
		return nil, fmt.Errorf("receiver: at least one nullifier required")
	}

	// Convert []common.Hash to [][32]byte for ABI encoding
	nullifiers := make([][32]byte, len(params.Nullifiers))
	for i, n := range params.Nullifiers {
		nullifiers[i] = n
	}

	data, err := withdrawABI.Pack(
		"withdraw",
		params.Proof,
		[32]byte(params.MerkleRoot),
		nullifiers,
		params.TotalAmount,
		params.ReceivingAddress,
		params.TreeID,
	)
	if err != nil {
		return nil, fmt.Errorf("receiver: failed to encode withdraw calldata: %w", err)
	}

	// Deep copy nullifiers for return
	retNullifiers := make([]common.Hash, len(params.Nullifiers))
	copy(retNullifiers, params.Nullifiers)

	return &WithdrawCalldata{
		Data:        data,
		Nullifiers:  retNullifiers,
		TotalAmount: new(big.Int).Set(params.TotalAmount),
	}, nil
}
