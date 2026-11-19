package invisi

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
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

func TestInitialPaymentPreparesGrossDebitWithoutChangingFrozenNet(t *testing.T) {
	for _, fee := range []*credit.Fee{nil, {Recipient: common.HexToAddress("0x1111111111111111111111111111111111111111"), Amount: big.NewInt(1)}} {
		r := fixture(t, "credit")
		r.Credit.InitialPayment = &ipp.InitialPaymentTerms{Amount: "5"}
		expected := int64(5)
		if fee != nil {
			expected = 6
		}
		target := Target[string]{Context: protocolrpc.DeploymentContext{ChainID: 8453, ProtocolAddress: common.HexToAddress(r.Credit.ProtocolAddress)}, PrepareCreateCredit: func(_ context.Context, p sdk.CreateCreditPreparation) (string, error) {
			if p.InitialPayment == nil || p.InitialPayment.Amount.Cmp(big.NewInt(expected)) != 0 {
				t.Fatal("initial gross debit was lost")
			}
			if p.Spender != r.Credit.Spender || p.CreationABI != "initial-payment" {
				t.Fatal("spender or deployment selection changed")
			}
			if fee != nil && (p.Fee == nil || p.Fee.Recipient != fee.Recipient || p.Fee.Amount.Cmp(fee.Amount) != 0) {
				t.Fatal("actual fee was lost")
			}
			return "credit", nil
		}}
		_, err := PrepareRequest(context.Background(), r, target, Options{CreationABI: "initial-payment", InitialPaymentFee: fee, Now: func() time.Time { return time.Date(2026, 11, 17, 8, 3, 0, 0, time.UTC) }})
		if err != nil {
			t.Fatal(err)
		}
		if r.Credit.InitialPayment.Amount != "5" {
			t.Fatal("mutated the merchant frozen amount")
		}
		if fee != nil && fee.Amount.Cmp(big.NewInt(1)) != 0 {
			t.Fatal("mutated caller fee")
		}
	}
}
func TestInitialPaymentFeeValidationDoesNotPrepareInvalidRequests(t *testing.T) {
	r := fixture(t, "credit")
	r.Credit.InitialPayment = &ipp.InitialPaymentTerms{Amount: "5"}
	target := Target[string]{Context: protocolrpc.DeploymentContext{ChainID: 8453, ProtocolAddress: common.HexToAddress(r.Credit.ProtocolAddress)}, PrepareCreateCredit: func(_ context.Context, _ sdk.CreateCreditPreparation) (string, error) {
		t.Fatal("prepared invalid fee")
		return "", nil
	}}
	for _, fee := range []*credit.Fee{{Amount: nil}, {Amount: big.NewInt(-1)}, {Amount: big.NewInt(1)}, {Recipient: common.HexToAddress("0x1111111111111111111111111111111111111111"), Amount: big.NewInt(0)}, {Recipient: common.HexToAddress("0x1111111111111111111111111111111111111111"), Amount: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 126), big.NewInt(5))}} {
		_, err := PrepareRequest(context.Background(), r, target, Options{InitialPaymentFee: fee, Now: func() time.Time { return time.Date(2026, 11, 17, 8, 3, 0, 0, time.UTC) }})
		if err == nil {
			t.Fatal("invalid local fee accepted")
		}
	}
	r.Credit.InitialPayment = nil
	if _, err := PrepareRequest(context.Background(), r, target, Options{InitialPaymentFee: &credit.Fee{Amount: big.NewInt(0)}, Now: func() time.Time { return time.Date(2026, 11, 17, 8, 3, 0, 0, time.UTC) }}); err == nil {
		t.Fatal("fee accepted without an initial payment")
	}
}
