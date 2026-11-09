package transaction

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/native"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"reflect"
	"strings"
	"testing"
)

func TestWalletRejectsSubstitutedPublicResponse(t *testing.T) {
	e, _ := native.NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	transport := wallet.ExecutorFunc(func(ctx context.Context, r wallet.Request) (wallet.Response, error) {
		v, err := e.Execute(ctx, r)
		v.ID = "substituted"
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
		if typ.Field(i).Type.Implements(reflect.TypeOf((*wallet.Executor)(nil)).Elem()) {
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
	response, err := local.Execute(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	otherRequest, _ := wallet.Prepare("other", other.OwnerAddress(), wallet.Public, struct{}{})
	otherResponse, _ := other.Execute(ctx, otherRequest)
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
