package session

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestDetachedCreditRequestsBuildWithoutManagedStore(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, 2000000000)
	provider := creditValidationProvider{sessionProvider: s.options.Provider.(sessionProvider)}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: s.rpc})
	address, _ := s.wallet.GetAddress()
	identity := sdk.TransactionIdentity{Wallet: s.wallet, Protocol: s.protocol}
	prepared, err := invisi.PrepareCreateCredit(ctx, sdk.ExplicitCreateCreditPreparation{TransactionIdentity: identity, CreatePreparation: credit.CreatePreparation{TokenAddress: token, Spender: address, Expiry: 2100000000, AvailableCredit: big.NewInt(400)}})
	if err != nil {
		t.Fatal(err)
	}
	i, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if i.CreatedState == nil || len(i.Deliveries) != 2 || i.Create.RevocationShard != 1 || i.Create.RevocationSlot != 1 {
		t.Fatal("detached request omitted allocation/opening/deliveries")
	}
	_, err = invisi.PrepareUpdateCredit(ctx, sdk.ExplicitUpdateCreditPreparation{TransactionIdentity: identity, UpdatePreparation: credit.UpdatePreparation{TokenAddress: token, Credit: i.Create.CreditCommitment}})
	var invisiError *sdk.InvisiError
	if !errors.As(err, &invisiError) || invisiError.Code != sdk.PreparationFailed {
		t.Fatal("missing detached input did not use stable preparation error", err)
	}
	if _, err = invisi.PrepareRevokeCredit(ctx, sdk.ExplicitRevokeCreditPreparation{TransactionIdentity: identity, RevokePreparation: credit.RevokePreparation{TokenAddress: token, Shard: 1, Slot: 1}}); err != nil {
		t.Fatal(err)
	}
	current := syncGrantState()
	current.ProtocolContext = s.coreContext()
	current.TokenAddress = primitives.BigIntToElement(token.Big())
	current.OwnerWalletIdentity = primitives.BigIntToElement(s.wallet.OwnerAddress().Big())
	current.SpenderWalletIdentity = current.OwnerWalletIdentity
	current.Expiry = primitives.ElementFromUint64(2100000000)
	leaf, _ := current.Commitment()
	provider.leaves.Store(uint64(2), common.Hash(leaf.Bytes()))
	keys := primitives.ReceivingKeys{ViewingPublicKey: s.wallet.ViewingPublicKey(), ClueKey: s.wallet.ClueKey()}
	inputs := []sdk.CreditInput{{Delivery: credit.DeliveryState{State: current, OwnerReceivingKeys: keys, SpenderReceivingKeys: keys}, Commitment: common.Hash(leaf.Bytes()), TreeID: 2}}
	update, err := invisi.PrepareUpdateCredit(ctx, sdk.ExplicitUpdateCreditPreparation{TransactionIdentity: identity, UpdatePreparation: credit.UpdatePreparation{TokenAddress: token, Credit: common.Hash(leaf.Bytes())}, Inputs: inputs})
	if err != nil {
		t.Fatal(err)
	}
	updateIntent, _ := update.ExportIntent()
	if updateIntent.UpdateChange.Policy != current.Policy || updateIntent.UpdateChange.Expiry != current.Expiry {
		t.Fatal("detached update failed optional preservation")
	}
	draw, err := invisi.PrepareDrawCredit(ctx, sdk.ExplicitDrawCreditPreparation{TransactionIdentity: identity, DrawPreparation: credit.DrawPreparation{TokenAddress: token, Credits: []credit.DrawMember{{Credit: common.Hash(leaf.Bytes()), Amount: big.NewInt(7)}}, Recipient: address}, Inputs: inputs})
	if err != nil {
		t.Fatal(err)
	}
	drawIntent, _ := draw.ExportIntent()
	if drawIntent.OutputNote == nil || primitives.ElementToBigInt(drawIntent.OutputNote.Amount).Int64() != 7 {
		t.Fatal("detached draw omitted ordinary Note")
	}
	state, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.PostpaidCredit.Operations) != 0 {
		t.Fatal("detached Credit preparation persisted managed state")
	}
}
