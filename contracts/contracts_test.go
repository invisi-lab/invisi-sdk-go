package contracts_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/contracts"
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

	required := []string{"payOpen", "payBlind", "depositCredit", "withdraw", "isCommitmentPaid", "isNullifierSpent"}
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
	if len(method.Inputs) != 2 {
		t.Errorf("getInstanceAddress: expected 2 inputs, got %d", len(method.Inputs))
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
