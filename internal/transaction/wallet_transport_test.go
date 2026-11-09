package transaction

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	native "github.com/invisi-lab/invisi-sdk-go/internal/testsupport"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func TestWalletRejectsSubstitutedPublicResponse(t *testing.T) {
	e, _ := native.NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	transport := wallet.ClientFunc(func(ctx context.Context, r []wallet.Request) ([]wallet.Response, error) {
		v, err := e.Execute(ctx, r)
		v[0].ID = "substituted"
		return v, err
	})
	if _, err := ConnectWallet(context.Background(), e.OwnerAddress(), transport); err == nil {
		t.Fatal("accepted replayed response")
	}
}
func TestWalletRetainsOnlyPublicMaterial(t *testing.T) {
	e, _ := native.NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	w, err := ConnectWallet(context.Background(), e.OwnerAddress(), e)
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
	resolved, err := ConnectWallet(ctx, local.OwnerAddress(), local)
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

func TestWalletClientNeverReceivesMalformedRequest(t *testing.T) {
	w := &Wallet{identity: common.HexToAddress("0x1111111111111111111111111111111111111111")}
	called := false
	client := wallet.ClientFunc(func(context.Context, []wallet.Request) ([]wallet.Response, error) { called = true; return nil, nil })
	if _, err := w.Detect(context.Background(), client, [][]byte{{1}}); err == nil {
		t.Fatal("malformed request accepted")
	}
	if called {
		t.Fatal("malformed request crossed native transport boundary")
	}
}
