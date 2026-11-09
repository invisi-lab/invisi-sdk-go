package core

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func checkUint(value fr.Element, bits int, name string) error {
	if ElementToBigInt(value).BitLen() > bits {
		return fmt.Errorf("%s out of uint%d range", name, bits)
	}
	return nil
}
func validateProtocolContext(context ProtocolContext, tokenAddress fr.Element) error {
	for _, v := range []struct {
		value fr.Element
		bits  int
		name  string
	}{{context.ChainID, 64, "chainId"}, {context.ProtocolAddress, 160, "protocolAddress"}, {tokenAddress, 160, "tokenAddress"}} {
		if err := checkUint(v.value, v.bits, v.name); err != nil {
			return err
		}
	}
	return nil
}
