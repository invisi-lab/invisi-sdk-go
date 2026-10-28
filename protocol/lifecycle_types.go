package protocol

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

type BuildWithdrawalCalldataParams struct {
	Context     TransactionContext
	Proof       []byte
	Params      CrossTreeTransactParams
	Funding     FundingParams
	Caller      *common.Address
	SourceNotes []core.Note
}
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
