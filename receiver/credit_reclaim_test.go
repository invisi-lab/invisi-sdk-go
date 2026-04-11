package receiver

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBuildCreditReclaimCalldata_EncodesCorrectly(t *testing.T) {
	result, err := BuildCreditReclaimCalldata(CreditReclaimParams{
		Nullifier:        common.HexToHash("0xababababababababababababababababababababababababababababababababab"),
		Amount:           big.NewInt(100_000_000),
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		MerkleRoot:       common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Proof:            common.Hex2Bytes("deadbeef"),
		TreeID:           big.NewInt(1),
		Expiry:           big.NewInt(9999999999),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}

func TestBuildCreditReclaimCalldata_NilAmountErrors(t *testing.T) {
	_, err := BuildCreditReclaimCalldata(CreditReclaimParams{
		Nullifier:        common.Hash{},
		Amount:           nil,
		ReceivingAddress: common.Address{},
		MerkleRoot:       common.Hash{},
		Proof:            []byte{},
		TreeID:           big.NewInt(0),
		Expiry:           big.NewInt(0),
	})
	if err == nil {
		t.Error("expected error for nil amount")
	}
}

func TestBuildCreditReclaimCalldata_CalldataIsReproducible(t *testing.T) {
	params := CreditReclaimParams{
		Nullifier:        common.HexToHash("0xababababababababababababababababababababababababababababababababab"),
		Amount:           big.NewInt(100_000_000),
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		MerkleRoot:       common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Proof:            common.Hex2Bytes("deadbeef"),
		TreeID:           big.NewInt(1),
		Expiry:           big.NewInt(9999999999),
	}

	a, err := BuildCreditReclaimCalldata(params)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BuildCreditReclaimCalldata(params)
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
