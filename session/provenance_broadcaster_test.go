package session

import (
	"context"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
	"testing"
	"time"
)

func TestRequiredProvenanceMissingCapabilityPreservesPrepared(t *testing.T) {
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	prepared, e := s.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, prepared, []int64{5}, false)
	if e = c.Quote(context.Background(), prepared, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)}); e != nil {
		t.Fatal("quotation must not require local Source capability", e)
	}
}
