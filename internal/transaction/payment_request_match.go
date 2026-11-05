package transaction

import (
	"bytes"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

func matchPaymentRequest(p *OutputIntent) error {
	if p == nil || p.IPP == "" {
		return errors.New("missing original IPR")
	}
	o, e := DecodePaymentRequest(p.IPP)
	if e != nil {
		return e
	}
	kind, owner, _, nonce, envelope, e := requestParts(o)
	if e != nil {
		return e
	}
	if o.ChainID.Cmp(core.ElementToBigInt(p.Context.ChainID)) != 0 || o.ProtocolAddress != common.BigToAddress(core.ElementToBigInt(p.Context.ProtocolAddress)) || o.TokenAddress != p.Token || o.Amount.Cmp(p.Amount) != 0 {
		return errors.New("IPR output context or amount mismatch")
	}
	if kind == 3 {
		if p.Note != nil || p.Pending || len(p.Delivery.Ciphertext) != 0 || len(p.Delivery.Clue) != 0 || o.Recipient.(IdentityRecipient).Identity != p.Identity {
			return errors.New("IPR identity output mismatch")
		}
		return nil
	}
	n := p.Note
	if n == nil || n.ProtocolContext != p.Context || common.BigToAddress(core.ElementToBigInt(n.TokenAddress)) != p.Token || core.ElementToBigInt(n.Amount).Cmp(p.Amount) != 0 || common.BigToAddress(core.ElementToBigInt(n.OwnerAddress)) != owner || !bytes.Equal(n.Payload, o.Payload) {
		return errors.New("IPR note opening mismatch")
	}
	if nonce != nil && nonce.Cmp(core.ElementToBigInt(n.Nonce)) != 0 {
		return errors.New("IPR fixed nonce mismatch")
	}
	identity, e := core.ComputeNoteIdentity(n.OwnerAddress, n.Nonce)
	if e != nil {
		return e
	}
	if FieldHash(identity) != p.Identity {
		return errors.New("IPR note identity mismatch")
	}
	if !envelope && (p.Pending || len(p.Delivery.Ciphertext) != 0 || len(p.Delivery.Clue) != 0) {
		return errors.New("IPR envelope mismatch")
	}
	return nil
}
