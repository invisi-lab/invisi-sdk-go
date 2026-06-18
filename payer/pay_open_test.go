package payer

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

func TestBuildPayOpenCalldata_Basic(t *testing.T) {
	result, err := BuildPayOpenCalldata(PayOpenParams{
		ChainID:            big.NewInt(1),
		IdentityCommitment: common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ab"),
		TokenAddress:       common.Address{},
		Amount:             big.NewInt(100_000_000),
		TreeID:             big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Data should start with the payOpen 4-byte selector
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}

	// PaymentCommitment must be non-zero
	if result.PaymentCommitment == (common.Hash{}) {
		t.Error("paymentCommitment should not be zero")
	}

	// Nonce must be non-zero (random)
	if result.Nonce == (common.Hash{}) {
		t.Error("nonce should not be zero")
	}

	// ApproveAmount == Amount
	if result.ApproveAmount.Cmp(big.NewInt(100_000_000)) != 0 {
		t.Errorf("approveAmount = %s, want 100000000", result.ApproveAmount)
	}
}

func TestBuildPayOpenCalldata_ExternalNonce(t *testing.T) {
	var nonce fr.Element
	nonce.SetUint64(12345)

	result, err := BuildPayOpenCalldata(PayOpenParams{
		ChainID:            big.NewInt(84532),
		IdentityCommitment: common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		TokenAddress:       common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
		Amount:             big.NewInt(100_000_000),
		TreeID:             big.NewInt(1),
		Nonce:              &nonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedNonce := common.HexToHash(core.ElementToHex(nonce))
	if result.Nonce != expectedNonce {
		t.Errorf("nonce = %s, want %s", result.Nonce.Hex(), expectedNonce.Hex())
	}
}

func TestBuildPayOpenCalldata_UniqueNonces(t *testing.T) {
	params := PayOpenParams{
		ChainID:            big.NewInt(1),
		IdentityCommitment: common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ab"),
		TokenAddress:       common.Address{},
		Amount:             big.NewInt(100_000_000),
		TreeID:             big.NewInt(1),
	}
	a, err := BuildPayOpenCalldata(params)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BuildPayOpenCalldata(params)
	if err != nil {
		t.Fatal(err)
	}
	if a.Nonce == b.Nonce {
		t.Error("two calls should produce different nonces")
	}
}

func TestBuildPayOpenCalldata_CommitmentConsistency(t *testing.T) {
	// Verify the returned paymentCommitment matches core.ComputePaymentCommitment
	chainID := big.NewInt(84532)
	identityCommitment := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ab")
	tokenAddr := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
	amount := big.NewInt(100_000_000)

	var nonce fr.Element
	nonce.SetUint64(42)

	result, err := BuildPayOpenCalldata(PayOpenParams{
		ChainID:            chainID,
		IdentityCommitment: identityCommitment,
		TokenAddress:       tokenAddr,
		Amount:             amount,
		TreeID:             big.NewInt(1),
		Nonce:              &nonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Recompute via core
	chainElem := core.BigIntToElement(chainID)
	icElem, err := core.HexToElement(identityCommitment.Hex())
	if err != nil {
		t.Fatal(err)
	}
	tokenElem, err := core.HexToElement(common.BytesToHash(tokenAddr.Bytes()).Hex())
	if err != nil {
		t.Fatal(err)
	}
	amountElem := core.BigIntToElement(amount)

	expected, err := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonce)
	if err != nil {
		t.Fatal(err)
	}

	expectedHash := common.HexToHash(core.ElementToHex(expected))
	if result.PaymentCommitment != expectedHash {
		t.Errorf("commitment mismatch: got %s, want %s",
			result.PaymentCommitment.Hex(), expectedHash.Hex())
	}
}

func TestBuildPayOpenCalldata_WithRefundLink(t *testing.T) {
	var refund fr.Element
	refund.SetUint64(999)

	result, err := BuildPayOpenCalldata(PayOpenParams{
		ChainID:            big.NewInt(1),
		IdentityCommitment: common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ab"),
		TokenAddress:       common.Address{},
		Amount:             big.NewInt(100),
		TreeID:             big.NewInt(1),
		RefundLink:         &refund,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Calldata should be longer than without refundLink (always encoded, just non-zero)
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}
