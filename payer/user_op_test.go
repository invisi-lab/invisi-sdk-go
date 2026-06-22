package payer

import (
	"math/big"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

func TestBuildPayOpenUserOp(t *testing.T) {
	var nonce fr.Element
	nonce.SetUint64(1)

	result, err := BuildPayOpenUserOp(PayOpenUserOpParams{
		IdentityCommitment: common.HexToHash("0x" + strings.Repeat("00", 31) + "ab"),
		TokenAddress:       common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
		Amount:             big.NewInt(100_000_000),
		TreeID:             big.NewInt(1),
		Nonce:              &nonce,
		ProtocolAddress:    common.HexToAddress("0xdeadbeef"),
		Sender:             common.HexToAddress("0xcafebabe"),
		ChainID:            big.NewInt(84532),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify BaseUserOp fields
	if result.To != common.HexToAddress("0xdeadbeef") {
		t.Errorf("To = %s, want protocol address", result.To.Hex())
	}
	if result.Sender != common.HexToAddress("0xcafebabe") {
		t.Errorf("Sender = %s, want sender address", result.Sender.Hex())
	}
	if len(result.CallData) < 4 {
		t.Fatal("callData too short")
	}

	// Gas estimates should match PayOpen defaults
	if result.CallGasLimit != GasEstimatePayOpen.CallGasLimit {
		t.Errorf("callGasLimit = %d, want %d", result.CallGasLimit, GasEstimatePayOpen.CallGasLimit)
	}
	if result.VerificationGasLimit != GasEstimatePayOpen.VerificationGasLimit {
		t.Errorf("verificationGasLimit = %d, want %d",
			result.VerificationGasLimit, GasEstimatePayOpen.VerificationGasLimit)
	}

	// PayOpen-specific fields
	if result.PaymentCommitment == (common.Hash{}) {
		t.Error("paymentCommitment should not be zero")
	}
	expectedNonce := common.HexToHash(core.ElementToHex(nonce))
	if result.Nonce != expectedNonce {
		t.Errorf("nonce = %s, want %s", result.Nonce.Hex(), expectedNonce.Hex())
	}

	// Paymaster should be nil by default
	if result.Paymaster != nil {
		t.Error("paymaster should be nil when not provided")
	}
}

func TestBuildPayOpenUserOp_WithPaymaster(t *testing.T) {
	var nonce fr.Element
	nonce.SetUint64(1)

	pm := &PaymasterConfig{URL: "https://pm.example.com", PolicyID: "test-policy"}
	result, err := BuildPayOpenUserOp(PayOpenUserOpParams{
		IdentityCommitment: common.HexToHash("0x01"),
		TokenAddress:       common.HexToAddress("0x01"),
		Amount:             big.NewInt(100),
		TreeID:             big.NewInt(1),
		Nonce:              &nonce,
		ProtocolAddress:    common.HexToAddress("0x02"),
		Sender:             common.HexToAddress("0x03"),
		ChainID:            big.NewInt(1),
		Paymaster:          pm,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Paymaster == nil {
		t.Fatal("paymaster should be set")
	}
	if result.Paymaster.URL != "https://pm.example.com" {
		t.Errorf("paymaster URL = %s", result.Paymaster.URL)
	}
}

func TestBuildPayBlindUserOp(t *testing.T) {
	result, err := BuildPayBlindUserOp(PayBlindUserOpParams{
		Proof:             []byte{0xde, 0xad},
		PaymentCommitment: common.HexToHash("0x" + "aa" + strings.Repeat("00", 31)),
		ChangeCommitment:  common.HexToHash("0x" + "bb" + strings.Repeat("00", 31)),
		InputAmount:       big.NewInt(200),
		TreeID:            big.NewInt(1),
		ProtocolAddress:   common.HexToAddress("0xdeadbeef"),
		Sender:            common.HexToAddress("0xcafebabe"),
		ChainID:           big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.CallGasLimit != GasEstimatePayBlind.CallGasLimit {
		t.Errorf("callGasLimit = %d, want %d", result.CallGasLimit, GasEstimatePayBlind.CallGasLimit)
	}
	if len(result.CallData) < 4 {
		t.Fatal("callData too short")
	}
}

func TestBuildDepositCreditUserOp(t *testing.T) {
	result, err := BuildDepositCreditUserOp(DepositCreditUserOpParams{
		IdentityCommitment: common.HexToHash("0x01"),
		PayerCommitment:    common.HexToHash("0x02"),
		Amount:             big.NewInt(100),
		Nonce:              common.HexToHash("0x03"),
		Expiry:             big.NewInt(999),
		NullifierKeyHash:   common.HexToHash("0x04"),
		TreeID:             big.NewInt(1),
		ProtocolAddress:    common.HexToAddress("0xdeadbeef"),
		Sender:             common.HexToAddress("0xcafebabe"),
		ChainID:            big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.CallGasLimit != GasEstimateDepositCredit.CallGasLimit {
		t.Errorf("callGasLimit = %d, want %d",
			result.CallGasLimit, GasEstimateDepositCredit.CallGasLimit)
	}
	if len(result.CallData) < 4 {
		t.Fatal("callData too short")
	}
}
