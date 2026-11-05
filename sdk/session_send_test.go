package sdk

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
	"strings"
	"sync"
	"testing"
)

type sendingProvider struct {
	sessionProvider
	mu      sync.Mutex
	sent    []*types.Transaction
	entered chan struct{}
	release chan struct{}
}

func (p *sendingProvider) PendingCodeAt(ctx context.Context, address common.Address) ([]byte, error) {
	return p.CodeAt(ctx, address, nil)
}
func (p *sendingProvider) PendingNonceAt(context.Context, common.Address) (uint64, error) {
	return 7, nil
}
func (p *sendingProvider) SuggestGasPrice(context.Context) (*big.Int, error) {
	return big.NewInt(100), nil
}
func (p *sendingProvider) SuggestGasTipCap(context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
}
func (p *sendingProvider) EstimateGas(context.Context, ethereum.CallMsg) (uint64, error) {
	return 123456, nil
}
func (p *sendingProvider) SendTransaction(_ context.Context, tx *types.Transaction) error {
	if p.entered != nil {
		close(p.entered)
		<-p.release
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent = append(p.sent, tx)
	return nil
}
func (p *sendingProvider) FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error) {
	return nil, nil
}
func (p *sendingProvider) SubscribeFilterLogs(context.Context, ethereum.FilterQuery, chan<- types.Log) (ethereum.Subscription, error) {
	return event.NewSubscription(func(quit <-chan struct{}) error { <-quit; return nil }), nil
}
func readySendingOperation(t *testing.T) (*WalletSession, *Operation, *sendingProvider) {
	t.Helper()
	ctx := context.Background()
	session, token, _ := testSession(t)
	provider := &sendingProvider{sessionProvider: session.options.Provider.(sessionProvider)}
	session.options.Provider = provider
	key, _ := crypto.HexToECDSA(strings.Repeat("0", 63) + "1")
	session.options.Signer, _ = bind.NewKeyedTransactorWithChainID(key, big.NewInt(31337))
	address, _ := session.wallet.GetAddress()
	operation, err := session.PrepareTransfer(ctx, TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if err != nil {
		t.Fatal(err)
	}
	operation, err = operation.Authorize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	operation, err = operation.Prove(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return session, operation, provider
}
func TestSessionSendDelegatesToGethAndTracksHash(t *testing.T) {
	session, ready, provider := readySendingOperation(t)
	transaction, err := session.SendTransaction(context.Background(), ready)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.sent) != 1 || provider.sent[0] != transaction || transaction.Nonce() != 7 || transaction.Gas() != 123456 {
		t.Fatal("geth transaction was not returned unchanged")
	}
	state, _ := session.options.Store.ReadSnapshot(context.Background(), session.scope)
	record := state.Operations[ready.ID()]
	if !record.Externalized || record.TransactionHash != transaction.Hash() || record.Status != "uncertain" {
		t.Fatalf("hash not tracked: %+v", record)
	}
	if _, err = session.SendTransaction(context.Background(), ready); err == nil {
		t.Fatal("stale handle sent twice")
	}
}
func TestSessionPlainCallLeavesOperationAssociationToApplication(t *testing.T) {
	session, ready, _ := readySendingOperation(t)
	call, err := ready.ExportTransaction(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := session.SendTransaction(context.Background(), call, TransactionOptions{GasLimit: 200000, Nonce: big.NewInt(12)})
	if err != nil {
		t.Fatal(err)
	}
	if transaction.Nonce() != 12 || transaction.Gas() != 200000 {
		t.Fatal("geth overrides ignored")
	}
	state, _ := session.options.Store.ReadSnapshot(context.Background(), session.scope)
	if state.Operations[ready.ID()].TransactionHash != (common.Hash{}) {
		t.Fatal("plain call inferred operation association")
	}
	call.To = common.Address{1}
	if _, err = session.SendTransaction(context.Background(), call); err == nil {
		t.Fatal("wrong protocol accepted")
	}
}
func TestSessionRejectsOverlappingManagedSend(t *testing.T) {
	session, ready, provider := readySendingOperation(t)
	provider.entered = make(chan struct{})
	provider.release = make(chan struct{})
	done := make(chan error, 1)
	go func() { _, err := session.SendTransaction(context.Background(), ready); done <- err }()
	<-provider.entered
	if _, err := session.SendTransaction(context.Background(), ready); err == nil {
		t.Fatal("overlapping send accepted")
	}
	close(provider.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

type rejectingAssociationStore struct{ WalletStore }

func (s rejectingAssociationStore) Update(ctx context.Context, scope WalletScope, apply func(*WalletState) error) error {
	return s.WalletStore.Update(ctx, scope, func(state *WalletState) error {
		if err := apply(state); err != nil {
			return err
		}
		for _, record := range state.Operations {
			if record.TransactionHash != (common.Hash{}) {
				return errors.New("storage association failed")
			}
		}
		return nil
	})
}
func TestSendReturnsBroadcastHashWhenTrackingFails(t *testing.T) {
	session, ready, _ := readySendingOperation(t)
	store := rejectingAssociationStore{session.options.Store}
	session.options.Store = store
	ready.session.options.Store = store
	transaction, err := session.SendTransaction(context.Background(), ready)
	if transaction == nil || err == nil || !strings.Contains(err.Error(), transaction.Hash().Hex()) {
		t.Fatalf("lost sent transaction: %v %v", transaction, err)
	}
	state, _ := store.ReadSnapshot(context.Background(), session.scope)
	if !state.Operations[ready.ID()].Externalized {
		t.Fatal("tracking failure released exposure guard")
	}
	for _, note := range state.Notes {
		if note.ReservedBy == "" {
			t.Fatal("broadcast input released")
		}
	}
}
func TestManagedExportRejectsExpiredOperation(t *testing.T) {
	_, ready, provider := readySendingOperation(t)
	provider.head.Time = ready.plan.Deadline + 1
	if _, err := ready.ExportTransaction(context.Background()); err == nil {
		t.Fatal("expired operation exported")
	}
}
