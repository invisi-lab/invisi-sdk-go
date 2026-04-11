package payer

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// invisiProtocolABI is the parsed ABI for InvisiProtocol contract.
// Only the methods used by this package are included.
var invisiProtocolABI abi.ABI

// erc20ABI is the parsed ABI for ERC20 approve.
var erc20ABI abi.ABI

func init() {
	var err error
	invisiProtocolABI, err = abi.JSON(strings.NewReader(invisiProtocolABIJSON))
	if err != nil {
		panic(fmt.Sprintf("payer: failed to parse InvisiProtocol ABI: %v", err))
	}
	erc20ABI, err = abi.JSON(strings.NewReader(erc20ABIJSON))
	if err != nil {
		panic(fmt.Sprintf("payer: failed to parse ERC20 ABI: %v", err))
	}
}

const invisiProtocolABIJSON = `[
  {
    "type": "function",
    "name": "payOpen",
    "inputs": [
      {"name": "identityCommitment", "type": "bytes32"},
      {"name": "amount", "type": "uint256"},
      {"name": "nonce", "type": "bytes32"},
      {"name": "treeId", "type": "uint256"},
      {"name": "refundLink", "type": "bytes32"}
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  },
  {
    "type": "function",
    "name": "payBlind",
    "inputs": [
      {"name": "proof", "type": "bytes"},
      {"name": "paymentCommitment", "type": "bytes32"},
      {"name": "changeCommitment", "type": "bytes32"},
      {"name": "inputAmount", "type": "uint256"},
      {"name": "treeId", "type": "uint256"},
      {"name": "refundLink", "type": "bytes32"}
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  },
  {
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
  },
  {
    "type": "function",
    "name": "depositCredit",
    "inputs": [
      {"name": "identityCommitment", "type": "bytes32"},
      {"name": "payerCommitment", "type": "bytes32"},
      {"name": "amount", "type": "uint256"},
      {"name": "nonce", "type": "bytes32"},
      {"name": "expiry", "type": "uint256"},
      {"name": "nullifierKeyHash", "type": "bytes32"},
      {"name": "treeId", "type": "uint256"}
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  },
  {
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
  }
]`

const erc20ABIJSON = `[
  {
    "type": "function",
    "name": "approve",
    "inputs": [
      {"name": "spender", "type": "address"},
      {"name": "amount", "type": "uint256"}
    ],
    "outputs": [{"name": "", "type": "bool"}],
    "stateMutability": "nonpayable"
  }
]`

// packCalldata ABI-encodes a function call. Returns the 4-byte selector + encoded args.
func packCalldata(contractABI abi.ABI, method string, args ...interface{}) ([]byte, error) {
	data, err := contractABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("payer: failed to encode %s: %w", method, err)
	}
	return data, nil
}

// elementToBytes32 converts an fr.Element hex string to a [32]byte for ABI encoding.
func elementToBytes32(hex string) [32]byte {
	return common.HexToHash(hex)
}

// zeroBytes32 returns a zero-valued [32]byte.
func zeroBytes32() [32]byte {
	return [32]byte{}
}

// addressToTokenElement converts common.Address to the field element representation
// used by Poseidon2 hashing (left-padded 32-byte big-endian).
func addressToTokenElement(addr common.Address) common.Hash {
	return common.BytesToHash(addr.Bytes())
}
