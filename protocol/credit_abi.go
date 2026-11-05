package protocol

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func creditNumber(v [32]byte) *big.Int { return new(big.Int).SetBytes(v[:]) }

func creditNumbers(values [][32]byte) []*big.Int {
	v := make([]*big.Int, len(values))
	for i, w := range values {
		v[i] = creditNumber(w)
	}
	return v
}

func creditPadded(v []*big.Int, n int) []*big.Int {
	out := make([]*big.Int, n)
	for i := range out {
		out[i] = new(big.Int)
		if i < len(v) {
			out[i].Set(v[i])
		}
	}
	return out
}

func creditWords(v []*big.Int) [][32]byte {
	out := make([][32]byte, len(v))
	for i, n := range v {
		out[i] = Word(n)
	}
	return out
}

func creditABIType(name string) abi.Type {
	v, err := abi.NewType(name, "", nil)
	if err != nil {
		panic(err)
	}
	return v
}

func creditMethodTuple(name string, index int) (abi.Type, error) {
	a, err := ProtocolABI()
	if err != nil {
		return abi.Type{}, err
	}
	m, ok := a.Methods[name]
	if !ok || index >= len(m.Inputs) {
		return abi.Type{}, fmt.Errorf("current ABI missing %s tuple", name)
	}
	return m.Inputs[index].Type, nil
}
