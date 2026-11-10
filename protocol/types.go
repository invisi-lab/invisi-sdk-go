package protocol

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

// Contract tuples use the generated, current ABI types.
type Delivery = contracts.IInvisiProtocolDelivery
type InputSource = contracts.IInvisiProtocolInputSource
type TokenPermit = contracts.IMetaTxTokenPermit
type FundingParams = contracts.IInvisiProtocolFundingParams
type ShieldParams = contracts.IInvisiProtocolShieldParams
type ShieldOutput = contracts.IInvisiProtocolShieldOutput
type CrossTreeTransactParams = contracts.IInvisiProtocolCrossTreeTransactParams
type TransactRequest = contracts.IInvisiProtocolTransactRequest

type TransactionContext struct {
	ChainID         *big.Int
	ProtocolAddress common.Address
	TokenAddress    common.Address
	Now             *big.Int // Chain block timestamp; required for Credit draw statements, optional for other deadline/expiry checks.
}
type Calldata struct {
	Data          []byte
	ApproveAmount *big.Int
	Capacity      int
}
type BuildTransactCalldataParams struct {
	Context    TransactionContext
	Proof      []byte
	Params     CrossTreeTransactParams
	Deliveries []Delivery
	Funding    FundingParams
	Caller     *common.Address
}
type BuildTransactAggregateCalldataParams struct {
	Context  TransactionContext
	Proof    []byte
	Requests []TransactRequest
	Caller   *common.Address
}
type BuildShieldCalldataParams struct {
	Context    TransactionContext
	Params     ShieldParams
	Outputs    []ShieldOutput
	Deliveries []Delivery
	Funding    FundingParams
	Caller     *common.Address
}

type ApproveParams struct {
	ProtocolAddress, Token common.Address
	Amount                 *big.Int
}
type ApproveCalldata struct {
	To       common.Address
	CallData []byte
}

type AggregateRequest struct {
	Proof    []byte
	Requests []TransactRequest
	Caller   *common.Address
}
type TransactionCall struct {
	To    common.Address
	Data  []byte
	Value *big.Int
}
