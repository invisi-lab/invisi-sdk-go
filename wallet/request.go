// Package wallet defines portable wallet requests and independently validates responses.
package wallet

import (
	"encoding/json"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	wire "github.com/invisi-lab/invisi-sdk-go/wallet/protocol"
)

type Method = wire.Method

const (
	Public        = wire.Public
	Detect        = wire.Detect
	DecryptNote   = wire.DecryptNote
	DecryptCredit = wire.DecryptCredit
	Authorize     = wire.Authorize
)

type Request = wire.Request
type Response = wire.Response
type ResponseError = wire.ResponseError

const NotRecipient = wire.NotRecipient
const MaxBatchItems = wire.MaxBatchItems
const MaxTransportBytes = wire.MaxTransportBytes

type Client = wire.Client
type ClientFunc = wire.ClientFunc
type ProtocolContext = wire.ProtocolContext
type DeliveryContext = wire.DeliveryContext
type PublicResult = wire.PublicResult
type DetectParams = wire.DetectParams
type DetectResult = wire.DetectResult
type DecryptParams = wire.DecryptParams
type DecryptResult = wire.DecryptResult
type AuthorizeParams = wire.AuthorizeParams
type AuthorizeResult = wire.AuthorizeResult

func Encode(v any) json.RawMessage                         { return wire.Encode(v) }
func Decode(raw json.RawMessage, v any) error              { return wire.Decode(raw, v) }
func Hex(b []byte) string                                  { return wire.Hex(b) }
func Bytes(s string, n int) ([]byte, error)                { return wire.Bytes(s, n) }
func Element(s string) (fr.Element, error)                 { return wire.Element(s) }
func Context(c primitives.ProtocolContext) ProtocolContext { return wire.Context(c) }
func Authorization(r primitives.WalletAuthorizationRequest) AuthorizeParams {
	return wire.Authorization(r)
}
func ValidateRequest(r Request) error              { return wire.ValidateRequest(r) }
func ValidateResponse(q Request, r Response) error { return wire.ValidateResponse(q, r) }
func Prepare(id string, owner common.Address, method Method, params any) (Request, error) {
	return wire.Prepare(id, owner, method, params)
}

func ValidateRequests(q []Request) error { return wire.ValidateRequests(q) }
func ValidateResponses(q []Request, r []Response) ([]Response, error) {
	return wire.ValidateResponses(q, r)
}

// AcceptBatch matches by ID and validates all successes before returning any decoded result.
func AcceptBatch(q []Request, r []Response, viewingPublicKey string) ([]any, error) {
	ordered, err := ValidateResponses(q, r)
	if err != nil {
		return nil, err
	}
	out := make([]any, len(q))
	for i := range q {
		out[i], err = Accept(q[i], ordered[i], viewingPublicKey)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
