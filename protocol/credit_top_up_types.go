package protocol

import "github.com/ethereum/go-ethereum/common"

type BuildTopUpCreditCalldataParams struct {
	Context    TransactionContext
	Proof      []byte
	Params     TopUpCreditParams
	Deliveries []Delivery
	Funding    FundingParams

	Caller *common.Address
}
