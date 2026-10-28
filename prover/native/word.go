package native

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func word(s string) (common.Hash, error) {
	base := 10
	if strings.HasPrefix(s, "0x") {
		base = 16
		s = s[2:]
	}
	n, ok := new(big.Int).SetString(s, base)
	if !ok || n.Sign() < 0 || n.BitLen() > 256 {
		return common.Hash{}, fmt.Errorf("invalid proof word")
	}
	return common.BigToHash(n), nil
}
