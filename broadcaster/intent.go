// Package broadcaster authenticates registered Invisi transaction relays.
package broadcaster

import (
	"bytes"
	"errors"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

type Trust struct {
	ChainID  *big.Int
	Protocol common.Address
	Signer   common.Address
}
type Quote struct {
	Token        common.Address `json:"token"`
	FeeRecipient common.Address `json:"feeRecipient"`
	FeeAmount    string         `json:"feeAmount"`
	Deadline     string         `json:"deadline"`
	MaxGas       string         `json:"maxGas"`
	RequestHash  common.Hash    `json:"requestHash"`
}
type SignedQuote struct {
	Quote              Quote  `json:"quote"`
	Signature          string `json:"signature"`
	Context            *Trust `json:"-"`
	ProvenanceRequired bool   `json:"provenanceRequired,omitempty"`
	RequestHashFormat  string `json:"requestHashFormat,omitempty"`
}
type Intent struct {
	Method              string
	Token, FeeRecipient common.Address
	FeeAmount, Deadline *big.Int
}

func field(v reflect.Value, name string) reflect.Value {
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v.FieldByName(name)
}
func DecodeIntent(data []byte) (Intent, error) { _, _, i, e := decode(data, false); return i, e }
func decode(data []byte, strip bool) (*abi.Method, []any, Intent, error) {
	var out Intent
	a, e := contracts.InvisiProtocolMetaData.GetAbi()
	if e != nil {
		return nil, nil, out, e
	}
	if len(data) < 4 {
		return nil, nil, out, errors.New("missing calldata")
	}
	m, e := a.MethodById(data[:4])
	if e != nil {
		return nil, nil, out, e
	}
	switch m.RawName {
	case "shield", "transact", "createCredit", "creditDraw", "updateCredit", "revokeCredit":
	default:
		return nil, nil, out, errors.New("unsupported broadcaster method")
	}
	if m.RawName == "transact" && (len(m.Inputs) < 2 || m.Inputs[1].Type.T != abi.TupleTy) {
		return nil, nil, out, errors.New("unsupported aggregate")
	}
	args, e := m.Inputs.Unpack(data[4:])
	if e != nil {
		return nil, nil, out, e
	}
	encoded, e := m.Inputs.Pack(args...)
	if e != nil || !bytes.Equal(encoded, data[4:]) {
		return nil, nil, out, errors.New("noncanonical broadcaster calldata")
	}
	var p reflect.Value
	if m.RawName == "revokeCredit" {
		out = Intent{m.RawName, args[0].(common.Address), args[5].(common.Address), args[6].(*big.Int), args[4].(*big.Int)}
	} else {
		if m.RawName == "shield" {
			p = reflect.ValueOf(args[0])
		} else if m.RawName == "transact" {
			p = field(reflect.ValueOf(args[1]), "Params")
		} else {
			p = reflect.ValueOf(args[1])
		}
		token, recipient, amount, deadline := field(p, "Token"), field(p, "FeeRecipient"), field(p, "FeeAmount"), field(p, "Deadline")
		if !token.IsValid() || !recipient.IsValid() || !amount.IsValid() || !deadline.IsValid() {
			return nil, nil, out, errors.New("unsupported fee parameters")
		}
		out = Intent{m.RawName, token.Interface().(common.Address), recipient.Interface().(common.Address), amount.Interface().(*big.Int), deadline.Interface().(*big.Int)}
	}
	if strip {
		if m.RawName != "shield" && m.RawName != "revokeCredit" {
			args[0] = []byte{}
		}
		if m.RawName == "revokeCredit" {
			args[7] = []byte{}
		}
		if m.RawName == "transact" {
			v := reflect.ValueOf(args[1])
			copy := reflect.New(v.Type()).Elem()
			copy.Set(v)
			f := field(copy, "Funding")
			sig := field(f, "Signature")
			sig.SetBytes(nil)
			args[1] = copy.Interface()
		} else if m.RawName == "shield" {
			v := reflect.ValueOf(args[3])
			copy := reflect.New(v.Type()).Elem()
			copy.Set(v)
			field(copy, "Signature").SetBytes(nil)
			args[3] = copy.Interface()
		}
	}
	return m, args, out, nil
}
func RequestHash(data []byte) (common.Hash, error) {
	m, args, _, e := decode(data, true)
	if e != nil {
		return common.Hash{}, e
	}
	b, e := m.Inputs.Pack(args...)
	if e != nil {
		return common.Hash{}, e
	}
	s, _ := abi.NewType("string", "", nil)
	bt, _ := abi.NewType("bytes", "", nil)
	b, e = (abi.Arguments{{Type: s}, {Type: s}, {Type: bt}}).Pack("InvisiBroadcasterIntentV1", m.RawName, append(append([]byte{}, m.ID...), b...))
	return crypto.Keccak256Hash(b), e
}
func decimal(s string) (*big.Int, error) {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.Sign() < 0 || s == "" {
		return nil, errors.New("invalid quote integer")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return nil, errors.New("invalid quote integer")
		}
	}
	return n, nil
}
func QuoteTypedData(q Quote, t Trust) (apitypes.TypedData, error) {
	if t.ChainID == nil || t.ChainID.Sign() <= 0 || t.Protocol == (common.Address{}) || t.Signer == (common.Address{}) {
		return apitypes.TypedData{}, errors.New("invalid broadcaster trust")
	}
	return apitypes.TypedData{Types: apitypes.Types{"EIP712Domain": {{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}}, "Quote": {{Name: "token", Type: "address"}, {Name: "feeRecipient", Type: "address"}, {Name: "feeAmount", Type: "uint256"}, {Name: "deadline", Type: "uint256"}, {Name: "maxGas", Type: "uint256"}, {Name: "requestHash", Type: "bytes32"}}}, PrimaryType: "Quote", Domain: apitypes.TypedDataDomain{Name: "InvisiBroadcaster", Version: "1", ChainId: (*math.HexOrDecimal256)(new(big.Int).Set(t.ChainID)), VerifyingContract: t.Protocol.Hex()}, Message: apitypes.TypedDataMessage{"token": q.Token.Hex(), "feeRecipient": q.FeeRecipient.Hex(), "feeAmount": q.FeeAmount, "deadline": q.Deadline, "maxGas": q.MaxGas, "requestHash": q.RequestHash.Hex()}}, nil
}
func VerifyQuote(s SignedQuote, t Trust, data []byte, now uint64) error {
	if (s.ProvenanceRequired && s.RequestHashFormat != ProvenanceRequestHashFormat) || (!s.ProvenanceRequired && s.RequestHashFormat != "") {
		return errors.New("broadcaster provenance requirement mismatch")
	}
	h, e := RequestHash(data)
	if e != nil {
		return e
	}
	if s.ProvenanceRequired {
		bound, err := provenance.ProvenanceIntentHash(h.Hex())
		if err != nil {
			return err
		}
		h = common.HexToHash(bound)
	}
	return verifyQuoteWithHash(s, t, data, now, h)
}
func verifyQuoteWithHash(s SignedQuote, t Trust, data []byte, now uint64, h common.Hash) error {
	i, e := DecodeIntent(data)
	if e != nil {
		return e
	}
	amount, e := decimal(s.Quote.FeeAmount)
	if e != nil {
		return e
	}
	deadline, e := decimal(s.Quote.Deadline)
	if e != nil {
		return e
	}
	gas, e := decimal(s.Quote.MaxGas)
	if e != nil {
		return e
	}
	if s.Quote.RequestHash != h || s.Quote.Token != i.Token || s.Quote.FeeRecipient != i.FeeRecipient || amount.Cmp(i.FeeAmount) != 0 || amount.Sign() <= 0 || amount.BitLen() > 128 || gas.Sign() <= 0 || gas.BitLen() > 256 || deadline.BitLen() > 256 || deadline.Cmp(i.Deadline) > 0 || deadline.Cmp(new(big.Int).SetUint64(now)) < 0 {
		return errors.New("quote does not match final intent or bounds")
	}
	td, e := QuoteTypedData(s.Quote, t)
	if e != nil {
		return e
	}
	digest, _, e := apitypes.TypedDataAndHash(td)
	if e != nil {
		return e
	}
	sig := common.FromHex(s.Signature)
	if len(sig) != 65 || len(s.Signature) != 132 || !strings.HasPrefix(s.Signature, "0x") {
		return errors.New("invalid quote signature")
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	pub, e := crypto.SigToPub(digest, sig)
	if e != nil {
		return e
	}
	if crypto.PubkeyToAddress(*pub) != t.Signer {
		return errors.New("untrusted quote signature")
	}
	return nil
}
