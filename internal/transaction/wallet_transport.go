package transaction

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func (w *Wallet) request(ctx context.Context, executor wallet.Executor, method wallet.Method, params any, result any) error {
	if w == nil || executor == nil {
		return errors.New("wallet executor required")
	}
	q := wallet.Request{Version: 1, ID: "go-" + rand.Text(), OwnerAddress: strings.ToLower(w.identity.Hex()), Method: method, Params: wallet.Encode(params)}
	// Preserve the original transcript even if an in-process executor mutates its copy.
	sent := q
	sent.Params = append([]byte(nil), q.Params...)
	response, err := executor.Execute(ctx, sent)
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if _, err = w.AcceptResponse(q, response); err != nil {
		return err
	}
	return wallet.Decode(response.Result, result)
}
func ConnectWallet(ctx context.Context, owner common.Address, executor wallet.Executor) (*Wallet, error) {
	w := &Wallet{identity: owner}
	var result wallet.PublicResult
	if err := w.request(ctx, executor, wallet.Public, struct{}{}, &result); err != nil {
		return nil, err
	}
	public, err := wallet.Bytes(result.SpendingPublicKey, 65)
	if err != nil {
		return nil, err
	}
	identity, err := core.ComputeOwnerAddress(public)
	if err != nil || identity != core.BigIntToElement(owner.Big()) {
		return nil, errors.New("wallet public identity mismatch")
	}
	viewing, err := wallet.Bytes(result.ViewingPublicKey, 33)
	if err != nil {
		return nil, err
	}
	if _, err = crypto.DecompressPubkey(viewing); err != nil {
		return nil, err
	}
	clue, err := wallet.Bytes(result.ClueKey, 32)
	if err != nil {
		return nil, err
	}
	if err = core.ValidateClueKey(clue); err != nil {
		return nil, err
	}
	w.keys = core.ReceivingKeys{ViewingPublicKey: viewing, ClueKey: clue}
	w.address, err = core.EncodeLongAddress(core.ReceivingAddress{OwnerAddress: identity, ViewingPublicKey: viewing, ClueKey: clue})
	return w, err
}
func (w *Wallet) Detect(ctx context.Context, executor wallet.Executor, clues [][]byte) ([]bool, error) {
	values := make([]string, len(clues))
	for i, v := range clues {
		values[i] = wallet.Hex(v)
	}
	var result wallet.DetectResult
	if err := w.request(ctx, executor, wallet.Detect, wallet.DetectParams{Clues: values}, &result); err != nil {
		return nil, err
	}
	if len(result.Matches) != len(clues) {
		return nil, errors.New("wallet detection response length mismatch")
	}
	return result.Matches, nil
}
func deliveryParams(envelope []byte, c core.ProtocolContext, token, commitment string, role string) wallet.DecryptParams {
	return wallet.DecryptParams{Envelope: wallet.Hex(envelope), Context: wallet.DeliveryContext{ProtocolContext: wallet.Context(c), TokenAddress: token, OutputCommitment: commitment, Role: role}}
}
func (w *Wallet) DecryptNote(ctx context.Context, executor wallet.Executor, envelope []byte, c core.NoteDecryptionContext) (core.Note, error) {
	c.OwnerAddress = core.BigIntToElement(w.identity.Big())
	var result wallet.DecryptResult
	commitment := c.OutputCommitment.Bytes()
	if err := w.request(ctx, executor, wallet.DecryptNote, deliveryParams(envelope, c.ProtocolContext, c.TokenAddress.String(), wallet.Hex(commitment[:]), ""), &result); err != nil {
		return core.Note{}, err
	}
	plain, err := wallet.Bytes(result.Plaintext, -1)
	if err != nil {
		return core.Note{}, err
	}
	return core.DecodeNotePlaintext(plain, c)
}
func (w *Wallet) DecryptCreditState(ctx context.Context, executor wallet.Executor, envelope []byte, c credit.DeliveryContext, role credit.Role) (credit.DeliveryState, error) {
	c.RecipientIdentity = core.BigIntToElement(w.identity.Big())
	r := "owner"
	if role == credit.Spender {
		r = "spender"
	} else if role != credit.Owner {
		return credit.DeliveryState{}, errors.New("invalid credit role")
	}
	var result wallet.DecryptResult
	commitment := c.Commitment.Bytes()
	if err := w.request(ctx, executor, wallet.DecryptCredit, deliveryParams(envelope, c.ProtocolContext, c.TokenAddress.String(), wallet.Hex(commitment[:]), r), &result); err != nil {
		return credit.DeliveryState{}, err
	}
	plain, err := wallet.Bytes(result.Plaintext, -1)
	if err != nil {
		return credit.DeliveryState{}, err
	}
	return credit.DecodeDeliveryPlaintext(plain, c, role, w.ViewingPublicKey())
}

// PrepareRequest creates a request for external transport, without executing cryptography.
func (w *Wallet) PrepareRequest(id string, method wallet.Method, params any) (wallet.Request, error) {
	if w == nil {
		return wallet.Request{}, errors.New("wallet required")
	}
	return wallet.Prepare(id, w.identity, method, params)
}

// AcceptResponse validates an externally executed request against this public wallet.
func (w *Wallet) AcceptResponse(q wallet.Request, r wallet.Response) (any, error) {
	if w == nil || !strings.EqualFold(q.OwnerAddress, w.identity.Hex()) {
		return nil, errors.New("wallet request identity mismatch")
	}
	value, err := wallet.Accept(q, r, wallet.Hex(w.ViewingPublicKey()))
	if err != nil {
		return nil, err
	}
	if q.Method == wallet.Public && len(w.keys.ViewingPublicKey) > 0 {
		p := value.(wallet.PublicResult)
		viewing, _ := wallet.Bytes(p.ViewingPublicKey, 33)
		clue, _ := wallet.Bytes(p.ClueKey, 32)
		if !bytes.Equal(viewing, w.keys.ViewingPublicKey) || !bytes.Equal(clue, w.keys.ClueKey) {
			return nil, errors.New("resolved wallet receiving keys mismatch")
		}
	}
	return value, nil
}

// WalletFromResponse resolves a public wallet from a validated public-material response.
func WalletFromResponse(q wallet.Request, r wallet.Response) (*Wallet, error) {
	if q.Method != wallet.Public {
		return nil, errors.New("public material request required")
	}
	value, err := wallet.Accept(q, r, "")
	if err != nil {
		return nil, err
	}
	p := value.(wallet.PublicResult)
	viewing, _ := wallet.Bytes(p.ViewingPublicKey, 33)
	clue, _ := wallet.Bytes(p.ClueKey, 32)
	owner := common.HexToAddress(q.OwnerAddress)
	address, err := core.EncodeLongAddress(core.ReceivingAddress{OwnerAddress: core.BigIntToElement(owner.Big()), ViewingPublicKey: viewing, ClueKey: clue})
	if err != nil {
		return nil, err
	}
	return &Wallet{identity: owner, address: address, keys: core.ReceivingKeys{ViewingPublicKey: viewing, ClueKey: clue}}, nil
}
