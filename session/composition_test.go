package session

import (
	"context"
	"math/big"
	"reflect"
	"sync"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestConnectComposesIndependentStoresAroundNeutralInfrastructure(t *testing.T) {
	source, token, _ := testSession(t)
	invisi := sdk.NewInvisi(source.options.InvisiOptions)
	firstStore, secondStore := NewMemoryWalletStore(), NewMemoryWalletStore()
	first, err := Connect(context.Background(), Options{Invisi: invisi, Wallet: source.wallet, Protocol: source.protocol, Store: firstStore})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Connect(context.Background(), Options{Invisi: invisi, Wallet: source.wallet, Protocol: source.protocol, Store: secondStore})
	if err != nil {
		t.Fatal(err)
	}
	address, _ := source.wallet.GetAddress()
	operation, err := first.PrepareShield(context.Background(), sdk.ShieldPreparation{
		TokenAddress: token,
		Outputs:      testOutputs(first, []testPaymentOutput{{To: address, Amount: big.NewInt(1)}}),
		Funding:      sdk.FundingSource{Payer: source.wallet.OwnerAddress(), Amount: big.NewInt(1), Authorization: "caller"},
	})
	if err != nil {
		t.Fatal(err)
	}
	firstState, err := firstStore.ReadSnapshot(context.Background(), first.Scope())
	if err != nil {
		t.Fatal(err)
	}
	secondState, err := secondStore.ReadSnapshot(context.Background(), second.Scope())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := firstState.Operations[operation.ID()]; !ok || len(secondState.Operations) != 0 {
		t.Fatal("independent sessions shared persisted operations")
	}
}

func TestSessionOptionsRequireOnlyPublicWallet(t *testing.T) {
	typ := reflect.TypeOf(Options{})
	for _, name := range []string{"Viewing", "Authorization", "WalletProvider"} {
		if _, ok := typ.FieldByName(name); ok {
			t.Fatal("obsolete injected capability", name)
		}
	}
}

// Metadata reads must remain safe while lifecycle methods update the handle.
func TestOperationMetadataConcurrentWithStateChanges(t *testing.T) {
	operation := &Operation{}
	operation.plan.ID = "metadata-race"
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		for i := 0; i < 1000; i++ {
			operation.mu.Lock()
			snapshot := operation.execution()
			snapshot.Status = "updated"
			operation.apply(snapshot)
			operation.mu.Unlock()
		}
	}()
	go func() {
		defer workers.Done()
		for i := 0; i < 1000; i++ {
			if operation.ID() != "metadata-race" {
				t.Error("operation ID changed")
			}
			_ = operation.Kind()
			_ = operation.String()
			_ = operation.GoString()
		}
	}()
	workers.Wait()
}
