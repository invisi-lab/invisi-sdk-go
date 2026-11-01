package sdk

import (
	"reflect"
	"testing"
)

func TestIndependentSigningSurface(t *testing.T) {
	typ := reflect.TypeOf((*PreparedTransaction)(nil))
	for _, name := range []string{"Authorize", "AuthorizeFunding"} {
		if _, ok := typ.MethodByName(name); ok {
			t.Errorf("independent signing entry remains: %s", name)
		}
	}
	for _, name := range []string{"ExportWalletAuthorization", "WithWalletSignature", "ExportFundingAuthorization", "WithFundingSignature"} {
		if _, ok := typ.MethodByName(name); !ok {
			t.Errorf("missing portable signing entry: %s", name)
		}
	}
	for _, name := range []string{"Signer", "FundingSigner"} {
		if _, ok := reflect.TypeOf(InvisiOptions{}).FieldByName(name); ok {
			t.Errorf("infrastructure signer remains: %s", name)
		}
	}
	for _, name := range []string{"PrepareWithdraw", "PrepareSweep"} {
		if _, ok := reflect.TypeOf((*Invisi)(nil)).MethodByName(name); ok {
			t.Errorf("retired entry remains: %s", name)
		}
	}
}
