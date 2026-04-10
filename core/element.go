package core

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func HexToElement(h string) (fr.Element, error) {
	var e fr.Element
	h = strings.TrimPrefix(h, "0x")
	h = strings.TrimPrefix(h, "0X")
	if len(h)%2 != 0 {
		h = "0" + h
	}
	b, err := hex.DecodeString(h)
	if err != nil {
		return e, fmt.Errorf("invalid hex: %w", err)
	}
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	if err := e.SetBytesCanonical(padded); err != nil {
		return e, fmt.Errorf("value out of BN254 field range: %w", err)
	}
	return e, nil
}

func ElementToHex(e fr.Element) string {
	b := e.Bytes()
	return "0x" + hex.EncodeToString(b[:])
}

func BigIntToElement(b *big.Int) fr.Element {
	var e fr.Element
	e.SetBigInt(b)
	return e
}

func ElementToBigInt(e fr.Element) *big.Int {
	var b big.Int
	e.BigInt(&b)
	return &b
}

func ElementFromUint64(v uint64) fr.Element {
	var e fr.Element
	e.SetUint64(v)
	return e
}
