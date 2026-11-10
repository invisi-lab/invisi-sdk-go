package settlement

import (
	"bytes"
	_ "embed"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed fees.abi.json
var feesABIJSON string
var feesABI = func() abi.ABI {
	a, e := abi.JSON(strings.NewReader(feesABIJSON))
	if e != nil {
		panic(e)
	}
	return a
}()

func FeeWithdrawal(token, recipient common.Address) ([]byte, error) {
	return feesABI.Pack("withdrawBroadcasterFees", token, recipient)
}
func FeeBalanceCall(token, recipient common.Address) ([]byte, error) {
	return feesABI.Pack("broadcasterFeeBalance", token, recipient)
}
func FeeBalance(data []byte) (*big.Int, error) {
	out, e := feesABI.Unpack("broadcasterFeeBalance", data)
	if e != nil || len(out) != 1 {
		return nil, errors.New("invalid fee balance response")
	}
	return new(big.Int).Set(out[0].(*big.Int)), nil
}
func DecodeFeeWithdrawal(data []byte) (common.Address, common.Address, error) {
	var zero common.Address
	method := feesABI.Methods["withdrawBroadcasterFees"]
	if len(data) < 4 || !bytes.Equal(data[:4], method.ID) {
		return zero, zero, errors.New("invalid fee withdrawal selector")
	}
	fields, e := method.Inputs.Unpack(data[4:])
	if e != nil {
		return zero, zero, e
	}
	encoded, e := feesABI.Pack(method.Name, fields...)
	if e != nil || !bytes.Equal(encoded, data) {
		return zero, zero, errors.New("noncanonical fee withdrawal")
	}
	return fields[0].(common.Address), fields[1].(common.Address), nil
}
