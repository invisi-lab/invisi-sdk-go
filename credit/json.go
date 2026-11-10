package credit

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func canonicalField(value string) (fr.Element, error) {
	n, ok := new(big.Int).SetString(value, 10)
	if !ok || n.Sign() < 0 || n.Cmp(fr.Modulus()) >= 0 || n.String() != value {
		return fr.Element{}, fmt.Errorf("noncanonical Credit field")
	}
	return primitives.BigIntToElement(n), nil
}
func fieldsJSON(values map[string]fr.Element) ([]byte, error) {
	out := map[string]string{}
	for k, v := range values {
		out[k] = v.String()
	}
	return json.Marshal(out)
}
func parseFields(raw []byte, names ...string) (map[string]fr.Element, error) {
	var values map[string]string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	if len(values) != len(names) {
		return nil, fmt.Errorf("unsupported or missing Credit fields")
	}
	out := map[string]fr.Element{}
	for _, k := range names {
		v, ok := values[k]
		if !ok {
			return nil, fmt.Errorf("missing Credit field %s", k)
		}
		e, err := canonicalField(v)
		if err != nil {
			return nil, err
		}
		out[k] = e
	}
	return out, nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	return fieldsJSON(map[string]fr.Element{"singleLimit": p.SingleLimit, "dailyLimit": p.DailyLimit, "monthlyLimit": p.MonthlyLimit})
}
func (p *Policy) UnmarshalJSON(raw []byte) error {
	v, err := parseFields(raw, "singleLimit", "dailyLimit", "monthlyLimit")
	if err != nil {
		return err
	}
	*p = Policy{v["singleLimit"], v["dailyLimit"], v["monthlyLimit"]}
	return nil
}
func (s State) MarshalJSON() ([]byte, error) {
	values := map[string]string{"chainId": s.ChainID.String(), "protocolAddress": s.ProtocolAddress.String(), "tokenAddress": s.TokenAddress.String(), "spenderWalletIdentity": s.SpenderWalletIdentity.String(), "ownerWalletIdentity": s.OwnerWalletIdentity.String(), "revocationShard": s.RevocationShard.String(), "revocationSlot": s.RevocationSlot.String(), "nonce": s.Nonce.String(), "expiry": s.Expiry.String(), "creditSharedKey": s.SharedKey.String(), "availableCredit": s.AvailableCredit.String()}
	for i, k := range []string{"singleLimit", "dailyLimit", "monthlyLimit"} {
		values[k] = s.Policy.Fields()[i].String()
	}
	for i, k := range []string{"lastDay", "dailySpent", "lastMonth", "monthlySpent", "totalSpent"} {
		values[k] = s.Usage.Fields()[i].String()
	}
	return json.Marshal(struct {
		Version   int               `json:"version"`
		Unlimited bool              `json:"unlimited"`
		Fields    map[string]string `json:"fields"`
		Payload   []byte            `json:"payload,omitempty"`
	}{2, s.Unlimited, values, s.Payload})
}
func (s *State) UnmarshalJSON(raw []byte) error {
	var envelope struct {
		Version   int             `json:"version"`
		Unlimited *bool           `json:"unlimited"`
		Fields    json.RawMessage `json:"fields"`
		Payload   []byte          `json:"payload"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	if envelope.Version != 2 || envelope.Unlimited == nil {
		return fmt.Errorf("unsupported Credit state schema")
	}
	v, err := parseFields(envelope.Fields, "chainId", "protocolAddress", "tokenAddress", "spenderWalletIdentity", "ownerWalletIdentity", "revocationShard", "revocationSlot", "nonce", "expiry", "creditSharedKey", "singleLimit", "dailyLimit", "monthlyLimit", "availableCredit", "lastDay", "dailySpent", "lastMonth", "monthlySpent", "totalSpent")
	if err != nil {
		return err
	}
	next := State{ProtocolContext: primitives.ProtocolContext{ChainID: v["chainId"], ProtocolAddress: v["protocolAddress"]}, TokenAddress: v["tokenAddress"], SpenderWalletIdentity: v["spenderWalletIdentity"], OwnerWalletIdentity: v["ownerWalletIdentity"], RevocationShard: v["revocationShard"], RevocationSlot: v["revocationSlot"], Nonce: v["nonce"], Expiry: v["expiry"], SharedKey: v["creditSharedKey"], Policy: Policy{v["singleLimit"], v["dailyLimit"], v["monthlyLimit"]}, AvailableCredit: v["availableCredit"], Unlimited: *envelope.Unlimited, Usage: Usage{v["lastDay"], v["dailySpent"], v["lastMonth"], v["monthlySpent"], v["totalSpent"]}, Payload: append([]byte(nil), envelope.Payload...)}
	if err = next.Validate(); err != nil {
		return err
	}
	*s = next
	return nil
}
func (c Context) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ChainID, ProtocolAddress string
		TokenAddress             common.Address
		Timestamp                uint64
	}{c.ChainID.String(), c.ProtocolAddress.String(), c.TokenAddress, c.Timestamp})
}
func (c *Context) UnmarshalJSON(raw []byte) error {
	var value struct {
		ChainID, ProtocolAddress string
		TokenAddress             common.Address
		Timestamp                uint64
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	chain, err := canonicalField(value.ChainID)
	if err != nil {
		return err
	}
	address, err := canonicalField(value.ProtocolAddress)
	if err != nil {
		return err
	}
	*c = Context{ProtocolContext: primitives.ProtocolContext{ChainID: chain, ProtocolAddress: address}, TokenAddress: value.TokenAddress, Timestamp: value.Timestamp}
	return nil
}

type locatedJSON struct {
	State     State
	TreeID    uint64
	LeafIndex uint16
	Path      [16]common.Hash
	Debit     string
}

func (s LocatedState) MarshalJSON() ([]byte, error) {
	return json.Marshal(locatedJSON{s.State, s.TreeID, s.LeafIndex, s.Path, s.Debit.String()})
}
func (s *LocatedState) UnmarshalJSON(raw []byte) error {
	var value locatedJSON
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	debit, err := canonicalField(value.Debit)
	if err != nil {
		return err
	}
	*s = LocatedState{value.State, value.TreeID, value.LeafIndex, value.Path, debit}
	return nil
}

type intentAlias Intent
type intentJSON struct {
	*intentAlias
	OutputIdentity string
}

func (i Intent) MarshalJSON() ([]byte, error) {
	return json.Marshal(intentJSON{(*intentAlias)(&i), i.OutputIdentity.String()})
}
func (i *Intent) UnmarshalJSON(raw []byte) error {
	var next Intent
	value := intentJSON{intentAlias: (*intentAlias)(&next)}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	identity, err := canonicalField(value.OutputIdentity)
	if err != nil {
		return err
	}
	next.OutputIdentity = identity
	*i = next
	return nil
}

type changeJSON struct {
	Policy                  Policy
	Expiry                  string
	Unlimited               *bool `json:",omitempty"`
	IncreaseAvailableCredit string
	DecreaseAvailableCredit string
}

func (c Change) MarshalJSON() ([]byte, error) {
	return json.Marshal(changeJSON{c.Policy, c.Expiry.String(), c.Unlimited, c.IncreaseAvailableCredit.String(), c.DecreaseAvailableCredit.String()})
}
func (c *Change) UnmarshalJSON(raw []byte) error {
	var v changeJSON
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	expiry, err := canonicalField(v.Expiry)
	if err != nil {
		return err
	}
	inc, err := canonicalField(v.IncreaseAvailableCredit)
	if err != nil {
		return err
	}
	dec, err := canonicalField(v.DecreaseAvailableCredit)
	if err != nil {
		return err
	}
	*c = Change{v.Policy, expiry, v.Unlimited, inc, dec}
	return nil
}
