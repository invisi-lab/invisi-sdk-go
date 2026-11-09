package session

import (
	"testing"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func walletAddress(w sdk.WalletReference) string {
	var value string
	var e error
	switch w := w.(type) {
	case *sdk.Wallet:
		value, e = w.GetAddress()
	case *sdk.PublicWallet:
		value, e = w.GetAddress()
	}
	if e != nil {
		panic(e)
	}
	return value
}

func detachedForTest(t *testing.T, o *Operation, state WalletState) *sdk.PreparedTransaction {
	t.Helper()
	execution := o.execution()
	runtime := *execution.Runtime
	runtime.ExplicitInventory = &transaction.Inventory{Notes: state.Notes}
	runtime.InventoryReader = nil
	runtime.OnComplete = nil
	runtime.OnPersist = nil
	runtime.OnLease = nil
	runtime.OnInvalidate = nil
	runtime.OnReprice = nil
	runtime.OnExternalize = nil
	execution.Runtime = &runtime
	p, e := transaction.NewDetachedExecution(execution)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
