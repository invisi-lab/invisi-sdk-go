package session

import (
	"reflect"
	"testing"

	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestIndependentSigningSurface(t *testing.T) {
	typ := reflect.TypeOf((*sdk.PreparedTransaction)(nil))
	for _, name := range []string{"ExportWalletAuthorization", "WithWalletSignature", "ExportFundingAuthorization", "WithFundingSignature"} {
		if _, ok := typ.MethodByName(name); !ok {
			t.Errorf("missing portable signing entry: %s", name)
		}
	}
	for _, name := range []string{"PrepareTransfer", "PrepareShield"} {
		if _, ok := reflect.TypeOf((*sdk.Invisi)(nil)).MethodByName(name); !ok {
			t.Errorf("missing preparation entry: %s", name)
		}
	}
}
