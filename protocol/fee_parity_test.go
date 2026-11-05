package protocol

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestFeePairAndEmptyFundsDelivery(t *testing.T) {
	if err := ValidateDelivery(Delivery{}, false, false); err != nil {
		t.Fatalf("funds empty pair: %v", err)
	}
	if err := ValidateDelivery(Delivery{}, true, false); err == nil {
		t.Fatal("credit delivery required")
	}
	p := CrossTreeTransactParams{Token: common.HexToAddress("0x1"), Nullifiers: [][32]byte{Word(big.NewInt(1))}, Sources: []InputSource{{TreeId: big.NewInt(1), MerkleRoot: Word(big.NewInt(2))}}, WithdrawalAddress: common.HexToAddress("0x2"), WithdrawalAmount: big.NewInt(1), Deadline: big.NewInt(2), FeeRecipient: common.HexToAddress("0x3"), FeeAmount: big.NewInt(4)}
	w, e := TransactPublicInputs(TransactionContext{ChainID: big.NewInt(1)}, p, nil, FundingParams{})
	if e != nil {
		t.Fatal(e)
	}
	if len(w) != 57 || w[len(w)-1] != Word(big.NewInt(4)) {
		t.Fatal("fee missing from statement")
	}
}

func TestFeeNormalizationRejectsUnpairedAndOutOfRange(t *testing.T) {
	for _, test := range []struct {
		recipient common.Address
		amount    *big.Int
	}{{common.HexToAddress("0x1"), nil}, {common.Address{}, big.NewInt(1)}, {common.HexToAddress("0x1"), new(big.Int).Lsh(big.NewInt(1), 128)}} {
		if _, err := normalizeFee(test.recipient, test.amount); err == nil {
			t.Fatal("invalid fee accepted")
		}
	}
	if err := ValidateDelivery(Delivery{Ciphertext: []byte{1}}, false, false); err == nil {
		t.Fatal("half delivery accepted")
	}
}
