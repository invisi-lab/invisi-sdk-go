package receiver

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBuildRefundPayOpenCalldata_EncodesCorrectly(t *testing.T) {
	result, err := BuildRefundPayOpenCalldata(RefundPayOpenParams{
		IdentityCommitment: common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		RefundAmount:       big.NewInt(500_000),
		Nonce:              common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		TreeID:             big.NewInt(1),
		RefundLink:         common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Data must start with a 4-byte function selector
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}

func TestBuildRefundPayOpenCalldata_NilAmountErrors(t *testing.T) {
	_, err := BuildRefundPayOpenCalldata(RefundPayOpenParams{
		IdentityCommitment: common.Hash{},
		RefundAmount:       nil,
		Nonce:              common.Hash{},
		TreeID:             big.NewInt(0),
		RefundLink:         common.Hash{},
	})
	if err == nil {
		t.Error("expected error for nil refund amount")
	}
}

func TestBuildRefundPayOpenCalldata_ZeroAmountErrors(t *testing.T) {
	_, err := BuildRefundPayOpenCalldata(RefundPayOpenParams{
		IdentityCommitment: common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		RefundAmount:       big.NewInt(0),
		Nonce:              common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		TreeID:             big.NewInt(1),
		RefundLink:         common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
	})
	if err == nil {
		t.Error("expected error for zero refund amount")
	}
}

func TestBuildRefundPayOpenCalldata_CalldataIsReproducible(t *testing.T) {
	params := RefundPayOpenParams{
		IdentityCommitment: common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		RefundAmount:       big.NewInt(500_000),
		Nonce:              common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		TreeID:             big.NewInt(1),
		RefundLink:         common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
	}

	a, err := BuildRefundPayOpenCalldata(params)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BuildRefundPayOpenCalldata(params)
	if err != nil {
		t.Fatal(err)
	}

	if len(a.Data) != len(b.Data) {
		t.Fatal("calldata length mismatch")
	}
	for i := range a.Data {
		if a.Data[i] != b.Data[i] {
			t.Fatalf("calldata byte mismatch at index %d", i)
		}
	}
}
