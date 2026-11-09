package credit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"io"
)

type AuditContext struct {
	ChainID         string         `json:"chainId"`
	ProtocolAddress common.Address `json:"protocolAddress"`
	TokenAddress    common.Address `json:"tokenAddress"`
}
type AuditExport struct {
	Context      AuditContext      `json:"context"`
	Fields       map[string]string `json:"state"`
	GrantViewKey string            `json:"grantViewKey"`
	Commitment   common.Hash       `json:"commitment"`
}
type AuditDisclosure struct{ material AuditExport }

func (d *AuditDisclosure) String() string   { return "GrantAuditDisclosure([redacted])" }
func (d *AuditDisclosure) GoString() string { return d.String() }
func (d *AuditDisclosure) MarshalJSON() ([]byte, error) {
	return []byte(`{"kind":"grant-audit","sensitive":true}`), nil
}

// Disclose exports authority to follow one grant's financial history, without wallet signing keys or recipient Notes.
func Disclose(state State) (*AuditDisclosure, error) {
	commitment, err := state.Commitment()
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	defer clear(raw)
	var encoded struct {
		Fields map[string]string `json:"fields"`
	}
	if err = json.Unmarshal(raw, &encoded); err != nil {
		return nil, err
	}
	delete(encoded.Fields, "chainId")
	delete(encoded.Fields, "protocolAddress")
	delete(encoded.Fields, "tokenAddress")
	delete(encoded.Fields, "creditSharedKey")
	keyHash, err := core.Poseidon2Hash(core.ElementFromUint64(7), state.SharedKey)
	if err != nil {
		return nil, err
	}
	encoded.Fields["keyHash"] = keyHash.String()
	return &AuditDisclosure{AuditExport{AuditContext{state.ChainID.String(), common.BigToAddress(core.ElementToBigInt(state.ProtocolAddress)), common.BigToAddress(core.ElementToBigInt(state.TokenAddress))}, encoded.Fields, state.SharedKey.String(), common.Hash(commitment.Bytes())}}, nil
}
func (d *AuditDisclosure) Export() AuditExport {
	copy := d.material
	copy.Fields = map[string]string{}
	for k, v := range d.material.Fields {
		copy.Fields[k] = v
	}
	return copy
}

// Verify requires an independently accepted on-chain commitment anchor.
func (d *AuditDisclosure) Verify(expected common.Hash) (State, error) {
	if d == nil || expected == (common.Hash{}) || d.material.Commitment != expected {
		return State{}, fmt.Errorf("grant audit commitment anchor mismatch")
	}
	m := d.Export()
	key, err := canonicalField(m.GrantViewKey)
	if err != nil || key.IsZero() {
		return State{}, fmt.Errorf("invalid grant view key")
	}
	keyHash, err := core.Poseidon2Hash(core.ElementFromUint64(7), key)
	if err != nil {
		return State{}, err
	}
	if m.Fields["keyHash"] != keyHash.String() {
		return State{}, fmt.Errorf("grant view key hash mismatch")
	}
	delete(m.Fields, "keyHash")
	m.Fields["chainId"] = m.Context.ChainID
	m.Fields["protocolAddress"] = m.Context.ProtocolAddress.Big().String()
	m.Fields["tokenAddress"] = m.Context.TokenAddress.Big().String()
	m.Fields["creditSharedKey"] = m.GrantViewKey
	raw, err := json.Marshal(struct {
		Version int               `json:"version"`
		Fields  map[string]string `json:"fields"`
	}{2, m.Fields})
	if err != nil {
		return State{}, err
	}
	defer clear(raw)
	var state State
	if err = json.Unmarshal(raw, &state); err != nil {
		return State{}, err
	}
	commitment, err := state.Commitment()
	if err != nil {
		return State{}, err
	}
	if common.Hash(commitment.Bytes()) != expected {
		return State{}, fmt.Errorf("grant audit opening mismatch")
	}
	return state, nil
}
func DisclosureFromExport(raw []byte) (*AuditDisclosure, error) {
	if len(raw) > 16384 {
		return nil, fmt.Errorf("grant audit exceeds resource bound")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var material AuditExport
	if err := decoder.Decode(&material); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("trailing grant audit data")
	}
	result := &AuditDisclosure{material}
	if _, err := result.Verify(material.Commitment); err != nil {
		return nil, err
	}
	return result, nil
}
