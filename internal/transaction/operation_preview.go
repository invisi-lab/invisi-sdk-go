package transaction

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
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
	CreditOutputs                   []CreditPreview
}

type PaymentPreview struct {
	To     string
	Amount *big.Int
	Change bool
}

type CreditPreview struct {
	Owner, Spender, OwnerPayout, SpenderPayout                      common.Address
	Amount, Expiry                                                  *big.Int
	SingleLimit, DailyLimit, MonthlyLimit, DailySpent, MonthlySpent *big.Int
	LastDay, LastMonth                                              uint64
}

func PublicOperationKind(p Plan) string {
	if p.Kind == "credit-draw" {
		return "draw-credit"
	}
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
	if p.DrawCredit != nil {
		amount.Set(p.DrawCredit.TotalAmount)
		address = p.DrawCredit.Spender
	}
	if p.CancelCredit != nil {
		amount.Set(p.CancelCredit.Amount)
		address = p.CancelCredit.Owner
	}
	feeRecipient, fee := PlanFee(p)
	if p.CancelCredit != nil {
		amount.Sub(amount, fee)
	}
	context, _ := Frozen(p.Context)
	funding := new(big.Int)
	if p.Funding.Amount != nil {
		funding.Set(p.Funding.Amount)
	}
	result := OperationPreview{ID: p.ID, Kind: PublicOperationKind(p), Status: o.Status, Target: p.Target, Context: context, TokenAddress: p.Context.TokenAddress, Capacity: p.Capacity,
		InputCount: len(p.Inputs) + len(p.CreditInputs), OutputCount: len(p.Outputs), FundingAmount: funding, FundingPayer: p.Funding.Payer,
		WithdrawalAmount: amount, WithdrawalAddress: address, NeedsWalletAuthorization: p.NeedsAuthorization, Deadline: p.Deadline, FeeRecipient: feeRecipient, FeeAmount: fee,
		Permissions: OperationPermissions{p.NeedsAuthorization, p.Kind != "shield", p.FundingMode == "signature"}}
	for _, out := range p.Outputs {
		result.Outputs = append(result.Outputs, PaymentPreview{out.Address, new(big.Int).Set(out.Amount), out.Change})
	}
	appendCredit := func(note core.CreditNote) {
		result.CreditOutputs = append(result.CreditOutputs, CreditPreview{Owner: common.BigToAddress(core.ElementToBigInt(note.OwnerWalletIdentity)), Spender: common.BigToAddress(core.ElementToBigInt(note.SpenderWalletIdentity)),
			OwnerPayout: common.BigToAddress(core.ElementToBigInt(note.OwnerPayout)), SpenderPayout: common.BigToAddress(core.ElementToBigInt(note.SpenderPayout)), Amount: core.ElementToBigInt(note.Amount), Expiry: core.ElementToBigInt(note.Expiry), SingleLimit: core.ElementToBigInt(note.SingleLimit), DailyLimit: core.ElementToBigInt(note.DailyLimit), MonthlyLimit: core.ElementToBigInt(note.MonthlyLimit), DailySpent: core.ElementToBigInt(note.DailySpent), MonthlySpent: core.ElementToBigInt(note.MonthlySpent), LastDay: core.ElementToBigInt(note.LastDay).Uint64(), LastMonth: core.ElementToBigInt(note.LastMonth).Uint64()})
	}
	if p.CreditOutput != nil {
		appendCredit(p.CreditOutput.Note)
	}
	if p.DrawCredit != nil {
		for i, hash := range p.DrawCredit.RemainderCommitments {
			if hash != ([32]byte{}) {
				child := p.CreditInputs[i].State.Note
				if updated, err := core.ApplyCreditDrawUsage(child, core.BigIntToElement(p.DrawAmounts[i]), p.Block.Timestamp); err == nil {
					child = updated
				}
				amount := new(big.Int).Sub(core.ElementToBigInt(child.Amount), p.DrawAmounts[i])
				child.Amount = core.BigIntToElement(amount)
				appendCredit(child)
			}
		}
	}
	return result
}
