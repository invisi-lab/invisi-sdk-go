package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
	"strings"
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
	required, e := c.WithProvenancePolicy(common.HexToHash("0x77"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = prepared.WithBroadcaster(context.Background(), required, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)}); e == nil || !strings.Contains(e.Error(), "capability") {
		t.Fatal("missing capability did not fail closed before quote", e)
	}
	if prepared.Status() != "prepared" {
		t.Fatal("failed source preparation consumed handle")
	}
}
