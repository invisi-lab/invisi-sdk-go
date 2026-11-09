package credit

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestCreditFeePairAndWidth(t *testing.T) {
	for _, tc := range []struct {
		recipient common.Address
		amount    *big.Int
	}{{common.Address{}, big.NewInt(1)}, {common.HexToAddress("0x01"), nil}, {common.HexToAddress("0x01"), new(big.Int).Lsh(big.NewInt(1), 128)}, {common.Address{}, big.NewInt(-1)}} {
		if _, err := normalizeFee(tc.recipient, tc.amount); err == nil {
			t.Fatal("accepted invalid Credit fee pair")
		}
	}
	if fee, err := normalizeFee(common.Address{}, nil); err != nil || fee.Sign() != 0 {
		t.Fatal("zero fee default", err)
	}
}
