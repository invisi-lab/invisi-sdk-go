package sdk

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

type CreateCreditPreparation struct {
	Payload                               []byte
	TokenAddress                          common.Address
	Spender                               string
	Amount, Expiry                        *big.Int
	SingleLimit, DailyLimit, MonthlyLimit *big.Int
	Funding                               *FundingSource
	Deadline                              uint64
}
type CreditDrawInput struct {
	// Omission creates an empty payload; consumed metadata is never inherited.
	Payload []byte
	Credit  common.Hash
	Amount  *big.Int
}
type DrawCreditPreparation struct {
	TokenAddress common.Address
	Credits      []CreditDrawInput
	Deadline     uint64
}
type CancelCreditPreparation struct {
	TokenAddress common.Address
	Credit       common.Hash
	Deadline     uint64
}
type TopUpCreditPreparation struct {
	// Omission creates an empty payload; consumed metadata is never inherited.
	Payload                  []byte
	TokenAddress             common.Address
	Credit                   common.Hash
	AdditionalAmount, Expiry *big.Int
	// Nil inherits the existing limit; zero explicitly disables its check without clearing usage.
	SingleLimit, DailyLimit, MonthlyLimit *big.Int
	Funding                               *FundingSource
	Deadline                              uint64
}

func creditCommitment(note core.CreditNote) (common.Hash, error) {
	key, e := core.ComputeCreditSharedKeyHash(note.CreditSharedKey)
	if e != nil {
		return common.Hash{}, e
	}
	leaf, e := core.ComputeCreditCommitment(note.ChainID, note.SpenderWalletIdentity, note.OwnerWalletIdentity, note.SpenderPayout, note.OwnerPayout, note.TokenAddress, note.Amount, note.Nonce, note.Expiry, key, note.CreditSpendingState)
	return fieldHash(leaf), e
}
func creditDeliveries(state core.CreditDeliveryState) ([]protocol.Delivery, error) {
	leaf, e := creditCommitment(state.Note)
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
func sourcesFor(memberships []protocolrpc.Membership) ([]protocol.InputSource, error) {
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
func (s *WalletSession) creditInput(ctx context.Context, token common.Address, hash common.Hash, role string, block protocolrpc.CanonicalBlock) (StoredCredit, protocolrpc.Membership, error) {
	state, e := s.preparationInventory(ctx)
	if e != nil {
		return StoredCredit{}, protocolrpc.Membership{}, e
	}
	credit, ok := state.Credits[creditKey(hash, role)]
	if !ok || credit.Spent || credit.ReservedBy != "" || core.ElementToBigInt(credit.State.Note.TokenAddress).Cmp(token.Big()) != 0 {
		return credit, protocolrpc.Membership{}, errors.New("available role-scoped Credit not found")
	}
	identity := credit.State.Note.OwnerWalletIdentity
	if role == "spender" {
		identity = credit.State.Note.SpenderWalletIdentity
	}
	if credit.State.Note.ProtocolContext != s.coreContext() || identity != core.BigIntToElement(s.wallet.identity.Big()) {
		return credit, protocolrpc.Membership{}, errors.New("stored Credit deployment or role mismatch")
	}
	leaf, e := creditCommitment(credit.State.Note)
	if e != nil || leaf != hash {
		return credit, protocolrpc.Membership{}, errors.New("stored Credit commitment mismatch")
	}
	nf, e := core.ComputeCreditNullifier(credit.State.Note.ChainID, credit.State.Note.CreditSharedKey, core.BigIntToElement(leaf.Big()))
	if e != nil || fieldHash(nf) != credit.Nullifier {
		return credit, protocolrpc.Membership{}, errors.New("stored Credit nullifier mismatch")
	}
	contract, e := contracts.NewInvisiProtocolCaller(s.scope.ProtocolAddress, s.options.Provider)
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
	membership, e := s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: credit.TreeID, LeafIndex: credit.LeafIndex, Commitment: hash}, block)
	return credit, membership, e
}
func (s *WalletSession) prepareCreditCreation(ctx context.Context, p CreateCreditPreparation, previous *StoredCredit, previousMembership *protocolrpc.Membership) (*Operation, error) {
	if err := validatePayload(p.Payload); err != nil {
		return nil, err
	}
	c, block, deadline, e := s.frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	if e = positive(p.Amount, 126); e != nil {
		return nil, e
	}
	if p.Expiry == nil || !p.Expiry.IsUint64() || p.Expiry.Uint64() <= block.Timestamp {
		return nil, errors.New("Credit expiry must be in the future")
	}
	funding, mode, e := s.funding(ctx, p.Funding, block, deadline)
	if e != nil {
		return nil, e
	}
	if previous != nil {
		fresh, membership, e := s.creditInput(ctx, p.TokenAddress, previous.Commitment, "owner", block)
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
		spender, e := s.resolve(ctx, p.Spender)
		if e != nil {
			return nil, e
		}
		ownerAddress, e := s.wallet.GetAddress()
		if e != nil {
			return nil, e
		}
		owner, e := s.resolve(ctx, ownerAddress)
		if e != nil {
			return nil, e
		}
		key, e := core.GenerateCreditSharedKey()
		if e != nil {
			return nil, e
		}
		state = core.CreditDeliveryState{Note: core.CreditNote{ProtocolContext: s.coreContext(), SpenderWalletIdentity: spender.OwnerAddress, OwnerWalletIdentity: owner.OwnerAddress, SpenderPayout: spender.OwnerAddress, OwnerPayout: owner.OwnerAddress, TokenAddress: core.BigIntToElement(p.TokenAddress.Big()), CreditSharedKey: key}, OwnerReceivingKeys: core.ReceivingKeys{ViewingPublicKey: owner.ViewingPublicKey, ClueKey: owner.ClueKey}, SpenderReceivingKeys: core.ReceivingKeys{ViewingPublicKey: spender.ViewingPublicKey, ClueKey: spender.ClueKey}}
	}
	state.Note.Payload = clonePayload(p.Payload)
	need := new(big.Int).Sub(spending, funding.Amount)
	if need.Sign() < 0 {
		return nil, errors.New("funding exceeds Credit addition")
	}
	inputs, memberships, e := s.selectInputs(ctx, p.TokenAddress, need, nil, block)
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
	leaf, e := creditCommitment(state.Note)
	if e != nil {
		return nil, e
	}
	deliveries, e := creditDeliveries(state)
	if e != nil {
		return nil, e
	}
	sources, e := sourcesFor(memberships)
	if e != nil {
		return nil, e
	}
	creation := protocol.CreateCreditParams{Token: p.TokenAddress, Sources: sources, CreditCommitment: leaf, Owner: s.wallet.identity, Expiry: new(big.Int).Set(p.Expiry), Deadline: new(big.Int).SetUint64(deadline)}
	sum := new(big.Int).Set(funding.Amount)
	for _, input := range inputs {
		creation.Nullifiers = append(creation.Nullifiers, input.Nullifier)
		sum.Add(sum, core.ElementToBigInt(input.Note.Amount))
	}
	change := new(big.Int).Sub(sum, spending)
	outputs := []preparedOutput{}
	if change.Sign() > 0 {
		address, e := s.wallet.GetAddress()
		if e != nil {
			return nil, e
		}
		if s.changeAddress != "" {
			address = s.changeAddress
		}
		recipient, err := s.resolve(ctx, address)
		if err != nil {
			return nil, err
		}
		if recipient.OwnerAddress != core.BigIntToElement(s.wallet.identity.Big()) {
			return nil, errors.New("Credit change must belong to the owner")
		}
		out, e := s.changeOutput(ctx, p.TokenAddress, address, change)
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
	plan := operationPlan{Kind: kind, Target: "evm", Context: c, Block: block, Deadline: deadline, Inputs: inputs, Memberships: memberships, Outputs: outputs, Funding: funding, FundingMode: mode, Capacity: 10, NeedsAuthorization: true, CreateCredit: &creation, CreditOutput: &state, Deliveries: deliveries}
	if previous != nil {
		plan.CreditInputs = []StoredCredit{*previous}
		plan.CreditMemberships = []protocolrpc.Membership{*previousMembership}
		plan.TopUpCredit = &protocol.TopUpCreditParams{Creation: creation, CreditSource: protocol.InputSource{TreeId: new(big.Int).SetUint64(previous.TreeID), MerkleRoot: previousMembership.Root}, CreditNullifier: previous.Nullifier, PreviousExpiry: core.ElementToBigInt(previous.State.Note.Expiry)}
	}
	return s.persistPlan(ctx, plan)
}
func (s *WalletSession) PrepareCreateCredit(ctx context.Context, p CreateCreditPreparation) (result *Operation, failure error) {
	defer func() { failure = wrapPreparation("prepareCreateCredit", failure) }()
	var err error
	p, err = frozen(p)
	if err != nil {
		return nil, err
	}
	return s.prepareCreditCreation(ctx, p, nil, nil)
}
func (s *WalletSession) PrepareTopUpCredit(ctx context.Context, p TopUpCreditPreparation) (result *Operation, failure error) {
	defer func() { failure = wrapPreparation("prepareTopUpCredit", failure) }()
	var err error
	p, err = frozen(p)
	if err != nil {
		return nil, err
	}
	_, block, _, e := s.frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	old, membership, e := s.creditInput(ctx, p.TokenAddress, p.Credit, "owner", block)
	if e != nil {
		return nil, e
	}
	if e = positive(p.AdditionalAmount, 126); e != nil {
		return nil, e
	}
	amount := new(big.Int).Add(core.ElementToBigInt(old.State.Note.Amount), p.AdditionalAmount)
	expiry := p.Expiry
	if expiry == nil {
		expiry = core.ElementToBigInt(old.State.Note.Expiry)
	}
	return s.prepareCreditCreation(ctx, CreateCreditPreparation{TokenAddress: p.TokenAddress, Amount: amount, Expiry: expiry, Funding: p.Funding, Deadline: p.Deadline, Payload: p.Payload, SingleLimit: p.SingleLimit, DailyLimit: p.DailyLimit, MonthlyLimit: p.MonthlyLimit}, &old, &membership)
}
func (s *WalletSession) PrepareCancelCredit(ctx context.Context, p CancelCreditPreparation) (result *Operation, failure error) {
	defer func() { failure = wrapPreparation("prepareCancelCredit", failure) }()
	c, block, deadline, e := s.frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	old, m, e := s.creditInput(ctx, p.TokenAddress, p.Credit, "owner", block)
	if e != nil {
		return nil, e
	}
	params := protocol.CancelCreditParams{Token: p.TokenAddress, Source: protocol.InputSource{TreeId: new(big.Int).SetUint64(old.TreeID), MerkleRoot: m.Root}, Nullifier: old.Nullifier, Owner: s.wallet.identity, Amount: core.ElementToBigInt(old.State.Note.Amount), Deadline: new(big.Int).SetUint64(deadline)}
	funding, _ := protocol.NormalizeFunding(protocol.FundingParams{})
	return s.persistPlan(ctx, operationPlan{Kind: "cancel-credit", Target: "evm", Context: c, Block: block, Deadline: deadline, Funding: funding, FundingMode: "none", Capacity: 1, NeedsAuthorization: true, CreditInputs: []StoredCredit{old}, CreditMemberships: []protocolrpc.Membership{m}, CancelCredit: &params})
}
func (s *WalletSession) PrepareDrawCredit(ctx context.Context, p DrawCreditPreparation) (result *Operation, failure error) {
	defer func() { failure = wrapPreparation("prepareDrawCredit", failure) }()
	var err error
	p, err = frozen(p)
	if err != nil {
		return nil, err
	}
	c, block, deadline, e := s.frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	profile, e := protocol.SelectCreditDrawCircuit(len(p.Credits))
	if e != nil {
		return nil, e
	}
	params := protocol.CreditDrawParams{Token: p.TokenAddress, Spender: s.wallet.identity, TotalAmount: new(big.Int), Deadline: new(big.Int).SetUint64(deadline)}
	inputs := []StoredCredit{}
	memberships := []protocolrpc.Membership{}
	deliveries := []protocol.Delivery{}
	amounts := []*big.Int{}
	seen := map[common.Hash]bool{}
	for _, entry := range p.Credits {
		if err := validatePayload(entry.Payload); err != nil {
			return nil, err
		}
		if seen[entry.Credit] {
			return nil, errors.New("duplicate Credit")
		}
		seen[entry.Credit] = true
		if e = positive(entry.Amount, 126); e != nil {
			return nil, e
		}
		old, m, e := s.creditInput(ctx, p.TokenAddress, entry.Credit, "spender", block)
		if e != nil {
			return nil, e
		}
		note := old.State.Note
		if core.ElementToBigInt(note.Expiry).Uint64() <= block.Timestamp || entry.Amount.Cmp(core.ElementToBigInt(note.Amount)) > 0 {
			return nil, errors.New("draw exceeds an unexpired Credit")
		}
		updated, e := core.ApplyCreditDrawUsage(note, core.BigIntToElement(entry.Amount), block.Timestamp)
		if e != nil {
			return nil, e
		}
		remaining := new(big.Int).Sub(core.ElementToBigInt(note.Amount), entry.Amount)
		remainder := common.Hash{}
		recovery := common.Hash{}
		if remaining.Sign() > 0 {
			nonce, e := core.DeriveRemainderNonce(note.ChainID, note.ProtocolAddress, note.TokenAddress, core.BigIntToElement(old.Nullifier.Big()), note.CreditSharedKey)
			if e != nil {
				return nil, e
			}
			child := old.State
			child.Note = updated
			child.Note.Payload = clonePayload(entry.Payload)
			child.Note.Amount = core.BigIntToElement(remaining)
			child.Note.Nonce = nonce
			remainder, e = creditCommitment(child.Note)
			if e != nil {
				return nil, e
			}
			pad, e := core.ComputeCreditRecoveryPad(note.ChainID, note.ProtocolAddress, note.TokenAddress, core.BigIntToElement(old.Nullifier.Big()), note.CreditSharedKey)
			if e != nil {
				return nil, e
			}
			value := core.BigIntToElement(remaining)
			value.Add(&value, &pad)
			recovery = fieldHash(value)
			ds, e := creditDeliveries(child)
			if e != nil {
				return nil, e
			}
			deliveries = append(deliveries, ds...)
		} else {
			deliveries = append(deliveries, protocol.Delivery{}, protocol.Delivery{})
		}
		inputs = append(inputs, old)
		memberships = append(memberships, m)
		amounts = append(amounts, new(big.Int).Set(entry.Amount))
		params.Nullifiers = append(params.Nullifiers, old.Nullifier)
		params.RemainderCommitments = append(params.RemainderCommitments, remainder)
		params.RecoveryAmounts = append(params.RecoveryAmounts, recovery)
		params.Expiries = append(params.Expiries, core.ElementToBigInt(note.Expiry))
		params.TotalAmount.Add(params.TotalAmount, entry.Amount)
	}
	sources, e := sourcesFor(memberships)
	if e != nil {
		return nil, e
	}
	params.Sources = sources
	funding, _ := protocol.NormalizeFunding(protocol.FundingParams{})
	return s.persistPlan(ctx, operationPlan{Kind: "credit-draw", Target: "evm", Context: c, Block: block, Deadline: deadline, Funding: funding, FundingMode: "none", Capacity: profile.MaxSize, NeedsAuthorization: true, CreditInputs: inputs, CreditMemberships: memberships, DrawCredit: &params, Deliveries: deliveries, DrawAmounts: amounts})
}
func (o *Operation) creditPrivateInputs(private map[string]any) (map[string]any, error) {
	if o.plan.Kind == "cancel-credit" {
		note := o.plan.CreditInputs[0].State.Note
		m := o.plan.CreditMemberships[0]
		private["spender_payout"] = note.SpenderPayout.String()
		private["spender_wallet_identity"] = note.SpenderWalletIdentity.String()
		private["owner_wallet_identity"] = note.OwnerWalletIdentity.String()
		private["input_nonce"] = note.Nonce.String()
		private["expiry"] = note.Expiry.String()
		private["credit_shared_key"] = note.CreditSharedKey.String()
		addCreditStateWitness(private, note.CreditSpendingState)
		private["leaf_index"] = new(big.Int).SetUint64(uint64(m.LeafIndex)).String()
		path := []string{}
		for _, h := range m.Path {
			path = append(path, h.Big().String())
		}
		private["merkle_path"] = path
		return private, nil
	}
	capacity := o.plan.Capacity
	amounts, nonces, indices, sources := []string{}, []string{}, []string{}, []string{}
	paths := [][]string{}
	if o.plan.Kind == "credit-draw" {
		identities, payouts, keys, draws := []string{}, []string{}, []string{}, []string{}
		for i, input := range o.plan.CreditInputs {
			note := input.State.Note
			amounts = append(amounts, note.Amount.String())
			nonces = append(nonces, note.Nonce.String())
			indices = append(indices, new(big.Int).SetUint64(uint64(input.LeafIndex)).String())
			path := []string{}
			for _, h := range o.plan.CreditMemberships[i].Path {
				path = append(path, h.Big().String())
			}
			paths = append(paths, path)
			idx := -1
			for j, source := range o.plan.DrawCredit.Sources {
				if source.TreeId.Uint64() == input.TreeID {
					idx = j
				}
			}
			if idx < 0 {
				return nil, errors.New("missing Credit source")
			}
			sources = append(sources, big.NewInt(int64(idx)).String())
			identities = append(identities, note.OwnerWalletIdentity.String())
			payouts = append(payouts, note.OwnerPayout.String())
			keys = append(keys, note.CreditSharedKey.String())
			draws = append(draws, o.plan.DrawAmounts[i].String())
		}
		private["spender_wallet_identity"] = o.plan.CreditInputs[0].State.Note.SpenderWalletIdentity.String()
		private["owner_payouts"] = padStrings(payouts, capacity)
		private["owner_wallet_identities"] = padStrings(identities, capacity)
		private["credit_shared_keys"] = padStrings(keys, capacity)
		private["draw_amounts"] = padStrings(draws, capacity)
		for column, name := range []string{"single_limits", "daily_limits", "monthly_limits", "last_days", "daily_spents", "last_months", "monthly_spents"} {
			values := []string{}
			for _, input := range o.plan.CreditInputs {
				values = append(values, input.State.Note.CreditSpendingState.Fields()[column].String())
			}
			private[name] = padStrings(values, capacity)
		}
	} else {
		for i, input := range o.plan.Inputs {
			amounts = append(amounts, input.Note.Amount.String())
			nonces = append(nonces, input.Note.Nonce.String())
			indices = append(indices, new(big.Int).SetUint64(uint64(input.LeafIndex)).String())
			path := []string{}
			for _, h := range o.plan.Memberships[i].Path {
				path = append(path, h.Big().String())
			}
			paths = append(paths, path)
			idx := -1
			for j, source := range o.plan.CreateCredit.Sources {
				if source.TreeId.Uint64() == input.TreeID {
					idx = j
				}
			}
			if idx < 0 {
				return nil, errors.New("missing funds source")
			}
			sources = append(sources, big.NewInt(int64(idx)).String())
		}
		note := o.plan.CreditOutput.Note
		private["spender_payout"] = note.SpenderPayout.String()
		private["spender_wallet_identity"] = note.SpenderWalletIdentity.String()
		private["owner_wallet_identity"] = note.OwnerWalletIdentity.String()
		private["credit_amount"] = note.Amount.String()
		private["credit_nonce"] = note.Nonce.String()
		private["credit_shared_key"] = note.CreditSharedKey.String()
		if o.plan.Kind == "create-credit" {
			private["single_limit"] = note.SingleLimit.String()
			private["daily_limit"] = note.DailyLimit.String()
			private["monthly_limit"] = note.MonthlyLimit.String()
		} else {
			addCreditStateWitness(private, note.CreditSpendingState)
		}
		private["change_amount"] = "0"
		private["change_nonce"] = "0"
		if len(o.plan.Outputs) > 0 {
			private["change_amount"] = o.plan.Outputs[0].Note.Amount.String()
			private["change_nonce"] = o.plan.Outputs[0].Note.Nonce.String()
		}
		if o.plan.Kind == "top-up-credit" {
			previous := o.plan.CreditInputs[0]
			private["previous_single_limit"] = previous.State.Note.SingleLimit.String()
			private["previous_daily_limit"] = previous.State.Note.DailyLimit.String()
			private["previous_monthly_limit"] = previous.State.Note.MonthlyLimit.String()
			private["previous_credit_amount"] = previous.State.Note.Amount.String()
			private["previous_credit_nonce"] = previous.State.Note.Nonce.String()
			private["previous_credit_leaf_index"] = new(big.Int).SetUint64(uint64(previous.LeafIndex)).String()
			path := []string{}
			for _, h := range o.plan.CreditMemberships[0].Path {
				path = append(path, h.Big().String())
			}
			private["previous_credit_merkle_path"] = path
		}
	}
	for len(paths) < capacity {
		paths = append(paths, zeroPath())
	}
	private["input_amounts"] = padStrings(amounts, capacity)
	private["input_nonces"] = padStrings(nonces, capacity)
	private["leaf_indices"] = padStrings(indices, capacity)
	private["merkle_paths"] = paths
	private["input_source_indices"] = padStrings(sources, capacity)
	return private, nil
}
func (o *Operation) encodeCredit(proof []byte, caller *common.Address) (*protocol.Calldata, error) {
	switch o.plan.Kind {
	case "create-credit":
		return protocol.BuildCreateCreditCalldata(protocol.BuildCreateCreditCalldataParams{Context: o.plan.Context, Proof: proof, Params: *o.plan.CreateCredit, Deliveries: o.deliveries(), Funding: o.plan.Funding, Caller: caller})
	case "credit-draw":
		return protocol.BuildCreditDrawCalldata(protocol.BuildCreditDrawCalldataParams{Context: o.plan.Context, Proof: proof, Params: *o.plan.DrawCredit, Deliveries: o.deliveries()})
	case "cancel-credit":
		return protocol.BuildCancelCreditCalldata(protocol.BuildCancelCreditCalldataParams{Context: o.plan.Context, Proof: proof, Params: *o.plan.CancelCredit})
	case "top-up-credit":
		return protocol.BuildTopUpCreditCalldata(protocol.BuildTopUpCreditCalldataParams{Context: o.plan.Context, Proof: proof, Params: *o.plan.TopUpCredit, Deliveries: o.deliveries(), Funding: o.plan.Funding, Caller: caller})
	}
	return nil, errors.New("unsupported Credit operation")
}

func addCreditStateWitness(private map[string]any, state core.CreditSpendingState) {
	for i, name := range []string{"single_limit", "daily_limit", "monthly_limit", "last_day", "daily_spent", "last_month", "monthly_spent"} {
		private[name] = state.Fields()[i].String()
	}
}
