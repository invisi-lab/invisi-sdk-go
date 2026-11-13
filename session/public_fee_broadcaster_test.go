package session

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestBroadcasterReplacesExistingPublicFee(t *testing.T) {
	cases := []struct {
		kind            string
		source          sdk.BroadcasterFeeSource
		funding, output int64
	}{
		{"shield", "", 37, 30}, {"shield", sdk.BroadcasterFeePublicFunding, 37, 30},
		{"transfer", "", 0, 30}, {"transfer", sdk.BroadcasterFeePrivateInputs, 0, 30},
		{"transfer", "", 20, 30}, {"transfer", sdk.BroadcasterFeePrivateInputs, 20, 30}, {"transfer", sdk.BroadcasterFeePublicFunding, 20, 30},
		{"transfer", sdk.BroadcasterFeePrivateInputs, 0, 93},
	}
	for _, tc := range cases {
		for _, replacement := range []int64{3, 7, 10} {
			if tc.output == 93 && replacement > 7 {
				continue
			}
			t.Run(fmt.Sprintf("%s_%s_%d_%d_%d", tc.kind, tc.source, tc.funding, tc.output, replacement), func(t *testing.T) {
				ctx := context.Background()
				s, token, _ := testSession(t, uint64(time.Now().Unix()))
				address, _ := s.wallet.GetAddress()
				outputs := testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(tc.output)}})
				fee := &sdk.Fee{Recipient: common.HexToAddress("0x1234"), Amount: big.NewInt(7)}
				var original *Operation
				var err error
				if tc.kind == "shield" {
					original, err = s.PrepareShield(ctx, sdk.ShieldPreparation{TokenAddress: token, Outputs: outputs, Fee: fee, Funding: sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(tc.funding), Authorization: "caller"}})
				} else {
					p := sdk.TransferPreparation{TokenAddress: token, Outputs: outputs, Fee: fee}
					if tc.funding > 0 {
						p.Funding = &sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(tc.funding), Authorization: "caller"}
					}
					original, err = s.PrepareTransfer(ctx, p)
				}
				if err != nil {
					t.Fatal(err)
				}
				before, err := frozen(original.plan)
				if err != nil {
					t.Fatal(err)
				}
				client, _ := testBroadcaster(t, original, []int64{replacement}, false)
				next, err := original, client.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(20), FeeSource: tc.source})
				if err != nil {
					t.Fatal(err)
				}
				_, previousFee := transaction.PlanFee(before)
				if next.Preview().FeeAmount.Int64() != replacement || previousFee.Int64() != 7 {
					t.Fatal("fee was added instead of replaced")
				}
				in := new(big.Int).Set(next.plan.Funding.Amount)
				for _, n := range next.plan.Inputs {
					in.Add(in, primitives.ElementToBigInt(n.Note.Amount))
				}
				out := big.NewInt(replacement)
				for _, n := range next.plan.Outputs {
					out.Add(out, n.Amount)
				}
				if in.Cmp(out) != 0 {
					t.Fatalf("replacement broke conservation: in=%s out=%s", in, out)
				}
				expectedFunding := tc.funding
				if tc.kind == "shield" || tc.source == sdk.BroadcasterFeePublicFunding {
					expectedFunding += replacement - 7
				}
				if next.plan.Funding.Amount.Int64() != expectedFunding {
					t.Fatalf("gross funding %s; wanted %d", next.plan.Funding.Amount, expectedFunding)
				}
				if before.Outputs[0].Commitment != next.plan.Outputs[0].Commitment {
					t.Fatal("business output changed")
				}
			})
		}
	}
}

func TestExistingFeePermitCallbackUsesFinalReplacementGross(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, err := s.PrepareShield(ctx, sdk.ShieldPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}}), Fee: &sdk.Fee{Recipient: common.HexToAddress("0x1234"), Amount: big.NewInt(7)}, Funding: sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(37), Authorization: "caller", Permit: &protocol.TokenPermit{Deadline: big.NewInt(time.Now().Unix() + 1000)}}})
	if err != nil {
		t.Fatal(err)
	}
	client, _ := testBroadcaster(t, original, []int64{5, 8, 8, 8}, false)
	calls := 0
	next, err := original, client.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10), FundingPermit: func(_ context.Context, r sdk.FundingPermitRequest) (*protocol.TokenPermit, error) {
		calls++
		if r.Amount.Int64() != 38 {
			t.Fatalf("permit gross=%s; wanted output30+replacement8", r.Amount)
		}
		return &protocol.TokenPermit{Deadline: big.NewInt(time.Now().Unix() + 1000)}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || next.Preview().FundingAmount.Int64() != 38 {
		t.Fatal("permit requested before final replacement fee", calls, next.Preview())
	}
}
