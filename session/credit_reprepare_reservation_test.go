package session

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	creditsession "github.com/invisi-labs/invisi-sdk-go/session/credit"
)

// A reprepare must use the original operation's reserved input, without making
// that input selectable by an ordinary preparation or a different reservation.
func TestCreditReprepareWalletInputKeepsReservationAndScopeFence(t *testing.T) {
	ctx := context.Background()
	wallet, token, _ := testSession(t, 2000000000)
	base := wallet.options.Provider.(sessionProvider)
	wallet.options.Provider = creditValidationProvider{sessionProvider: base}
	wallet.rpc = wallet.rpc.WithProvider(wallet.options.Provider)
	current := syncGrantState()
	current.ProtocolContext = wallet.coreContext()
	current.TokenAddress = primitives.BigIntToElement(token.Big())
	current.OwnerWalletIdentity = primitives.BigIntToElement(wallet.wallet.OwnerAddress().Big())
	current.SpenderWalletIdentity = current.OwnerWalletIdentity
	current.Expiry = primitives.ElementFromUint64(2100000000)
	leaf, err := current.Commitment()
	if err != nil {
		t.Fatal(err)
	}
	commitment := common.Hash(leaf.Bytes())
	base.leaves.Store(uint64(2), commitment)
	var preparationCtx context.Context
	// Keep the cached wallet frame fixed and stop after its real builder has
	// verified the input. Session Credit tests cover replacing the fresh journal.
	stopAfterInput := errors.New("wallet input inspected; stop before journal replacement")
	coreContext := wallet.coreContext()
	managed, err := creditsession.New(creditsession.Options{ChainTimestamp: func(context.Context) (uint64, error) { return 2000000000 + 86400, nil }, Store: creditStoreAdapter{session: wallet}, OwnerAddress: wallet.scope.OwnerAddress, Context: &coreContext, Build: func(ctx context.Context, p any) (*credit.Prepared, error) {
		preparationCtx = ctx
		prepared, err := wallet.buildCreditRequest(ctx, p)
		if err == nil && creditsession.PreparationReservation(ctx) != "" {
			return nil, stopAfterInput
		}
		return prepared, err
	}})
	if err != nil {
		t.Fatal(err)
	}
	keys := primitives.ReceivingKeys{ViewingPublicKey: wallet.wallet.ViewingPublicKey(), ClueKey: wallet.wallet.ClueKey()}
	nf, err := current.Nullifier()
	if err != nil {
		t.Fatal(err)
	}
	if err = managed.Track(ctx, creditsession.StoredState{GrantStatus: 1, Delivery: credit.DeliveryState{State: current, OwnerReceivingKeys: keys, SpenderReceivingKeys: keys}, Commitment: commitment, Nullifier: common.Hash(nf.Bytes()), TreeID: 2, Role: credit.Spender}); err != nil {
		t.Fatal(err)
	}
	if err = wallet.RefreshState(ctx); err != nil {
		t.Fatal(err)
	}
	request := credit.DrawPreparation{TokenAddress: token, Credits: []credit.DrawMember{{Credit: commitment, Amount: big.NewInt(7)}}, Deadline: 2000000300}
	op, err := managed.PrepareDraw(ctx, "original", request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = managed.PrepareDraw(ctx, "ordinary", request); err == nil || !strings.Contains(err.Error(), "Credit role unavailable") {
		t.Fatalf("ordinary selection bypassed reservation: %v", err)
	}
	request.Deadline = 2000000400
	if err = op.ReprepareDraw(ctx, request); !errors.Is(err, stopAfterInput) {
		t.Fatal("same operation cannot rebuild its reserved input", err)
	}
	authorizedCtx := preparationCtx
	// Use the context actually supplied by ReprepareDraw; no test-only way to
	// manufacture reservation authority is exposed by the managed Credit package.
	if _, err = wallet.buildCreditRequest(authorizedCtx, request); err != nil {
		t.Fatal("authorized input failed membership reconstruction", err)
	}
	key := creditsession.GrantKey(commitment, credit.Spender)
	snapshot, err := wallet.options.Store.ReadSnapshot(ctx, wallet.scope)
	if err != nil {
		t.Fatal(err)
	}
	original := snapshot.PostpaidCredit.Grants[key]
	if original.ReservedBy != "original" {
		t.Fatalf("reservation changed to %q", original.ReservedBy)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*creditsession.StoredState)
		want   string
	}{
		{"different reservation", func(g *creditsession.StoredState) { g.ReservedBy = "other" }, "Credit role unavailable"},
		{"consumed", func(g *creditsession.StoredState) { g.Consumed = true }, "Credit role unavailable"},
		{"revoked", func(g *creditsession.StoredState) { g.GrantStatus = 2; g.Revoked = true }, "Credit role unavailable"},
		{"wrong token", func(g *creditsession.StoredState) { g.Delivery.State.TokenAddress = primitives.ElementFromUint64(99) }, "deployment mismatch"},
		{"wrong role identity", func(g *creditsession.StoredState) {
			g.Delivery.State.SpenderWalletIdentity = primitives.ElementFromUint64(99)
		}, "deployment mismatch"},
		{"wrong deployment", func(g *creditsession.StoredState) {
			g.Delivery.State.ProtocolContext.ChainID = primitives.ElementFromUint64(99)
		}, "deployment mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := wallet.options.Store.Update(ctx, wallet.scope, func(v *WalletState) error { g := original; tc.mutate(&g); v.PostpaidCredit.Grants[key] = g; return nil }); err != nil {
				t.Fatal(err)
			}
			if _, err := wallet.buildCreditRequest(authorizedCtx, request); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("input fence bypassed: %v", err)
			}
		})
	}
	t.Run("membership still authenticated", func(t *testing.T) {
		if err := wallet.options.Store.Update(ctx, wallet.scope, func(v *WalletState) error {
			v.PostpaidCredit.Grants[key] = original
			m := v.PreparationSnapshot.Memberships[commitment]
			m.Path[0] = common.HexToHash("0x1")
			v.PreparationSnapshot.Memberships[commitment] = m
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := wallet.buildCreditRequest(authorizedCtx, request); err == nil || !strings.Contains(err.Error(), "invalid cached membership path") {
			t.Fatalf("reservation authority bypassed membership authentication: %v", err)
		}
	})
}
