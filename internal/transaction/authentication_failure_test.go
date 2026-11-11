package transaction

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"testing"
)

func TestAmbiguousAuthenticationRemainsRetryable(t *testing.T) {
	key, _ := crypto.GenerateKey()
	envelope := make([]byte, 78)
	envelope[0] = 1
	copy(envelope[1:34], crypto.CompressPubkey(&key.PublicKey))
	owner := common.Address{1}
	d := wallet.DeliveryDecryptorFuncs{Note: func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error) {
		return nil, wallet.ErrAuthenticationFailed
	}}
	_, e := decryptNoteForOwner(context.Background(), owner, d, envelope, primitives.NoteDecryptionContext{})
	if !errors.Is(e, wallet.ErrAuthenticationFailed) {
		t.Fatalf("ambiguity lost: %v", e)
	}
}
