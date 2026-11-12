package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
	"math/big"
	"testing"
)

func TestManagedAllCreditPreparationsUseSnapshotOffline(t *testing.T) {
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
			c := credit.Context{ProtocolContext: current.ProtocolContext, TokenAddress: token, Timestamp: 2000000000}
			leaf, _ := current.Commitment()
			base.leaves.Store(uint64(2), common.Hash(leaf.Bytes()))
			root := leaf
			for range 16 {
				root, _ = primitives.Poseidon2Hash(root, primitives.ElementFromUint64(0))
			}
			source := protocol.InputSource{TreeId: big.NewInt(2), MerkleRoot: common.Hash(root.Bytes())}
			clue := make([]byte, 68)
			clue[64] = 24
			deliveries := []protocol.Delivery{{Ciphertext: []byte{0x82}, Clue: clue}, {Ciphertext: []byte{0x82}, Clue: clue}}
			var prepared *credit.Prepared
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
			membership, ok := decoded.PreparationSnapshot.Memberships[common.Hash(leaf.Bytes())]
			if !ok {
				t.Fatal("normal refresh omitted Credit path")
			}
			source = protocol.InputSource{TreeId: new(big.Int).SetUint64(membership.TreeID), MerkleRoot: membership.Root}
			located := credit.LocatedState{State: current, TreeID: membership.TreeID, LeafIndex: uint16(membership.LeafIndex)}
			copy(located.Path[:], membership.Path)

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
			switch kind {
			case "create":
				created := current
				created.RevocationSlot = primitives.ElementFromUint64(1)
				commitment, _ := created.Commitment()
				prepared, err = credit.PrepareCreate(c, credit.CreateParams{Token: token, Owner: wallet.wallet.OwnerAddress(), RevocationShard: 1, RevocationSlot: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}, deliveries, created)
			case "update":
				change := credit.Change{Policy: current.Policy, Expiry: current.Expiry, IncreaseAvailableCredit: primitives.ElementFromUint64(5)}
				next, _ := current.Update(change, 2000000300)
				prepared, err = credit.PrepareUpdate(c, credit.UpdateParams{Token: token, Owner: wallet.wallet.OwnerAddress(), Source: source, Nullifier: common.Hash(next.Nullifier.Bytes()), CreditCommitment: common.Hash(next.StateCommitment.Bytes()), RevocationShard: 1, Deadline: big.NewInt(2000000300)}, deliveries, located, change)
			case "draw":
				debit := primitives.ElementFromUint64(7)
				next, _ := current.Draw(debit, c.Timestamp)
				identity := primitives.ElementFromUint64(1)
				output, _ := primitives.ComputeNoteCommitment(c.ChainID, c.ProtocolAddress, identity, current.TokenAddress, debit)
				drawInput := located
				drawInput.Debit = debit
				prepared, err = credit.PrepareDraw(c, credit.DrawParams{Token: token, Sources: []protocol.InputSource{source}, Payers: []credit.Payer{{Payer: wallet.wallet.OwnerAddress(), Amount: big.NewInt(7)}}, GrantRefs: []credit.GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(next.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(next.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(next.RecoveryAmount.Bytes())}, NoteCommitment: common.Hash(output.Bytes()), TotalAmount: big.NewInt(7), Deadline: big.NewInt(2000000300)}, append(deliveries, protocol.Delivery{}), []credit.LocatedState{drawInput}, identity)
			case "revoke":
				prepared, err = credit.PrepareRevoke(c, credit.RevokeParams{Token: token, Owner: wallet.wallet.OwnerAddress(), Shard: 1, Deadline: big.NewInt(2000000300)})
			}
			if err != nil {
				t.Fatal(err)
			}
			op, err := managed.Prepare(ctx, kind, prepared)
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
