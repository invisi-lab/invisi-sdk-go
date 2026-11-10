package transaction

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

// PaymentNote is immutable validated offline delivery data; it does not assert on-chain receipt.
type PaymentNote struct {
	data PaymentNoteData
	ipn  string
}

func PaymentNoteFromIPN(ipn string) (*PaymentNote, error) {
	d, e := DecodePaymentNote(ipn)
	if e != nil {
		return nil, e
	}
	n := paymentNoteOpening(d)
	identity, e := primitives.ComputeNoteIdentity(n.OwnerAddress, n.Nonce)
	if e != nil {
		return nil, e
	}
	commitment, e := primitives.ComputeNoteCommitment(n.ChainID, n.ProtocolAddress, identity, n.TokenAddress, n.Amount)
	if e != nil {
		return nil, e
	}
	if FieldHash(commitment) != d.Commitment {
		return nil, errors.New("payment note commitment mismatch")
	}
	return &PaymentNote{clonePaymentNote(d), ipn}, nil
}
func paymentNoteOpening(d PaymentNoteData) primitives.Note {
	return primitives.Note{ProtocolContext: primitives.ProtocolContext{ChainID: primitives.BigIntToElement(d.ChainID), ProtocolAddress: primitives.BigIntToElement(d.ProtocolAddress.Big())}, OwnerAddress: primitives.BigIntToElement(d.OwnerAddress.Big()), TokenAddress: primitives.BigIntToElement(d.TokenAddress.Big()), Amount: primitives.BigIntToElement(d.Amount), Nonce: primitives.BigIntToElement(d.Nonce)}
}
func (n *PaymentNote) Data() PaymentNoteData           { return clonePaymentNote(n.data) }
func (n *PaymentNote) IPN() string                     { return n.ipn }
func (n *PaymentNote) Amount() *big.Int                { return new(big.Int).Set(n.data.Amount) }
func (n *PaymentNote) Nonce() *big.Int                 { return new(big.Int).Set(n.data.Nonce) }
func (n *PaymentNote) ChainID() *big.Int               { return new(big.Int).Set(n.data.ChainID) }
func (n *PaymentNote) OwnerAddress() common.Address    { return n.data.OwnerAddress }
func (n *PaymentNote) ProtocolAddress() common.Address { return n.data.ProtocolAddress }
func (n *PaymentNote) TokenAddress() common.Address    { return n.data.TokenAddress }
func (n *PaymentNote) Commitment() common.Hash         { return n.data.Commitment }
func (n *PaymentNote) String() string                  { return "PaymentNote([private opening redacted])" }
func (n *PaymentNote) GoString() string                { return n.String() }

func exportPaymentNote(note *primitives.Note, commitment common.Hash) (string, error) {
	if note == nil {
		return "", errors.New("identity outputs cannot export an IPN")
	}
	return EncodePaymentNote(PaymentNoteData{ChainID: primitives.ElementToBigInt(note.ChainID), ProtocolAddress: common.BigToAddress(primitives.ElementToBigInt(note.ProtocolAddress)), TokenAddress: common.BigToAddress(primitives.ElementToBigInt(note.TokenAddress)), OwnerAddress: common.BigToAddress(primitives.ElementToBigInt(note.OwnerAddress)), Amount: primitives.ElementToBigInt(note.Amount), Nonce: primitives.ElementToBigInt(note.Nonce), Commitment: commitment})
}
func (n NoteOutput) ExportPaymentNote() (string, error) {
	if n.saved == nil {
		return "", errors.New("uninitialized NoteOutput")
	}
	return exportPaymentNote(n.saved.Note, n.saved.Commitment)
}
