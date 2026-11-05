package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Explicit canonical field serialization avoids exposing gnark's internal Montgomery limbs.
func fieldsJSON(values map[string]fr.Element, payload []byte) ([]byte, error) {
	if err := ValidatePayload(payload); err != nil {
		return nil, err
	}
	out := map[string]string{}
	if len(payload) > 0 {
		out["payload"] = base64.StdEncoding.EncodeToString(payload)
	}
	for name, value := range values {
		out[name] = value.String()
	}
	return json.Marshal(out)
}
func parseFields(data []byte, names ...string) (map[string]fr.Element, error) {
	var input map[string]string
	if e := json.Unmarshal(data, &input); e != nil {
		return nil, e
	}
	result := map[string]fr.Element{}
	for _, name := range names {
		raw, ok := input[name]
		if !ok {
			return nil, fmt.Errorf("missing field %s", name)
		}
		n, ok := new(big.Int).SetString(raw, 10)
		if !ok || n.Sign() < 0 || n.Cmp(fr.Modulus()) >= 0 {
			return nil, fmt.Errorf("noncanonical field %s", name)
		}
		result[name] = BigIntToElement(n)
	}
	return result, nil
}
func (n Note) MarshalJSON() ([]byte, error) {
	return fieldsJSON(map[string]fr.Element{"chainId": n.ChainID, "protocolAddress": n.ProtocolAddress, "ownerAddress": n.OwnerAddress, "tokenAddress": n.TokenAddress, "amount": n.Amount, "nonce": n.Nonce}, n.Payload)
}
func (n *Note) UnmarshalJSON(data []byte) error {
	v, e := parseFields(data, "chainId", "protocolAddress", "ownerAddress", "tokenAddress", "amount", "nonce")
	if e != nil {
		return e
	}
	*n = Note{ProtocolContext{v["chainId"], v["protocolAddress"]}, v["ownerAddress"], v["tokenAddress"], v["amount"], v["nonce"], nil}
	n.Payload, e = parsePayloadJSON(data)
	return e
}
func (n CreditNote) MarshalJSON() ([]byte, error) {
	return fieldsJSON(map[string]fr.Element{"chainId": n.ChainID, "protocolAddress": n.ProtocolAddress, "spenderWalletIdentity": n.SpenderWalletIdentity, "ownerWalletIdentity": n.OwnerWalletIdentity, "spenderPayout": n.SpenderPayout, "ownerPayout": n.OwnerPayout, "tokenAddress": n.TokenAddress, "amount": n.Amount, "nonce": n.Nonce, "expiry": n.Expiry, "creditSharedKey": n.CreditSharedKey, "singleLimit": n.SingleLimit, "dailyLimit": n.DailyLimit, "monthlyLimit": n.MonthlyLimit, "lastDay": n.LastDay, "dailySpent": n.DailySpent, "lastMonth": n.LastMonth, "monthlySpent": n.MonthlySpent}, n.Payload)
}
func (n *CreditNote) UnmarshalJSON(data []byte) error {
	v, e := parseFields(data, "chainId", "protocolAddress", "spenderWalletIdentity", "ownerWalletIdentity", "spenderPayout", "ownerPayout", "tokenAddress", "amount", "nonce", "expiry", "creditSharedKey")
	if e != nil {
		return e
	}
	var raw map[string]string
	if e = json.Unmarshal(data, &raw); e != nil {
		return e
	}
	for _, name := range []string{"singleLimit", "dailyLimit", "monthlyLimit", "lastDay", "dailySpent", "lastMonth", "monthlySpent"} {
		if _, ok := raw[name]; !ok {
			raw[name] = "0"
		}
	}
	canonical, e := json.Marshal(raw)
	if e != nil {
		return e
	}
	state, e := parseFields(canonical, "singleLimit", "dailyLimit", "monthlyLimit", "lastDay", "dailySpent", "lastMonth", "monthlySpent")
	if e != nil {
		return e
	}
	*n = CreditNote{ProtocolContext: ProtocolContext{v["chainId"], v["protocolAddress"]}, SpenderWalletIdentity: v["spenderWalletIdentity"], OwnerWalletIdentity: v["ownerWalletIdentity"], SpenderPayout: v["spenderPayout"], OwnerPayout: v["ownerPayout"], TokenAddress: v["tokenAddress"], Amount: v["amount"], Nonce: v["nonce"], Expiry: v["expiry"], CreditSharedKey: v["creditSharedKey"], CreditSpendingState: CreditSpendingState{state["singleLimit"], state["dailyLimit"], state["monthlyLimit"], state["lastDay"], state["dailySpent"], state["lastMonth"], state["monthlySpent"]}}
	if e = ValidateCreditSpendingState(n.CreditSpendingState); e != nil {
		return e
	}
	n.Payload, e = parsePayloadJSON(data)
	return e
}

func parsePayloadJSON(data []byte) ([]byte, error) {
	var input struct {
		Payload []byte `json:"payload"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}
	if err := ValidatePayload(input.Payload); err != nil {
		return nil, err
	}
	return CloneMemo(input.Payload), nil
}
