package transaction

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

// DecryptNote validates App plaintext against this public wallet and the output commitment.
func (w *Wallet) DecryptNote(ctx context.Context, decryptor wallet.DeliveryDecryptor, envelope []byte, c core.NoteDecryptionContext) (core.Note, error) {
	if w == nil || decryptor == nil {
		return core.Note{}, errors.New("App delivery decryptor required")
	}
	if err := ctx.Err(); err != nil {
		return core.Note{}, err
	}
	if err := wallet.ValidateDeliveryEnvelope(envelope, 0x01); err != nil {
		return core.Note{}, err
	}
	c.OwnerAddress = core.BigIntToElement(w.identity.Big())
	plain, err := decryptor.DecryptNote(ctx, append([]byte(nil), envelope...), c)
	if err := ctx.Err(); err != nil {
		return core.Note{}, err
	}
	if err != nil {
		// App errors must not alias the SDK nonrecipient marker generated from nil plaintext.
		return core.Note{}, fmt.Errorf("App Note decryption failed: %v", err)
	}
	if plain == nil {
		return core.Note{}, wallet.ErrNotRecipient
	}
	if len(plain)+78 > core.MaxCiphertextBytes {
		return core.Note{}, errors.New("plaintext exceeds resource bound")
	}
	return core.DecodeNotePlaintext(append([]byte(nil), plain...), c)
}
func (w *Wallet) DecryptCreditState(ctx context.Context, decryptor wallet.DeliveryDecryptor, envelope []byte, c credit.DeliveryContext, role credit.Role) (credit.DeliveryState, error) {
	if w == nil || decryptor == nil {
		return credit.DeliveryState{}, errors.New("App delivery decryptor required")
	}
	if err := ctx.Err(); err != nil {
		return credit.DeliveryState{}, err
	}
	if err := wallet.ValidateDeliveryEnvelope(envelope, 0x82); err != nil {
		return credit.DeliveryState{}, err
	}
	if role != credit.Owner && role != credit.Spender {
		return credit.DeliveryState{}, errors.New("invalid Credit role")
	}
	c.RecipientIdentity = core.BigIntToElement(w.identity.Big())
	plain, err := decryptor.DecryptCredit(ctx, append([]byte(nil), envelope...), c, role)
	if err := ctx.Err(); err != nil {
		return credit.DeliveryState{}, err
	}
	if err != nil {
		// App errors must not alias the SDK nonrecipient marker generated from nil plaintext.
		return credit.DeliveryState{}, fmt.Errorf("App Credit decryption failed: %v", err)
	}
	if plain == nil {
		return credit.DeliveryState{}, wallet.ErrNotRecipient
	}
	if len(plain)+78 > core.MaxCiphertextBytes {
		return credit.DeliveryState{}, errors.New("plaintext exceeds resource bound")
	}
	return credit.DecodeDeliveryPlaintext(append([]byte(nil), plain...), c, role, w.ViewingPublicKey())
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
