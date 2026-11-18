package session

import (
	"context"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"math/big"
	"testing"
)

func TestExplicitCreditKeySurvivesManagedJournalAndResume(t *testing.T) {
	ctx := context.Background()
	wallet, token, _ := testSession(t, 2000000000)
	wallet.options.Provider = creditValidationProvider{sessionProvider: wallet.options.Provider.(sessionProvider)}
	managed, e := wallet.Credit()
	if e != nil {
		t.Fatal(e)
	}
	address, e := wallet.wallet.GetAddress()
	if e != nil {
		t.Fatal(e)
	}
	_, e = managed.PrepareCreate(ctx, "ipp-original", credit.CreatePreparation{TokenAddress: token, Spender: address, AvailableCredit: big.NewInt(400), Expiry: 2100000000, Deadline: 2000000300, CreditSharedKey: big.NewInt(123456789)})
	if e != nil {
		t.Fatal(e)
	}
	state, e := wallet.options.Store.ReadSnapshot(ctx, wallet.scope)
	if e != nil {
		t.Fatal(e)
	}
	before := state.PostpaidCredit.Operations["ipp-original"].Intent
	if before.CreatedState == nil || before.CreatedState.CreditSharedKey != primitives.ElementFromUint64(123456789) {
		t.Fatal("journal changed explicit key")
	}
	codec, e := NewEncryptedWalletCodec(make([]byte, 32))
	if e != nil {
		t.Fatal(e)
	}
	ciphertext, e := codec.Seal(wallet.scope, state)
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := codec.Open(wallet.scope, ciphertext)
	if e != nil {
		t.Fatal(e)
	}
	if e = wallet.options.Store.Update(ctx, wallet.scope, func(v *WalletState) error { *v = decoded; return nil }); e != nil {
		t.Fatal(e)
	}
	if _, e = managed.Resume(ctx, "ipp-original", nil); e != nil {
		t.Fatal(e)
	}
	state, e = wallet.options.Store.ReadSnapshot(ctx, wallet.scope)
	if e != nil {
		t.Fatal(e)
	}
	after := state.PostpaidCredit.Operations["ipp-original"].Intent
	if after.CreatedState.CreditSharedKey != before.CreatedState.CreditSharedKey || after.Create.CreditCommitment != before.Create.CreditCommitment {
		t.Fatal("resume changed key or commitment")
	}
	for i := range before.Deliveries {
		if string(after.Deliveries[i].Ciphertext) != string(before.Deliveries[i].Ciphertext) {
			t.Fatal("resume changed ciphertext")
		}
	}
}
