package transaction

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type FundingSource struct {
	SigNonce      *big.Int
	Deadline      uint64
	Payer         common.Address
	Amount        *big.Int
	Authorization string
	Permit        *protocol.TokenPermit
}

type ShieldPreparation struct {
	TokenAddress common.Address
	Outputs      []NoteOutput
	Funding      FundingSource
	Deadline     uint64
}

type TransferPreparation struct {
	TokenAddress common.Address
	Outputs      []NoteOutput
	Funding      *FundingSource
	Deadline     uint64
	Withdrawal   *Withdrawal
	Commitments  []common.Hash
}

type Withdrawal struct {
	To     common.Address
	Amount *big.Int
}

type PreparedOutput struct {
	Address    string
	Note       *core.Note
	Amount     *big.Int
	Identity   common.Hash
	Commitment common.Hash
	Delivery   protocol.Delivery
	Change     bool
}

type Plan struct {
	Version            int
	OwnerAddress       common.Address
	Namespace          string
	Kind, ID, Target   string
	Context            protocol.TransactionContext
	Block              protocolrpc.CanonicalBlock
	Deadline           uint64
	Inputs             []NoteRecord
	Memberships        []protocolrpc.Membership
	Outputs            []PreparedOutput
	Funding            protocol.FundingParams
	FundingMode        string
	Params             protocol.CrossTreeTransactParams
	Shield             protocol.ShieldParams
	Capacity           int
	NeedsAuthorization bool
	CreateCredit       *protocol.CreateCreditParams
	DrawCredit         *protocol.CreditDrawParams
	CancelCredit       *protocol.CancelCreditParams
	TopUpCredit        *protocol.TopUpCreditParams
	CreditInputs       []CreditRecord
	CreditMemberships  []protocolrpc.Membership
	CreditOutput       *core.CreditDeliveryState
	Deliveries         []protocol.Delivery
	DrawAmounts        []*big.Int
	Broadcaster        *BroadcasterPlan
}

func ValidatePayload(payload []byte) error { return core.ValidatePayload(payload) }

func Positive(amount *big.Int, bits int) error {
	if amount == nil || amount.Sign() <= 0 || amount.BitLen() > bits {
		return errors.New("amount must be a Positive bounded integer")
	}
	return nil
}

func (s *Runtime) Frame(ctx context.Context, token common.Address, deadline uint64) (protocol.TransactionContext, protocolrpc.CanonicalBlock, uint64, error) {
	var block protocolrpc.CanonicalBlock
	c, e := s.Protocol.ForToken(token)
	if e != nil {
		return c, block, 0, e
	}
	if e = s.Protocol.ValidateDeployment(ctx, s.Options.Provider); e != nil {
		return c, block, 0, e
	}
	head, e := s.Options.Provider.HeaderByNumber(ctx, nil)
	if e != nil {
		return c, block, 0, e
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() {
		return c, block, 0, errors.New("invalid latest block")
	}
	block = protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	if deadline == 0 {
		deadline = head.Time + 300
	}
	if deadline < head.Time || deadline-head.Time > 1<<53 {
		return c, block, 0, errors.New("invalid deadline")
	}
	c.Now = new(big.Int).SetUint64(head.Time)
	code, e := s.Options.Provider.CodeAt(ctx, token, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return c, block, 0, e
	}
	if len(code) == 0 {
		return c, block, 0, errors.New("token address must have contract code")
	}
	state, e := s.PreparationInventory(ctx)
	if e != nil {
		return c, block, 0, e
	}
	if state.Invalidated {
		return c, block, 0, errors.New("wallet requires reset refresh")
	}
	if state.ObservedThrough != nil {
		if e = s.Canonical(ctx, *state.ObservedThrough); e != nil {
			_ = s.Invalidate(ctx)
			return c, block, 0, e
		}
	}
	if state.CheckedThrough != nil {
		if e = s.Canonical(ctx, *state.CheckedThrough); e != nil {
			_ = s.Invalidate(ctx)
			return c, block, 0, e
		}
	}
	return c, block, deadline, nil
}

func (s *Runtime) PrepareFunds(ctx context.Context, p TransferPreparation, withdrawal common.Address, withdrawAmount *big.Int, commitments []common.Hash) (*Execution, error) {
	c, block, deadline, e := s.Frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	funding, mode, e := s.Funding(ctx, p.Funding, block, deadline)
	if e != nil {
		return nil, e
	}
	target := "evm"
	total := new(big.Int)
	for index, out := range p.Outputs {
		if e := Positive(out.Amount(), 128); e != nil {
			return nil, OutputFailure(e, index, out.To(), "amount")
		}
		total.Add(total, out.Amount())
	}
	if withdrawAmount == nil {
		withdrawAmount = new(big.Int)
	} else {
		if e := Positive(withdrawAmount, 128); e != nil {
			return nil, e
		}
		withdrawAmount = new(big.Int).Set(withdrawAmount)
	}
	total.Add(total, withdrawAmount)
	need := new(big.Int).Sub(total, funding.Amount)
	if need.Sign() < 0 {
		return nil, errors.New("funding exceeds selected outputs")
	}
	inputs, memberships, e := s.SelectInputs(ctx, p.TokenAddress, need, commitments, block)
	if e != nil {
		return nil, e
	}
	sum := new(big.Int).Set(funding.Amount)
	for _, n := range inputs {
		sum.Add(sum, core.ElementToBigInt(n.Note.Amount))
	}
	outputs := []PreparedOutput{}
	for index, out := range p.Outputs {
		o, e := s.Output(ctx, p.TokenAddress, out)
		if e != nil {
			return nil, OutputFailure(e, index, out.To(), "")
		}
		outputs = append(outputs, o)
	}
	change := new(big.Int).Sub(sum, total)
	if change.Sign() > 0 {
		address, e := s.Wallet.GetAddress()
		if e != nil {
			return nil, e
		}
		if s.ChangeAddress != "" {
			address = s.ChangeAddress
		}
		o, e := s.ChangeOutput(ctx, p.TokenAddress, address, change)
		if e != nil {
			return nil, e
		}
		o.Change = true
		outputs = append(outputs, o)
	}
	capacity, e := protocol.SelectTransactCapacity(len(inputs), len(outputs))
	if e != nil {
		return nil, e
	}
	params := protocol.CrossTreeTransactParams{Token: p.TokenAddress, WithdrawalAddress: withdrawal, WithdrawalAmount: withdrawAmount, Deadline: new(big.Int).SetUint64(deadline)}
	roots := map[uint64]common.Hash{}
	for i, n := range inputs {
		params.Nullifiers = append(params.Nullifiers, n.Nullifier)
		root := memberships[i].Root
		if old, ok := roots[n.TreeID]; ok && old != root {
			return nil, errors.New("inconsistent roots for one input tree")
		}
		roots[n.TreeID] = root
	}
	ids := []uint64{}
	for id := range roots {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		params.Sources = append(params.Sources, protocol.InputSource{TreeId: new(big.Int).SetUint64(id), MerkleRoot: roots[id]})
	}
	for _, out := range outputs {
		params.OutputCommitments = append(params.OutputCommitments, out.Commitment)
	}
	plan := Plan{Kind: "transact", Target: target, Context: c, Block: block, Deadline: deadline, Inputs: inputs, Memberships: memberships, Outputs: outputs, Funding: funding, FundingMode: mode, Params: params, Capacity: capacity, NeedsAuthorization: len(inputs) > 0 && !(withdrawal == s.Wallet.identity && withdrawAmount.Cmp(sum) == 0 && len(outputs) == 0 && funding.Amount.Sign() == 0)}
	return s.PersistPlan(ctx, plan)
}

func (s *Runtime) PrepareTransfer(ctx context.Context, p TransferPreparation) (result *Execution, failure error) {
	defer func() { failure = WrapPreparation("prepareTransfer", failure) }()
	var err error
	p, err = Frozen(p)
	if err != nil {
		return nil, err
	}
	if len(p.Outputs) == 0 && p.Withdrawal == nil {
		return nil, errors.New("transfer requires an output or withdrawal")
	}
	var to common.Address
	var amount *big.Int
	if p.Withdrawal != nil {
		to, amount = p.Withdrawal.To, p.Withdrawal.Amount
	}
	return s.PrepareFunds(ctx, p, to, amount, p.Commitments)
}

func (s *Runtime) PrepareShield(ctx context.Context, p ShieldPreparation) (result *Execution, failure error) {
	defer func() { failure = WrapPreparation("prepareShield", failure) }()
	var err error
	p, err = Frozen(p)
	if err != nil {
		return nil, err
	}
	c, block, deadline, e := s.Frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	funding, mode, e := s.Funding(ctx, &p.Funding, block, deadline)
	if e != nil {
		return nil, e
	}
	if len(p.Outputs) < 1 || len(p.Outputs) > 10 {
		return nil, errors.New("shield requires 1..10 outputs")
	}
	plan := Plan{Kind: "shield", Context: c, Block: block, Deadline: deadline, Funding: funding, FundingMode: mode, Shield: protocol.ShieldParams{Token: p.TokenAddress, Deadline: new(big.Int).SetUint64(deadline)}}
	sum := new(big.Int)
	for index, out := range p.Outputs {
		o, e := s.Output(ctx, p.TokenAddress, out)
		if e != nil {
			return nil, OutputFailure(e, index, out.To(), "")
		}
		plan.Outputs = append(plan.Outputs, o)
		plan.Shield.OutputCommitments = append(plan.Shield.OutputCommitments, o.Commitment)
		sum.Add(sum, out.Amount())
	}
	if sum.Cmp(funding.Amount) != 0 {
		return nil, errors.New("shield funding must equal output sum")
	}
	return s.PersistPlan(ctx, plan)
}

func RandomOperationID() (string, error) {
	var raw [16]byte
	if _, e := rand.Read(raw[:]); e != nil {
		return "", e
	}
	return hex.EncodeToString(raw[:]), nil
}
