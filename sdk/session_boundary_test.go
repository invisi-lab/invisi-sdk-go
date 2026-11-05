package sdk_test

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestSDKDoesNotOwnSessions(t *testing.T) {
	for _, name := range []string{"Store", "Signer", "FundingSigner", "Wallet", "Viewing", "Authorization", "Namespace"} {
		if _, ok := reflect.TypeOf(sdk.InvisiOptions{}).FieldByName(name); ok {
			t.Errorf("wallet state and permissions must belong to session: %s", name)
		}
	}
	if _, ok := reflect.TypeOf((*sdk.Invisi)(nil)).MethodByName("Connect"); ok {
		t.Error("sdk must not construct managed sessions")
	}
	for _, name := range []string{"Authorize", "AuthorizeFunding", "Resume", "Reconcile", "Abandon", "TrackTransaction"} {
		if _, ok := reflect.TypeOf((*sdk.PreparedTransaction)(nil)).MethodByName(name); ok {
			t.Errorf("managed API leaked into stateless transaction: %s", name)
		}
	}
}

func TestSDKDependenciesExcludeManagedWalletState(t *testing.T) {
	output, err := exec.Command("go", "list", "-deps", "github.com/invisi-lab/invisi-sdk-go/sdk").CombinedOutput()
	if err != nil {
		t.Fatalf("inspect dependencies: %s: %v", output, err)
	}
	for _, line := range strings.Fields(string(output)) {
		if strings.Contains(line, "invisi-sdk-go/session") || strings.HasPrefix(line, "modernc.org/sqlite") {
			t.Errorf("stateless SDK imports managed persistence: %s", line)
		}
	}
}
