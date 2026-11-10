package shared

import (
	"context"
	"errors"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
)

// AppDeliveryDecryptor is an integration placeholder; replace its methods with secure App decryption.
// Return nil, nil for nonrecipients; unavailable service or failed decryption must return an error.
type AppDeliveryDecryptor struct{}

func (AppDeliveryDecryptor) DecryptNote(context.Context, []byte, core.NoteDecryptionContext) ([]byte, error) {
	return nil, errors.New("implement App Note delivery decryption")
}
func (AppDeliveryDecryptor) DecryptCredit(context.Context, []byte, credit.DeliveryContext, credit.Role) ([]byte, error) {
	return nil, errors.New("implement App Credit delivery decryption")
}
