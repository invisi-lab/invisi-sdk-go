package transaction

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestPublicFeeNormalization(t *testing.T) {
	recipient := common.HexToAddress("0x1234")
	amount := big.NewInt(9)
	fee, err := normalizePublicFee(&Fee{Recipient: recipient, Amount: amount})
	if err != nil || fee.Amount.Cmp(amount) != 0 {
		t.Fatal(fee, err)
	}
	amount.SetInt64(12)
	if fee.Amount.Int64() != 9 {
		t.Fatal("fee retained caller amount")
	}
	for _, value := range []*Fee{{Recipient: recipient}, {Amount: big.NewInt(1)}, {Recipient: recipient, Amount: big.NewInt(-1)}} {
		if _, err := normalizePublicFee(value); err == nil {
			t.Fatal("accepted malformed fee")
		}
	}
}
