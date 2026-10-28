package sdk

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type PaymentOutput struct {
	To     string
	Amount *big.Int
}
type FundingSource struct {
	Payer         common.Address
	Amount        *big.Int
	Authorization string
	Permit        *protocol.TokenPermit
}
type ShieldPreparation struct {
	TokenAddress common.Address
	Outputs      []PaymentOutput
	Funding      FundingSource
	Deadline     uint64
}
type TransferPreparation struct {
	TokenAddress common.Address
	Outputs      []PaymentOutput
	Funding      *FundingSource
	Deadline     uint64
	Target       string
}
type WithdrawalPreparation struct {
	TokenAddress common.Address
	To           common.Address
	Amount       *big.Int
	Deadline     uint64
	Target       string
}
type SweepPreparation struct {
	TokenAddress common.Address
	Commitments  []common.Hash
	Deadline     uint64
	Target       string
}
type preparedOutput struct {
	Address    string
	Note       core.Note
	Identity   common.Hash
	Commitment common.Hash
	Delivery   protocol.Delivery
	Change     bool
}
type operationPlan struct {
	Version            int
	WalletIdentity     common.Address
	Namespace          string
	Kind, ID, Target   string
	Context            protocol.TransactionContext
	Block              protocolrpc.CanonicalBlock
	Deadline           uint64
	Inputs             []StoredNote
	Memberships        []protocolrpc.Membership
	Outputs            []preparedOutput
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
	CreditInputs       []StoredCredit
	CreditMemberships  []protocolrpc.Membership
	CreditOutput       *core.CreditDeliveryState
	Deliveries         []protocol.Delivery
	DrawAmounts        []*big.Int
}

func positive(amount *big.Int, bits int) error {
	if amount == nil || amount.Sign() <= 0 || amount.BitLen() > bits {
		return errors.New("amount must be a positive bounded integer")
	}
	return nil
}
func (s *WalletSession) frame(ctx context.Context, token common.Address, deadline uint64) (protocol.TransactionContext, protocolrpc.CanonicalBlock, uint64, error) {
	var block protocolrpc.CanonicalBlock
	c, e := s.protocol.ForToken(token)
	if e != nil {
		return c, block, 0, e
	}
	if e = s.protocol.ValidateDeployment(ctx, s.options.Provider); e != nil {
		return c, block, 0, e
	}
	head, e := s.options.Provider.HeaderByNumber(ctx, nil)
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
	code, e := s.options.Provider.CodeAt(ctx, token, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return c, block, 0, e
	}
	if len(code) == 0 {
		return c, block, 0, errors.New("token address must have contract code")
	}
	state, e := s.preparationInventory(ctx)
	if e != nil {
		return c, block, 0, e
	}
	if state.Invalidated {
		return c, block, 0, errors.New("wallet requires reset refresh")
	}
	if state.ObservedThrough != nil {
		if e = s.canonical(ctx, *state.ObservedThrough); e != nil {
			_ = s.invalidate(ctx)
			return c, block, 0, e
		}
	}
	if state.CheckedThrough != nil {
		if e = s.canonical(ctx, *state.CheckedThrough); e != nil {
			_ = s.invalidate(ctx)
			return c, block, 0, e
		}
	}
	return c, block, deadline, nil
}
func (s *WalletSession) resolve(ctx context.Context, address string) (core.ReceivingAddress, error) {
	return core.ResolveAddress(address, func(identity fr.Element) (*core.ReceivingKeys, error) {
		deployment := s.protocol.Deployment()
		if deployment.Registry == nil {
			return nil, errors.New("registry deployment required for short address")
		}
		if e := s.protocol.ValidateDeployment(ctx, s.options.Provider); e != nil {
			return nil, e
		}
		code, e := s.options.Provider.CodeAt(ctx, deployment.Registry.Address, nil)
		if e != nil {
			return nil, e
		}
		if hashCode(code) != deployment.Registry.RuntimeCodeHash {
			return nil, errors.New("registry runtime mismatch")
		}
		caller, e := contracts.NewInvisiAddressRegistryCaller(deployment.Registry.Address, s.options.Provider)
		if e != nil {
			return nil, e
		}
		return protocol.CreateRegistryReceivingKeyLookup(caller, &bind.CallOpts{Context: ctx})(identity)
	})
}
func (s *WalletSession) output(ctx context.Context, token common.Address, output PaymentOutput) (preparedOutput, error) {
	var result preparedOutput
	if e := positive(output.Amount, 128); e != nil {
		return result, e
	}
	recipient, e := s.resolve(ctx, output.To)
	if e != nil {
		return result, e
	}
	nonce, e := core.GenerateNonce()
	if e != nil {
		return result, e
	}
	note := core.Note{ProtocolContext: s.coreContext(), WalletIdentity: recipient.WalletIdentity, TokenAddress: core.BigIntToElement(token.Big()), Amount: core.BigIntToElement(output.Amount), Nonce: nonce}
	identity, e := core.ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
	if e != nil {
		return result, e
	}
	commitment, e := core.ComputeNoteCommitment(note.ChainID, identity, note.TokenAddress, note.Amount)
	if e != nil {
		return result, e
	}
	cipher, e := core.EncryptPrivateTransferNote(recipient.ViewingPublicKey, note, core.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: commitment})
	if e != nil {
		return result, e
	}
	clue, e := core.CreateClue(recipient.ClueKey)
	if e != nil {
		return result, e
	}
	return preparedOutput{Address: output.To, Note: note, Identity: fieldHash(identity), Commitment: fieldHash(commitment), Delivery: protocol.Delivery{Ciphertext: cipher, Clue: clue}}, nil
}
func (s *WalletSession) funding(source *FundingSource) (protocol.FundingParams, string, error) {
	if source == nil {
		f, e := protocol.NormalizeFunding(protocol.FundingParams{})
		return f, "none", e
	}
	if e := positive(source.Amount, 128); e != nil {
		return protocol.FundingParams{}, "", e
	}
	mode := source.Authorization
	if mode == "" {
		mode = "signature"
	}
	if mode != "signature" && mode != "caller" {
		return protocol.FundingParams{}, "", errors.New("invalid funding authorization mode")
	}
	f := protocol.FundingParams{Payer: source.Payer, Amount: new(big.Int).Set(source.Amount)}
	if source.Permit != nil {
		f.Permit = *source.Permit
	}
	f, e := protocol.NormalizeFunding(f)
	return f, mode, e
}
func (s *WalletSession) selectInputs(ctx context.Context, token common.Address, amount *big.Int, all bool, commitments []common.Hash, block protocolrpc.CanonicalBlock) ([]StoredNote, []protocolrpc.Membership, error) {
	state, e := s.preparationInventory(ctx)
	if e != nil {
		return nil, nil, e
	}
	if s.explicitInventory != nil {
		for _, n := range state.Notes {
			if core.ElementToBigInt(n.Note.TokenAddress).Cmp(token.Big()) != 0 {
				return nil, nil, errors.New("explicit note token mismatch")
			}
		}
	}
	inputs := []StoredNote{}
	memberships := []protocolrpc.Membership{}
	sum := new(big.Int)
	selected := map[common.Hash]bool{}
	for _, hash := range commitments {
		if selected[hash] {
			return nil, nil, errors.New("duplicate commitment")
		}
		selected[hash] = true
	}
	for _, n := range orderedNotes(state, token) {
		if n.Note.ProtocolContext != s.coreContext() || n.Note.WalletIdentity != core.BigIntToElement(s.wallet.identity.Big()) {
			return nil, nil, errors.New("stored note deployment or owner mismatch")
		}
		expected, e := s.wallet.ExportNullifier().ForNote(n.Note)
		if e != nil || expected != n.Nullifier {
			return nil, nil, errors.New("stored note nullifier mismatch")
		}
		identity, e := core.ComputeNoteIdentity(n.Note.WalletIdentity, n.Note.Nonce)
		if e != nil {
			return nil, nil, e
		}
		leaf, e := core.ComputeNoteCommitment(n.Note.ChainID, identity, n.Note.TokenAddress, n.Note.Amount)
		if e != nil || fieldHash(leaf) != n.Commitment {
			return nil, nil, errors.New("stored note commitment mismatch")
		}

		if len(selected) > 0 && !selected[n.Commitment] {
			continue
		}
		if s.explicitInventory == nil && !all && sum.Cmp(amount) >= 0 {
			break
		}
		contract, e := contracts.NewInvisiProtocolCaller(s.scope.ProtocolAddress, s.options.Provider)
		if e != nil {
			return nil, nil, e
		}
		spent, e := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(block.Number)}, n.Nullifier)
		if e != nil {
			return nil, nil, e
		}
		if spent {
			return nil, nil, errors.New("stored note was consumed; refresh wallet")
		}
		m, e := s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: n.TreeID, LeafIndex: n.LeafIndex, Commitment: n.Commitment}, block)
		if e != nil {
			return nil, nil, e
		}
		inputs = append(inputs, n)
		memberships = append(memberships, m)
		sum.Add(sum, core.ElementToBigInt(n.Note.Amount))
	}
	if (!all && sum.Cmp(amount) < 0) || (all && len(inputs) == 0) || (len(selected) > 0 && len(inputs) != len(selected)) {
		return nil, nil, errors.New("insufficient available known notes")
	}
	return inputs, memberships, nil
}
func (s *WalletSession) prepareFunds(ctx context.Context, p TransferPreparation, withdrawal common.Address, withdrawAmount *big.Int, sweep bool, commitments []common.Hash) (*Operation, error) {
	c, block, deadline, e := s.frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	funding, mode, e := s.funding(p.Funding)
	if e != nil {
		return nil, e
	}
	target := p.Target
	if target == "" {
		target = "evm"
	}
	if target != "evm" && target != "noir-recursive" {
		return nil, errors.New("unsupported proof target")
	}
	total := new(big.Int)
	for _, out := range p.Outputs {
		if e := positive(out.Amount, 128); e != nil {
			return nil, e
		}
		total.Add(total, out.Amount)
	}
	if withdrawAmount == nil {
		withdrawAmount = new(big.Int)
	} else {
		if e := positive(withdrawAmount, 128); e != nil {
			return nil, e
		}
		withdrawAmount = new(big.Int).Set(withdrawAmount)
	}
	total.Add(total, withdrawAmount)
	need := new(big.Int).Sub(total, funding.Amount)
	if need.Sign() < 0 {
		return nil, errors.New("funding exceeds selected outputs")
	}
	inputs, memberships, e := s.selectInputs(ctx, p.TokenAddress, need, sweep, commitments, block)
	if e != nil {
		return nil, e
	}
	sum := new(big.Int).Set(funding.Amount)
	for _, n := range inputs {
		sum.Add(sum, core.ElementToBigInt(n.Note.Amount))
	}
	outputs := []preparedOutput{}
	if sweep {
		if len(p.Outputs) > 0 || funding.Amount.Sign() != 0 {
			return nil, errors.New("sweep requires full notes without outputs or funding")
		}
		withdrawal = s.wallet.identity
		withdrawAmount.Set(sum)
		total.Set(sum)
	}
	for _, out := range p.Outputs {
		o, e := s.output(ctx, p.TokenAddress, out)
		if e != nil {
			return nil, e
		}
		outputs = append(outputs, o)
	}
	change := new(big.Int).Sub(sum, total)
	if change.Sign() > 0 {
		address, e := s.wallet.GetAddress()
		if e != nil {
			return nil, e
		}
		if s.changeAddress != "" {
			address = s.changeAddress
		}
		o, e := s.output(ctx, p.TokenAddress, PaymentOutput{address, change})
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
	if target == "noir-recursive" && capacity != 14 {
		return nil, errors.New("recursive aggregation only supports slots14")
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
	plan := operationPlan{Kind: "transact", Target: target, Context: c, Block: block, Deadline: deadline, Inputs: inputs, Memberships: memberships, Outputs: outputs, Funding: funding, FundingMode: mode, Params: params, Capacity: capacity, NeedsAuthorization: len(inputs) > 0 && !sweep}
	return s.persistPlan(ctx, plan)
}
func (s *WalletSession) PrepareTransfer(ctx context.Context, p TransferPreparation) (*Operation, error) {
	return s.prepareFunds(ctx, p, common.Address{}, nil, false, nil)
}
func (s *WalletSession) PrepareWithdraw(ctx context.Context, p WithdrawalPreparation) (*Operation, error) {
	if p.To == (common.Address{}) {
		return nil, errors.New("withdrawal recipient required")
	}
	return s.prepareFunds(ctx, TransferPreparation{TokenAddress: p.TokenAddress, Deadline: p.Deadline, Target: p.Target}, p.To, p.Amount, false, nil)
}
func (s *WalletSession) PrepareSweep(ctx context.Context, p SweepPreparation) (*Operation, error) {
	return s.prepareFunds(ctx, TransferPreparation{TokenAddress: p.TokenAddress, Deadline: p.Deadline, Target: p.Target}, s.wallet.identity, nil, true, p.Commitments)
}
func (s *WalletSession) PrepareShield(ctx context.Context, p ShieldPreparation) (*Operation, error) {
	c, block, deadline, e := s.frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	funding, mode, e := s.funding(&p.Funding)
	if e != nil {
		return nil, e
	}
	if len(p.Outputs) < 1 || len(p.Outputs) > 10 {
		return nil, errors.New("shield requires 1..10 outputs")
	}
	plan := operationPlan{Kind: "shield", Context: c, Block: block, Deadline: deadline, Funding: funding, FundingMode: mode, Shield: protocol.ShieldParams{Token: p.TokenAddress, Deadline: new(big.Int).SetUint64(deadline)}}
	sum := new(big.Int)
	for _, out := range p.Outputs {
		o, e := s.output(ctx, p.TokenAddress, out)
		if e != nil {
			return nil, e
		}
		plan.Outputs = append(plan.Outputs, o)
		plan.Shield.OutputCommitments = append(plan.Shield.OutputCommitments, o.Commitment)
		sum.Add(sum, out.Amount)
	}
	if sum.Cmp(funding.Amount) != 0 {
		return nil, errors.New("shield funding must equal output sum")
	}
	return s.persistPlan(ctx, plan)
}
func randomOperationID() (string, error) {
	var raw [16]byte
	if _, e := rand.Read(raw[:]); e != nil {
		return "", e
	}
	return hex.EncodeToString(raw[:]), nil
}
