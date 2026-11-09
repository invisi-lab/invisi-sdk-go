// Package wallet defines portable wallet requests and independently validates responses.
package wallet

import (
	"encoding/json"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
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
type Executor = wire.Executor
type ExecutorFunc = wire.ExecutorFunc
type ProtocolContext = wire.ProtocolContext
type DeliveryContext = wire.DeliveryContext
type PublicResult = wire.PublicResult
type DetectParams = wire.DetectParams
type DetectResult = wire.DetectResult
type DecryptParams = wire.DecryptParams
type DecryptResult = wire.DecryptResult
type AuthorizeParams = wire.AuthorizeParams
type AuthorizeResult = wire.AuthorizeResult

func Encode(v any) json.RawMessage                                    { return wire.Encode(v) }
func Decode(raw json.RawMessage, v any) error                         { return wire.Decode(raw, v) }
func Hex(b []byte) string                                             { return wire.Hex(b) }
func Bytes(s string, n int) ([]byte, error)                           { return wire.Bytes(s, n) }
func Element(s string) (fr.Element, error)                            { return wire.Element(s) }
func Context(c core.ProtocolContext) ProtocolContext                  { return wire.Context(c) }
func Authorization(r core.WalletAuthorizationRequest) AuthorizeParams { return wire.Authorization(r) }
func ValidateRequest(r Request) error                                 { return wire.ValidateRequest(r) }
func ValidateResponse(q Request, r Response) error                    { return wire.ValidateResponse(q, r) }
func Prepare(id string, owner common.Address, method Method, params any) (Request, error) {
	return wire.Prepare(id, owner, method, params)
}
