package protocol

import "github.com/ethereum/go-ethereum/common"

type BuildCreditDrawCalldataParams struct {
	Context    TransactionContext
	Proof      []byte
	Params     CreditDrawParams
	Deliveries []Delivery
}
type BuildCancelCreditCalldataParams struct {
	Context TransactionContext
	Proof   []byte
	Params  CancelCreditParams
}
type BuildTopUpCreditCalldataParams struct {
	Context    TransactionContext
	Proof      []byte
	Params     TopUpCreditParams
	Deliveries []Delivery
	Funding    FundingParams

	Caller *common.Address
}
