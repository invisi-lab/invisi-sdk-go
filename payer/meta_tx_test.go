package payer

import (
	"math/big"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
)

func TestGetEIP712Domain(t *testing.T) {
	domain := GetEIP712Domain(EIP712DomainParams{
		ChainID:           big.NewInt(1),
		VerifyingContract: common.HexToAddress("0xdeadbeef"),
	})
	if domain.Name != "InvisibleProtocol" {
		t.Errorf("domain.Name = %q, want InvisibleProtocol", domain.Name)
	}
	if domain.Version != "1" {
		t.Errorf("domain.Version = %q, want 1", domain.Version)
	}
	if domain.ChainId == nil {
		t.Fatal("domain.ChainId is nil")
	}
	chainID := (*big.Int)(domain.ChainId)
	if chainID.Int64() != 1 {
		t.Errorf("domain.ChainId = %d, want 1", chainID.Int64())
	}
}

func TestBuildSignPayOpenTypedData(t *testing.T) {
	result, err := BuildSignPayOpenTypedData(SignPayOpenParams{
		ChainID:            big.NewInt(84532),
		VerifyingContract:  common.HexToAddress("0xdeadbeef"),
		IdentityCommitment: common.HexToHash("0x" + strings.Repeat("00", 31) + "ab"),
		TokenAddress:       common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
		Amount:             big.NewInt(100_000_000),
		TreeID:             big.NewInt(1),
		SigNonce:           big.NewInt(0),
		Deadline:           big.NewInt(1700000000),
	})
	if err != nil {
		t.Fatal(err)
	}

	// PaymentCommitment should be non-zero
	if result.PaymentCommitment == (common.Hash{}) {
		t.Error("paymentCommitment should not be zero")
	}

	// Nonce should be non-zero
	if result.Nonce == (common.Hash{}) {
		t.Error("nonce should not be zero")
	}

	// TypedData should have correct primary type
	if result.TypedData.PrimaryType != "PayOpenWithSig" {
		t.Errorf("primaryType = %q, want PayOpenWithSig", result.TypedData.PrimaryType)
	}

	// Domain should be set
	if result.TypedData.Domain.Name != "InvisibleProtocol" {
		t.Error("domain name not set")
	}

	// Message should contain the expected fields
	msg := result.TypedData.Message
	if _, ok := msg["identityCommitment"]; !ok {
		t.Error("message missing identityCommitment")
	}
	if _, ok := msg["sigNonce"]; !ok {
		t.Error("message missing sigNonce")
	}
	if _, ok := msg["deadline"]; !ok {
		t.Error("message missing deadline")
	}
}

func TestBuildSignPayBlindTypedData(t *testing.T) {
	result, err := BuildSignPayBlindTypedData(SignPayBlindParams{
		ChainID:           big.NewInt(1),
		VerifyingContract: common.HexToAddress("0xbeef"),
		Proof:             []byte{0xde, 0xad},
		PaymentCommitment: common.HexToHash("0x" + "aa" + strings.Repeat("00", 31)),
		ChangeCommitment:  common.HexToHash("0x" + "bb" + strings.Repeat("00", 31)),
		InputAmount:       big.NewInt(200),
		TreeID:            big.NewInt(1),
		SigNonce:          big.NewInt(1),
		Deadline:          big.NewInt(1700000000),
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.TypedData.PrimaryType != "PayBlindWithSig" {
		t.Errorf("primaryType = %q, want PayBlindWithSig", result.TypedData.PrimaryType)
	}
}

func TestBuildSignDepositCreditTypedData(t *testing.T) {
	result, err := BuildSignDepositCreditTypedData(SignDepositCreditParams{
		ChainID:            big.NewInt(1),
		VerifyingContract:  common.HexToAddress("0xbeef"),
		IdentityCommitment: common.HexToHash("0x01"),
		PayerCommitment:    common.HexToHash("0x02"),
		Amount:             big.NewInt(100),
		Nonce:              common.HexToHash("0x03"),
		Expiry:             big.NewInt(999),
		NullifierKeyHash:   common.HexToHash("0x04"),
		TreeID:             big.NewInt(1),
		SigNonce:           big.NewInt(0),
		Deadline:           big.NewInt(1700000000),
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.TypedData.PrimaryType != "DepositCreditWithSig" {
		t.Errorf("primaryType = %q, want DepositCreditWithSig", result.TypedData.PrimaryType)
	}

	msg := result.TypedData.Message
	if _, ok := msg["nullifierKeyHash"]; !ok {
		t.Error("message missing nullifierKeyHash")
	}
	if _, ok := msg["expiry"]; !ok {
		t.Error("message missing expiry")
	}
}

func TestBuildSignPayOpenTypedData_NilRefundDefaultsToZero(t *testing.T) {
	result, err := BuildSignPayOpenTypedData(SignPayOpenParams{
		ChainID:            big.NewInt(1),
		VerifyingContract:  common.HexToAddress("0x01"),
		IdentityCommitment: common.HexToHash("0x01"),
		TokenAddress:       common.HexToAddress("0x01"),
		Amount:             big.NewInt(100),
		TreeID:             big.NewInt(1),
		SigNonce:           big.NewInt(0),
		Deadline:           big.NewInt(999),
	})
	if err != nil {
		t.Fatal(err)
	}

	refundLink, ok := result.TypedData.Message["refundLink"].(string)
	if !ok {
		t.Fatal("refundLink missing or not string")
	}
	// Should be zero hash hex
	zeroHash := common.Hash{}.Hex()
	if refundLink != zeroHash {
		t.Errorf("refundLink = %s, want %s", refundLink, zeroHash)
	}
}

// TestBuildSignPayOpenTypedData_InjectNonce verifies that when an explicit Nonce
// is supplied, the result is deterministic and the nonce is reflected in both
// SignPayOpenResult.Nonce and TypedData.Message["nonce"].
func TestBuildSignPayOpenTypedData_InjectNonce(t *testing.T) {
	// Build a fixed fr.Element from a known big.Int.
	var fixedNonce fr.Element
	fixedNonce.SetBigInt(big.NewInt(0xdeadbeef))

	params := SignPayOpenParams{
		ChainID:            big.NewInt(84532),
		VerifyingContract:  common.HexToAddress("0xdeadbeef"),
		IdentityCommitment: common.HexToHash("0x" + strings.Repeat("00", 31) + "ab"),
		TokenAddress:       common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
		Amount:             big.NewInt(100_000_000),
		TreeID:             big.NewInt(1),
		Nonce:              &fixedNonce,
		SigNonce:           big.NewInt(0),
		Deadline:           big.NewInt(1700000000),
	}

	result1, err := BuildSignPayOpenTypedData(params)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := BuildSignPayOpenTypedData(params)
	if err != nil {
		t.Fatal(err)
	}

	// Injected nonce → results must be deterministic.
	if result1.PaymentCommitment != result2.PaymentCommitment {
		t.Errorf("PaymentCommitment not deterministic: %s vs %s",
			result1.PaymentCommitment.Hex(), result2.PaymentCommitment.Hex())
	}
	if result1.Nonce != result2.Nonce {
		t.Errorf("Nonce not deterministic: %s vs %s", result1.Nonce.Hex(), result2.Nonce.Hex())
	}

	// result.Nonce must be non-zero and must match TypedData.Message["nonce"].
	if result1.Nonce == (common.Hash{}) {
		t.Error("Nonce should not be zero")
	}
	msgNonce, ok := result1.TypedData.Message["nonce"].(string)
	if !ok {
		t.Fatal("TypedData.Message[\"nonce\"] missing or not string")
	}
	if msgNonce != result1.Nonce.Hex() {
		t.Errorf("TypedData nonce %s != result.Nonce %s", msgNonce, result1.Nonce.Hex())
	}
}

// TestBuildSignPayOpenTypedData_ParityWithCalldata is the key parity guarantee:
// for the same (chainId, identityCommitment, tokenAddress, amount, treeId, nonce),
// BuildSignPayOpenTypedData and BuildPayOpenCalldata must produce the same
// PaymentCommitment. This proves a relay-mode signature covers the same
// commitment the server computed via the calldata builder.
func TestBuildSignPayOpenTypedData_ParityWithCalldata(t *testing.T) {
	var fixedNonce fr.Element
	fixedNonce.SetBigInt(big.NewInt(0xcafe1234))

	chainID := big.NewInt(84532)
	ic := common.HexToHash("0x" + strings.Repeat("00", 31) + "cd")
	token := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
	amount := big.NewInt(500_000_000)
	treeID := big.NewInt(2)

	signResult, err := BuildSignPayOpenTypedData(SignPayOpenParams{
		ChainID:            chainID,
		VerifyingContract:  common.HexToAddress("0xdeadbeef"),
		IdentityCommitment: ic,
		TokenAddress:       token,
		Amount:             amount,
		TreeID:             treeID,
		Nonce:              &fixedNonce,
		SigNonce:           big.NewInt(0),
		Deadline:           big.NewInt(1700000000),
	})
	if err != nil {
		t.Fatalf("BuildSignPayOpenTypedData: %v", err)
	}

	calldataResult, err := BuildPayOpenCalldata(PayOpenParams{
		ChainID:            chainID,
		IdentityCommitment: ic,
		TokenAddress:       token,
		Amount:             amount,
		TreeID:             treeID,
		Nonce:              &fixedNonce,
	})
	if err != nil {
		t.Fatalf("BuildPayOpenCalldata: %v", err)
	}

	// Core parity assertion: same inputs + same nonce → same PaymentCommitment.
	if signResult.PaymentCommitment != calldataResult.PaymentCommitment {
		t.Errorf("PaymentCommitment mismatch:\n  sign builder:     %s\n  calldata builder: %s",
			signResult.PaymentCommitment.Hex(), calldataResult.PaymentCommitment.Hex())
	}
}
