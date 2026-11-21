package contracts

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// CreditStatusValue validates the three protocol statuses at the ABI boundary.
func (c *InvisiCoreCaller) CreditStatusValue(opts *bind.CallOpts, owner, token common.Address, grantID *big.Int) (uint8, error) {
	if grantID == nil || grantID.Sign() <= 0 || grantID.BitLen() > 128 {
		return 0, fmt.Errorf("invalid positive uint128 grant ID")
	}
	status, err := c.CreditStatus(opts, owner, token, new(big.Int).Set(grantID))
	if err != nil {
		return 0, err
	}
	if status == nil || status.Sign() < 0 || !status.IsUint64() || status.Uint64() > 2 {
		return 0, fmt.Errorf("invalid Credit status")
	}
	return uint8(status.Uint64()), nil
}
