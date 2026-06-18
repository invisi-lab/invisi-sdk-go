package receiver

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

func makeProposeParams() ProposeParams {
	return ProposeParams{
		IdentityCommitment: common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		TokenAddress:       common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
		Amount:             big.NewInt(100_000_000),
		ChainID:            big.NewInt(84532),
	}
}

func TestComputePaymentCommitmentForPropose_ReturnsNonceAndCommitment(t *testing.T) {
	result, err := ComputePaymentCommitmentForPropose(makeProposeParams())
	if err != nil {
		t.Fatal(err)
	}

	// Nonce must not be zero
	if result.Nonce == (common.Hash{}) {
		t.Error("Nonce is zero")
	}

	// PaymentCommitment must not be zero
	if result.PaymentCommitment == (common.Hash{}) {
		t.Error("PaymentCommitment is zero")
	}
}

func TestComputePaymentCommitmentForPropose_DifferentNoncesEachCall(t *testing.T) {
	params := makeProposeParams()

	a, err := ComputePaymentCommitmentForPropose(params)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ComputePaymentCommitmentForPropose(params)
	if err != nil {
		t.Fatal(err)
	}

	if a.Nonce == b.Nonce {
		t.Error("Two calls produced the same nonce")
	}
	if a.PaymentCommitment == b.PaymentCommitment {
		t.Error("Two calls produced the same paymentCommitment")
	}
}

func TestComputePaymentCommitmentForPropose_MatchesCoreComputation(t *testing.T) {
	params := makeProposeParams()

	result, err := ComputePaymentCommitmentForPropose(params)
	if err != nil {
		t.Fatal(err)
	}

	// Re-derive using core functions
	nonceElem := core.BigIntToElement(result.Nonce.Big())
	icElem, err := core.HexToElement(params.IdentityCommitment.Hex())
	if err != nil {
		t.Fatal(err)
	}
	tokenElem := core.BigIntToElement(new(big.Int).SetBytes(params.TokenAddress.Bytes()))
	amountElem := core.BigIntToElement(params.Amount)
	chainElem := core.BigIntToElement(params.ChainID)

	expectedPC, err := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonceElem)
	if err != nil {
		t.Fatal(err)
	}

	expectedPCHash := common.HexToHash(core.ElementToHex(expectedPC))
	if result.PaymentCommitment != expectedPCHash {
		t.Errorf("PaymentCommitment mismatch: got %s, want %s",
			result.PaymentCommitment.Hex(), expectedPCHash.Hex())
	}
}
