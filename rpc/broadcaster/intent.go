// Package broadcaster authenticates registered Invisi transaction relays.
package broadcaster

import (
	"errors"
	"math/big"
	"strings"

	"github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
	"github.com/invisi-lab/invisi-sdk-go/protocol/settlement"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type Trust struct {
	ChainID  *big.Int
	Protocol common.Address
	Signer   common.Address
}
type Schedule = settlement.Schedule

type Quote struct {
	Schedule     *Schedule      `json:"schedule,omitempty"`
	Token        common.Address `json:"token"`
	FeeRecipient common.Address `json:"feeRecipient"`
	FeeAmount    string         `json:"feeAmount"`
	Deadline     string         `json:"deadline"`
	MaxGas       string         `json:"maxGas"`
	RequestHash  common.Hash    `json:"requestHash"`
	Execution    string         `json:"execution"`
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

func DecodeIntent(data []byte) (Intent, error) {
	r, err := settlement.DecodeSingle(data)
	if err != nil {
		return Intent{}, err
	}
	p := r.Statement()
	return Intent{r.Method, p.Token, p.FeeRecipient, p.FeeAmount, p.Deadline}, nil
}
func RequestHash(data []byte) (common.Hash, error) {
	return settlement.RequestHash(data)
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
	if q.Execution != "direct" && q.Execution != "aggregate" {
		return apitypes.TypedData{}, errors.New("invalid broadcaster quote execution")
	}
	if t.ChainID == nil || t.ChainID.Sign() <= 0 || t.Protocol == (common.Address{}) || t.Signer == (common.Address{}) {
		return apitypes.TypedData{}, errors.New("invalid broadcaster trust")
	}
	td := apitypes.TypedData{Types: apitypes.Types{"EIP712Domain": {{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}}, "Quote": {{Name: "token", Type: "address"}, {Name: "feeRecipient", Type: "address"}, {Name: "feeAmount", Type: "uint256"}, {Name: "deadline", Type: "uint256"}, {Name: "maxGas", Type: "uint256"}, {Name: "requestHash", Type: "bytes32"}, {Name: "execution", Type: "string"}}}, PrimaryType: "Quote", Domain: apitypes.TypedDataDomain{Name: "InvisiBroadcaster", Version: "1", ChainId: (*math.HexOrDecimal256)(new(big.Int).Set(t.ChainID)), VerifyingContract: t.Protocol.Hex()}, Message: apitypes.TypedDataMessage{"token": q.Token.Hex(), "feeRecipient": q.FeeRecipient.Hex(), "feeAmount": q.FeeAmount, "deadline": q.Deadline, "maxGas": q.MaxGas, "requestHash": q.RequestHash.Hex(), "execution": q.Execution}}
	if q.Schedule != nil {
		if err := q.Schedule.Validate(); err != nil {
			return apitypes.TypedData{}, err
		}
		td.PrimaryType = "ScheduledQuote"
		td.Types["ScheduledQuote"] = append(td.Types["Quote"], apitypes.Type{Name: "publishAfter", Type: "uint256"}, apitypes.Type{Name: "executeBefore", Type: "uint256"}, apitypes.Type{Name: "cancelTokenHash", Type: "bytes32"})
		delete(td.Types, "Quote")
		td.Message["publishAfter"] = q.Schedule.PublishAfter
		td.Message["executeBefore"] = q.Schedule.ExecuteBefore
		td.Message["cancelTokenHash"] = q.Schedule.CancelTokenHash
	}
	return td, nil
}
func VerifyQuote(s SignedQuote, t Trust, data []byte, now uint64) error {
	if (s.ProvenanceRequired && s.RequestHashFormat != ProvenanceRequestHashFormat) || (!s.ProvenanceRequired && s.RequestHashFormat != "") {
		return errors.New("broadcaster provenance requirement mismatch")
	}
	if s.Quote.Execution == "aggregate" {
		if _, e := aggregateCapacity(data); e != nil {
			return e
		}
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
