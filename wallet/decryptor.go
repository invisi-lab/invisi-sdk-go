package wallet

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

// DeliveryDecryptor is implemented by the App. Only nil plaintext with no error is a nonrecipient.
// The App authenticates ciphertext and owns key storage, unlocking and consent.
type DeliveryDecryptor interface {
	DecryptNote(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error)
	DecryptCredit(context.Context, []byte, credit.DeliveryContext, credit.Role) ([]byte, error)
}

type DeliveryDecryptorFuncs struct {
	Note   func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error)
	Credit func(context.Context, []byte, credit.DeliveryContext, credit.Role) ([]byte, error)
}

func (f DeliveryDecryptorFuncs) DecryptNote(ctx context.Context, envelope []byte, c primitives.NoteDecryptionContext) ([]byte, error) {
	if f.Note == nil {
		return nil, errors.New("App Note decryptor required")
	}
	return f.Note(ctx, envelope, c)
}
func (f DeliveryDecryptorFuncs) DecryptCredit(ctx context.Context, envelope []byte, c credit.DeliveryContext, role credit.Role) ([]byte, error) {
	if f.Credit == nil {
		return nil, errors.New("App Credit decryptor required")
	}
	return f.Credit(ctx, envelope, c, role)
}

var ErrNotRecipient = errors.New("delivery is not addressed to this wallet")

// ValidateDeliveryEnvelope rejects malformed payloads before invoking App cryptography.
func ValidateDeliveryEnvelope(envelope []byte, header byte) error {
	if len(envelope) < 78 || len(envelope) > 4096 || envelope[0] != header {
		return errors.New("invalid delivery envelope")
	}
	if _, err := crypto.DecompressPubkey(envelope[1:34]); err != nil {
		return errors.New("invalid envelope ephemeral key")
	}
	return nil
}
