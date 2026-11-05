package transaction

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

const maxRequestPayload = 3936

func requestParts(o PaymentRequest) (kind byte, owner common.Address, keys []byte, nonce *big.Int, envelope bool, err error) {
	if o.ChainID == nil || o.ChainID.Sign() <= 0 || o.ChainID.BitLen() > 64 || o.ProtocolAddress == (common.Address{}) {
		err = errors.New("invalid note output context")
		return
	}
	if err = Positive(o.Amount, 128); err != nil {
		return
	}
	if len(o.Payload) > maxRequestPayload {
		err = errors.New("payment request payload too large")
		return
	}
	switch r := o.Recipient.(type) {
	case EVMRecipient:
		kind = 1
		owner = r.Address
		nonce = r.Nonce
		envelope = r.Envelope == nil || *r.Envelope
		if owner == (common.Address{}) {
			err = errors.New("invalid EVM recipient")
			return
		}
		if r.AddressLookup != "" && r.AddressLookup != "chain" && r.AddressLookup != "protocol-rpc" {
			err = errors.New("invalid recipient addressLookup")
			return
		}
	case InvisiRecipient:
		kind = 2
		nonce = r.Nonce
		envelope = true
		var d core.DecodedAddress
		d, err = core.DecodeAddress(r.Address)
		if err != nil {
			return
		}
		if d.Kind != "long" {
			err = errors.New("recipient address type mismatch")
			return
		}
		owner = common.BigToAddress(core.ElementToBigInt(d.OwnerAddress))
		keys = append(append([]byte{}, d.ViewingPublicKey...), d.ClueKey...)
	case IdentityRecipient:
		kind = 3
		if !CanonicalField(r.Identity.Big()) || r.Identity == (common.Hash{}) {
			err = errors.New("identity must be a nonzero canonical BN254 field element")
			return
		}
	default:
		err = errors.New("invalid recipient type")
		return
	}
	if nonce != nil && !CanonicalField(nonce) {
		err = errors.New("nonce must be a canonical BN254 field element")
		return
	}
	if !envelope && o.Payload != nil {
		err = errors.New("payload requires an encrypted envelope")
	}
	return
}

// EncodePaymentRequest performs pure canonical encoding, without selecting a nonce or querying registration.
func EncodePaymentRequest(o PaymentRequest) (string, error) {
	kind, owner, keys, nonce, envelope, e := requestParts(o)
	if e != nil {
		return "", e
	}
	b := bytes.NewBuffer([]byte{1, kind})
	binary.Write(b, binary.BigEndian, o.ChainID.Uint64())
	b.Write(o.ProtocolAddress[:])
	b.Write(o.TokenAddress[:])
	b.Write(o.Amount.FillBytes(make([]byte, 16)))
	if kind == 3 {
		r := o.Recipient.(IdentityRecipient)
		b.Write(r.Identity[:])
	} else {
		b.Write(owner[:])
		if kind == 1 {
			if envelope {
				b.WriteByte(1)
			} else {
				b.WriteByte(0)
			}
		} else {
			b.Write(keys)
		}
		if nonce == nil {
			b.WriteByte(0)
		} else {
			b.WriteByte(1)
			b.Write(nonce.FillBytes(make([]byte, 32)))
		}
		if envelope {
			binary.Write(b, binary.BigEndian, uint16(len(o.Payload)))
			b.Write(o.Payload)
		}
	}
	return "ipr:" + base64.RawURLEncoding.EncodeToString(b.Bytes()), nil
}

// DecodePaymentRequest rejects ambiguous encodings, unknown flags and extra bytes.
func DecodePaymentRequest(encoded string) (PaymentRequest, error) {
	var o PaymentRequest
	if len(encoded) > 5500 || !strings.HasPrefix(encoded, "ipr:") {
		return o, errors.New("invalid payment request encoding")
	}
	raw, e := base64.RawURLEncoding.Strict().DecodeString(strings.TrimPrefix(encoded, "ipr:"))
	if e != nil {
		return o, e
	}
	if len(raw) < 66 || len(raw) > 4122 || raw[0] != 1 || raw[1] < 1 || raw[1] > 3 {
		return o, errors.New("unsupported payment request version, type or length")
	}
	r := bytes.NewReader(raw[2:])
	take := func(n int) ([]byte, error) { b := make([]byte, n); _, e := io.ReadFull(r, b); return b, e }
	read := func(n int) *big.Int {
		b, e := take(n)
		if e != nil {
			return nil
		}
		return new(big.Int).SetBytes(b)
	}
	o.ChainID = read(8)
	b, _ := take(20)
	o.ProtocolAddress = common.BytesToAddress(b)
	b, _ = take(20)
	o.TokenAddress = common.BytesToAddress(b)
	o.Amount = read(16)
	kind := raw[1]
	if kind == 3 {
		b, e = take(32)
		if e != nil {
			return o, e
		}
		o.Recipient = IdentityRecipient{Identity: common.BytesToHash(b)}
	} else {
		b, e = take(20)
		if e != nil {
			return o, e
		}
		owner := common.BytesToAddress(b)
		envelope := true
		var keys []byte
		if kind == 1 {
			flag, e := r.ReadByte()
			if e != nil || flag > 1 {
				return o, errors.New("invalid envelope flag")
			}
			envelope = flag == 1
		} else {
			keys, e = take(65)
			if e != nil {
				return o, e
			}
		}
		flag, e := r.ReadByte()
		if e != nil || flag > 1 {
			return o, errors.New("invalid nonce flag")
		}
		var nonce *big.Int
		if flag == 1 {
			b, e = take(32)
			if e != nil {
				return o, e
			}
			nonce = new(big.Int).SetBytes(b)
		}
		if kind == 1 {
			o.Recipient = EVMRecipient{Address: owner, Nonce: nonce, Envelope: &envelope}
		} else {
			address, e := core.EncodeLongAddress(core.ReceivingAddress{OwnerAddress: core.BigIntToElement(owner.Big()), ViewingPublicKey: keys[:33], ClueKey: keys[33:]})
			if e != nil {
				return o, e
			}
			o.Recipient = InvisiRecipient{Address: address, Nonce: nonce}
		}
		if envelope {
			b, e = take(2)
			if e != nil {
				return o, e
			}
			size := int(binary.BigEndian.Uint16(b))
			if size > maxRequestPayload {
				return o, errors.New("payment request payload too large")
			}
			if size > 0 {
				o.Payload, e = take(size)
				if e != nil {
					return o, e
				}
			}
		}
	}
	if r.Len() != 0 {
		return o, errors.New("trailing payment request bytes")
	}
	canonical, e := EncodePaymentRequest(o)
	if e != nil {
		return o, e
	}
	if canonical != encoded {
		return o, fmt.Errorf("noncanonical payment request encoding")
	}
	return o, nil
}
