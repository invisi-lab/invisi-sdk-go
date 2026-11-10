package transaction

import (
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func CanonicalField(v *big.Int) bool { return v != nil && v.Sign() >= 0 && v.Cmp(fr.Modulus()) < 0 }

func createNoteOutput(o PaymentRequest) (NoteOutput, error) {
	fail := func(e error) (NoteOutput, error) { return NoteOutput{}, e }
	if o.ChainID == nil || o.ChainID.Sign() <= 0 || o.ChainID.BitLen() > 64 || o.ProtocolAddress == (common.Address{}) {
		return fail(errors.New("invalid note output context"))
	}
	if e := Positive(o.Amount, 128); e != nil {
		return fail(e)
	}
	if e := ValidatePayload(o.Payload); e != nil {
		return fail(e)
	}
	p := &OutputIntent{Context: primitives.ProtocolContext{ChainID: primitives.BigIntToElement(o.ChainID), ProtocolAddress: primitives.BigIntToElement(o.ProtocolAddress.Big())}, Token: o.TokenAddress, Amount: new(big.Int).Set(o.Amount)}
	var recipient primitives.ReceivingAddress
	var nonce *big.Int
	encrypted := false
	switch r := o.Recipient.(type) {
	case IdentityRecipient:
		if !CanonicalField(r.Identity.Big()) || r.Identity == (common.Hash{}) {
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
		recipient.OwnerAddress = primitives.BigIntToElement(r.Address.Big())
		p.Pending = encrypted
	case InvisiRecipient:
		d, e := primitives.DecodeAddress(r.Address)
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
			nf, e = primitives.GenerateNonce()
			if e != nil {
				return fail(e)
			}
		} else {
			if !CanonicalField(nonce) {
				return fail(errors.New("nonce must be a canonical BN254 field element"))
			}
			nf = primitives.BigIntToElement(nonce)
		}
		note := primitives.Note{ProtocolContext: p.Context, OwnerAddress: recipient.OwnerAddress, TokenAddress: primitives.BigIntToElement(o.TokenAddress.Big()), Amount: primitives.BigIntToElement(o.Amount), Nonce: nf, Payload: append([]byte(nil), o.Payload...)}
		p.Note = &note
		id, e := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
		if e != nil {
			return fail(e)
		}
		p.Identity = FieldHash(id)
	}
	commitment, e := primitives.ComputeNoteCommitment(p.Context.ChainID, p.Context.ProtocolAddress, primitives.BigIntToElement(p.Identity.Big()), primitives.BigIntToElement(o.TokenAddress.Big()), primitives.BigIntToElement(o.Amount))
	if e != nil {
		return fail(e)
	}
	p.Commitment = FieldHash(commitment)
	if encrypted && !p.Pending {
		p.Delivery, e = EncryptOutput(recipient, *p.Note, p.Commitment)
		if e != nil {
			return fail(e)
		}
	}
	return NoteOutput{saved: p}, nil
}
