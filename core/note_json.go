package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Explicit canonical field serialization avoids exposing gnark's internal Montgomery limbs.
func fieldsJSON(values map[string]fr.Element, memo []byte) ([]byte, error) {
	if err := ValidateMemo(memo); err != nil {
		return nil, err
	}
	out := map[string]string{}
	if len(memo) > 0 {
		out["memo"] = base64.StdEncoding.EncodeToString(memo)
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
	return fieldsJSON(map[string]fr.Element{"chainId": n.ChainID, "protocolAddress": n.ProtocolAddress, "walletIdentity": n.WalletIdentity, "tokenAddress": n.TokenAddress, "amount": n.Amount, "nonce": n.Nonce}, n.Memo)
}
func (n *Note) UnmarshalJSON(data []byte) error {
	v, e := parseFields(data, "chainId", "protocolAddress", "walletIdentity", "tokenAddress", "amount", "nonce")
	if e != nil {
		return e
	}
	*n = Note{ProtocolContext{v["chainId"], v["protocolAddress"]}, v["walletIdentity"], v["tokenAddress"], v["amount"], v["nonce"], nil}
	n.Memo, e = parseMemoJSON(data)
	return e
}
func (n CreditNote) MarshalJSON() ([]byte, error) {
	return fieldsJSON(map[string]fr.Element{"chainId": n.ChainID, "protocolAddress": n.ProtocolAddress, "spenderWalletIdentity": n.SpenderWalletIdentity, "ownerWalletIdentity": n.OwnerWalletIdentity, "spenderPayout": n.SpenderPayout, "ownerPayout": n.OwnerPayout, "tokenAddress": n.TokenAddress, "amount": n.Amount, "nonce": n.Nonce, "expiry": n.Expiry, "creditSharedKey": n.CreditSharedKey}, n.Memo)
}
func (n *CreditNote) UnmarshalJSON(data []byte) error {
	v, e := parseFields(data, "chainId", "protocolAddress", "spenderWalletIdentity", "ownerWalletIdentity", "spenderPayout", "ownerPayout", "tokenAddress", "amount", "nonce", "expiry", "creditSharedKey")
	if e != nil {
		return e
	}
	*n = CreditNote{ProtocolContext{v["chainId"], v["protocolAddress"]}, v["spenderWalletIdentity"], v["ownerWalletIdentity"], v["spenderPayout"], v["ownerPayout"], v["tokenAddress"], v["amount"], v["nonce"], v["expiry"], v["creditSharedKey"], nil}
	n.Memo, e = parseMemoJSON(data)
	return e
}

func parseMemoJSON(data []byte) ([]byte, error) {
	var input struct {
		Memo []byte `json:"memo"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}
	if err := ValidateMemo(input.Memo); err != nil {
		return nil, err
	}
	return CloneMemo(input.Memo), nil
}
