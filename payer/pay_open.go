package payer

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// BuildPayOpenCalldata computes the paymentCommitment and encodes
// payOpen(identityCommitment, amount, nonce, treeId, refundLink) calldata.
func BuildPayOpenCalldata(params PayOpenParams) (*PayOpenCalldata, error) {
	if params.ChainID == nil {
		return nil, fmt.Errorf("payer: ChainID is required")
	}
	if params.Amount == nil {
		return nil, fmt.Errorf("payer: Amount is required")
	}
	if params.TreeID == nil {
		return nil, fmt.Errorf("payer: TreeID is required")
	}

	// Resolve nonce: caller must supply it per the payment mode. This builder
	// never generates one.
	if params.Nonce == nil {
		return nil, fmt.Errorf("payer: Nonce is required")
	}
	nonce := *params.Nonce

	// Convert params to field elements for Poseidon2
	chainElem := core.BigIntToElement(params.ChainID)

	icElem, err := core.HexToElement(params.IdentityCommitment.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid identityCommitment: %w", err)
	}

	tokenHash := addressToTokenElement(params.TokenAddress)
	tokenElem, err := core.HexToElement(tokenHash.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid tokenAddress: %w", err)
	}

	amountElem := core.BigIntToElement(params.Amount)

	// Compute paymentCommitment = Poseidon2(D_PAY, chainId, identityCommitment, tokenAddress, amount, nonce)
	commitment, err := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonce)
	if err != nil {
		return nil, fmt.Errorf("payer: failed to compute paymentCommitment: %w", err)
	}

	// Convert to common.Hash for ABI encoding
	nonceHash := common.HexToHash(core.ElementToHex(nonce))

	// RefundLink: default to zero
	var refundLinkHash [32]byte
	if params.RefundLink != nil {
		refundLinkHash = common.HexToHash(core.ElementToHex(*params.RefundLink))
	}

	// ABI encode: payOpen(identityCommitment, amount, nonce, treeId, refundLink)
	data, err := packCalldata(
		invisiProtocolABI,
		"payOpen",
		params.IdentityCommitment, // bytes32
		params.Amount,             // uint256
		nonceHash,                 // bytes32
		params.TreeID,             // uint256
		refundLinkHash,            // bytes32
	)
	if err != nil {
		return nil, err
	}

	commitmentHash := common.HexToHash(core.ElementToHex(commitment))

	return &PayOpenCalldata{
		PaymentCommitment: commitmentHash,
		Nonce:             nonceHash,
		Data:              data,
		ApproveAmount:     new(big.Int).Set(params.Amount),
	}, nil
}
