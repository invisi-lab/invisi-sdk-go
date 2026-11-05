package transaction

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type CreateCreditPreparation struct {
	Payload                               []byte
	TokenAddress                          common.Address
	Spender                               string
	Amount, Expiry                        *big.Int
	SingleLimit, DailyLimit, MonthlyLimit *big.Int
	Funding                               *FundingSource
	Deadline                              uint64
}

type CreditDrawInput struct {
	// Omission creates an empty payload; consumed metadata is never inherited.
	Payload []byte
	Credit  common.Hash
	Amount  *big.Int
}

type DrawCreditPreparation struct {
	TokenAddress common.Address
	Credits      []CreditDrawInput
	Deadline     uint64
}

type CancelCreditPreparation struct {
	TokenAddress common.Address
	Credit       common.Hash
	Deadline     uint64
}

type TopUpCreditPreparation struct {
	// Omission creates an empty payload; consumed metadata is never inherited.
	Payload                  []byte
	TokenAddress             common.Address
	Credit                   common.Hash
	AdditionalAmount, Expiry *big.Int
	// Nil inherits the existing limit; zero explicitly disables its check without clearing usage.
	SingleLimit, DailyLimit, MonthlyLimit *big.Int
	Funding                               *FundingSource
	Deadline                              uint64
}
