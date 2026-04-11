package payer

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-labs/invisi-sdk-go/core"
)

// BuildPayBlindCalldata encodes payBlind(proof, paymentCommitment, changeCommitment, inputAmount, treeId, refundLink).
// The proof must be obtained from a remote prover — this function only assembles calldata.
func BuildPayBlindCalldata(params PayBlindParams) (*PayBlindCalldata, error) {
	if params.InputAmount == nil {
		return nil, fmt.Errorf("payer: InputAmount is required")
	}
	if params.TreeID == nil {
		return nil, fmt.Errorf("payer: TreeID is required")
	}

	// RefundLink: default to zero
	var refundLinkHash [32]byte
	if params.RefundLink != nil {
		refundLinkHash = common.HexToHash(core.ElementToHex(*params.RefundLink))
	}

	// ABI encode: payBlind(proof, paymentCommitment, changeCommitment, inputAmount, treeId, refundLink)
	data, err := packCalldata(
		invisiProtocolABI,
		"payBlind",
		params.Proof,             // bytes
		params.PaymentCommitment, // bytes32
		params.ChangeCommitment,  // bytes32
		params.InputAmount,       // uint256
		params.TreeID,            // uint256
		refundLinkHash,           // bytes32
	)
	if err != nil {
		return nil, err
	}

	return &PayBlindCalldata{
		Data:          data,
		ApproveAmount: new(big.Int).Set(params.InputAmount),
	}, nil
}
