package transaction

import (
	"context"
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

func CreditCommitment(note core.CreditNote) (common.Hash, error) {
	key, e := core.ComputeCreditSharedKeyHash(note.CreditSharedKey)
	if e != nil {
		return common.Hash{}, e
	}
	leaf, e := core.ComputeCreditCommitment(note.ChainID, note.SpenderWalletIdentity, note.OwnerWalletIdentity, note.SpenderPayout, note.OwnerPayout, note.TokenAddress, note.Amount, note.Nonce, note.Expiry, key, note.CreditSpendingState)
	return FieldHash(leaf), e
}

func CreditDeliveries(state core.CreditDeliveryState) ([]protocol.Delivery, error) {
	leaf, e := CreditCommitment(state.Note)
	if e != nil {
		return nil, e
	}
	result := []protocol.Delivery{}
	for _, entry := range []struct {
		role core.CreditDeliveryRole
		keys core.ReceivingKeys
	}{{core.CreditOwner, state.OwnerReceivingKeys}, {core.CreditSpender, state.SpenderReceivingKeys}} {
		cipher, e := core.EncryptCreditNote(entry.keys.ViewingPublicKey, state, core.CreditNoteEncryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: state.Note.ProtocolContext, OutputCommitment: core.BigIntToElement(leaf.Big())}, Role: entry.role})
		if e != nil {
			return nil, e
		}
		clue, e := core.CreateClue(entry.keys.ClueKey)
		if e != nil {
			return nil, e
		}
		result = append(result, protocol.Delivery{Ciphertext: cipher, Clue: clue})
	}
	return result, nil
}

func SourcesFor(memberships []protocolrpc.Membership) ([]protocol.InputSource, error) {
	roots := map[uint64]common.Hash{}
	for _, m := range memberships {
		if root, ok := roots[m.TreeID]; ok && root != m.Root {
			return nil, errors.New("inconsistent roots for input tree")
		}
		roots[m.TreeID] = m.Root
	}
	ids := []uint64{}
	for id := range roots {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	result := []protocol.InputSource{}
	for _, id := range ids {
		result = append(result, protocol.InputSource{TreeId: new(big.Int).SetUint64(id), MerkleRoot: roots[id]})
	}
	return result, nil
}

func (s *Runtime) CreditInput(ctx context.Context, token common.Address, hash common.Hash, role string, block protocolrpc.CanonicalBlock) (CreditRecord, protocolrpc.Membership, error) {
	state, e := s.PreparationInventory(ctx)
	if e != nil {
		return CreditRecord{}, protocolrpc.Membership{}, e
	}
	credit, ok := state.Credits[CreditKey(hash, role)]
	if !ok || credit.Spent || credit.ReservedBy != "" || core.ElementToBigInt(credit.State.Note.TokenAddress).Cmp(token.Big()) != 0 {
		return credit, protocolrpc.Membership{}, errors.New("available role-scoped Credit not found")
	}
	identity := credit.State.Note.OwnerWalletIdentity
	if role == "spender" {
		identity = credit.State.Note.SpenderWalletIdentity
	}
	if credit.State.Note.ProtocolContext != s.CoreContext() || identity != core.BigIntToElement(s.Wallet.identity.Big()) {
		return credit, protocolrpc.Membership{}, errors.New("stored Credit deployment or role mismatch")
	}
	leaf, e := CreditCommitment(credit.State.Note)
	if e != nil || leaf != hash {
		return credit, protocolrpc.Membership{}, errors.New("stored Credit commitment mismatch")
	}
	nf, e := core.ComputeCreditNullifier(credit.State.Note.ChainID, credit.State.Note.CreditSharedKey, core.BigIntToElement(leaf.Big()))
	if e != nil || FieldHash(nf) != credit.Nullifier {
		return credit, protocolrpc.Membership{}, errors.New("stored Credit nullifier mismatch")
	}
	contract, e := contracts.NewInvisiProtocolCaller(s.Scope.ProtocolAddress, s.Options.Provider)
	if e != nil {
		return credit, protocolrpc.Membership{}, e
	}
	spent, e := contract.IsNullifierSpent(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(block.Number)}, credit.Nullifier)
	if e != nil {
		return credit, protocolrpc.Membership{}, e
	}
	if spent {
		return credit, protocolrpc.Membership{}, errors.New("Credit was consumed; refresh wallet")
	}
	membership, e := s.RPC.GetMembership(ctx, s.Protocol.Context(), protocolrpc.MembershipInput{TreeID: credit.TreeID, LeafIndex: credit.LeafIndex, Commitment: hash}, block)
	return credit, membership, e
}

func (s *Runtime) PrepareCreditCreation(ctx context.Context, p CreateCreditPreparation, previous *CreditRecord, previousMembership *protocolrpc.Membership) (*Execution, error) {
	if err := ValidatePayload(p.Payload); err != nil {
		return nil, err
	}
	c, block, deadline, e := s.Frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	if e = Positive(p.Amount, 126); e != nil {
		return nil, e
	}
	if p.Expiry == nil || !p.Expiry.IsUint64() || p.Expiry.Uint64() <= block.Timestamp {
		return nil, errors.New("Credit expiry must be in the future")
	}
	funding, mode, e := s.Funding(ctx, p.Funding, block, deadline)
	if e != nil {
		return nil, e
	}
	if previous != nil {
		fresh, membership, e := s.CreditInput(ctx, p.TokenAddress, previous.Commitment, "owner", block)
		if e != nil {
			return nil, e
		}
		previous = &fresh
		previousMembership = &membership
	}
	amount := new(big.Int).Set(p.Amount)
	spending := new(big.Int).Set(amount)
	var state core.CreditDeliveryState
	kind := "create-credit"
	if previous != nil {
		kind = "top-up-credit"
		if core.ElementToBigInt(previous.State.Note.Expiry).Uint64() <= block.Timestamp || p.Expiry.Cmp(core.ElementToBigInt(previous.State.Note.Expiry)) < 0 {
			return nil, errors.New("top-up requires unexpired Credit and unchanged or extended expiry")
		}
		state = previous.State
		spending.Sub(spending, core.ElementToBigInt(state.Note.Amount))
		if spending.Sign() <= 0 {
			return nil, errors.New("top-up must increase the Credit amount")
		}
	} else {
		spender, e := s.ResolveRecipient(ctx, p.Spender)
		if e != nil {
			return nil, e
		}
		ownerAddress, e := s.Wallet.GetAddress()
		if e != nil {
			return nil, e
		}
		owner, e := s.ResolveRecipient(ctx, ownerAddress)
		if e != nil {
			return nil, e
		}
		key, e := core.GenerateCreditSharedKey()
		if e != nil {
			return nil, e
		}
		state = core.CreditDeliveryState{Note: core.CreditNote{ProtocolContext: s.CoreContext(), SpenderWalletIdentity: spender.OwnerAddress, OwnerWalletIdentity: owner.OwnerAddress, SpenderPayout: spender.OwnerAddress, OwnerPayout: owner.OwnerAddress, TokenAddress: core.BigIntToElement(p.TokenAddress.Big()), CreditSharedKey: key}, OwnerReceivingKeys: core.ReceivingKeys{ViewingPublicKey: owner.ViewingPublicKey, ClueKey: owner.ClueKey}, SpenderReceivingKeys: core.ReceivingKeys{ViewingPublicKey: spender.ViewingPublicKey, ClueKey: spender.ClueKey}}
	}
	state.Note.Payload = ClonePayload(p.Payload)
	need := new(big.Int).Sub(spending, funding.Amount)
	if need.Sign() < 0 {
		return nil, errors.New("funding exceeds Credit addition")
	}
	inputs, memberships, e := s.SelectInputs(ctx, p.TokenAddress, need, nil, block)
	if e != nil {
		return nil, e
	}
	if len(inputs) > 10 {
		return nil, errors.New("Credit creation supports at most ten funds inputs")
	}
	nonce, e := core.GenerateNonce()
	if e != nil {
		return nil, e
	}
	for _, entry := range []struct {
		value  *big.Int
		target *fr.Element
	}{{p.SingleLimit, &state.Note.SingleLimit}, {p.DailyLimit, &state.Note.DailyLimit}, {p.MonthlyLimit, &state.Note.MonthlyLimit}} {
		if entry.value != nil {
			if err := protocol.CheckUint(entry.value, 126, "Credit limit"); err != nil {
				return nil, err
			}
			*entry.target = core.BigIntToElement(entry.value)
		}
	}
	state.Note.Amount = core.BigIntToElement(amount)
	state.Note.Nonce = nonce
	state.Note.Expiry = core.BigIntToElement(p.Expiry)
	leaf, e := CreditCommitment(state.Note)
	if e != nil {
		return nil, e
	}
	deliveries, e := CreditDeliveries(state)
	if e != nil {
		return nil, e
	}
	sources, e := SourcesFor(memberships)
	if e != nil {
		return nil, e
	}
	creation := protocol.CreateCreditParams{Token: p.TokenAddress, Sources: sources, CreditCommitment: leaf, Owner: s.Wallet.identity, Expiry: new(big.Int).Set(p.Expiry), Deadline: new(big.Int).SetUint64(deadline)}
	sum := new(big.Int).Set(funding.Amount)
	for _, input := range inputs {
		creation.Nullifiers = append(creation.Nullifiers, input.Nullifier)
		sum.Add(sum, core.ElementToBigInt(input.Note.Amount))
	}
	change := new(big.Int).Sub(sum, spending)
	outputs := []PreparedOutput{}
	if change.Sign() > 0 {
		address, e := s.Wallet.GetAddress()
		if e != nil {
			return nil, e
		}
		if s.ChangeAddress != "" {
			address = s.ChangeAddress
		}
		recipient, err := s.ResolveRecipient(ctx, address)
		if err != nil {
			return nil, err
		}
		if recipient.OwnerAddress != core.BigIntToElement(s.Wallet.identity.Big()) {
			return nil, errors.New("Credit change must belong to the owner")
		}
		out, e := s.ChangeOutput(ctx, p.TokenAddress, address, change)
		if e != nil {
			return nil, e
		}
		out.Change = true
		outputs = append(outputs, out)
		creation.ChangeCommitment = out.Commitment
		deliveries = append(deliveries, out.Delivery)
	} else {
		deliveries = append(deliveries, protocol.Delivery{})
	}
	plan := Plan{Kind: kind, Target: "evm", Context: c, Block: block, Deadline: deadline, Inputs: inputs, Memberships: memberships, Outputs: outputs, Funding: funding, FundingMode: mode, Capacity: 10, NeedsAuthorization: true, CreateCredit: &creation, CreditOutput: &state, Deliveries: deliveries}
	if previous != nil {
		plan.CreditInputs = []CreditRecord{*previous}
		plan.CreditMemberships = []protocolrpc.Membership{*previousMembership}
		plan.TopUpCredit = &protocol.TopUpCreditParams{Creation: creation, CreditSource: protocol.InputSource{TreeId: new(big.Int).SetUint64(previous.TreeID), MerkleRoot: previousMembership.Root}, CreditNullifier: previous.Nullifier, PreviousExpiry: core.ElementToBigInt(previous.State.Note.Expiry)}
	}
	return s.PersistPlan(ctx, plan)
}
