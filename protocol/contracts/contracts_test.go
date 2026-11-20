package contracts_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/protocol/contracts"
)

// TestBindingsCompile verifies that all generated ABI bindings parse correctly
// and that their MetaData is valid.
func TestBindingsCompile(t *testing.T) {
	t.Run("InvisiCore ABI parses", func(t *testing.T) {
		if contracts.InvisiCoreMetaData == nil {
			t.Fatal("InvisiCoreMetaData is nil")
		}
		_, err := contracts.InvisiCoreMetaData.GetAbi()
		if err != nil {
			t.Fatalf("InvisiCore ABI parse error: %v", err)
		}
	})

	t.Run("InvisiCoreFactory ABI parses", func(t *testing.T) {
		if contracts.InvisiCoreFactoryMetaData == nil {
			t.Fatal("InvisiCoreFactoryMetaData is nil")
		}
		_, err := contracts.InvisiCoreFactoryMetaData.GetAbi()
		if err != nil {
			t.Fatalf("InvisiCoreFactory ABI parse error: %v", err)
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

// TestInvisiCoreKeyMethods verifies that critical methods exist in the ABI.
func TestInvisiCoreKeyMethods(t *testing.T) {
	parsed, err := contracts.InvisiCoreMetaData.GetAbi()
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	required := []string{"shield", "transact", "createCredit", "creditDraw", "updateCredit", "revokeCredit", "creditAllocation", "creditRevocationBitmap", "activeNoteTreeId", "getNoteTreeRoot", "getNoteTreeNextLeafIndex", "isCommitmentRegistered", "isNullifierSpent"}
	for _, name := range required {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("InvisiCore ABI missing method: %s", name)
		}
	}
}

// TestInvisiCoreFactoryGetInstanceAddress verifies that getInstanceAddress is present.
func TestInvisiCoreFactoryGetInstanceAddress(t *testing.T) {
	parsed, err := contracts.InvisiCoreFactoryMetaData.GetAbi()
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	method, ok := parsed.Methods["getInstanceAddress"]
	if !ok {
		t.Fatal("InvisiCoreFactory ABI missing method: getInstanceAddress")
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

func TestCurrentABIArtifacts(t *testing.T) {
	for name, metadata := range map[string]*contracts.MetaData{"InvisiCore": contracts.InvisiCoreMetaData, "InvisiCoreFactory": contracts.InvisiCoreFactoryMetaData, "VerifierRouter": contracts.VerifierRouterMetaData, "CreditVerifierRouter": contracts.CreditVerifierRouterMetaData, "IERC20": contracts.IERC20MetaData, "InvisiAddressRegistry": contracts.InvisiAddressRegistryMetaData} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("abi/" + name + ".abi.json")
			if err != nil {
				t.Fatal(err)
			}
			var compact bytes.Buffer
			if err = json.Compact(&compact, source); err != nil {
				t.Fatal(err)
			}
			if compact.String() != metadata.ABI {
				t.Fatal("generated binding differs from source ABI; run go generate ./contracts")
			}
			if _, err = metadata.GetAbi(); err != nil {
				t.Fatal(err)
			}
		})
	}
	parsed, _ := contracts.InvisiCoreMetaData.GetAbi()
	for _, name := range []string{"CommitmentsInserted", "NoteOutputs", "TreeCreated", "NoteOperation", "CreditCreated"} {
		if _, ok := parsed.Events[name]; !ok {
			t.Errorf("missing current event %s", name)
		}
	}
	for event, field := range map[string]string{"NoteOutputs": "packedDeliveries", "CommitmentsInserted": "commitments"} {
		found := false
		for _, input := range parsed.Events[event].Inputs {
			found = found || input.Name == field
		}
		if !found {
			t.Errorf("%s must expose %s", event, field)
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

	protocol, err := contracts.InvisiCoreMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := protocol.Methods["addressRegistry"]; !ok {
		t.Fatal("missing immutable protocol registry getter")
	}
	if _, ok := protocol.Methods["setAddressRegistry"]; ok {
		t.Fatal("protocol registry must be immutable")
	}
	if len(protocol.Constructor.Inputs) != 2 || protocol.Constructor.Inputs[0].Name != "params" || protocol.Constructor.Inputs[1].Name != "_addressRegistry" {
		t.Fatal("Core constructor must take instance parameters and address registry")
	}
	registryAddress := common.HexToAddress("0x1234")
	params := contracts.InstanceParams{Factory: common.HexToAddress("0x11"), InitialOwner: common.HexToAddress("0x22"), GlobalEmergency: common.HexToAddress("0x33"), EmergencyAdmin: common.HexToAddress("0x44"), VerifierRouter: common.HexToAddress("0x55")}
	constructor, err := protocol.Pack("", params, registryAddress)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := protocol.Constructor.Inputs.Unpack(constructor)
	if err != nil || len(decoded) != 2 || decoded[1] != registryAddress {
		t.Fatalf("constructor registry did not round trip: %v", err)
	}
	factory, err := contracts.InvisiCoreFactoryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"deploy", "computeInstanceAddress"} {
		method, ok := factory.Methods[name]
		if !ok || len(method.Inputs) != 5 || method.Inputs[4].Type.String() != "address" || method.Inputs[4].Name != "addressRegistry_" {
			t.Fatalf("%s must take address registry as its fifth argument", name)
		}
		data, err := factory.Pack(name, []byte{0x60, 0x00}, [32]byte{1}, params.InitialOwner, params.EmergencyAdmin, registryAddress)
		if err != nil {
			t.Fatal(err)
		}
		selectors := map[string]string{"deploy": "72eff647", "computeInstanceAddress": "51dfabc3"}
		if common.Bytes2Hex(data[:4]) != selectors[name] {
			t.Fatalf("unexpected %s selector: %x", name, data[:4])
		}
		args, err := method.Inputs.Unpack(data[4:])
		if err != nil || len(args) != 5 || args[4] != registryAddress {
			t.Fatalf("%s registry did not round trip: %v", name, err)
		}
	}
}
