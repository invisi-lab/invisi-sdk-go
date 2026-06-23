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

func TestVerifyCreditCommitment_Valid(t *testing.T) {
	// Build a known set of credit-commitment inputs.
	var chainID, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash fr.Element
	chainID.SetUint64(84532)
	identityCommitment.SetUint64(0xab)
	payerCommitment.SetUint64(0xcd)
	tokenAddress.SetUint64(0xef)
	amount.SetUint64(1_000_000)
	nonce.SetUint64(42)
	expiry.SetUint64(1700000000)
	nullifierKeyHash.SetUint64(0x99)

	// Compute the expected commitment via core (8 inputs).
	expected, err := core.ComputeCreditCommitment(
		chainID, identityCommitment, payerCommitment, tokenAddress,
		amount, nonce, expiry, nullifierKeyHash,
	)
	if err != nil {
		t.Fatal(err)
	}

	params := VerifyCreditParams{
		ChainID:            chainID,
		IdentityCommitment: identityCommitment,
		PayerCommitment:    payerCommitment,
		TokenAddress:       tokenAddress,
		Amount:             amount,
		Nonce:              nonce,
		Expiry:             expiry,
		NullifierKeyHash:   nullifierKeyHash,
		CreditCommitment:   expected,
	}

	valid, err := VerifyCreditCommitment(params)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("expected credit commitment to be valid")
	}
}

func TestVerifyCreditCommitment_Invalid(t *testing.T) {
	var chainID, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash fr.Element
	chainID.SetUint64(84532)
	identityCommitment.SetUint64(0xab)
	payerCommitment.SetUint64(0xcd)
	tokenAddress.SetUint64(0xef)
	amount.SetUint64(1_000_000)
	nonce.SetUint64(42)
	expiry.SetUint64(1700000000)
	nullifierKeyHash.SetUint64(0x99)

	// Expected commitment is computed with the original nonce.
	expected, err := core.ComputeCreditCommitment(
		chainID, identityCommitment, payerCommitment, tokenAddress,
		amount, nonce, expiry, nullifierKeyHash,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper the nonce: the recomputed commitment must no longer match.
	var tamperedNonce fr.Element
	tamperedNonce.SetUint64(43)

	valid, err := VerifyCreditCommitment(VerifyCreditParams{
		ChainID:            chainID,
		IdentityCommitment: identityCommitment,
		PayerCommitment:    payerCommitment,
		TokenAddress:       tokenAddress,
		Amount:             amount,
		Nonce:              tamperedNonce,
		Expiry:             expiry,
		NullifierKeyHash:   nullifierKeyHash,
		CreditCommitment:   expected,
	})
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected credit commitment to be invalid after nonce tampering")
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
