package payer

import (
	"math/big"
	"strings"
	"testing"

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
