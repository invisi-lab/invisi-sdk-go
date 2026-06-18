package receiver

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

// ComputePaymentCommitmentForPropose generates a random nonce and computes
// the paymentCommitment for a propose response.
//
// This is called by the receiver's propose handler when a payer proposes
// an amount for an interactive payment, or internally by CreatePaymentRequest
// for mode 1a (pre-computed nonce).
func ComputePaymentCommitmentForPropose(params ProposeParams) (*ProposeResult, error) {
	// Generate random nonce
	nonce, err := core.GenerateNonce()
	if err != nil {
		return nil, err
	}

	// Convert params to field elements
	icElem, err := core.HexToElement(params.IdentityCommitment.Hex())
	if err != nil {
		return nil, err
	}
	tokenElem := core.BigIntToElement(new(big.Int).SetBytes(params.TokenAddress.Bytes()))
	amountElem := core.BigIntToElement(params.Amount)
	chainElem := core.BigIntToElement(params.ChainID)

	// paymentCommitment = Poseidon2(D_PAY, chainId, identityCommitment, tokenAddress, amount, nonce)
	pc, err := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonce)
	if err != nil {
		return nil, err
	}

	nonceHash := common.HexToHash(core.ElementToHex(nonce))
	pcHash := common.HexToHash(core.ElementToHex(pc))

	return &ProposeResult{
		Nonce:             nonceHash,
		PaymentCommitment: pcHash,
	}, nil
}
