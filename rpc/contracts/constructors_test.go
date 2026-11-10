package contracts_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
)

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
