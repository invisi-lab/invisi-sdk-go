package sdk

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
	"github.com/invisi-lab/invisi-sdk-go/core"
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
	ProtocolContext  core.ProtocolContext
	TokenAddress     fr.Element
	OutputCommitment common.Hash
	Role             string
	Ciphertext       []byte
}
type CreditOpening struct {
	// Memo is present only after authenticating a payload-key disclosure.
	Memo []byte
	core.ProtocolContext
	SpenderWalletIdentity, OwnerWalletIdentity, SpenderPayout, OwnerPayout, TokenAddress, Amount, Nonce, Expiry fr.Element
	CreditSharedKeyHash                                                                                         common.Hash
}
type DisclosureResult struct {
	Kind   string
	Note   *core.Note
	Credit *CreditOpening
}
type DisclosureScope struct {
	RevealsMemo                                                                             bool
	Version                                                                                 int
	Kind, Method                                                                            string
	PermitsOwnerSweep, RevealsNoteConsumption, RevealsCreditSharedKey, RevealsReceivingKeys bool
}

func (d *Disclosure) String() string               { return "Disclosure([redacted])" }
func (d *Disclosure) GoString() string             { return d.String() }
func (d *Disclosure) MarshalJSON() ([]byte, error) { return json.Marshal(d.GetScope()) }
func (d *Disclosure) GetScope() DisclosureScope {
	return DisclosureScope{
		Version: 1, Kind: d.material.Kind, Method: d.material.Method,
		PermitsOwnerSweep: d.material.Kind == "note", RevealsNoteConsumption: d.material.Kind == "note",
		RevealsCreditSharedKey: d.material.Kind == "credit" && d.material.Method == "payload-key",
		RevealsReceivingKeys:   d.material.Kind == "credit" && d.material.Method == "payload-key",
		RevealsMemo:            d.material.Method == "payload-key",
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
func disclosedContext(context core.ProtocolContext, token fr.Element, leaf common.Hash) DisclosureContext {
	return DisclosureContext{context.ChainID.String(), context.ProtocolAddress.String(), token.String(), leaf}
}
func ownerText(identity fr.Element) string {
	return strings.ToLower(common.BigToAddress(core.ElementToBigInt(identity)).Hex())
}
func DisclosureFromNote(note core.Note) (*Disclosure, error) {
	identity, e := core.ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
	if e != nil {
		return nil, e
	}
	leaf, e := core.ComputeNoteCommitment(note.ChainID, identity, note.TokenAddress, note.Amount)
	if e != nil {
		return nil, e
	}
	m := DisclosureExport{1, "note", "opening", disclosedContext(note.ProtocolContext, note.TokenAddress, fieldHash(leaf)), map[string]string{"walletIdentity": ownerText(note.WalletIdentity), "amount": note.Amount.String(), "nonce": note.Nonce.String()}}
	return &Disclosure{m}, nil
}
func DisclosureFromCredit(note core.CreditNote) (*Disclosure, error) {
	leaf, e := creditCommitment(note)
	if e != nil {
		return nil, e
	}
	key, e := core.ComputeCreditSharedKeyHash(note.CreditSharedKey)
	if e != nil {
		return nil, e
	}
	data := map[string]string{"spenderWalletIdentity": ownerText(note.SpenderWalletIdentity), "ownerWalletIdentity": ownerText(note.OwnerWalletIdentity), "spenderPayout": note.SpenderPayout.String(), "ownerPayout": note.OwnerPayout.String(), "amount": note.Amount.String(), "nonce": note.Nonce.String(), "expiry": note.Expiry.String(), "creditSharedKeyHash": fieldHash(key).Hex()}
	return &Disclosure{DisclosureExport{1, "credit", "opening", disclosedContext(note.ProtocolContext, note.TokenAddress, leaf), data}}, nil
}
func decimalField(text string, bits int) (fr.Element, error) {
	if len(text) > 78 || !regexp.MustCompile(`^(0|[1-9][0-9]*)$`).MatchString(text) {
		return fr.Element{}, errors.New("invalid disclosure decimal")
	}
	n, ok := new(big.Int).SetString(text, 10)
	if !ok || n.Sign() < 0 || n.BitLen() > bits || n.Cmp(core.FieldModulus()) >= 0 {
		return fr.Element{}, errors.New("disclosure value out of range")
	}
	return core.BigIntToElement(n), nil
}
func disclosedOwner(text string) (fr.Element, error) {
	if len(text) != 42 || !strings.HasPrefix(text, "0x") || !common.IsHexAddress(text) || common.HexToAddress(text) == (common.Address{}) {
		return fr.Element{}, errors.New("invalid disclosed wallet identity")
	}
	return core.BigIntToElement(common.HexToAddress(text).Big()), nil
}
func (d *Disclosure) output() (DisclosureOutput, error) {
	c := d.material.Context
	chain, e := decimalField(c.ChainID, 64)
	if e != nil {
		return DisclosureOutput{}, e
	}
	protocol, e := decimalField(c.ProtocolAddress, 160)
	if e != nil {
		return DisclosureOutput{}, e
	}
	token, e := decimalField(c.TokenAddress, 160)
	if e != nil {
		return DisclosureOutput{}, e
	}
	if _, e = core.HexToElement(c.OutputCommitment.Hex()); e != nil {
		return DisclosureOutput{}, e
	}
	role := "funds"
	if d.material.Kind == "credit" {
		role = "owner"
	}
	return DisclosureOutput{ProtocolContext: core.ProtocolContext{ChainID: chain, ProtocolAddress: protocol}, TokenAddress: token, OutputCommitment: c.OutputCommitment, Role: role}, nil
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
	if material.Version != 1 || (material.Kind != "note" && material.Kind != "credit") || (material.Method != "opening" && material.Method != "payload-key") {
		return nil, errors.New("unsupported disclosure format")
	}
	expected := []string{"walletIdentity", "amount", "nonce"}
	if material.Kind == "credit" {
		expected = []string{"spenderWalletIdentity", "ownerWalletIdentity", "spenderPayout", "ownerPayout", "amount", "nonce", "expiry", "creditSharedKeyHash"}
	}
	if material.Method == "payload-key" {
		expected = []string{"payloadKey", "walletIdentity", "role"}
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
		if len(material.Data["payloadKey"]) != 66 || !common.IsHexAddress(material.Data["walletIdentity"]) {
			return nil, errors.New("invalid payload disclosure")
		}
		if _, e = core.DecodePrivateKey(material.Data["payloadKey"]); e != nil {
			return nil, e
		}
		role := material.Data["role"]
		if material.Kind == "note" && role != "funds" || material.Kind == "credit" && role != "owner" && role != "spender" {
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
	if d.material.Kind == "note" && output.Role != "funds" || d.material.Kind == "credit" && output.Role != "owner" && output.Role != "spender" {
		return DisclosureResult{}, errors.New("disclosure role mismatch")
	}
	if d.material.Method == "payload-key" {
		if data["role"] != output.Role {
			return DisclosureResult{}, errors.New("selected envelope role differs")
		}
		key, e := core.DecodePrivateKey(data["payloadKey"])
		if e != nil {
			return DisclosureResult{}, e
		}
		defer clear(key)
		identity, e := disclosedOwner(data["walletIdentity"])
		if e != nil {
			return DisclosureResult{}, e
		}
		context := core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: output.ProtocolContext, OutputCommitment: core.BigIntToElement(output.OutputCommitment.Big())}, TokenAddress: output.TokenAddress, WalletIdentity: identity}
		if d.material.Kind == "note" {
			note, e := core.DecryptNotePayload(key, output.Ciphertext, context)
			return DisclosureResult{Kind: "note", Note: &note}, e
		}
		state, e := core.DecryptCreditPayload(key, output.Ciphertext, core.CreditNoteDecryptionContext{NoteDecryptionContext: context, Role: core.CreditDeliveryRole(output.Role)})
		if e != nil {
			return DisclosureResult{}, e
		}
		opening, e := creditOpening(state.Note)
		return DisclosureResult{Kind: "credit", Credit: &opening}, e
	}
	if d.material.Kind == "note" {
		owner, e := disclosedOwner(data["walletIdentity"])
		if e != nil {
			return DisclosureResult{}, e
		}
		amount, e := decimalField(data["amount"], 128)
		if e != nil {
			return DisclosureResult{}, e
		}
		nonce, e := decimalField(data["nonce"], 254)
		if e != nil {
			return DisclosureResult{}, e
		}
		note := core.Note{ProtocolContext: output.ProtocolContext, WalletIdentity: owner, TokenAddress: output.TokenAddress, Amount: amount, Nonce: nonce}
		identity, e := core.ComputeNoteIdentity(owner, nonce)
		if e != nil {
			return DisclosureResult{}, e
		}
		leaf, e := core.ComputeNoteCommitment(note.ChainID, identity, note.TokenAddress, note.Amount)
		if e != nil || fieldHash(leaf) != output.OutputCommitment {
			return DisclosureResult{}, errors.New("disclosed note commitment mismatch")
		}
		return DisclosureResult{Kind: "note", Note: &note}, nil
	}
	var opening CreditOpening
	opening.ProtocolContext = output.ProtocolContext
	opening.TokenAddress = output.TokenAddress
	opening.SpenderWalletIdentity, e = disclosedOwner(data["spenderWalletIdentity"])
	if e != nil {
		return DisclosureResult{}, e
	}
	opening.OwnerWalletIdentity, e = disclosedOwner(data["ownerWalletIdentity"])
	if e != nil {
		return DisclosureResult{}, e
	}
	for _, field := range []struct {
		name   string
		bits   int
		target *fr.Element
	}{{"spenderPayout", 160, &opening.SpenderPayout}, {"ownerPayout", 160, &opening.OwnerPayout}, {"amount", 126, &opening.Amount}, {"nonce", 254, &opening.Nonce}, {"expiry", 64, &opening.Expiry}} {
		*field.target, e = decimalField(data[field.name], field.bits)
		if e != nil {
			return DisclosureResult{}, e
		}
	}
	if !regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`).MatchString(data["creditSharedKeyHash"]) {
		return DisclosureResult{}, errors.New("invalid Credit shared key hash")
	}
	hash, e := core.HexToElement(data["creditSharedKeyHash"])
	if e != nil {
		return DisclosureResult{}, e
	}
	opening.CreditSharedKeyHash = fieldHash(hash)
	leaf, e := core.ComputeCreditCommitment(output.ProtocolContext.ChainID, opening.SpenderWalletIdentity, opening.OwnerWalletIdentity, opening.SpenderPayout, opening.OwnerPayout, opening.TokenAddress, opening.Amount, opening.Nonce, opening.Expiry, hash)
	if e != nil || fieldHash(leaf) != output.OutputCommitment {
		return DisclosureResult{}, errors.New("disclosed Credit commitment mismatch")
	}
	return DisclosureResult{Kind: "credit", Credit: &opening}, nil
}
func creditOpening(note core.CreditNote) (CreditOpening, error) {
	hash, e := core.ComputeCreditSharedKeyHash(note.CreditSharedKey)
	return CreditOpening{ProtocolContext: note.ProtocolContext, SpenderWalletIdentity: note.SpenderWalletIdentity, OwnerWalletIdentity: note.OwnerWalletIdentity, SpenderPayout: note.SpenderPayout, OwnerPayout: note.OwnerPayout, TokenAddress: note.TokenAddress, Amount: note.Amount, Nonce: note.Nonce, Expiry: note.Expiry, CreditSharedKeyHash: fieldHash(hash), Memo: core.CloneMemo(note.Memo)}, e
}
func (v *ViewingCapability) DiscloseNote(envelope []byte, context core.NoteDecryptionContext) (*Disclosure, error) {
	if _, e := v.decryptNote(envelope, context); e != nil {
		return nil, e
	}
	key, e := core.ExportPayloadKey(v.key, envelope, false)
	if e != nil {
		return nil, e
	}
	defer clear(key)
	return &Disclosure{DisclosureExport{1, "note", "payload-key", disclosedContext(context.ProtocolContext, context.TokenAddress, fieldHash(context.OutputCommitment)), map[string]string{"payloadKey": "0x" + common.Bytes2Hex(key), "walletIdentity": strings.ToLower(v.identity.Hex()), "role": "funds"}}}, nil
}
func (v *ViewingCapability) DiscloseCredit(envelope []byte, context core.CreditNoteDecryptionContext) (*Disclosure, error) {
	if _, e := v.decryptCredit(envelope, context); e != nil {
		return nil, e
	}
	key, e := core.ExportPayloadKey(v.key, envelope, true)
	if e != nil {
		return nil, e
	}
	defer clear(key)
	return &Disclosure{DisclosureExport{1, "credit", "payload-key", disclosedContext(context.ProtocolContext, context.TokenAddress, fieldHash(context.OutputCommitment)), map[string]string{"payloadKey": "0x" + common.Bytes2Hex(key), "walletIdentity": strings.ToLower(v.identity.Hex()), "role": string(context.Role)}}}, nil
}
