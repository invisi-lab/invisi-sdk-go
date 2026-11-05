package session

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func TestDroppedTrackedTransactionExpiresAndReleasesInputs(t *testing.T) {
	session, ready, provider := readySendingOperation(t)
	transaction, err := session.SendTransaction(context.Background(), ready)
	if err != nil {
		t.Fatal(err)
	}
	tracked, err := session.Resume(context.Background(), ready.ID())
	if err != nil {
		t.Fatal(err)
	}
	header := types.CopyHeader(provider.head)
	header.Number = big.NewInt(11)
	header.ParentHash = provider.head.Hash()
	header.Time = ready.plan.Deadline + 1
	provider.head = header
	resolved, err := tracked.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Status() != "expired" {
		t.Fatal(resolved.Status())
	}
	state, _ := session.options.Store.ReadSnapshot(context.Background(), session.scope)
	record := state.Operations[ready.ID()]
	if record.ResolvedAt == nil || record.TransactionHash != transaction.Hash() {
		t.Fatal("lost resolution")
	}
	for _, note := range state.Notes {
		if note.ReservedBy != "" {
			t.Fatal("expired input remains reserved")
		}
	}
	if err := restoreReservations(&state); err != nil {
		t.Fatal(err)
	}
	for _, note := range state.Notes {
		if note.ReservedBy != "" {
			t.Fatal("expired input re-locked")
		}
	}
}

type confirmingProvider struct {
	*sendingProvider
	receipt       *types.Receipt
	transaction   *types.Transaction
	receiptHeader *types.Header
	spent         bool
	hook          func()
}

func (p *confirmingProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return p.receipt, nil
}
func (p *confirmingProvider) TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error) {
	return p.transaction, false, nil
}
func (p *confirmingProvider) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if number != nil && p.receiptHeader != nil && number.Cmp(p.receiptHeader.Number) == 0 {
		return p.receiptHeader, nil
	}
	return p.sendingProvider.HeaderByNumber(ctx, number)
}
func (p *confirmingProvider) CallContract(ctx context.Context, message ethereum.CallMsg, number *big.Int) ([]byte, error) {
	contractABI, _ := contracts.InvisiProtocolMetaData.GetAbi()
	if len(message.Data) >= 4 {
		if method, err := contractABI.MethodById(message.Data[:4]); err == nil && method.RawName == "isNullifierSpent" {
			if p.hook != nil {
				hook := p.hook
				p.hook = nil
				hook()
			}
			return method.Outputs.Pack(p.spent)
		}
	}
	return p.sendingProvider.CallContract(ctx, message, number)
}
func confirmingFixture(t *testing.T) (*WalletSession, *Operation, *confirmingProvider) {
	t.Helper()
	session, ready, base := readySendingOperation(t)
	transaction, err := session.SendTransaction(context.Background(), ready)
	if err != nil {
		t.Fatal(err)
	}
	receiptHeader := types.CopyHeader(base.head)
	provider := &confirmingProvider{sendingProvider: base, transaction: transaction, receiptHeader: receiptHeader, spent: true,
		receipt: &types.Receipt{TxHash: transaction.Hash(), Status: types.ReceiptStatusSuccessful, BlockNumber: new(big.Int).Set(receiptHeader.Number), BlockHash: receiptHeader.Hash()}}
	latest := types.CopyHeader(receiptHeader)
	latest.Number = big.NewInt(11)
	latest.ParentHash = receiptHeader.Hash()
	latest.Time++
	base.head = latest
	session.options.Provider = provider
	operation, err := session.Resume(context.Background(), ready.ID())
	if err != nil {
		t.Fatal(err)
	}
	return session, operation, provider
}
func TestOlderReceiptConfirmsAtLatestWalletCheckpoint(t *testing.T) {
	session, operation, provider := confirmingFixture(t)
	block := protocolrpc.CanonicalBlock{Number: provider.head.Number.Uint64(), Hash: provider.head.Hash(), ParentHash: provider.head.ParentHash, Timestamp: provider.head.Time}
	err := session.options.Store.Update(context.Background(), session.scope, func(state *WalletState) error { state.CheckedThrough = &block; return nil })
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := operation.Reconcile(context.Background())
	if err != nil || resolved.Status() != "confirmed" {
		t.Fatalf("older receipt: %v %v", resolved, err)
	}
	state, _ := session.options.Store.ReadSnapshot(context.Background(), session.scope)
	for _, note := range state.Notes {
		if !note.Spent || note.ReservedBy != "" {
			t.Fatal("confirmation did not consume input")
		}
	}
	if state.CheckedThrough.Number != 11 {
		t.Fatal("checkpoint regressed")
	}
}
func TestRevertedTransactionKeepsProofReservedUntilExpiry(t *testing.T) {
	session, operation, provider := confirmingFixture(t)
	provider.spent = false
	provider.receipt.Status = types.ReceiptStatusFailed
	uncertain, err := operation.Reconcile(context.Background())
	if err != nil || uncertain.Status() != "uncertain" {
		t.Fatalf("revert: %v %v", uncertain, err)
	}
	state, _ := session.options.Store.ReadSnapshot(context.Background(), session.scope)
	for _, note := range state.Notes {
		if note.ReservedBy == "" {
			t.Fatal("live proof released after revert")
		}
	}
	latest := types.CopyHeader(provider.head)
	latest.Number = big.NewInt(12)
	latest.ParentHash = provider.head.Hash()
	latest.Time = operation.plan.Deadline + 1
	provider.head = latest
	resolved, err := uncertain.Reconcile(context.Background())
	if err != nil || resolved.Status() != "expired" {
		t.Fatalf("expired revert: %v %v", resolved, err)
	}
}

type cancelAfterResolutionStore struct {
	WalletStore
	cancel context.CancelFunc
	id     string
}

func (s cancelAfterResolutionStore) Update(ctx context.Context, scope WalletScope, apply func(*WalletState) error) error {
	err := s.WalletStore.Update(ctx, scope, apply)
	if err == nil {
		state, _ := s.WalletStore.ReadSnapshot(context.Background(), scope)
		if state.Operations[s.id].ResolvedAt != nil {
			s.cancel()
		}
	}
	return err
}
func TestCancellationAfterResolutionCommitDoesNotRollbackRecord(t *testing.T) {
	session, operation, _ := confirmingFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := cancelAfterResolutionStore{session.options.Store, cancel, operation.ID()}
	session.options.Store = store
	resolved, err := operation.Reconcile(ctx)
	if err != nil || resolved.Status() != "confirmed" {
		t.Fatalf("committed cancellation: %v %v", resolved, err)
	}
	if ctx.Err() == nil {
		t.Fatal("test did not cancel after commit")
	}
	state, _ := store.ReadSnapshot(context.Background(), session.scope)
	if state.Operations[operation.ID()].ResolvedAt == nil || state.Operations[operation.ID()].Status != "confirmed" {
		t.Fatal("committed record rolled back")
	}
}
func TestResumedWorkerFencesConfirmationInventoryWrite(t *testing.T) {
	for _, interleave := range []string{"resume", "reset"} {
		t.Run(interleave, func(t *testing.T) {
			session, operation, provider := confirmingFixture(t)
			provider.hook = func() {
				if interleave == "resume" {
					if _, err := session.Resume(context.Background(), operation.ID()); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := session.options.Store.Update(context.Background(), session.scope, func(state *WalletState) error {
						state.Notes = map[common.Hash]StoredNote{}
						state.Invalidated = true
						return nil
					}); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := operation.Reconcile(context.Background()); err == nil {
				t.Fatal("stale confirmation accepted")
			}
			state, _ := session.options.Store.ReadSnapshot(context.Background(), session.scope)
			if interleave == "reset" && len(state.Notes) != 0 {
				t.Fatal("reset inventory recreated")
			}
			for _, note := range state.Notes {
				if note.Spent || note.ReservedBy == "" {
					t.Fatal("stale confirmation changed inventory")
				}
			}
		})
	}
}

// Keep a compile-time assertion that a real geth backend remains the send boundary.
var _ bind.ContractBackend = (*sendingProvider)(nil)
