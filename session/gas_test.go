package session

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type estimateProvider struct {
	sessionProvider
	call  ethereum.CallMsg
	calls int
	err   error
}

func (p *estimateProvider) EstimateGas(_ context.Context, call ethereum.CallMsg) (uint64, error) {
	p.call = call
	p.calls++
	return 654321, p.err
}

func TestFinalGasEstimateUsesFrozenCallAndCaller(t *testing.T) {
	ctx := context.Background()
	s, ready, _ := readySendingOperation(t)
	provider := &estimateProvider{sessionProvider: s.options.Provider.(*sendingProvider).sessionProvider}
	s.options.Provider = provider
	ready.session.options.Provider = provider
	from := common.HexToAddress("0x1234")
	result, err := ready.EstimateGas(ctx, from)
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != "rpc" || result.GasUnits != 654321 || provider.call.From != from || provider.call.To == nil || *provider.call.To != ready.plan.Context.ProtocolAddress || !bytes.Equal(provider.call.Data, ready.data) || provider.call.Value.Sign() != 0 {
		t.Fatal("estimate did not use final call", result, provider.call)
	}
	state, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if !state.Operations[ready.ID()].Externalized {
		t.Fatal("estimate bypassed managed export fence")
	}
	detached, err := transaction.NewDetachedExecution(ready.execution())
	if err != nil {
		t.Fatal(err)
	}
	result, err = detached.EstimateGas(ctx, from)
	if err != nil || result.GasUnits != 654321 {
		t.Fatal(result, err)
	}
	provider.err = errors.New("estimate rejected")
	if _, err = detached.EstimateGas(ctx, from); !errors.Is(err, provider.err) {
		t.Fatal("provider error lost", err)
	}
}

func TestFinalGasEstimateRejectsUnreadyAndMissingCapability(t *testing.T) {
	ctx := context.Background()
	s, ready, _ := readySendingOperation(t)
	s.options.Provider = sessionProvider{}
	ready.session.options.Provider = sessionProvider{}
	if _, err := ready.EstimateGas(ctx, common.Address{}); err == nil {
		t.Fatal("missing estimator accepted")
	}
	invisi, input, _ := explicitFixture(t)
	prepared, err := invisi.PrepareTransfer(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	provider := &estimateProvider{}
	transaction.PreparedExecution(prepared).Runtime.Options.Provider = provider
	if _, err = prepared.EstimateGas(ctx, common.Address{}); err == nil || provider.calls != 0 {
		t.Fatal("unready estimate accepted")
	}
}
