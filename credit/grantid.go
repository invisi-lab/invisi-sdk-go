package credit

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// ValidGrantID accepts precisely the positive uint128 grant identifier range.
func ValidGrantID(id *big.Int) bool { return id != nil && id.Sign() > 0 && id.BitLen() <= 128 }

// ParseGrantID accepts the canonical decimal wire representation of a grant ID.
func ParseGrantID(raw string) (*big.Int, error) {
	id, ok := new(big.Int).SetString(raw, 10)
	if !ok || !ValidGrantID(id) || id.String() != raw {
		return nil, fmt.Errorf("grant ID must be canonical positive uint128")
	}
	return id, nil
}

// GenerateGrantID chooses a positive uint128 identifier with cryptographic entropy.
func GenerateGrantID() (*big.Int, error) {
	for {
		var raw [16]byte
		if _, err := rand.Read(raw[:]); err != nil {
			return nil, err
		}
		id := new(big.Int).SetBytes(raw[:])
		if id.Sign() > 0 {
			return id, nil
		}
	}
}
