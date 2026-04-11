package receiver

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBuildWithdrawCalldata_EncodesCorrectly(t *testing.T) {
	result, err := BuildWithdrawCalldata(WithdrawParams{
		Nullifiers:       []common.Hash{common.HexToHash("0xababababababababababababababababababababababababababababababababab")},
		TotalAmount:      big.NewInt(100_000_000),
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		MerkleRoot:       common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Proof:            common.Hex2Bytes("deadbeef"),
		TreeID:           big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Data must start with a 4-byte function selector
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}

	// Passthrough fields
	if len(result.Nullifiers) != 1 {
		t.Errorf("Nullifiers length = %d, want 1", len(result.Nullifiers))
	}
	if result.TotalAmount.Cmp(big.NewInt(100_000_000)) != 0 {
		t.Errorf("TotalAmount = %s, want 100000000", result.TotalAmount)
	}
}

func TestBuildWithdrawCalldata_MultipleNullifiers(t *testing.T) {
	nullifiers := []common.Hash{
		common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
		common.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333"),
	}
	result, err := BuildWithdrawCalldata(WithdrawParams{
		Nullifiers:       nullifiers,
		TotalAmount:      big.NewInt(350_000_000),
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		MerkleRoot:       common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Proof:            common.Hex2Bytes("deadbeef"),
		TreeID:           big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Nullifiers) != 3 {
		t.Errorf("Nullifiers length = %d, want 3", len(result.Nullifiers))
	}
	if result.TotalAmount.Cmp(big.NewInt(350_000_000)) != 0 {
		t.Errorf("TotalAmount = %s, want 350000000", result.TotalAmount)
	}
}

func TestBuildWithdrawCalldata_EmptyNullifiers(t *testing.T) {
	_, err := BuildWithdrawCalldata(WithdrawParams{
		Nullifiers:       []common.Hash{},
		TotalAmount:      big.NewInt(0),
		ReceivingAddress: common.Address{},
		MerkleRoot:       common.Hash{},
		Proof:            []byte{},
		TreeID:           big.NewInt(0),
	})
	if err == nil {
		t.Error("expected error for empty nullifiers")
	}
}
