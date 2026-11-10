package wallet

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	wire "github.com/invisi-lab/invisi-sdk-go/wallet/protocol"
)

// Accept independently verifies the response against the original request.
// It returns PublicResult, []bool, primitives.Note, credit.DeliveryState or
// primitives.WalletAuthorizationWitness according to the requested method.
// viewingPublicKey binds Credit receiving snapshots to previously resolved public material.
func Accept(q Request, r Response, viewingPublicKey string) (any, error) {
	if err := ValidateResponse(q, r); err != nil {
		return nil, err
	}
	if r.Error != nil {
		return nil, r.Error
	}
	owner := primitives.BigIntToElement(common.HexToAddress(q.OwnerAddress).Big())
	switch q.Method {
	case Public:
		var params struct{}
		if err := Decode(q.Params, &params); err != nil {
			return nil, err
		}
		var p PublicResult
		if err := Decode(r.Result, &p); err != nil {
			return nil, err
		}
		public, err := Bytes(p.SpendingPublicKey, 65)
		if err != nil {
			return nil, err
		}
		identity, err := primitives.ComputeOwnerAddress(public)
		if err != nil || identity != owner {
			return nil, errors.New("wallet public identity mismatch")
		}
		viewing, err := Bytes(p.ViewingPublicKey, 33)
		if err != nil {
			return nil, err
		}
		if _, err = crypto.DecompressPubkey(viewing); err != nil {
			return nil, err
		}
		clue, err := Bytes(p.ClueKey, 32)
		if err != nil {
			return nil, err
		}
		if err = primitives.ValidateClueKey(clue); err != nil {
			return nil, err
		}
		return p, nil
	case Detect:
		var p DetectParams
		if err := Decode(q.Params, &p); err != nil {
			return nil, err
		}
		if p.Clues == nil {
			return nil, errors.New("wallet detection clues required")
		}
		for _, s := range p.Clues {
			if _, err := Bytes(s, 68); err != nil {
				return nil, err
			}
		}

		var strict struct {
			Matches []json.RawMessage `json:"matches"`
		}
		if err := Decode(r.Result, &strict); err != nil {
			return nil, err
		}
		for _, entry := range strict.Matches {
			if string(entry) != "true" && string(entry) != "false" {
				return nil, errors.New("wallet detection result must contain booleans")
			}
		}
		var v DetectResult
		if err := Decode(r.Result, &v); err != nil {
			return nil, err
		}
		if v.Matches == nil || len(v.Matches) != len(p.Clues) {
			return nil, errors.New("wallet detection length mismatch")
		}
		return v.Matches, nil
	case Authorize:
		return wire.AcceptAuthorization(q, r)
	case DecryptNote, DecryptCredit:
		var p DecryptParams
		if err := Decode(q.Params, &p); err != nil {
			return nil, err
		}
		c, err := p.Context.ProtocolContext.Core()
		if err != nil {
			return nil, err
		}
		token, err := Element(p.Context.TokenAddress)
		if err != nil || primitives.ElementToBigInt(token).BitLen() > 160 {
			return nil, errors.New("invalid token address")
		}
		raw, err := Bytes(p.Context.OutputCommitment, 32)
		if err != nil {
			return nil, err
		}
		commitment, err := Element(common.BytesToHash(raw).Big().String())
		if err != nil {
			return nil, err
		}
		envelope, err := Bytes(p.Envelope, -1)
		if err != nil {
			return nil, err
		}
		minimum, version := 160, byte(1)
		if q.Method == DecryptCredit {
			minimum, version = 485, 0x82
		}
		if len(envelope) < minimum || len(envelope) > 4096 || envelope[0] != version {
			return nil, errors.New("invalid wallet delivery envelope")
		}
		if _, err = crypto.DecompressPubkey(envelope[1:34]); err != nil {
			return nil, errors.New("invalid envelope ephemeral key")
		}
		var v DecryptResult
		if err = Decode(r.Result, &v); err != nil {
			return nil, err
		}
		plain, err := Bytes(v.Plaintext, -1)
		if err != nil {
			return nil, err
		}
		if len(plain) > 4096 {
			return nil, errors.New("wallet plaintext exceeds bound")
		}
		if q.Method == DecryptNote {
			if p.Context.Role != "" {
				return nil, errors.New("note context cannot contain role")
			}
			return primitives.DecodeNotePlaintext(plain, primitives.NoteDecryptionContext{NoteEncryptionContext: primitives.NoteEncryptionContext{ProtocolContext: c, OutputCommitment: commitment}, TokenAddress: token, OwnerAddress: owner})
		}
		role := credit.Owner
		if p.Context.Role == "spender" {
			role = credit.Spender
		} else if p.Context.Role != "owner" {
			return nil, errors.New("invalid credit role")
		}
		viewing, err := Bytes(viewingPublicKey, 33)
		if err != nil {
			return nil, err
		}
		return credit.DecodeDeliveryPlaintext(plain, credit.DeliveryContext{ProtocolContext: c, TokenAddress: token, Commitment: commitment, RecipientIdentity: owner}, role, viewing)
	}
	return nil, fmt.Errorf("unsupported wallet method %s", q.Method)
}
