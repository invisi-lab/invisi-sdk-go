package session

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
	"testing"
)

func testPlan(t *testing.T) (*credit.Prepared, StoredState) {
	t.Helper()
	s := credit.State{AvailableCredit: core.ElementFromUint64(400), ProtocolContext: core.ProtocolContext{ChainID: core.ElementFromUint64(1), ProtocolAddress: core.ElementFromUint64(17)}, TokenAddress: core.ElementFromUint64(34), SpenderWalletIdentity: core.ElementFromUint64(33), OwnerWalletIdentity: core.ElementFromUint64(44), RevocationShard: core.ElementFromUint64(1), Nonce: core.ElementFromUint64(1), Expiry: core.ElementFromUint64(2100000000), SharedKey: core.ElementFromUint64(2)}
	transition, err := s.Draw(core.ElementFromUint64(7), 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	c := credit.Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	p := credit.DrawParams{Token: c.TokenAddress, Sources: []protocol.InputSource{{TreeId: big.NewInt(1), MerkleRoot: common.HexToHash("0x01")}}, Payers: []credit.Payer{{Payer: common.BigToAddress(big.NewInt(44)), Amount: big.NewInt(7)}}, GrantRefs: []credit.GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(transition.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(transition.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(transition.RecoveryAmount.Bytes())}, NoteCommitment: common.HexToHash("0x02"), TotalAmount: big.NewInt(7), Deadline: big.NewInt(2000000300)}
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: []byte{0x82}, Clue: clue}, {Ciphertext: []byte{0x82}, Clue: clue}, {}}
	prepared, err := credit.PrepareDraw(c, p, ds, []credit.LocatedState{{State: s, TreeID: 1, Debit: core.ElementFromUint64(7)}}, core.ElementFromUint64(1))
	if err != nil {
		t.Fatal(err)
	}
	wallet, _ := core.DeriveWallet(common.LeftPadBytes([]byte{1}, 32))
	keys := core.ReceivingKeys{ViewingPublicKey: wallet.ViewingPublicKey, ClueKey: wallet.ClueKey}
	commitment, _ := s.Commitment()
	return prepared, StoredState{Delivery: credit.DeliveryState{State: s, OwnerReceivingKeys: keys, SpenderReceivingKeys: keys}, Commitment: common.Hash(commitment.Bytes()), Nullifier: common.Hash(transition.Nullifier.Bytes()), TreeID: 1, Role: credit.Spender}
}
func TestAtomicReservationsAndExternalization(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	session, err := New(Options{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	prepared, state := testPlan(t)
	if err = session.Track(ctx, state); err != nil {
		t.Fatal(err)
	}
	first, err := session.Prepare(ctx, "first", prepared)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.Prepare(ctx, "second", prepared); err == nil {
		t.Fatal("double reservation")
	}
	if err = first.Cancel(ctx); err != nil {
		t.Fatal(err)
	}
	second, err := session.Prepare(ctx, "second", prepared)
	if err != nil {
		t.Fatal(err)
	}
	if err = second.Externalize(ctx, common.Hash{}); err == nil {
		t.Fatal("externalized without frozen calldata")
	}
	if err = second.Cancel(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestManagedScopeRejectsForeignGrant(t *testing.T) {
	_, state := testPlan(t)
	managed, err := New(Options{Store: NewMemoryStore(), OwnerAddress: common.HexToAddress("0xdead")})
	if err != nil {
		t.Fatal(err)
	}
	if err = managed.Track(context.Background(), state); err == nil {
		t.Fatal("tracked another wallet's private grant")
	}
}

type reconcileFunc func(context.Context, credit.Intent, []byte, common.Hash) (*Outcome, error)

func (f reconcileFunc) Reconcile(ctx context.Context, i credit.Intent, d []byte, h common.Hash) (*Outcome, error) {
	return f(ctx, i, d, h)
}
func TestRevocationJournalSurvivesUncertainSubmitAndResume(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	managed, err := New(Options{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	_, grant := testPlan(t)
	grant.Role = credit.Owner
	if err = managed.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	state := grant.Delivery.State
	c := credit.Context{ProtocolContext: state.ProtocolContext, TokenAddress: common.BigToAddress(core.ElementToBigInt(state.TokenAddress)), Timestamp: 2000000000}
	p := credit.RevokeParams{Token: c.TokenAddress, Owner: common.BigToAddress(core.ElementToBigInt(state.OwnerWalletIdentity)), Shard: 1, Slot: 0, Deadline: big.NewInt(2000000300)}
	prepared, err := credit.PrepareRevoke(c, p)
	if err != nil {
		t.Fatal(err)
	}
	op, err := managed.Prepare(ctx, "revoke", prepared)
	if err != nil {
		t.Fatal(err)
	}
	data, err := op.Calldata(ctx, p.Owner)
	if err != nil {
		t.Fatal(err)
	}
	hash := common.HexToHash("0x0123")
	if err = op.Externalize(ctx, hash); err != nil {
		t.Fatal(err)
	}
	if err = op.Cancel(ctx); err == nil {
		t.Fatal("released uncertain Credit journal")
	}
	resumed, err := managed.Resume(ctx, "revoke", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = op.Cancel(ctx); err == nil {
		t.Fatal("old worker retained lease after resume")
	}
	reconciler := reconcileFunc(func(context.Context, credit.Intent, []byte, common.Hash) (*Outcome, error) {
		return &Outcome{TransactionHash: hash, Calldata: data, BlockHash: common.HexToHash("0xbeef"), Success: true, RevokedOwner: p.Owner, RevokedToken: p.Token, RevokedShard: 1}, nil
	})
	if err = resumed.Reconcile(ctx, reconciler); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.Grants[GrantKey(grant.Commitment, grant.Role)].Revoked || snapshot.Operations["revoke"].Status != "confirmed" {
		t.Fatal("canonical revocation did not persist")
	}
}

func TestManagedValidationFailureDoesNotReserveInputs(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	prepared, grant := testPlan(t)
	reject := false
	managed, err := New(Options{Store: store, Validate: func(context.Context, credit.Intent) error {
		if reject {
			return fmt.Errorf("canonical grant revoked")
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err = managed.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	reject = true
	if _, err = managed.Prepare(ctx, "invalid", prepared); err == nil {
		t.Fatal("ignored current chain validation")
	}
	snapshot, err := store.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Operations) != 0 || snapshot.Grants[GrantKey(grant.Commitment, grant.Role)].ReservedBy != "" {
		t.Fatal("failed preparation reserved a grant")
	}
}

func TestUpdateJournalPreservesDeltaAndResumesWithoutReapplying(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	managed, err := New(Options{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	draw, grant := testPlan(t)
	grant.Role = credit.Owner
	if err = managed.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	original, err := draw.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	current := grant.Delivery.State
	change := credit.Change{Policy: current.Policy, Expiry: current.Expiry, IncreaseAvailableCredit: core.ElementFromUint64(5)}
	next, err := current.Update(change, 2000000300)
	if err != nil {
		t.Fatal(err)
	}
	p := credit.UpdateParams{Token: original.Context.TokenAddress, Owner: common.BigToAddress(core.ElementToBigInt(current.OwnerWalletIdentity)), Source: original.Draw.Sources[0], Nullifier: common.Hash(next.Nullifier.Bytes()), CreditCommitment: common.Hash(next.StateCommitment.Bytes()), RevocationShard: 1, Deadline: big.NewInt(2000000300)}
	prepared, err := credit.PrepareUpdate(original.Context, p, original.Deliveries[:2], credit.LocatedState{State: current, TreeID: 1}, change)
	if err != nil {
		t.Fatal(err)
	}
	_, err = managed.Prepare(ctx, "quota-update", prepared)
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := managed.Resume(ctx, "quota-update", nil)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.ReadSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	record := snapshot.Operations["quota-update"]
	if record.Intent.UpdateChange.IncreaseAvailableCredit != core.ElementFromUint64(5) || record.Intent.Inputs[0].State.AvailableCredit != core.ElementFromUint64(400) {
		t.Fatal("journal replaced delta with a target balance")
	}
	history, err := managed.GetHistory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Operations) != 1 || history.Operations[0].Preview.AvailableCredit.Cmp(big.NewInt(405)) != 0 {
		t.Fatal("resume applied quota delta twice")
	}
	if err = resumed.Cancel(ctx); err != nil {
		t.Fatal(err)
	}
}
