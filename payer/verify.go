package payer

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-labs/invisi-sdk-go/core"
)

// VerifyPaymentCommitment recomputes a paymentCommitment from the given params
// and returns true if it matches the provided paymentCommitment.
func VerifyPaymentCommitment(params VerifyParams) (bool, error) {
	if params.ChainID == nil {
		return false, fmt.Errorf("payer: ChainID is required")
	}
	if params.Amount == nil {
		return false, fmt.Errorf("payer: Amount is required")
	}

	chainElem := core.BigIntToElement(params.ChainID)
	icElem, err := core.HexToElement(params.IdentityCommitment.Hex())
	if err != nil {
		return false, fmt.Errorf("payer: invalid identityCommitment: %w", err)
	}
	tokenHash := addressToTokenElement(params.TokenAddress)
	tokenElem, err := core.HexToElement(tokenHash.Hex())
	if err != nil {
		return false, fmt.Errorf("payer: invalid tokenAddress: %w", err)
	}
	amountElem := core.BigIntToElement(params.Amount)
	nonceElem, err := core.HexToElement(params.Nonce.Hex())
	if err != nil {
		return false, fmt.Errorf("payer: invalid nonce: %w", err)
	}

	computed, err := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonceElem)
	if err != nil {
		return false, fmt.Errorf("payer: failed to compute commitment: %w", err)
	}

	computedHash := common.HexToHash(core.ElementToHex(computed))
	return computedHash == params.PaymentCommitment, nil
}
