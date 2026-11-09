// Package native runs wallet cryptography locally using host-supplied private material.
// Private keys never appear in wallet requests or responses.
package native

import (
	"context"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

type Executor struct{ material *core.InvisiWallet }

func NewExecutor(privateKey string) (*Executor, error) {
	raw, err := core.DecodePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	defer clear(raw)
	material, err := core.DeriveWallet(raw)
	if err != nil {
		return nil, err
	}
	return &Executor{material}, nil
}
func (e *Executor) OwnerAddress() common.Address {
	return common.BigToAddress(core.ElementToBigInt(e.material.OwnerAddress))
}
func (e *Executor) String() string               { return "NativeWalletExecutor([redacted])" }
func (e *Executor) GoString() string             { return e.String() }
func (e *Executor) MarshalJSON() ([]byte, error) { return []byte(`{"nativeWallet":"[redacted]"}`), nil }
func (e *Executor) Execute(ctx context.Context, r wallet.Request) (wallet.Response, error) {
	var result any
	if err := ctx.Err(); err != nil {
		return wallet.Response{}, err
	}
	if err := wallet.ValidateRequest(r); err != nil {
		return wallet.Response{}, err
	}
	if !strings.EqualFold(r.OwnerAddress, e.OwnerAddress().Hex()) {
		return wallet.Response{}, errors.New("wallet owner mismatch")
	}
	switch r.Method {
	case wallet.Public:
		var params struct{}
		if err := wallet.Decode(r.Params, &params); err != nil {
			return wallet.Response{}, err
		}
		result = wallet.PublicResult{ViewingPublicKey: wallet.Hex(e.material.ViewingPublicKey), SpendingPublicKey: wallet.Hex(e.material.SpendingPublicKey), ClueKey: wallet.Hex(e.material.ClueKey)}
	case wallet.Detect:
		var p wallet.DetectParams
		if err := wallet.Decode(r.Params, &p); err != nil {
			return wallet.Response{}, err
		}
		if p.Clues == nil {
			return wallet.Response{}, errors.New("wallet detection clues required")
		}
		matches := make([]bool, len(p.Clues))
		for i, s := range p.Clues {
			clue, err := wallet.Bytes(s, 68)
			if err != nil {
				return wallet.Response{}, err
			}
			matches[i], err = core.DetectClue(e.material.DetectionKey, clue, 24)
			if err != nil {
				matches[i] = false
			}
		}
		result = wallet.DetectResult{Matches: matches}
	case wallet.Authorize:
		var p wallet.AuthorizeParams
		if err := wallet.Decode(r.Params, &p); err != nil {
			return wallet.Response{}, err
		}
		request, err := p.Core()
		if err != nil {
			return wallet.Response{}, err
		}
		signature, err := core.SignWalletAuthorization(e.material.SpendingPrivateKey, request)
		if err != nil {
			return wallet.Response{}, err
		}
		digest, err := core.ComputeWalletAuthorizationDigest(request)
		if err != nil {
			return wallet.Response{}, err
		}
		result = wallet.AuthorizeResult{SpendingPublicKey: wallet.Hex(e.material.SpendingPublicKey), Signature: wallet.Hex(signature), Digest: digest.Hex()}
	case wallet.DecryptNote, wallet.DecryptCredit:
		var p wallet.DecryptParams
		if err := wallet.Decode(r.Params, &p); err != nil {
			return wallet.Response{}, err
		}
		c, err := p.Context.ProtocolContext.Core()
		if err != nil {
			return wallet.Response{}, err
		}
		token, err := wallet.Element(p.Context.TokenAddress)
		if err != nil {
			return wallet.Response{}, err
		}
		commitment, err := wallet.Bytes(p.Context.OutputCommitment, 32)
		if err != nil {
			return wallet.Response{}, err
		}
		field, err := wallet.Element(common.BytesToHash(commitment).Big().String())
		if err != nil {
			return wallet.Response{}, err
		}
		envelope, err := wallet.Bytes(p.Envelope, -1)
		if err != nil {
			return wallet.Response{}, err
		}
		var plain []byte
		if r.Method == wallet.DecryptNote {
			if p.Context.Role != "" {
				return wallet.Response{}, errors.New("note context cannot contain role")
			}
			note, err := core.DecryptPrivateTransferNote(e.material.ViewingPrivateKey, envelope, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: c, OutputCommitment: field}, TokenAddress: token, OwnerAddress: e.material.OwnerAddress})
			if err != nil {
				return wallet.Response{}, err
			}
			plain = core.EncodeNotePlaintext(note)
		} else {
			role := credit.Owner
			if p.Context.Role == "spender" {
				role = credit.Spender
			} else if p.Context.Role != "owner" {
				return wallet.Response{}, errors.New("invalid credit role")
			}
			state, err := credit.Decrypt(e.material.ViewingPrivateKey, envelope, credit.DeliveryContext{ProtocolContext: c, TokenAddress: token, Commitment: field, RecipientIdentity: e.material.OwnerAddress}, role)
			if err != nil {
				return wallet.Response{}, err
			}
			plain = credit.EncodeDeliveryPlaintext(state)
		}
		result = wallet.DecryptResult{Plaintext: wallet.Hex(plain)}
	}
	return wallet.Response{Version: 1, ID: r.ID, OwnerAddress: r.OwnerAddress, Method: r.Method, Result: wallet.Encode(result)}, nil
}
