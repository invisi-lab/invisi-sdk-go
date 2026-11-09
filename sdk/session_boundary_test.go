package sdk_test

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestSDKOptionsContainOnlyInfrastructure(t *testing.T) {
	allowed := map[string]bool{"ProtocolRPC": true, "Provider": true, "Prover": true, "FundingVerifier": true}
	typ := reflect.TypeOf(sdk.InvisiOptions{})
	if typ.NumField() != len(allowed) {
		t.Fatal("unexpected infrastructure option count", typ.NumField())
	}
	for i := 0; i < typ.NumField(); i++ {
		if !allowed[typ.Field(i).Name] {
			t.Error("non-infrastructure option", typ.Field(i).Name)
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
