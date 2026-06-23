package payer

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
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

// VerifyCreditParams carries the inputs needed to recompute a credit commitment,
// plus the server-returned CreditCommitment value to verify against.
type VerifyCreditParams struct {
	ChainID            fr.Element
	IdentityCommitment fr.Element
	PayerCommitment    fr.Element
	TokenAddress       fr.Element
	Amount             fr.Element
	Nonce              fr.Element
	Expiry             fr.Element
	NullifierKeyHash   fr.Element
	CreditCommitment   fr.Element // server-returned value to verify
}

// VerifyCreditCommitment recomputes the credit commitment locally and compares it
// to the server-returned value. A mismatch means the server folded a wrong nonce
// (or other parameter) — the caller MUST abort and not depositCredit.
func VerifyCreditCommitment(p VerifyCreditParams) (bool, error) {
	got, err := core.ComputeCreditCommitment(
		p.ChainID, p.IdentityCommitment, p.PayerCommitment, p.TokenAddress,
		p.Amount, p.Nonce, p.Expiry, p.NullifierKeyHash,
	)
	if err != nil {
		return false, fmt.Errorf("payer: compute credit commitment: %w", err)
	}
	return got.Equal(&p.CreditCommitment), nil
}
