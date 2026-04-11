package receiver

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// reclaimABI is the ABI fragment for InvisiProtocol.reclaim.
// reclaim(bytes proof, bytes32 merkleRoot, bytes32 nullifier, uint256 amount, address receivingAddress, uint256 treeId, uint256 expiry)
var reclaimABI abi.ABI

func init() {
	var err error
	reclaimABI, err = abi.JSON(strings.NewReader(`[{
		"type": "function",
		"name": "reclaim",
		"inputs": [
			{"name": "proof", "type": "bytes"},
			{"name": "merkleRoot", "type": "bytes32"},
			{"name": "nullifier", "type": "bytes32"},
			{"name": "amount", "type": "uint256"},
			{"name": "receivingAddress", "type": "address"},
			{"name": "treeId", "type": "uint256"},
			{"name": "expiry", "type": "uint256"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	}]`))
	if err != nil {
		panic(fmt.Sprintf("receiver: failed to parse reclaim ABI: %v", err))
	}
}

// BuildCreditReclaimCalldata encodes a reclaim call for an expired credit.
// This is used by the payer to reclaim funds from an expired credit line.
func BuildCreditReclaimCalldata(params CreditReclaimParams) (*CreditReclaimCalldata, error) {
	if params.Amount == nil {
		return nil, fmt.Errorf("receiver: amount is required for credit reclaim")
	}

	data, err := reclaimABI.Pack(
		"reclaim",
		params.Proof,
		[32]byte(params.MerkleRoot),
		[32]byte(params.Nullifier),
		params.Amount,
		params.ReceivingAddress,
		params.TreeID,
		params.Expiry,
	)
	if err != nil {
		return nil, fmt.Errorf("receiver: failed to encode reclaim calldata: %w", err)
	}

	return &CreditReclaimCalldata{Data: data}, nil
}
