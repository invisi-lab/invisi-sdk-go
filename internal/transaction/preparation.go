package transaction

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
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

type Fee = credit.Fee

func normalizePublicFee(input *Fee) (Fee, error) {
	fee := Fee{Amount: new(big.Int)}
	if input != nil {
		fee.Recipient = input.Recipient
		if input.Amount != nil {
			fee.Amount.Set(input.Amount)
		}
	}
	if fee.Amount.Sign() < 0 || fee.Amount.BitLen() > 128 || (fee.Recipient == (common.Address{})) != (fee.Amount.Sign() == 0) {
		return Fee{}, errors.New("fee requires a nonzero recipient and positive uint128 amount, or an empty zero pair")
	}
	return fee, nil
}

type ShieldPreparation struct {
	Fee          *Fee
	TokenAddress common.Address
	Outputs      []NoteOutput
	Funding      FundingSource
	Deadline     uint64
}

type TransferPreparation struct {
	Fee          *Fee
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
	Note       *primitives.Note
	Amount     *big.Int
	Identity   common.Hash
	Commitment common.Hash
	Delivery   protocol.Delivery
	Change     bool
}

type Plan struct {
	SubmissionValidation string `json:",omitempty"`
	Version              int
	OwnerAddress         common.Address
	Namespace            string
	Kind, ID, Target     string
	Context              protocol.TransactionContext
	Block                protocolrpc.CanonicalBlock
	Deadline             uint64
	Inputs               []NoteRecord
	Memberships          []protocolrpc.Membership
	Outputs              []PreparedOutput
	Funding              protocol.FundingParams
	FundingMode          string
	Params               protocol.CrossTreeTransactParams
	Shield               protocol.ShieldParams
	Capacity             int
	NeedsAuthorization   bool

	Deliveries  []protocol.Delivery
	DrawAmounts []*big.Int
	Broadcaster *BroadcasterPlan
	Provenance  *ProvenancePlan
}

func ValidatePayload(payload []byte) error { return primitives.ValidatePayload(payload) }

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
	var head *types.Header
	_, e = MapPreparation(ctx, 2, func(ctx context.Context, i int) (struct{}, error) {
		if i == 0 {
			return struct{}{}, s.Protocol.ValidateDeployment(ctx, s.Options.Provider)
		}
		var err error
		head, err = s.Options.Provider.HeaderByNumber(ctx, nil)
		return struct{}{}, err
	})
	if e != nil {
		return c, block, 0, e
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() {
		return c, block, 0, errors.New("invalid latest block")
	}
	block = protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	timestamp := head.Time
	if s.Snapshot != nil && time.Now().Unix() > 0 && uint64(time.Now().Unix()) > timestamp {
		timestamp = uint64(time.Now().Unix())
	}
	if deadline == 0 {
		deadline = timestamp + 300
	}
	if deadline < timestamp || deadline-timestamp > 1<<53 {
		return c, block, 0, errors.New("invalid deadline")
	}
	c.Now = new(big.Int).SetUint64(timestamp)
	code, e := s.Options.Provider.CodeAt(ctx, token, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return c, block, 0, e
	}
	if len(code) == 0 {
		return c, block, 0, errors.New("token address must have contract code")
	}
	readState := s.PreparationInventory
	if s.SummaryReader != nil {
		readState = s.SummaryReader
	}
	state, e := readState(ctx)
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
	s = s.preparationRuntime()
	c, block, deadline, e := s.Frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	funding, mode, e := s.Funding(ctx, p.Funding, block, deadline)
	if e != nil {
		return nil, e
	}
	fee, e := normalizePublicFee(p.Fee)
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
	total.Add(total, fee.Amount)
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
		sum.Add(sum, primitives.ElementToBigInt(n.Note.Amount))
	}
	outputs, e := MapPreparation(ctx, len(p.Outputs), func(ctx context.Context, index int) (PreparedOutput, error) {
		out := p.Outputs[index]
		o, err := s.Output(ctx, p.TokenAddress, out)
		if err != nil {
			return PreparedOutput{}, OutputFailure(err, index, out.To(), "")
		}
		return o, nil
	})
	if e != nil {
		return nil, e
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
	params := protocol.CrossTreeTransactParams{FeeRecipient: fee.Recipient, FeeAmount: fee.Amount, Token: p.TokenAddress, WithdrawalAddress: withdrawal, WithdrawalAmount: withdrawAmount, Deadline: new(big.Int).SetUint64(deadline)}
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
		params.NoteCommitments = append(params.NoteCommitments, out.Commitment)
	}
	plan := Plan{Kind: "transact", Target: target, Context: c, Block: block, Deadline: deadline, Inputs: inputs, Memberships: memberships, Outputs: outputs, Funding: funding, FundingMode: mode, Params: params, Capacity: capacity, NeedsAuthorization: len(inputs) > 0 && !(withdrawal == s.Wallet.ownerAddress && withdrawAmount.Cmp(sum) == 0 && len(outputs) == 0 && funding.Amount.Sign() == 0 && fee.Amount.Sign() == 0)}
	return s.PersistPlan(ctx, plan)
}

func (s *Runtime) PrepareTransfer(ctx context.Context, p TransferPreparation) (result *Execution, failure error) {
	defer func() { failure = WrapPreparation("prepareTransfer", failure) }()
	var err error
	p, err = Frozen(p)
	if err != nil {
		return nil, err
	}
	if len(p.Outputs) == 0 && p.Withdrawal == nil && p.Fee == nil {
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
	s = s.preparationRuntime()
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
	fee, e := normalizePublicFee(p.Fee)
	if e != nil {
		return nil, e
	}
	plan := Plan{Kind: "shield", Context: c, Block: block, Deadline: deadline, Funding: funding, FundingMode: mode, Shield: protocol.ShieldParams{FeeRecipient: fee.Recipient, FeeAmount: fee.Amount, Token: p.TokenAddress, Deadline: new(big.Int).SetUint64(deadline)}}
	sum := new(big.Int)
	plan.Outputs, e = MapPreparation(ctx, len(p.Outputs), func(ctx context.Context, index int) (PreparedOutput, error) {
		out := p.Outputs[index]
		o, err := s.Output(ctx, p.TokenAddress, out)
		if err != nil {
			return PreparedOutput{}, OutputFailure(err, index, out.To(), "")
		}
		return o, nil
	})
	if e != nil {
		return nil, e
	}
	for _, out := range plan.Outputs {
		plan.Shield.NoteCommitments = append(plan.Shield.NoteCommitments, out.Commitment)
		sum.Add(sum, out.Amount)
	}
	sum.Add(sum, fee.Amount)
	if sum.Cmp(funding.Amount) != 0 {
		return nil, errors.New("shield funding must equal output sum plus fee")
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
