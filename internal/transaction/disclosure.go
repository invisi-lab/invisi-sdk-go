package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

type DisclosureContext struct {
	ChainID          string      `json:"chainId"`
	ProtocolAddress  string      `json:"protocolAddress"`
	TokenAddress     string      `json:"tokenAddress"`
	OutputCommitment common.Hash `json:"outputCommitment"`
}

type DisclosureExport struct {
	Version int               `json:"version"`
	Kind    string            `json:"kind"`
	Method  string            `json:"method"`
	Context DisclosureContext `json:"context"`
	Data    map[string]string `json:"data"`
}

type Disclosure struct{ material DisclosureExport }

type DisclosureOutput struct {
	ProtocolContext  primitives.ProtocolContext
	TokenAddress     fr.Element
	OutputCommitment common.Hash
	Role             string
	Ciphertext       []byte
}

type DisclosureResult struct {
	Kind string
	Note *primitives.Note
}

type DisclosureScope struct {
	PermitsOwnerSweep, RevealsNoteConsumption bool
	RevealsPayload                            bool
	Version                                   int
	Kind, Method                              string
}

func (d *Disclosure) String() string { return "Disclosure([redacted])" }

func (d *Disclosure) GoString() string { return d.String() }

func (d *Disclosure) MarshalJSON() ([]byte, error) { return json.Marshal(d.GetScope()) }

func (d *Disclosure) GetScope() DisclosureScope {
	return DisclosureScope{
		Version: 1, Kind: d.material.Kind, Method: d.material.Method,
		PermitsOwnerSweep: d.material.Kind == "note", RevealsNoteConsumption: d.material.Kind == "note",
		RevealsPayload: d.material.Method == "payload-key",
	}
}

func (d *Disclosure) Export() DisclosureExport {
	m := d.material
	m.Data = map[string]string{}
	for k, v := range d.material.Data {
		m.Data[k] = v
	}
	return m
}

func DisclosedContext(context primitives.ProtocolContext, token fr.Element, leaf common.Hash) DisclosureContext {
	return DisclosureContext{context.ChainID.String(), context.ProtocolAddress.String(), token.String(), leaf}
}

func OwnerText(identity fr.Element) string {
	return strings.ToLower(common.BigToAddress(primitives.ElementToBigInt(identity)).Hex())
}

func DisclosureFromNote(note primitives.Note) (*Disclosure, error) {
	identity, e := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
	if e != nil {
		return nil, e
	}
	leaf, e := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
	if e != nil {
		return nil, e
	}
	m := DisclosureExport{1, "note", "opening", DisclosedContext(note.ProtocolContext, note.TokenAddress, FieldHash(leaf)), map[string]string{"ownerAddress": OwnerText(note.OwnerAddress), "amount": note.Amount.String(), "nonce": note.Nonce.String()}}
	return &Disclosure{m}, nil
}

func DecimalField(text string, bits int) (fr.Element, error) {
	if len(text) > 78 || !regexp.MustCompile(`^(0|[1-9][0-9]*)$`).MatchString(text) {
		return fr.Element{}, errors.New("invalid disclosure decimal")
	}
	n, ok := new(big.Int).SetString(text, 10)
	if !ok || n.Sign() < 0 || n.BitLen() > bits || n.Cmp(primitives.FieldModulus()) >= 0 {
		return fr.Element{}, errors.New("disclosure value out of range")
	}
	return primitives.BigIntToElement(n), nil
}

func DisclosedOwner(text string) (fr.Element, error) {
	if len(text) != 42 || !strings.HasPrefix(text, "0x") || !common.IsHexAddress(text) || common.HexToAddress(text) == (common.Address{}) {
		return fr.Element{}, errors.New("invalid disclosed wallet identity")
	}
	return primitives.BigIntToElement(common.HexToAddress(text).Big()), nil
}

func (d *Disclosure) output() (DisclosureOutput, error) {
	c := d.material.Context
	chain, e := DecimalField(c.ChainID, 64)
	if e != nil {
		return DisclosureOutput{}, e
	}
	protocol, e := DecimalField(c.ProtocolAddress, 160)
	if e != nil {
		return DisclosureOutput{}, e
	}
	token, e := DecimalField(c.TokenAddress, 160)
	if e != nil {
		return DisclosureOutput{}, e
	}
	if _, e = primitives.HexToElement(c.OutputCommitment.Hex()); e != nil {
		return DisclosureOutput{}, e
	}
	role := "funds"

	return DisclosureOutput{ProtocolContext: primitives.ProtocolContext{ChainID: chain, ProtocolAddress: protocol}, TokenAddress: token, OutputCommitment: c.OutputCommitment, Role: role}, nil
}

func DisclosureFromExport(data []byte) (*Disclosure, error) {
	var material DisclosureExport
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&material); e != nil {
		return nil, e
	}
	var trailing any
	if e := decoder.Decode(&trailing); e != io.EOF {
		return nil, errors.New("trailing disclosure data")
	}
	if material.Version != 1 || material.Kind != "note" || (material.Method != "opening" && material.Method != "payload-key") {
		return nil, errors.New("unsupported disclosure format")
	}
	expected := []string{"ownerAddress", "amount", "nonce"}

	if material.Method == "payload-key" {
		expected = []string{"payloadKey", "ownerAddress", "role"}
	}
	if len(material.Data) != len(expected) {
		return nil, errors.New("unexpected disclosure fields")
	}
	for _, name := range expected {
		if _, ok := material.Data[name]; !ok {
			return nil, fmt.Errorf("missing disclosure field %s", name)
		}
	}
	disclosure := &Disclosure{material}
	output, e := disclosure.output()
	if e != nil {
		return nil, e
	}
	if material.Method == "opening" {
		if _, e = disclosure.Verify(output); e != nil {
			return nil, e
		}
	} else {
		if len(material.Data["payloadKey"]) != 66 || !common.IsHexAddress(material.Data["ownerAddress"]) {
			return nil, errors.New("invalid payload disclosure")
		}
		if _, e = primitives.DecodePrivateKey(material.Data["payloadKey"]); e != nil {
			return nil, e
		}
		role := material.Data["role"]
		if role != "funds" {
			return nil, errors.New("invalid disclosure role")
		}
	}
	return disclosure, nil
}

func (d *Disclosure) Verify(output DisclosureOutput) (DisclosureResult, error) {
	expected, e := d.output()
	if e != nil {
		return DisclosureResult{}, e
	}
	if output.ProtocolContext != expected.ProtocolContext || output.TokenAddress != expected.TokenAddress || output.OutputCommitment != expected.OutputCommitment {
		return DisclosureResult{}, errors.New("disclosure deployment or output commitment mismatch")
	}
	data := d.material.Data
	if output.Role != "funds" {
		return DisclosureResult{}, errors.New("disclosure role mismatch")
	}
	if d.material.Method == "payload-key" {
		if data["role"] != output.Role {
			return DisclosureResult{}, errors.New("selected envelope role differs")
		}
		key, e := primitives.DecodePrivateKey(data["payloadKey"])
		if e != nil {
			return DisclosureResult{}, e
		}
		defer clear(key)
		identity, e := DisclosedOwner(data["ownerAddress"])
		if e != nil {
			return DisclosureResult{}, e
		}
		context := primitives.NoteDecryptionContext{NoteEncryptionContext: primitives.NoteEncryptionContext{ProtocolContext: output.ProtocolContext, OutputCommitment: primitives.BigIntToElement(output.OutputCommitment.Big())}, TokenAddress: output.TokenAddress, OwnerAddress: identity}
		if d.material.Kind == "note" {
			note, e := primitives.DecryptNotePayload(key, output.Ciphertext, context)
			return DisclosureResult{Kind: "note", Note: &note}, e
		}
		return DisclosureResult{}, errors.New("unsupported Note disclosure")
	}
	if d.material.Kind == "note" {
		owner, e := DisclosedOwner(data["ownerAddress"])
		if e != nil {
			return DisclosureResult{}, e
		}
		amount, e := DecimalField(data["amount"], 128)
		if e != nil {
			return DisclosureResult{}, e
		}
		nonce, e := DecimalField(data["nonce"], 254)
		if e != nil {
			return DisclosureResult{}, e
		}
		note := primitives.Note{ProtocolContext: output.ProtocolContext, OwnerAddress: owner, TokenAddress: output.TokenAddress, Amount: amount, Nonce: nonce}
		identity, e := primitives.ComputeNoteIdentity(owner, nonce)
		if e != nil {
			return DisclosureResult{}, e
		}
		leaf, e := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
		if e != nil || FieldHash(leaf) != output.OutputCommitment {
			return DisclosureResult{}, errors.New("disclosed note commitment mismatch")
		}
		return DisclosureResult{Kind: "note", Note: &note}, nil
	}
	return DisclosureResult{}, errors.New("unsupported Note disclosure")
}

func (v *ViewingCapability) DiscloseNote(envelope []byte, context primitives.NoteDecryptionContext) (*Disclosure, error) {
	if _, e := v.decryptNote(envelope, context); e != nil {
		return nil, e
	}
	key, e := primitives.ExportPayloadKey(v.key, envelope)
	if e != nil {
		return nil, e
	}
	defer clear(key)
	return &Disclosure{DisclosureExport{1, "note", "payload-key", DisclosedContext(context.ProtocolContext, context.TokenAddress, FieldHash(context.OutputCommitment)), map[string]string{"payloadKey": "0x" + common.Bytes2Hex(key), "ownerAddress": strings.ToLower(v.identity.Hex()), "role": "funds"}}}, nil
}
