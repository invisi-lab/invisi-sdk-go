package transaction

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	native "github.com/invisi-lab/invisi-sdk-go/internal/testsupport"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func TestWalletRejectsSubstitutedPublicResponse(t *testing.T) {
	e, _ := native.NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	transport := wallet.ClientFunc(func(ctx context.Context, r []wallet.Request) ([]wallet.Response, error) {
		v, err := e.Execute(ctx, r)
		v[0].ID = "substituted"
		return v, err
	})
	if _, err := connectTestWallet(context.Background(), e.OwnerAddress(), transport); err == nil {
		t.Fatal("accepted replayed response")
	}
}
func TestWalletRetainsOnlyPublicMaterial(t *testing.T) {
	e, _ := native.NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	w, err := connectTestWallet(context.Background(), e.OwnerAddress(), e)
	if err != nil {
		t.Fatal(err)
	}
	if w.OwnerAddress() == (common.Address{}) {
		t.Fatal("missing owner")
	}
	typ := reflect.TypeOf(*w)
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Type.Implements(reflect.TypeOf((*wallet.Client)(nil)).Elem()) {
			t.Fatal("business wallet retained private executor")
		}
	}
	for _, name := range []string{"GetViewer", "GetDetector", "ExportAuthorization", "ExportKey"} {
		if _, exists := reflect.TypeOf(w).MethodByName(name); exists {
			t.Fatal("business wallet exposes private capability", name)
		}
	}
}

func TestResolvedWalletRejectsSubstitutedReceivingKeys(t *testing.T) {
	ctx := context.Background()
	local, _ := native.NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	other, _ := native.NewExecutor("0x" + strings.Repeat("0", 63) + "2")
	resolved, err := connectTestWallet(ctx, local.OwnerAddress(), local)
	if err != nil {
		t.Fatal(err)
	}
	q, err := resolved.PrepareRequest("public-refresh", wallet.Public, struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	response, err := local.ExecuteOne(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	otherRequest, _ := wallet.Prepare("other", other.OwnerAddress(), wallet.Public, struct{}{})
	otherResponse, _ := other.ExecuteOne(ctx, otherRequest)
	var original, foreign wallet.PublicResult
	wallet.Decode(response.Result, &original)
	wallet.Decode(otherResponse.Result, &foreign)
	original.ViewingPublicKey = foreign.ViewingPublicKey
	original.ClueKey = foreign.ClueKey
	response.Result = wallet.Encode(original)
	if _, err = resolved.AcceptResponse(q, response); err == nil {
		t.Fatal("accepted changed receiving keys for resolved wallet")
	}
}

func TestAppDecryptorNeverReceivesMalformedEnvelope(t *testing.T) {
	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	called := false
	decryptor := wallet.DeliveryDecryptorFuncs{Note: func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error) {
		called = true
		return nil, nil
	}}
	if _, err := decryptNoteForOwner(context.Background(), owner, decryptor, []byte{1}, primitives.NoteDecryptionContext{}); err == nil {
		t.Fatal("malformed envelope accepted")
	}
	if called {
		t.Fatal("malformed ciphertext reached App decryptor")
	}
}

func TestAppNonrecipientErrorsRemainFailures(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	envelope := make([]byte, 78)
	copy(envelope[1:34], crypto.CompressPubkey(&key.PublicKey))
	owner := common.Address{1}
	for _, appErr := range []error{wallet.ErrNotRecipient, fmt.Errorf("App failure: %w", wallet.ErrNotRecipient)} {
		decryptor := wallet.DeliveryDecryptorFuncs{
			Note:   func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error) { return nil, appErr },
			Credit: func(context.Context, []byte, credit.DeliveryContext, credit.Role) ([]byte, error) { return nil, appErr },
		}
		envelope[0] = 1
		if _, err = decryptNoteForOwner(context.Background(), owner, decryptor, envelope, primitives.NoteDecryptionContext{}); err == nil || errors.Is(err, wallet.ErrNotRecipient) {
			t.Fatal("App Note error became nonrecipient", err)
		}
		envelope[0] = 0x82
		if _, err = decryptCreditForOwner(context.Background(), owner, nil, decryptor, envelope, credit.DeliveryContext{}, credit.Owner); err == nil || errors.Is(err, wallet.ErrNotRecipient) {
			t.Fatal("App Credit error became nonrecipient", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	decryptor := wallet.DeliveryDecryptorFuncs{Note: func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error) {
		cancel()
		return nil, wallet.ErrNotRecipient
	}}
	envelope[0] = 1
	if _, err = decryptNoteForOwner(ctx, owner, decryptor, envelope, primitives.NoteDecryptionContext{}); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation was hidden by App error", err)
	}
}
