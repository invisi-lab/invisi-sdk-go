package invisi

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/ipp"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
)

func fixture(t *testing.T, kind string) *ipp.Request {
	t.Helper()
	b, e := os.ReadFile("../testdata/ipp-wire-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var corpus struct {
		Cases []struct {
			Kind  string
			Valid bool
			JSON  string
		}
	}
	json.Unmarshal(b, &corpus)
	for _, c := range corpus.Cases {
		if c.Kind != "request" || !c.Valid {
			continue
		}
		r, e := ipp.ParsePaymentRequest([]byte(c.JSON), "")
		if e == nil && string(r.Mode) == kind {
			return r
		}
	}
	t.Fatal("missing fixture")
	return nil
}
func TestPrepareOnlyTrustedActiveRequests(t *testing.T) {
	r := fixture(t, "payment")
	calls := 0
	target := Target[string]{Context: protocolrpc.DeploymentContext{ChainID: 8453, ProtocolAddress: common.HexToAddress(r.Payment.ProtocolAddress)}, PrepareTransfer: func(ctx context.Context, p sdk.TransferPreparation) (string, error) {
		calls++
		if p.Deadline != 1794903600 || len(p.Outputs) != 1 || p.Outputs[0].Amount().Cmp(big.NewInt(10000000)) != 0 || p.Outputs[0].Identity() != common.HexToHash(r.Payment.Recipient.Identity) {
			t.Fatal("wrong terms")
		}
		return "prepared", nil
	}}
	out, e := PrepareRequest(context.Background(), r, target, Options{Now: func() time.Time { return time.Date(2026, 11, 17, 8, 3, 0, 0, time.UTC) }})
	if e != nil || out != "prepared" || calls != 1 {
		t.Fatal(out, e, calls)
	}
	r.Status = "processing"
	if _, e = PrepareRequest(context.Background(), r, target, Options{}); e == nil {
		t.Fatal("processing accepted")
	}
	r.Status = "requires_action"
	target.Context.ChainID = 1
	if _, e = PrepareRequest(context.Background(), r, target, Options{Now: func() time.Time { return time.Date(2026, 11, 17, 8, 3, 0, 0, time.UTC) }}); e == nil {
		t.Fatal("wrong deployment")
	}
	if calls != 1 {
		t.Fatal("unauthorized preparation")
	}
}
func TestCreditKeyPassesExactly(t *testing.T) {
	r := fixture(t, "credit")
	target := Target[string]{Context: protocolrpc.DeploymentContext{ChainID: 8453, ProtocolAddress: common.HexToAddress(r.Credit.ProtocolAddress)}, PrepareCreateCredit: func(ctx context.Context, p sdk.CreateCreditPreparation) (string, error) {
		want, _ := new(big.Int).SetString(r.Credit.CreditSharedKey[2:], 16)
		if p.CreditSharedKey.Cmp(want) != 0 || p.Spender != r.Credit.Spender {
			t.Fatal("key or spender changed")
		}
		return "credit", nil
	}}
	if _, e := PrepareRequest(context.Background(), r, target, Options{Now: func() time.Time { return time.Date(2026, 11, 17, 8, 3, 0, 0, time.UTC) }}); e != nil {
		t.Fatal(e)
	}
}
