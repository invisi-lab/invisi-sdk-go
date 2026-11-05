package transaction

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// PaymentNoteData is a complete funds Note opening, with no payload or wallet keys.
type PaymentNoteData struct {
	ChainID                                     *big.Int
	ProtocolAddress, TokenAddress, OwnerAddress common.Address
	Amount, Nonce                               *big.Int
	Commitment                                  common.Hash
}

func validatePaymentNote(d PaymentNoteData) error {
	if d.ChainID == nil || d.ChainID.Sign() <= 0 || d.ChainID.BitLen() > 64 || d.ProtocolAddress == (common.Address{}) || d.OwnerAddress == (common.Address{}) {
		return errors.New("invalid payment note context or owner")
	}
	if e := Positive(d.Amount, 128); e != nil {
		return e
	}
	if !CanonicalField(d.Nonce) || !CanonicalField(d.Commitment.Big()) || d.Commitment == (common.Hash{}) {
		return errors.New("invalid payment note nonce or commitment")
	}
	return nil
}
func clonePaymentNote(d PaymentNoteData) PaymentNoteData {
	if d.ChainID != nil {
		d.ChainID = new(big.Int).Set(d.ChainID)
	}
	if d.Amount != nil {
		d.Amount = new(big.Int).Set(d.Amount)
	}
	if d.Nonce != nil {
		d.Nonce = new(big.Int).Set(d.Nonce)
	}
	return d
}

// EncodePaymentNote performs canonical structural encoding. PaymentNoteFromIPN also verifies the commitment.
func EncodePaymentNote(d PaymentNoteData) (string, error) {
	if e := validatePaymentNote(d); e != nil {
		return "", e
	}
	b := bytes.NewBuffer([]byte{1})
	binary.Write(b, binary.BigEndian, d.ChainID.Uint64())
	b.Write(d.ProtocolAddress[:])
	b.Write(d.TokenAddress[:])
	b.Write(d.Amount.FillBytes(make([]byte, 16)))
	b.Write(d.OwnerAddress[:])
	b.Write(d.Nonce.FillBytes(make([]byte, 32)))
	b.Write(d.Commitment[:])
	return "ipn:" + base64.RawURLEncoding.EncodeToString(b.Bytes()), nil
}
func DecodePaymentNote(encoded string) (PaymentNoteData, error) {
	var d PaymentNoteData
	if len(encoded) != 203 || !strings.HasPrefix(encoded, "ipn:") {
		return d, errors.New("invalid payment note encoding")
	}
	b, e := base64.RawURLEncoding.Strict().DecodeString(encoded[4:])
	if e != nil {
		return d, e
	}
	if len(b) != 149 || b[0] != 1 {
		return d, errors.New("unsupported payment note version or length")
	}
	d.ChainID = new(big.Int).SetBytes(b[1:9])
	d.ProtocolAddress = common.BytesToAddress(b[9:29])
	d.TokenAddress = common.BytesToAddress(b[29:49])
	d.Amount = new(big.Int).SetBytes(b[49:65])
	d.OwnerAddress = common.BytesToAddress(b[65:85])
	d.Nonce = new(big.Int).SetBytes(b[85:117])
	d.Commitment = common.BytesToHash(b[117:149])
	canonical, e := EncodePaymentNote(d)
	if e != nil {
		return d, e
	}
	if canonical != encoded {
		return d, errors.New("noncanonical payment note encoding")
	}
	return d, nil
}
