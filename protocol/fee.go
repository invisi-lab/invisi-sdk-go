package protocol

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func normalizeFee(recipient common.Address, amount *big.Int) (*big.Int, error) {
	if amount == nil {
		amount = new(big.Int)
	}
	if err := CheckUint(amount, 128, "feeAmount"); err != nil {
		return nil, err
	}
	if (recipient == common.Address{}) != (amount.Sign() == 0) {
		return nil, fmt.Errorf("invalid fee context")
	}
	return new(big.Int).Set(amount), nil
}

func feeAmount(v *big.Int) *big.Int {
	if v == nil {
		return new(big.Int)
	}
	return v
}
