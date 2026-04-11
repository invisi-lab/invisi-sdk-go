package payer

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-labs/invisi-sdk-go/core"
	"github.com/invisi-labs/invisi-sdk-go/ipp"
)

// BuildPaymentFromRequest bridges an ipp.IPPRequest to payment calldata.
//
// Flow:
//  1. Parse chainId from request.Network (CAIP-2)
//  2. Verify the provided paymentCommitment matches the request parameters
//  3. Build payOpen calldata using the request's identity commitment, token, treeId
//
// This is the main integration point for payer applications receiving IPP payment requests.
func BuildPaymentFromRequest(params BuildFromRequestParams) (*BuildFromRequestResult, error) {
	if params.Amount == nil {
		return nil, fmt.Errorf("payer: Amount is required")
	}

	// Step 1: Parse chainId from CAIP-2 network string
	chainID, err := ipp.CAIP2ToChainID(params.Request.Network)
	if err != nil {
		return nil, fmt.Errorf("payer: invalid network: %w", err)
	}

	p := params.Request.Payload

	// Step 2: Verify paymentCommitment
	icHash := common.HexToHash(p.IdentityCommitment)
	tokenAddr := common.HexToAddress(p.TokenAddress)

	valid, err := VerifyPaymentCommitment(VerifyParams{
		IdentityCommitment: icHash,
		TokenAddress:       tokenAddr,
		Amount:             params.Amount,
		Nonce:              params.Nonce,
		PaymentCommitment:  params.PaymentCommitment,
		ChainID:            chainID,
	})
	if err != nil {
		return nil, fmt.Errorf("payer: commitment verification failed: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("payer: paymentCommitment verification failed: server returned incorrect parameters")
	}

	// Step 3: Build payOpen calldata
	treeID, ok := new(big.Int).SetString(p.TreeID, 10)
	if !ok {
		return nil, fmt.Errorf("payer: invalid treeId: %q", p.TreeID)
	}

	nonceElem, err := core.HexToElement(params.Nonce.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid nonce: %w", err)
	}

	result, err := BuildPayOpenCalldata(PayOpenParams{
		ChainID:            chainID,
		IdentityCommitment: icHash,
		TokenAddress:       tokenAddr,
		Amount:             params.Amount,
		TreeID:             treeID,
		Nonce:              &nonceElem,
	})
	if err != nil {
		return nil, fmt.Errorf("payer: failed to build calldata: %w", err)
	}

	return &BuildFromRequestResult{
		Calldata:      result.Data,
		ApproveAmount: result.ApproveAmount,
	}, nil
}
