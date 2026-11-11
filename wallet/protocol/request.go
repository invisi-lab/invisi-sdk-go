// Package wallet defines the portable, private-key-free wallet transport contract.
package protocol

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

type Method string

const (
	Public        Method = "wallet.public"
	Detect        Method = "wallet.detect"
	DecryptNote   Method = "wallet.decryptNote"
	DecryptCredit Method = "wallet.decryptCredit"
	Authorize     Method = "wallet.authorize"
)

type Request struct {
	Version      int             `json:"version"`
	ID           string          `json:"id"`
	OwnerAddress string          `json:"ownerAddress"`
	Method       Method          `json:"method"`
	Params       json.RawMessage `json:"params"`
}
type Response struct {
	Version      int             `json:"version"`
	ID           string          `json:"id"`
	OwnerAddress string          `json:"ownerAddress"`
	Method       Method          `json:"method"`
	Result       json.RawMessage `json:"result,omitempty"`
	Error        *ResponseError  `json:"error,omitempty"`
}

// ResponseError carries a stable native-wallet failure code without private diagnostics.
type ResponseError struct {
	Code string `json:"code"`
}

func (e *ResponseError) Error() string { return "wallet execution failed: " + e.Code }

const NotRecipient = "WALLET_NOT_RECIPIENT"

type Client interface {
	Execute(context.Context, []Request) ([]Response, error)
}
type ClientFunc func(context.Context, []Request) ([]Response, error)

func (f ClientFunc) Execute(c context.Context, r []Request) ([]Response, error) { return f(c, r) }

type ProtocolContext struct {
	ChainID         string `json:"chainId"`
	ProtocolAddress string `json:"protocolAddress"`
}
type DeliveryContext struct {
	ProtocolContext
	TokenAddress     string `json:"tokenAddress"`
	OutputCommitment string `json:"outputCommitment"`
	Owner            string `json:"owner,omitempty"`
	Clue             string `json:"clue,omitempty"`
	Role             string `json:"role,omitempty"`
}
type PublicResult struct {
	ViewingPublicKey  string `json:"viewingPublicKey"`
	SpendingPublicKey string `json:"spendingPublicKey"`
	ClueKey           string `json:"clueKey"`
}
type DetectParams struct {
	Clues []string `json:"clues"`
}
type DetectResult struct {
	Matches []bool `json:"matches"`
}
type DecryptParams struct {
	Envelope string          `json:"envelope"`
	Context  DeliveryContext `json:"context"`
}
type DecryptResult struct {
	Plaintext string `json:"plaintext"`
}
type AuthorizeParams struct {
	Operation primitives.WalletOperation `json:"operation"`
	Capacity  int                        `json:"capacity"`
	Context   ProtocolContext            `json:"context"`
	Words     []string                   `json:"words"`
}
type AuthorizeResult struct {
	SpendingPublicKey string `json:"spendingPublicKey"`
	Signature         string `json:"signature"`
	Digest            string `json:"digest"`
}

func Encode(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func Decode(raw json.RawMessage, v any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return errors.New("trailing JSON")
	}
	return nil
}
func Hex(b []byte) string { return "0x" + hex.EncodeToString(b) }
func Bytes(s string, n int) ([]byte, error) {
	if !strings.HasPrefix(s, "0x") || len(s)%2 != 0 || s != strings.ToLower(s) {
		return nil, errors.New("expected hexadecimal bytes")
	}
	b, e := hex.DecodeString(s[2:])
	if e != nil {
		return nil, e
	}
	if n >= 0 && len(b) != n {
		return nil, fmt.Errorf("expected %d bytes", n)
	}
	return b, nil
}
func Element(s string) (fr.Element, error) {
	var zero fr.Element
	if s == "" || len(s) > 1 && s[0] == '0' {
		return zero, errors.New("noncanonical decimal field")
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return zero, errors.New("invalid decimal field")
		}
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok || v.Cmp(fr.Modulus()) >= 0 {
		return zero, errors.New("field out of range")
	}
	return primitives.BigIntToElement(v), nil
}
func (c ProtocolContext) Core() (primitives.ProtocolContext, error) {
	chain, e := Element(c.ChainID)
	if e != nil {
		return primitives.ProtocolContext{}, e
	}
	address, e := Element(c.ProtocolAddress)
	if e != nil {
		return primitives.ProtocolContext{}, e
	}
	if primitives.ElementToBigInt(chain).BitLen() > 64 || primitives.ElementToBigInt(address).BitLen() > 160 || address.IsZero() {
		return primitives.ProtocolContext{}, errors.New("invalid protocol context")
	}
	return primitives.ProtocolContext{ChainID: chain, ProtocolAddress: address}, nil
}
func Context(c primitives.ProtocolContext) ProtocolContext {
	return ProtocolContext{c.ChainID.String(), c.ProtocolAddress.String()}
}
func (p AuthorizeParams) Core() (primitives.WalletAuthorizationRequest, error) {
	c, e := p.Context.Core()
	if e != nil {
		return primitives.WalletAuthorizationRequest{}, e
	}
	words := make([]fr.Element, len(p.Words))
	for i, s := range p.Words {
		words[i], e = Element(s)
		if e != nil {
			return primitives.WalletAuthorizationRequest{}, e
		}
	}
	r := primitives.WalletAuthorizationRequest{Context: c, Operation: p.Operation, Capacity: p.Capacity, Words: words}
	_, e = primitives.ComputeWalletAuthorizationDigest(r)
	return r, e
}
func Authorization(r primitives.WalletAuthorizationRequest) AuthorizeParams {
	w := make([]string, len(r.Words))
	for i, v := range r.Words {
		w[i] = v.String()
	}
	return AuthorizeParams{r.Operation, r.Capacity, Context(r.Context), w}
}
func ValidateRequest(r Request) error {
	if r.Version != 1 || len(r.ID) < 1 || len(r.ID) > 128 {
		return errors.New("invalid wallet request version or id")
	}
	for _, r := range r.ID {
		if r < 0x21 || r > 0x7e {
			return errors.New("wallet request id must be printable ASCII")
		}
	}
	if _, e := Bytes(r.OwnerAddress, 20); e != nil || common.HexToAddress(r.OwnerAddress) == (common.Address{}) {
		return errors.New("invalid wallet owner address")
	}
	switch r.Method {
	case Public, Detect, DecryptNote, DecryptCredit, Authorize:
	default:
		return errors.New("unsupported wallet method")
	}
	if len(r.Params) == 0 || !json.Valid(r.Params) || r.Params[0] != '{' {
		return errors.New("wallet params must be an object")
	}
	return validateParams(r)
}
func ValidateResponse(q Request, r Response) error {
	if e := ValidateRequest(q); e != nil {
		return e
	}
	if r.Version != q.Version || r.ID != q.ID || r.Method != q.Method || r.OwnerAddress != q.OwnerAddress {
		return errors.New("wallet response request binding mismatch")
	}
	if r.Error != nil {
		if len(r.Result) != 0 || len(r.Error.Code) == 0 || len(r.Error.Code) > 128 {
			return errors.New("invalid wallet error response")
		}
		if r.Error.Code[0] < 'A' || r.Error.Code[0] > 'Z' {
			return errors.New("invalid wallet error code")
		}
		for _, c := range r.Error.Code {
			if !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') && c != '_' {
				return errors.New("invalid wallet error code")
			}
		}
		return nil
	}
	if len(r.Result) == 0 || !json.Valid(r.Result) || r.Result[0] != '{' {
		return errors.New("wallet result must be an object")
	}
	return nil
}

// Prepare constructs a private-key-free request with an application-owned correlation ID.
func Prepare(id string, owner common.Address, method Method, params any) (Request, error) {
	q := Request{Version: 1, ID: id, OwnerAddress: strings.ToLower(owner.Hex()), Method: method, Params: Encode(params)}
	if err := ValidateRequest(q); err != nil {
		return Request{}, err
	}
	return q, nil
}

// AcceptAuthorization verifies the owner, request transcript, digest and canonical signature.
func AcceptAuthorization(q Request, r Response) (primitives.WalletAuthorizationWitness, error) {
	var witness primitives.WalletAuthorizationWitness
	if q.Method != Authorize {
		return witness, errors.New("wallet authorization response required")
	}
	if err := ValidateResponse(q, r); err != nil {
		return witness, err
	}
	if r.Error != nil {
		return witness, r.Error
	}
	var p AuthorizeParams
	if err := Decode(q.Params, &p); err != nil {
		return witness, err
	}
	request, err := p.Core()
	if err != nil {
		return witness, err
	}
	var value AuthorizeResult
	if err = Decode(r.Result, &value); err != nil {
		return witness, err
	}
	public, err := Bytes(value.SpendingPublicKey, 65)
	if err != nil {
		return witness, err
	}
	signature, err := Bytes(value.Signature, 64)
	if err != nil {
		return witness, err
	}
	returned, err := Bytes(value.Digest, 32)
	if err != nil {
		return witness, err
	}
	digest, err := primitives.ComputeWalletAuthorizationDigest(request)
	if err != nil {
		return witness, err
	}
	if !bytes.Equal(digest[:], returned) {
		return witness, errors.New("wallet digest mismatch")
	}
	owner := primitives.BigIntToElement(common.HexToAddress(q.OwnerAddress).Big())
	valid, err := primitives.VerifyWalletAuthorization(owner, public, signature, request)
	if err != nil {
		return witness, err
	}
	if !valid {
		return witness, errors.New("wallet signature mismatch")
	}
	copy(witness.PublicKeyX[:], public[1:33])
	copy(witness.PublicKeyY[:], public[33:])
	copy(witness.Signature[:], signature)
	return witness, nil
}
