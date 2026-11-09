// Package credit models private postpaid grants independently of ordinary payment Notes.
package credit

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"math/big"
)

type Policy struct{ SingleLimit, DailyLimit, MonthlyLimit, LifetimeLimit fr.Element }
type Usage struct{ LastDay, DailySpent, LastMonth, MonthlySpent, TotalSpent fr.Element }
type State struct {
	core.ProtocolContext
	Policy
	Usage
	TokenAddress, SpenderWalletIdentity, OwnerWalletIdentity  fr.Element
	RevocationShard, RevocationSlot, Nonce, Expiry, SharedKey fr.Element
	Payload                                                   []byte
}
type Transition struct {
	State                                      State
	Nullifier, StateCommitment, RecoveryAmount fr.Element
}

func (p Policy) Fields() []fr.Element {
	return []fr.Element{p.SingleLimit, p.DailyLimit, p.MonthlyLimit, p.LifetimeLimit}
}
func (u Usage) Fields() []fr.Element {
	return []fr.Element{u.LastDay, u.DailySpent, u.LastMonth, u.MonthlySpent, u.TotalSpent}
}
func (s State) Clone() State { s.Payload = append([]byte(nil), s.Payload...); return s }
func uintField(v fr.Element, bits int, name string) error {
	if core.ElementToBigInt(v).BitLen() > bits {
		return fmt.Errorf("%s exceeds uint%d", name, bits)
	}
	return nil
}
func (s State) Validate() error {
	for _, x := range []struct {
		v        fr.Element
		bits     int
		name     string
		positive bool
	}{
		{s.ChainID, 64, "chain ID", false}, {s.ProtocolAddress, 160, "protocol", false}, {s.TokenAddress, 160, "token", false},
		{s.SpenderWalletIdentity, 160, "spender", true}, {s.OwnerWalletIdentity, 160, "owner", true},
		{s.RevocationShard, 64, "revocation shard", true}, {s.RevocationSlot, 16, "revocation slot", false}, {s.Expiry, 64, "expiry", true}} {
		if err := uintField(x.v, x.bits, x.name); err != nil {
			return err
		}
		if x.positive && x.v.IsZero() {
			return fmt.Errorf("%s must be positive", x.name)
		}
	}
	if s.SharedKey.IsZero() {
		return fmt.Errorf("grant shared key must be nonzero")
	}
	for _, v := range s.Policy.Fields() {
		if err := uintField(v, 126, "limit"); err != nil {
			return err
		}
	}
	for _, v := range []fr.Element{s.DailySpent, s.MonthlySpent, s.TotalSpent} {
		if err := uintField(v, 128, "usage"); err != nil {
			return err
		}
	}
	for _, v := range []fr.Element{s.LastDay, s.LastMonth} {
		if err := uintField(v, 64, "period"); err != nil {
			return err
		}
	}
	if s.LastDay.IsZero() && !s.DailySpent.IsZero() || s.LastMonth.IsZero() && !s.MonthlySpent.IsZero() {
		return fmt.Errorf("unused period requires zero usage")
	}
	return core.ValidatePayload(s.Payload)
}
func (s State) Commitment() (fr.Element, error) {
	if err := s.Validate(); err != nil {
		return fr.Element{}, err
	}
	key, err := core.Poseidon2Hash(core.ElementFromUint64(7), s.SharedKey)
	if err != nil {
		return fr.Element{}, err
	}
	fields := []fr.Element{core.ElementFromUint64(5), s.ChainID, s.ProtocolAddress, s.SpenderWalletIdentity, s.OwnerWalletIdentity, s.TokenAddress, s.RevocationShard, s.RevocationSlot, s.Nonce, s.Expiry, key}
	fields = append(fields, s.Policy.Fields()...)
	fields = append(fields, s.Usage.Fields()...)
	return core.Poseidon2Hash(fields...)
}
func (s State) Nullifier() (fr.Element, error) {
	c, err := s.Commitment()
	if err != nil {
		return fr.Element{}, err
	}
	return core.Poseidon2Hash(core.ElementFromUint64(6), s.ChainID, s.SharedKey, c)
}
func (s State) secrets(nf fr.Element) (nonce, mask fr.Element, err error) {
	args := []fr.Element{s.ChainID, s.ProtocolAddress, s.TokenAddress, nf, s.SharedKey}
	nonce, err = core.Poseidon2Hash(append([]fr.Element{core.ElementFromUint64(10)}, args...)...)
	if err != nil {
		return
	}
	mask, err = core.Poseidon2Hash(append([]fr.Element{core.ElementFromUint64(9)}, args...)...)
	if err == nil && nonce == s.Nonce {
		err = fmt.Errorf("successor nonce must change")
	}
	return
}
func (s State) drawUsage(debit fr.Element, timestamp uint64) (Usage, error) {
	if err := s.Validate(); err != nil {
		return Usage{}, err
	}
	if err := uintField(debit, 126, "debit"); err != nil {
		return Usage{}, err
	}
	if debit.IsZero() {
		return Usage{}, fmt.Errorf("debit must be positive")
	}
	if !s.SingleLimit.IsZero() && debit.Cmp(&s.SingleLimit) > 0 {
		return Usage{}, fmt.Errorf("single limit exceeded")
	}
	p := Periods(timestamp)
	u := s.Usage
	update := func(limit, last, spent fr.Element, current uint64) (fr.Element, fr.Element, error) {
		if current == 0 || current < core.ElementToBigInt(last).Uint64() {
			return fr.Element{}, fr.Element{}, fmt.Errorf("period rollback")
		}
		total := new(big.Int).Set(core.ElementToBigInt(debit))
		if current == core.ElementToBigInt(last).Uint64() {
			total.Add(total, core.ElementToBigInt(spent))
		}
		if total.BitLen() > 128 || !limit.IsZero() && total.Cmp(core.ElementToBigInt(limit)) > 0 {
			return fr.Element{}, fr.Element{}, fmt.Errorf("period usage exceeds limit or uint128")
		}
		return core.ElementFromUint64(current), core.BigIntToElement(total), nil
	}
	var err error
	u.LastDay, u.DailySpent, err = update(s.DailyLimit, s.LastDay, s.DailySpent, p.CurrentDay)
	if err != nil {
		return Usage{}, err
	}
	u.LastMonth, u.MonthlySpent, err = update(s.MonthlyLimit, s.LastMonth, s.MonthlySpent, p.CurrentMonth)
	if err != nil {
		return Usage{}, err
	}
	total := new(big.Int).Add(core.ElementToBigInt(s.TotalSpent), core.ElementToBigInt(debit))
	if total.BitLen() > 128 || !s.LifetimeLimit.IsZero() && total.Cmp(core.ElementToBigInt(s.LifetimeLimit)) > 0 {
		return Usage{}, fmt.Errorf("lifetime usage exceeds limit or uint128")
	}
	u.TotalSpent = core.BigIntToElement(total)
	return u, nil
}
func (s State) Draw(debit fr.Element, timestamp uint64) (Transition, error) {
	usage, err := s.drawUsage(debit, timestamp)
	if err != nil {
		return Transition{}, err
	}
	nf, err := s.Nullifier()
	if err != nil {
		return Transition{}, err
	}
	nonce, mask, err := s.secrets(nf)
	if err != nil {
		return Transition{}, err
	}
	next := s.Clone()
	next.Usage = usage
	next.Nonce = nonce
	c, err := next.Commitment()
	if err != nil {
		return Transition{}, err
	}
	var recovery fr.Element
	recovery.Add(&debit, &mask)
	return Transition{next, nf, c, recovery}, nil
}

// Recover authenticates a draw against the previous private grant and canonical event fields.
func (s State) Recover(nf, commitment, recovery fr.Element, timestamp uint64) (State, error) {
	expected, err := s.Nullifier()
	if err != nil {
		return State{}, err
	}
	if nf.IsZero() || nf != expected {
		return State{}, fmt.Errorf("previous grant nullifier mismatch")
	}
	_, mask, err := s.secrets(nf)
	if err != nil {
		return State{}, err
	}
	var debit fr.Element
	debit.Sub(&recovery, &mask)
	next, err := s.Draw(debit, timestamp)
	if err != nil {
		return State{}, err
	}
	if next.StateCommitment != commitment || commitment.IsZero() {
		return State{}, fmt.Errorf("recovered grant commitment mismatch")
	}
	next.State.Payload = nil
	return next.State, nil
}

// Update preserves grant identity and every usage counter, including when limits are disabled.
func (s State) Update(policy Policy, expiry fr.Element, deadline uint64) (Transition, error) {
	if err := s.Validate(); err != nil {
		return Transition{}, err
	}
	d := core.ElementFromUint64(deadline)
	if d.Cmp(&s.Expiry) > 0 || d.Cmp(&expiry) > 0 {
		return Transition{}, fmt.Errorf("expiry before request deadline")
	}
	nf, err := s.Nullifier()
	if err != nil {
		return Transition{}, err
	}
	nonce, _, err := s.secrets(nf)
	if err != nil {
		return Transition{}, err
	}
	next := s.Clone()
	next.Policy = policy
	next.Expiry = expiry
	next.Nonce = nonce
	c, err := next.Commitment()
	if err != nil {
		return Transition{}, err
	}
	return Transition{State: next, Nullifier: nf, StateCommitment: c}, nil
}
