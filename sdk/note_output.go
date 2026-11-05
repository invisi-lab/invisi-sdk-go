package sdk

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

// NoteRecipient selects a recoverable address or an opaque per-note identity.
type NoteRecipient interface{ noteRecipient() }
type EVMRecipient struct {
	Address       common.Address
	Nonce         *big.Int
	Envelope      *bool
	AddressLookup string
}
type InvisiRecipient struct {
	Address string
	Nonce   *big.Int
}
type IdentityRecipient struct{ Identity common.Hash }

func (EVMRecipient) noteRecipient()      {}
func (InvisiRecipient) noteRecipient()   {}
func (IdentityRecipient) noteRecipient() {}

type NoteOutputOptions struct {
	ChainID                       *big.Int
	ProtocolAddress, TokenAddress common.Address
	Amount                        *big.Int
	Recipient                     NoteRecipient
	Payload                       []byte
}

// NoteOutput is immutable. Its zero value is invalid; construct it with NewNoteOutput.
type NoteOutput struct{ saved *outputIntent }
type outputIntent struct {
	Context              core.ProtocolContext
	Token                common.Address
	Amount               *big.Int
	Address, Lookup      string
	Note                 *core.Note
	Identity, Commitment common.Hash
	Delivery             protocol.Delivery
	Pending              bool
}

var outputSealKey = func() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return key
}()

func (n NoteOutput) MarshalJSON() ([]byte, error) {
	if n.saved == nil {
		return nil, errors.New("outputs must be created with NewNoteOutput")
	}
	raw, e := json.Marshal(n.saved)
	if e != nil {
		return nil, e
	}
	mac := hmac.New(sha256.New, outputSealKey)
	mac.Write(raw)
	return json.Marshal(struct {
		Material json.RawMessage
		Seal     []byte
	}{raw, mac.Sum(nil)})
}
func (n *NoteOutput) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Material json.RawMessage
		Seal     []byte
	}
	if e := json.Unmarshal(raw, &wire); e != nil {
		return e
	}
	mac := hmac.New(sha256.New, outputSealKey)
	mac.Write(wire.Material)
	if !hmac.Equal(wire.Seal, mac.Sum(nil)) {
		return errors.New("outputs must be created with NewNoteOutput; invalid output snapshot")
	}
	var saved outputIntent
	if e := json.Unmarshal(wire.Material, &saved); e != nil {
		return e
	}
	n.saved = &saved
	return nil
}
func (n NoteOutput) Amount() *big.Int {
	if n.saved == nil {
		return nil
	}
	return new(big.Int).Set(n.saved.Amount)
}
func (n NoteOutput) To() string {
	if n.saved == nil {
		return ""
	}
	return n.saved.Address
}
func (n NoteOutput) Identity() common.Hash {
	if n.saved == nil {
		return common.Hash{}
	}
	return n.saved.Identity
}
func (n NoteOutput) Commitment() common.Hash {
	if n.saved == nil {
		return common.Hash{}
	}
	return n.saved.Commitment
}
func (n NoteOutput) DeliveryPending() bool { return n.saved != nil && n.saved.Pending }
func (n NoteOutput) Note() *core.Note {
	if n.saved == nil || n.saved.Note == nil {
		return nil
	}
	v := *n.saved.Note
	v.Payload = append([]byte(nil), v.Payload...)
	return &v
}
func (n NoteOutput) Delivery() protocol.Delivery {
	if n.saved == nil {
		return protocol.Delivery{}
	}
	d := n.saved.Delivery
	d.Ciphertext = append([]byte(nil), d.Ciphertext...)
	d.Clue = append([]byte(nil), d.Clue...)
	return d
}
func (n NoteOutput) AssertCommitment(expected common.Hash) error {
	if n.saved == nil || n.saved.Commitment != expected {
		return errors.New("note output commitment mismatch")
	}
	return nil
}
func canonicalField(v *big.Int) bool { return v != nil && v.Sign() >= 0 && v.Cmp(fr.Modulus()) < 0 }
func NewNoteOutput(o NoteOutputOptions) (NoteOutput, error) {
	fail := func(e error) (NoteOutput, error) { return NoteOutput{}, e }
	if o.ChainID == nil || o.ChainID.Sign() <= 0 || o.ChainID.BitLen() > 64 || o.ProtocolAddress == (common.Address{}) {
		return fail(errors.New("invalid note output context"))
	}
	if e := positive(o.Amount, 128); e != nil {
		return fail(e)
	}
	if e := validatePayload(o.Payload); e != nil {
		return fail(e)
	}
	p := &outputIntent{Context: core.ProtocolContext{ChainID: core.BigIntToElement(o.ChainID), ProtocolAddress: core.BigIntToElement(o.ProtocolAddress.Big())}, Token: o.TokenAddress, Amount: new(big.Int).Set(o.Amount)}
	var recipient core.ReceivingAddress
	var nonce *big.Int
	encrypted := false
	switch r := o.Recipient.(type) {
	case IdentityRecipient:
		if !canonicalField(r.Identity.Big()) || r.Identity == (common.Hash{}) {
			return fail(errors.New("identity must be a nonzero canonical BN254 field element"))
		}
		p.Identity = r.Identity
		p.Address = r.Identity.Hex()
	case EVMRecipient:
		if r.Address == (common.Address{}) {
			return fail(errors.New("invalid EVM recipient"))
		}
		p.Address = r.Address.Hex()
		nonce = r.Nonce
		encrypted = r.Envelope == nil || *r.Envelope
		p.Lookup = r.AddressLookup
		if p.Lookup == "" {
			p.Lookup = "protocol-rpc"
		}
		if p.Lookup != "protocol-rpc" && p.Lookup != "chain" {
			return fail(errors.New("invalid recipient addressLookup"))
		}
		recipient.OwnerAddress = core.BigIntToElement(r.Address.Big())
		p.Pending = encrypted
	case InvisiRecipient:
		d, e := core.DecodeAddress(r.Address)
		if e != nil {
			return fail(e)
		}
		if d.Kind != "long" {
			return fail(errors.New("recipient address type mismatch"))
		}
		recipient = d.ReceivingAddress
		p.Address = r.Address
		nonce = r.Nonce
		encrypted = true
	default:
		return fail(errors.New("invalid recipient type"))
	}
	if !encrypted && o.Payload != nil {
		return fail(errors.New("payload requires an encrypted envelope"))
	}
	if _, opaque := o.Recipient.(IdentityRecipient); !opaque {
		var nf fr.Element
		if nonce == nil {
			var e error
			nf, e = core.GenerateNonce()
			if e != nil {
				return fail(e)
			}
		} else {
			if !canonicalField(nonce) {
				return fail(errors.New("nonce must be a canonical BN254 field element"))
			}
			nf = core.BigIntToElement(nonce)
		}
		note := core.Note{ProtocolContext: p.Context, OwnerAddress: recipient.OwnerAddress, TokenAddress: core.BigIntToElement(o.TokenAddress.Big()), Amount: core.BigIntToElement(o.Amount), Nonce: nf, Payload: append([]byte(nil), o.Payload...)}
		p.Note = &note
		id, e := core.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
		if e != nil {
			return fail(e)
		}
		p.Identity = fieldHash(id)
	}
	commitment, e := core.ComputeNoteCommitment(p.Context.ChainID, core.BigIntToElement(p.Identity.Big()), core.BigIntToElement(o.TokenAddress.Big()), core.BigIntToElement(o.Amount))
	if e != nil {
		return fail(e)
	}
	p.Commitment = fieldHash(commitment)
	if encrypted && !p.Pending {
		p.Delivery, e = encryptOutput(recipient, *p.Note, p.Commitment)
		if e != nil {
			return fail(e)
		}
	}
	return NoteOutput{saved: p}, nil
}
func encryptOutput(recipient core.ReceivingAddress, note core.Note, commitment common.Hash) (protocol.Delivery, error) {
	cipher, e := core.EncryptPrivateTransferNote(recipient.ViewingPublicKey, note, core.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: core.BigIntToElement(commitment.Big())})
	if e != nil {
		return protocol.Delivery{}, e
	}
	clue, e := core.CreateClue(recipient.ClueKey)
	return protocol.Delivery{Ciphertext: cipher, Clue: clue}, e
}
func (s *WalletSession) output(ctx context.Context, token common.Address, n NoteOutput) (preparedOutput, error) {
	if n.saved == nil {
		return preparedOutput{}, errors.New("outputs must be created with NewNoteOutput")
	}
	p := n.saved
	if p.Context != s.coreContext() || p.Token != token {
		return preparedOutput{}, errors.New("note output chain, protocol or token mismatch")
	}
	commitment, e := core.ComputeNoteCommitment(p.Context.ChainID, core.BigIntToElement(p.Identity.Big()), core.BigIntToElement(token.Big()), core.BigIntToElement(p.Amount))
	if e != nil || fieldHash(commitment) != p.Commitment {
		return preparedOutput{}, errors.New("note output commitment mismatch")
	}
	d := n.Delivery()
	if p.Pending {
		recipient, e := s.resolveWithLookup(ctx, p.Address, p.Lookup)
		if e != nil {
			return preparedOutput{}, e
		}
		d, e = encryptOutput(recipient, *p.Note, p.Commitment)
		if e != nil {
			return preparedOutput{}, e
		}
	}
	return preparedOutput{Address: p.Address, Note: n.Note(), Amount: n.Amount(), Identity: p.Identity, Commitment: p.Commitment, Delivery: d}, nil
}
func (s *WalletSession) changeOutput(ctx context.Context, token common.Address, address string, amount *big.Int) (preparedOutput, error) {
	var r NoteRecipient = InvisiRecipient{Address: address}
	if common.IsHexAddress(address) {
		r = EVMRecipient{Address: common.HexToAddress(address)}
	}
	n, e := NewNoteOutput(NoteOutputOptions{ChainID: core.ElementToBigInt(s.coreContext().ChainID), ProtocolAddress: s.scope.ProtocolAddress, TokenAddress: token, Amount: amount, Recipient: r})
	if e != nil {
		return preparedOutput{}, e
	}
	return s.output(ctx, token, n)
}

// NoteOutputContext fixes the financial scope of a factory output.
type NoteOutputContext struct {
	ChainID                       *big.Int
	ProtocolAddress, TokenAddress common.Address
}

func (n NoteOutput) Context() NoteOutputContext {
	if n.saved == nil {
		return NoteOutputContext{}
	}
	return NoteOutputContext{ChainID: core.ElementToBigInt(n.saved.Context.ChainID), ProtocolAddress: common.BigToAddress(core.ElementToBigInt(n.saved.Context.ProtocolAddress)), TokenAddress: n.saved.Token}
}
func (n NoteOutput) Nonce() *big.Int {
	if n.saved == nil || n.saved.Note == nil {
		return nil
	}
	return core.ElementToBigInt(n.saved.Note.Nonce)
}
func (n NoteOutput) Payload() []byte {
	if n.saved == nil || n.saved.Note == nil {
		return nil
	}
	return append([]byte(nil), n.saved.Note.Payload...)
}
