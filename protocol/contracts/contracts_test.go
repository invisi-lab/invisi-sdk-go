package contracts_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

// TestBindingsCompile verifies that all generated ABI bindings parse correctly
// and that their MetaData is valid.
func TestBindingsCompile(t *testing.T) {
	t.Run("InvisiProtocol ABI parses", func(t *testing.T) {
		if contracts.InvisiProtocolMetaData == nil {
			t.Fatal("InvisiProtocolMetaData is nil")
		}
		_, err := contracts.InvisiProtocolMetaData.GetAbi()
		if err != nil {
			t.Fatalf("InvisiProtocol ABI parse error: %v", err)
		}
	})

	t.Run("InvisiFactory ABI parses", func(t *testing.T) {
		if contracts.InvisiFactoryMetaData == nil {
			t.Fatal("InvisiFactoryMetaData is nil")
		}
		_, err := contracts.InvisiFactoryMetaData.GetAbi()
		if err != nil {
			t.Fatalf("InvisiFactory ABI parse error: %v", err)
		}
	})

	t.Run("VerifierRouter ABI parses", func(t *testing.T) {
		if contracts.VerifierRouterMetaData == nil {
			t.Fatal("VerifierRouterMetaData is nil")
		}
		_, err := contracts.VerifierRouterMetaData.GetAbi()
		if err != nil {
			t.Fatalf("VerifierRouter ABI parse error: %v", err)
		}
	})

	t.Run("IERC20 ABI parses", func(t *testing.T) {
		if contracts.IERC20MetaData == nil {
			t.Fatal("IERC20MetaData is nil")
		}
		_, err := contracts.IERC20MetaData.GetAbi()
		if err != nil {
			t.Fatalf("IERC20 ABI parse error: %v", err)
		}
	})
}

// TestInvisiProtocolKeyMethods verifies that critical methods exist in the ABI.
func TestInvisiProtocolKeyMethods(t *testing.T) {
	parsed, err := contracts.InvisiProtocolMetaData.GetAbi()
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	required := []string{"shield", "transact", "createCredit", "creditDraw", "cancelCredit", "topUpCredit", "activeNoteTreeId", "getNoteTreeRoot", "getNoteTreeNextLeafIndex", "isCommitmentPaid", "isNullifierSpent"}
	for _, name := range required {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("InvisiProtocol ABI missing method: %s", name)
		}
	}
}

// TestInvisiFactoryGetInstanceAddress verifies that getInstanceAddress is present.
func TestInvisiFactoryGetInstanceAddress(t *testing.T) {
	parsed, err := contracts.InvisiFactoryMetaData.GetAbi()
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	method, ok := parsed.Methods["getInstanceAddress"]
	if !ok {
		t.Fatal("InvisiFactory ABI missing method: getInstanceAddress")
	}
	if len(method.Inputs) != 1 {
		t.Errorf("getInstanceAddress: expected 1 input, got %d", len(method.Inputs))
	}
	if len(method.Outputs) != 1 {
		t.Errorf("getInstanceAddress: expected 1 output, got %d", len(method.Outputs))
	}
}

// TestIERC20KeyMethods verifies that standard ERC20 methods are present.
func TestIERC20KeyMethods(t *testing.T) {
	parsed, err := contracts.IERC20MetaData.GetAbi()
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	required := []string{"balanceOf", "transfer", "transferFrom", "approve", "allowance"}
	for _, name := range required {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("IERC20 ABI missing method: %s", name)
		}
	}
}

// TestNewBinderConstructors verifies that the generated binder constructors
// compile and can be called with a zero address and nil caller.
// Actual RPC calls require a live backend; this test just checks the type system.
func TestNewBinderConstructors(t *testing.T) {
	zero := common.Address{}

	// Each constructor should succeed (nil caller is stored, not validated until call time).
	if _, err := contracts.NewInvisiProtocolCaller(zero, nil); err != nil {
		t.Errorf("NewInvisiProtocolCaller: %v", err)
	}
	if _, err := contracts.NewInvisiFactoryCaller(zero, nil); err != nil {
		t.Errorf("NewInvisiFactoryCaller: %v", err)
	}
	if _, err := contracts.NewVerifierRouterCaller(zero, nil); err != nil {
		t.Errorf("NewVerifierRouterCaller: %v", err)
	}
	if _, err := contracts.NewIERC20Caller(zero, nil); err != nil {
		t.Errorf("NewIERC20Caller: %v", err)
	}
}

// Ensure bind package is used (suppress unused import errors).
var _ = bind.Bind

func TestCurrentABIArtifacts(t *testing.T) {
	for name, metadata := range map[string]*bind.MetaData{"InvisiProtocol": contracts.InvisiProtocolMetaData, "InvisiFactory": contracts.InvisiFactoryMetaData, "VerifierRouter": contracts.VerifierRouterMetaData, "CreditVerifierRouter": contracts.CreditVerifierRouterMetaData, "IERC20": contracts.IERC20MetaData, "InvisiAddressRegistry": contracts.InvisiAddressRegistryMetaData} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("abi/" + name + ".abi.json")
			if err != nil {
				t.Fatal(err)
			}
			var compact bytes.Buffer
			if err = json.Compact(&compact, source); err != nil {
				t.Fatal(err)
			}
			if strings.ReplaceAll(compact.String(), " ", "") != metadata.ABI {
				t.Fatal("generated binding differs from source ABI; run go generate ./contracts")
			}
			if _, err = metadata.GetAbi(); err != nil {
				t.Fatal(err)
			}
		})
	}
	parsed, _ := contracts.InvisiProtocolMetaData.GetAbi()
	for _, name := range []string{"CommitmentsInserted", "BatchOutputs", "TreeCreated", "Shield", "CreditCreated"} {
		if _, ok := parsed.Events[name]; !ok {
			t.Errorf("missing current event %s", name)
		}
	}
	for _, name := range []string{"payOpen", "payBlind", "depositCredit", "withdraw"} {
		if _, ok := parsed.Methods[name]; ok {
			t.Errorf("legacy entry point retained: %s", name)
		}
	}
}

func TestFinalRegistryAndFactoryABIContract(t *testing.T) {
	registry, err := contracts.InvisiAddressRegistryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	if len(registry.Constructor.Inputs) != 0 {
		t.Fatal("registry constructor must take no verifier or other arguments")
	}
	for _, name := range []string{"registerKeys", "updateKeys"} {
		method := registry.Methods[name]
		if len(method.Inputs) != 2 || method.Inputs[1].Name != "signature" || method.Inputs[1].Type.String() != "bytes" {
			t.Fatalf("unexpected signature entrypoint %s", name)
		}
		tuple := method.Inputs[0].Type
		if len(tuple.TupleElems) != 4 || tuple.TupleElems[0].String() != "bytes32" || tuple.TupleElems[2].String() != "uint64" || tuple.TupleElems[3].String() != "uint64" {
			t.Fatalf("unexpected registration tuple %s", tuple.String())
		}
	}
	for _, name := range []string{"registrationVerifier", "verifier", "setRegistrationVerifier"} {
		if _, ok := registry.Methods[name]; ok {
			t.Fatalf("obsolete registry verifier API %s", name)
		}
	}
	protocol, err := contracts.InvisiProtocolMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := protocol.Methods["addressRegistry"]; !ok {
		t.Fatal("missing immutable protocol registry getter")
	}
	if _, ok := protocol.Methods["setAddressRegistry"]; ok {
		t.Fatal("protocol registry must be immutable")
	}
	factory, err := contracts.InvisiFactoryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"deploy", "computeInstanceAddress"} {
		method, ok := factory.Methods[name]
		if !ok || len(method.Inputs) != 6 || method.Inputs[5].Type.String() != "address" || method.Inputs[5].Name != "addressRegistry_" {
			t.Fatalf("%s must append address registry as its sixth argument", name)
		}
	}
}
