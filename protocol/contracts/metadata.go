package contracts

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// MetaData stores only the public ABI; network bindings live in the RPC module.
type MetaData struct{ ABI string }

func (m *MetaData) GetAbi() (*abi.ABI, error) {
	a, e := abi.JSON(strings.NewReader(m.ABI))
	if e != nil {
		return nil, e
	}
	return &a, nil
}
