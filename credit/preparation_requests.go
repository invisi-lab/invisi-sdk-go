package credit

import (
	"context"
	"fmt"
	"math/big"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

type Fee struct {
	Recipient common.Address
	Amount    *big.Int
}
type CreatePreparation struct {
	TokenAddress                          common.Address
	Spender                               string
	SingleLimit, DailyLimit, MonthlyLimit fr.Element
	Expiry                                uint64
	AvailableCredit                       *big.Int
	Unlimited                             bool
	Payload                               []byte
	Deadline                              uint64
	Fee                                   *Fee
}
type UpdatePreparation struct {
	TokenAddress common.Address
	Credit       common.Hash
	UpdatePatch
	Deadline uint64
	Fee      *Fee
}
type DrawMember struct {
	Credit  common.Hash
	Amount  *big.Int
	Payload []byte
}
type DrawPreparation struct {
	TokenAddress common.Address
	Credits      []DrawMember
	Recipient    string
	Payload      []byte
	Deadline     uint64
	Fee          *Fee
}
type RevokePreparation struct {
	TokenAddress common.Address
	Owner        common.Address
	Shard        uint64
	Slot         uint16
	Deadline     uint64
	Fee          *Fee
}

type PreparationInput struct {
	Located  LocatedState
	Delivery DeliveryState
	Source   protocol.InputSource
}

// PreparationBuilder resolves current or pinned facts; authorization stays with the application.
type PreparationBuilder struct {
	Frame      func(context.Context, common.Address, uint64) (Context, uint64, error)
	Resolve    func(context.Context, string) (primitives.ReceivingAddress, error)
	Owner      func() (string, error)
	Allocation func(context.Context, Context) (uint64, uint16, error)
	Input      func(context.Context, Context, common.Hash, Role) (PreparationInput, error)
}

func requestFee(f *Fee) (common.Address, *big.Int, error) {
	if f == nil {
		return common.Address{}, new(big.Int), nil
	}
	amount, err := normalizeFee(f.Recipient, f.Amount)
	return f.Recipient, amount, err
}
func Seal(d DeliveryState) ([]protocol.Delivery, error) {
	commitment, err := d.State.Commitment()
	if err != nil {
		return nil, err
	}
	ds := make([]protocol.Delivery, 2)
	for i, role := range []Role{Owner, Spender} {
		keys, err := roleKey(d, role)
		if err != nil {
			return nil, err
		}
		clue, err := primitives.CreateClue(keys.ClueKey)
		if err != nil {
			return nil, err
		}
		identity := d.State.OwnerWalletIdentity
		if role == Spender {
			identity = d.State.SpenderWalletIdentity
		}
		ciphertext, err := Encrypt(d, DeliveryContext{ProtocolContext: d.State.ProtocolContext, TokenAddress: d.State.TokenAddress, Commitment: commitment, RecipientIdentity: identity, Clue: clue}, role)
		if err != nil {
			return nil, err
		}
		ds[i] = protocol.Delivery{Ciphertext: ciphertext, Clue: clue}
	}
	return ds, nil
}
func (b PreparationBuilder) Create(ctx context.Context, p CreatePreparation) (*Prepared, error) {
	p = p.Clone()
	c, deadline, err := b.Frame(ctx, p.TokenAddress, p.Deadline)
	if err != nil {
		return nil, err
	}
	if p.Expiry <= c.Timestamp || p.Expiry < deadline {
		return nil, fmt.Errorf("Credit expiry must cover request deadline and be in the future")
	}
	r, f, err := requestFee(p.Fee)
	if err != nil {
		return nil, err
	}
	ownerAddress, err := b.Owner()
	if err != nil {
		return nil, err
	}
	owner, err := b.Resolve(ctx, ownerAddress)
	if err != nil {
		return nil, err
	}
	spender, err := b.Resolve(ctx, p.Spender)
	if err != nil {
		return nil, err
	}
	shard, slot, err := b.Allocation(ctx, c)
	if err != nil {
		return nil, err
	}
	key, err := GenerateSharedKey()
	if err != nil {
		return nil, err
	}
	nonce, err := primitives.GenerateNonce()
	if err != nil {
		return nil, err
	}
	amount := new(big.Int)
	if p.AvailableCredit != nil {
		amount.Set(p.AvailableCredit)
	}
	if amount.Sign() < 0 || amount.BitLen() > 128 {
		return nil, fmt.Errorf("available credit must be uint128")
	}
	state := State{ProtocolContext: c.ProtocolContext, Policy: Policy{SingleLimit: p.SingleLimit, DailyLimit: p.DailyLimit, MonthlyLimit: p.MonthlyLimit}, TokenAddress: primitives.BigIntToElement(p.TokenAddress.Big()), OwnerWalletIdentity: owner.OwnerAddress, SpenderWalletIdentity: spender.OwnerAddress, RevocationShard: primitives.ElementFromUint64(shard), RevocationSlot: primitives.ElementFromUint64(uint64(slot)), Nonce: nonce, Expiry: primitives.ElementFromUint64(p.Expiry), CreditSharedKey: key, AvailableCredit: primitives.BigIntToElement(amount), Unlimited: p.Unlimited, Payload: append([]byte(nil), p.Payload...)}
	d := DeliveryState{State: state, OwnerReceivingKeys: primitives.ReceivingKeys{ViewingPublicKey: owner.ViewingPublicKey, ClueKey: owner.ClueKey}, SpenderReceivingKeys: primitives.ReceivingKeys{ViewingPublicKey: spender.ViewingPublicKey, ClueKey: spender.ClueKey}}
	ds, err := Seal(d)
	if err != nil {
		return nil, err
	}
	commitment, err := state.Commitment()
	if err != nil {
		return nil, err
	}
	return PrepareCreate(c, CreateParams{Token: p.TokenAddress, Owner: common.BigToAddress(primitives.ElementToBigInt(owner.OwnerAddress)), RevocationShard: shard, RevocationSlot: slot, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: new(big.Int).SetUint64(deadline), FeeRecipient: r, FeeAmount: f}, ds, state)
}
func (b PreparationBuilder) Update(ctx context.Context, p UpdatePreparation) (*Prepared, error) {
	p = p.Clone()
	c, deadline, err := b.Frame(ctx, p.TokenAddress, p.Deadline)
	if err != nil {
		return nil, err
	}
	r, f, err := requestFee(p.Fee)
	if err != nil {
		return nil, err
	}
	input, err := b.Input(ctx, c, p.Credit, Owner)
	if err != nil {
		return nil, err
	}
	change := p.UpdatePatch.Change(input.Located.State)
	next, err := input.Located.State.Update(change, deadline)
	if err != nil {
		return nil, err
	}
	d := input.Delivery
	d.State = next.State
	ds, err := Seal(d)
	if err != nil {
		return nil, err
	}
	s := input.Located.State
	return PrepareUpdate(c, UpdateParams{Token: p.TokenAddress, Owner: common.BigToAddress(primitives.ElementToBigInt(s.OwnerWalletIdentity)), Source: input.Source, Nullifier: common.Hash(next.Nullifier.Bytes()), CreditCommitment: common.Hash(next.StateCommitment.Bytes()), RevocationShard: primitives.ElementToBigInt(s.RevocationShard).Uint64(), RevocationSlot: uint16(primitives.ElementToBigInt(s.RevocationSlot).Uint64()), Deadline: new(big.Int).SetUint64(deadline), FeeRecipient: r, FeeAmount: f}, ds, input.Located, change)
}
func (b PreparationBuilder) Revoke(ctx context.Context, p RevokePreparation) (*Prepared, error) {
	p = p.Clone()
	c, deadline, err := b.Frame(ctx, p.TokenAddress, p.Deadline)
	if err != nil {
		return nil, err
	}
	r, f, err := requestFee(p.Fee)
	if err != nil {
		return nil, err
	}
	owner := p.Owner
	if owner == (common.Address{}) {
		address, err := b.Owner()
		if err != nil {
			return nil, err
		}
		recipient, err := b.Resolve(ctx, address)
		if err != nil {
			return nil, err
		}
		owner = common.BigToAddress(primitives.ElementToBigInt(recipient.OwnerAddress))
	}
	return PrepareRevoke(c, RevokeParams{Token: p.TokenAddress, Owner: owner, Shard: p.Shard, Slot: p.Slot, Deadline: new(big.Int).SetUint64(deadline), FeeRecipient: r, FeeAmount: f})
}
func (b PreparationBuilder) Draw(ctx context.Context, p DrawPreparation) (*Prepared, error) {
	p = p.Clone()
	c, deadline, err := b.Frame(ctx, p.TokenAddress, p.Deadline)
	if err != nil {
		return nil, err
	}
	if _, err = DrawCapacity(len(p.Credits)); err != nil {
		return nil, err
	}
	r, f, err := requestFee(p.Fee)
	if err != nil {
		return nil, err
	}
	type member struct {
		input   PreparationInput
		next    Transition
		payload []byte
	}
	members := make([]member, len(p.Credits))
	total := new(big.Int)
	seen := map[common.Hash]bool{}
	for i, m := range p.Credits {
		if seen[m.Credit] || m.Amount == nil || m.Amount.Sign() <= 0 || m.Amount.BitLen() > 126 {
			return nil, fmt.Errorf("invalid or duplicate draw member")
		}
		seen[m.Credit] = true
		input, err := b.Input(ctx, c, m.Credit, Spender)
		if err != nil {
			return nil, err
		}
		if primitives.ElementToBigInt(input.Located.State.Expiry).Uint64() < deadline {
			return nil, fmt.Errorf("Credit expiry before request deadline")
		}
		input.Located.Debit = primitives.BigIntToElement(m.Amount)
		next, err := input.Located.State.Draw(input.Located.Debit, c.Timestamp)
		if err != nil {
			return nil, err
		}
		members[i] = member{input, next, m.Payload}
		total.Add(total, m.Amount)
	}
	sort.Slice(members, func(i, j int) bool {
		a, z := members[i].input.Located.State, members[j].input.Located.State
		if cmp := a.OwnerWalletIdentity.Cmp(&z.OwnerWalletIdentity); cmp != 0 {
			return cmp < 0
		}
		if cmp := a.RevocationShard.Cmp(&z.RevocationShard); cmp != 0 {
			return cmp < 0
		}
		return a.RevocationSlot.Cmp(&z.RevocationSlot) < 0
	})
	params := DrawParams{Token: p.TokenAddress, TotalAmount: total, Deadline: new(big.Int).SetUint64(deadline), FeeRecipient: r, FeeAmount: f}
	inputs := []LocatedState{}
	ds := []protocol.Delivery{}
	sources := map[uint64]protocol.InputSource{}
	for _, m := range members {
		s := m.input.Located.State
		payer := common.BigToAddress(primitives.ElementToBigInt(s.OwnerWalletIdentity))
		index := len(params.Payers) - 1
		if index < 0 || params.Payers[index].Payer != payer {
			params.Payers = append(params.Payers, Payer{Payer: payer, Amount: new(big.Int)})
			index++
		}
		params.Payers[index].Amount.Add(params.Payers[index].Amount, primitives.ElementToBigInt(m.input.Located.Debit))
		params.GrantRefs = append(params.GrantRefs, GrantRef{PayerIndex: big.NewInt(int64(index)), Shard: primitives.ElementToBigInt(s.RevocationShard).Uint64(), Slot: uint16(primitives.ElementToBigInt(s.RevocationSlot).Uint64())})
		params.Nullifiers = append(params.Nullifiers, common.Hash(m.next.Nullifier.Bytes()))
		params.CreditCommitments = append(params.CreditCommitments, common.Hash(m.next.StateCommitment.Bytes()))
		params.RecoveryAmounts = append(params.RecoveryAmounts, common.Hash(m.next.RecoveryAmount.Bytes()))
		d := m.input.Delivery
		d.State = m.next.State
		if m.payload != nil {
			d.State.Payload = append([]byte{}, m.payload...)
		}
		deliveries, err := Seal(d)
		if err != nil {
			return nil, err
		}
		ds = append(ds, deliveries...)
		inputs = append(inputs, m.input.Located)
		id := m.input.Located.TreeID
		if old, ok := sources[id]; ok && old.MerkleRoot != m.input.Source.MerkleRoot {
			return nil, fmt.Errorf("inconsistent Credit roots")
		}
		sources[id] = m.input.Source
	}
	ids := make([]uint64, 0, len(sources))
	for id := range sources {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		params.Sources = append(params.Sources, sources[id])
	}
	if total.Cmp(f) <= 0 {
		return nil, fmt.Errorf("Credit fee must leave a positive Note")
	}
	address := p.Recipient
	if address == "" {
		address, err = b.Owner()
		if err != nil {
			return nil, err
		}
	}
	recipient, err := b.Resolve(ctx, address)
	if err != nil {
		return nil, err
	}
	nonce, err := primitives.GenerateNonce()
	if err != nil {
		return nil, err
	}
	note := primitives.Note{ProtocolContext: c.ProtocolContext, OwnerAddress: recipient.OwnerAddress, TokenAddress: primitives.BigIntToElement(p.TokenAddress.Big()), Amount: primitives.BigIntToElement(new(big.Int).Sub(total, f)), Nonce: nonce, Payload: append([]byte(nil), p.Payload...)}
	keys := primitives.ReceivingKeys{ViewingPublicKey: recipient.ViewingPublicKey, ClueKey: recipient.ClueKey}
	return PrepareDrawToNote(c, params, ds, inputs, note, &keys)
}
