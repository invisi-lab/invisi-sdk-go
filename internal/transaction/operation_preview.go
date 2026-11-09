package transaction

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

type OperationPermissions struct{ WalletAuthorization, Proof, FundingAuthorization bool }

type OperationPreview struct {
	ID, Kind, Status, Target        string
	Context                         protocol.TransactionContext
	TokenAddress                    common.Address
	Capacity                        int
	InputCount, OutputCount         int
	FundingAmount, WithdrawalAmount *big.Int
	FundingPayer, WithdrawalAddress common.Address
	NeedsWalletAuthorization        bool
	Permissions                     OperationPermissions
	Deadline                        uint64
	FeeRecipient                    common.Address
	FeeAmount                       *big.Int
	Outputs                         []PaymentPreview
}

type PaymentPreview struct {
	To     string
	Amount *big.Int
	Change bool
}

func PublicOperationKind(p Plan) string {

	if p.Kind != "transact" {
		return p.Kind
	}
	return "transfer"
}

func (o *Execution) Kind() string { return PublicOperationKind(o.Plan) }

func (p *PreparedTransaction) ID() string { return p.operation.ID() }

func (p *PreparedTransaction) Kind() string { return p.operation.Kind() }

func (o *Execution) Preview() OperationPreview {
	o.mu.Lock()
	defer o.mu.Unlock()
	p := o.Plan
	amount := new(big.Int)
	address := p.Params.WithdrawalAddress
	if p.Params.WithdrawalAmount != nil {
		amount.Set(p.Params.WithdrawalAmount)
	}

	feeRecipient, fee := PlanFee(p)

	context, _ := Frozen(p.Context)
	funding := new(big.Int)
	if p.Funding.Amount != nil {
		funding.Set(p.Funding.Amount)
	}
	result := OperationPreview{ID: p.ID, Kind: PublicOperationKind(p), Status: o.Status, Target: p.Target, Context: context, TokenAddress: p.Context.TokenAddress, Capacity: p.Capacity,
		InputCount: len(p.Inputs), OutputCount: len(p.Outputs), FundingAmount: funding, FundingPayer: p.Funding.Payer,
		WithdrawalAmount: amount, WithdrawalAddress: address, NeedsWalletAuthorization: p.NeedsAuthorization, Deadline: p.Deadline, FeeRecipient: feeRecipient, FeeAmount: fee,
		Permissions: OperationPermissions{p.NeedsAuthorization, p.Kind != "shield", p.FundingMode == "signature"}}
	for _, out := range p.Outputs {
		result.Outputs = append(result.Outputs, PaymentPreview{out.Address, new(big.Int).Set(out.Amount), out.Change})
	}

	return result
}
