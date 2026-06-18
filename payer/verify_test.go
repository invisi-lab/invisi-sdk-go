package payer

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

func TestVerifyPaymentCommitment_Valid(t *testing.T) {
	// Set up known values
	chainID := big.NewInt(84532)
	amount := big.NewInt(100_000_000)

	var nonce fr.Element
	nonce.SetUint64(42)

	// Arbitrary identity commitment and token address
	identityCommitment := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000ab")
	tokenAddr := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")

	// Compute the expected commitment via core
	chainElem := core.BigIntToElement(chainID)
	icElem, _ := core.HexToElement(identityCommitment.Hex())
	tokenHash := common.BytesToHash(tokenAddr.Bytes())
	tokenElem, _ := core.HexToElement(tokenHash.Hex())
	amountElem := core.BigIntToElement(amount)

	expected, err := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonce)
	if err != nil {
		t.Fatal(err)
	}
	expectedHash := common.HexToHash(core.ElementToHex(expected))
	nonceHash := common.HexToHash(core.ElementToHex(nonce))

	valid, err := VerifyPaymentCommitment(VerifyParams{
		IdentityCommitment: identityCommitment,
		TokenAddress:       tokenAddr,
		Amount:             amount,
		Nonce:              nonceHash,
		PaymentCommitment:  expectedHash,
		ChainID:            chainID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected commitment to be valid")
	}
}

func TestVerifyPaymentCommitment_Invalid(t *testing.T) {
	valid, err := VerifyPaymentCommitment(VerifyParams{
		IdentityCommitment: common.HexToHash("0x01"),
		TokenAddress:       common.HexToAddress("0x01"),
		Amount:             big.NewInt(100),
		Nonce:              common.HexToHash("0x42"),
		PaymentCommitment:  common.HexToHash("0xdeadbeef"), // wrong
		ChainID:            big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected commitment to be invalid")
	}
}

func TestVerifyPaymentCommitment_RequiresChainID(t *testing.T) {
	_, err := VerifyPaymentCommitment(VerifyParams{
		IdentityCommitment: common.HexToHash("0x01"),
		TokenAddress:       common.HexToAddress("0x01"),
		Amount:             big.NewInt(100),
		Nonce:              common.HexToHash("0x42"),
		PaymentCommitment:  common.HexToHash("0x01"),
	})
	if err == nil {
		t.Error("expected error for nil ChainID")
	}
}
