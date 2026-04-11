package payer

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBuildApproveCalldata_Basic(t *testing.T) {
	token := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
	protocol := common.HexToAddress("0xdeadbeef")

	result, err := BuildApproveCalldata(ApproveParams{
		ProtocolAddress: protocol,
		Token:           token,
		Amount:          big.NewInt(100_000_000),
	})
	if err != nil {
		t.Fatal(err)
	}

	// To should be the token address
	if result.To != token {
		t.Errorf("To = %s, want %s", result.To.Hex(), token.Hex())
	}

	// CallData should start with approve selector: 0x095ea7b3
	if len(result.CallData) < 4 {
		t.Fatal("calldata too short")
	}
	// approve(address,uint256) = keccak256("approve(address,uint256)")[:4] = 0x095ea7b3
	expectedSelector := []byte{0x09, 0x5e, 0xa7, 0xb3}
	for i := 0; i < 4; i++ {
		if result.CallData[i] != expectedSelector[i] {
			t.Errorf("selector byte %d: got 0x%02x, want 0x%02x",
				i, result.CallData[i], expectedSelector[i])
		}
	}

	// Total length: 4 (selector) + 32 (address) + 32 (amount) = 68
	if len(result.CallData) != 68 {
		t.Errorf("calldata length = %d, want 68", len(result.CallData))
	}
}

func TestBuildApproveCalldata_RequiresAmount(t *testing.T) {
	_, err := BuildApproveCalldata(ApproveParams{
		ProtocolAddress: common.HexToAddress("0x01"),
		Token:           common.HexToAddress("0x02"),
	})
	if err == nil {
		t.Error("expected error for nil Amount")
	}
}
