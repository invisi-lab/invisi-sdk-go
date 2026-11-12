package session

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

func TestManagedCreditRequestsUseSnapshotOffline(t *testing.T) {
	for _, kind := range []string{"create", "update", "draw", "revoke"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			wallet, token, _ := testSession(t, 2000000000)
			base := wallet.options.Provider.(sessionProvider)
			transport := &offlineProvider{Provider: creditValidationProvider{sessionProvider: base}}
			wallet.options.Provider = transport
			wallet.rpc = wallet.rpc.WithProvider(transport)
			current := syncGrantState()
			current.ProtocolContext = wallet.coreContext()
			current.TokenAddress = primitives.BigIntToElement(token.Big())
			current.OwnerWalletIdentity = primitives.BigIntToElement(wallet.wallet.OwnerAddress().Big())
			current.SpenderWalletIdentity = current.OwnerWalletIdentity
			current.Expiry = primitives.ElementFromUint64(2100000000)
			leaf, _ := current.Commitment()
			base.leaves.Store(uint64(2), common.Hash(leaf.Bytes()))

			var err error

			managed, err := wallet.Credit()
			if err != nil {
				t.Fatal(err)
			}
			{
				keys := primitives.ReceivingKeys{ViewingPublicKey: wallet.wallet.ViewingPublicKey(), ClueKey: wallet.wallet.ClueKey()}
				nf, _ := current.Nullifier()
				role := credit.Owner
				if kind == "draw" {
					role = credit.Spender
				}
				if err = managed.Track(ctx, creditsession.StoredState{Delivery: credit.DeliveryState{State: current, OwnerReceivingKeys: keys, SpenderReceivingKeys: keys}, Commitment: common.Hash(leaf.Bytes()), Nullifier: common.Hash(nf.Bytes()), TreeID: 2, Role: role}); err != nil {
					t.Fatal(err)
				}
			}
			if err = wallet.RefreshState(ctx); err != nil {
				t.Fatal(err)
			}
			state, err := wallet.options.Store.ReadSnapshot(ctx, wallet.scope)
			if err != nil {
				t.Fatal(err)
			}
			codec, err := NewEncryptedWalletCodec(make([]byte, 32))
			if err != nil {
				t.Fatal(err)
			}
			ciphertext, err := codec.Seal(wallet.scope, state)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := codec.Open(wallet.scope, ciphertext)
			if err != nil {
				t.Fatal(err)
			}
			_, ok := decoded.PreparationSnapshot.Memberships[common.Hash(leaf.Bytes())]
			if !ok {
				t.Fatal("normal refresh omitted Credit path")
			}

			restoredStore := NewMemoryWalletStore()
			if err = restoredStore.Update(ctx, wallet.scope, func(v *WalletState) error { *v = decoded; return nil }); err != nil {
				t.Fatal(err)
			}
			transport.disabled.Store(true)
			transport.calls.Store(0)
			dead, _ := sdk.NewProtocolRpcClient("http://127.0.0.1:1/v1")
			reopenedWallet, err := Connect(ctx, Options{Invisi: sdk.NewInvisi(sdk.InvisiOptions{Provider: transport, ProtocolRPC: dead, Prover: &testProver{}}), Wallet: wallet.wallet, Protocol: wallet.protocol, Store: restoredStore, Namespace: wallet.scope.Namespace, PublicHistory: wallet.publicHistory})
			if err != nil {
				t.Fatal(err)
			}
			wallet = reopenedWallet
			managed, err = wallet.Credit()
			if err != nil {
				t.Fatal(err)
			}
			address, err := wallet.wallet.GetAddress()
			if err != nil {
				t.Fatal(err)
			}
			var op *creditsession.Operation
			switch kind {
			case "create":
				op, err = managed.PrepareCreate(ctx, kind, credit.CreatePreparation{TokenAddress: token, Spender: address, AvailableCredit: big.NewInt(400), Expiry: 2100000000, Deadline: 2000000300})
			case "update":
				op, err = managed.PrepareUpdate(ctx, kind, credit.UpdatePreparation{TokenAddress: token, Credit: common.Hash(leaf.Bytes()), UpdatePatch: credit.UpdatePatch{IncreaseAvailableCredit: primitives.ElementFromUint64(5)}, Deadline: 2000000300})
			case "draw":
				op, err = managed.PrepareDraw(ctx, kind, credit.DrawPreparation{TokenAddress: token, Credits: []credit.DrawMember{{Credit: common.Hash(leaf.Bytes()), Amount: big.NewInt(7)}}, Deadline: 2000000300})
			case "revoke":
				op, err = managed.PrepareRevoke(ctx, kind, credit.RevokePreparation{TokenAddress: token, Shard: 1, Deadline: 2000000300})
			}
			if err != nil {
				t.Fatal(err)
			}
			request, err := op.ExportSigningRequest(ctx)
			if err != nil {
				t.Fatal(err)
			}
			signature, err := signFixtureRequest(request)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "revoke" {
				err = wallet.WithCreditRevocationSignature(ctx, op, signature)
			} else {
				err = op.WithWalletSignature(ctx, signature)
				if err == nil {
					err = op.Prove(ctx, &testProver{}, prover.Options{})
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			reopened, e := wallet.Credit()
			if e != nil {
				t.Fatal(e)
			}
			op, e = reopened.Resume(ctx, kind, &testProver{})
			if e != nil {
				t.Fatal(e)
			}
			if _, err = op.Calldata(ctx, wallet.wallet.OwnerAddress()); err != nil {
				t.Fatal(err)
			}
			if transport.calls.Load() != 0 {
				t.Fatalf("offline operation made %d network reads", transport.calls.Load())
			}
			state, _ = wallet.options.Store.ReadSnapshot(ctx, wallet.scope)
			if state.PostpaidCredit.Operations[kind].SubmissionValidation != "required" {
				t.Fatal("missing submission marker")
			}
			if err = op.ValidateSubmission(ctx); err == nil {
				t.Fatal("offline live gate passed")
			}
			transport.disabled.Store(false)
			if err = op.ValidateSubmission(ctx); err != nil {
				t.Fatal("unchanged live state rejected", err)
			}
			fork := base
			fork.head = types.CopyHeader(base.head)
			fork.head.Extra = []byte("replaced preparation anchor")
			wallet.options.Provider = creditValidationProvider{sessionProvider: fork}
			if err = op.ValidateSubmission(ctx); err == nil {
				t.Fatal("cached Credit accepted a replaced preparation anchor")
			}
			wallet.options.Provider = creditValidationProvider{sessionProvider: base, revoked: true, allowance: big.NewInt(0)}
			if kind != "create" {
				if err = op.ValidateSubmission(ctx); err == nil {
					t.Fatal("revoked Credit passed live gate")
				}
			}
		})
	}
}
