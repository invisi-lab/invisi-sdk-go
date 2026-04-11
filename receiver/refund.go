package receiver

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// payOpenABI is the ABI fragment for InvisiProtocol.payOpen.
// payOpen(bytes32 identityCommitment, uint256 amount, bytes32 nonce, uint256 treeId, bytes32 refundLink)
var payOpenABI abi.ABI

func init() {
	var err error
	payOpenABI, err = abi.JSON(strings.NewReader(`[{
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
	}]`))
	if err != nil {
		panic(fmt.Sprintf("receiver: failed to parse payOpen ABI: %v", err))
	}
}

// BuildRefundPayOpenCalldata encodes a payOpen call for a refund transaction.
//
// This builds the calldata for the merchant to send funds back to the payer.
// The identityCommitment should be derived from the refund nullifier key and
// the payer's address. The refundLink anchors the refund to the original
// payment commitment.
//
// Note: This does NOT depend on the payer package. When the payer package
// is implemented, it will also have a BuildPayOpenCalldata function for
// payer-initiated payments. This receiver version is specifically for
// merchant-initiated refunds.
func BuildRefundPayOpenCalldata(params RefundPayOpenParams) (*RefundPayOpenCalldata, error) {
	if params.RefundAmount == nil {
		return nil, fmt.Errorf("receiver: refund amount is required")
	}
	if params.RefundAmount.Sign() <= 0 {
		return nil, fmt.Errorf("receiver: refund amount must be positive")
	}

	data, err := payOpenABI.Pack(
		"payOpen",
		[32]byte(params.IdentityCommitment),
		params.RefundAmount,
		[32]byte(params.Nonce),
		params.TreeID,
		[32]byte(params.RefundLink),
	)
	if err != nil {
		return nil, fmt.Errorf("receiver: failed to encode payOpen refund calldata: %w", err)
	}

	return &RefundPayOpenCalldata{Data: data}, nil
}
