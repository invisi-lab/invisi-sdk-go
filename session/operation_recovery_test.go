package session

import (
	"context"
	"math/big"
	"testing"

	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestResumeFencesPreviousWorkerAndRequiresFreshAuthorization(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t)
	prepared, e := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Withdrawal: &sdk.Withdrawal{To: session.wallet.OwnerAddress(), Amount: big.NewInt(30)}})
	if e != nil {
		t.Fatal(e)
	}
	if prepared, e = prepared.Authorize(ctx); e != nil {
		t.Fatal(e)
	}
	resumed, e := session.Resume(ctx, prepared.ID())
	if e != nil {
		t.Fatal(e)
	}
	if _, e = prepared.Prove(ctx); e == nil {
		t.Fatal("replaced worker still generated proofs")
	}
	if _, e = resumed.Prove(ctx); e == nil {
		t.Fatal("private signature was persisted")
	}
	if resumed, e = resumed.Authorize(ctx); e != nil {
		t.Fatal(e)
	}
	if _, e = resumed.Prove(ctx); e != nil {
		t.Fatal(e)
	}
}
