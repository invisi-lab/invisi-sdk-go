package primitives

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/invisi-lab/invisi-sdk-go/primitives/internal/cryptography/field"
)

func HexToElement(h string) (fr.Element, error) { return field.HexToElement(h) }
func ElementToHex(e fr.Element) string          { return field.ElementToHex(e) }
func BigIntToElement(b *big.Int) fr.Element     { return field.BigIntToElement(b) }
func ElementToBigInt(e fr.Element) *big.Int     { return field.ElementToBigInt(e) }
func ElementFromUint64(v uint64) fr.Element     { return field.ElementFromUint64(v) }
