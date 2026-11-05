package transaction

import (
	"errors"
	"math/big"

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

func (EVMRecipient) noteRecipient() {}

func (InvisiRecipient) noteRecipient() {}

func (IdentityRecipient) noteRecipient() {}

type PaymentRequest struct {
	ChainID                       *big.Int
	ProtocolAddress, TokenAddress common.Address
	Amount                        *big.Int
	Recipient                     NoteRecipient
	Payload                       []byte
}

// NoteOutput is immutable. Its zero value is invalid; construct it with NewNoteOutput.
type NoteOutput struct{ saved *OutputIntent }

type OutputIntent struct {
	IPP                  string
	Context              core.ProtocolContext
	Token                common.Address
	Amount               *big.Int
	Address, Lookup      string
	Note                 *core.Note
	Identity, Commitment common.Hash
	Delivery             protocol.Delivery
	Pending              bool
}

// IPP returns the original canonical request, without execution-selected nonces or lookup choices.
func (n NoteOutput) IPP() string {
	if n.saved == nil {
		return ""
	}
	return n.saved.IPP
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
