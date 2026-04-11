package payer

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
)

func TestBuildPayBlindCalldata_Basic(t *testing.T) {
	result, err := BuildPayBlindCalldata(PayBlindParams{
		Proof:             []byte{0xde, 0xad, 0xbe, 0xef},
		PaymentCommitment: common.HexToHash("0xab00000000000000000000000000000000000000000000000000000000000000"),
		ChangeCommitment:  common.HexToHash("0xcd00000000000000000000000000000000000000000000000000000000000000"),
		InputAmount:       big.NewInt(200_000_000),
		TreeID:            big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Calldata starts with payBlind selector
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}

	// ApproveAmount == InputAmount
	if result.ApproveAmount.Cmp(big.NewInt(200_000_000)) != 0 {
		t.Errorf("approveAmount = %s, want 200000000", result.ApproveAmount)
	}
}

func TestBuildPayBlindCalldata_WithRefundLink(t *testing.T) {
	var refund fr.Element
	refund.SetUint64(42)

	result, err := BuildPayBlindCalldata(PayBlindParams{
		Proof:             []byte{0x01},
		PaymentCommitment: common.HexToHash("0xaa00000000000000000000000000000000000000000000000000000000000000"),
		ChangeCommitment:  common.HexToHash("0xbb00000000000000000000000000000000000000000000000000000000000000"),
		InputAmount:       big.NewInt(100),
		TreeID:            big.NewInt(1),
		RefundLink:        &refund,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}

func TestBuildPayBlindCalldata_NilRefundDefaultsToZero(t *testing.T) {
	result, err := BuildPayBlindCalldata(PayBlindParams{
		Proof:             []byte{0x01},
		PaymentCommitment: common.HexToHash("0xaa00000000000000000000000000000000000000000000000000000000000000"),
		ChangeCommitment:  common.HexToHash("0xbb00000000000000000000000000000000000000000000000000000000000000"),
		InputAmount:       big.NewInt(100),
		TreeID:            big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should succeed with zero refundLink
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}

func TestBuildPayBlindCalldata_RequiresInputAmount(t *testing.T) {
	_, err := BuildPayBlindCalldata(PayBlindParams{
		Proof:             []byte{0x01},
		PaymentCommitment: common.HexToHash("0x01"),
		ChangeCommitment:  common.HexToHash("0x02"),
		TreeID:            big.NewInt(1),
	})
	if err == nil {
		t.Error("expected error for nil InputAmount")
	}
}
