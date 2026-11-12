package session

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestNotePublicFeeAccounting(t *testing.T) {
	for _, kind := range []string{"shield", "transfer"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t)
			address, _ := s.wallet.GetAddress()
			outputs := testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})
			fee := &sdk.Fee{Recipient: common.HexToAddress("0x1234"), Amount: big.NewInt(7)}
			var op *Operation
			var err error
			if kind == "shield" {
				op, err = s.PrepareShield(ctx, sdk.ShieldPreparation{TokenAddress: token, Outputs: outputs, Fee: fee, Funding: sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(37)}})
			} else {
				op, err = s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: outputs, Fee: fee})
			}
			if err != nil {
				t.Fatal(err)
			}
			if preview := op.Preview(); preview.FeeRecipient != fee.Recipient || preview.FeeAmount.Int64() != 7 {
				t.Fatal("public fee not exposed")
			}
			fee.Amount.SetInt64(90)
			if kind == "shield" {
				if op.plan.Shield.FeeAmount.Int64() != 7 || op.plan.Funding.Amount.Int64() != 37 {
					t.Fatal("shield fee or gross funding changed")
				}
			} else {
				inputTotal := new(big.Int)
				for _, input := range op.plan.Inputs {
					inputTotal.Add(inputTotal, primitives.ElementToBigInt(input.Note.Amount))
				}
				outputTotal := big.NewInt(7)
				for _, output := range op.plan.Outputs {
					outputTotal.Add(outputTotal, output.Amount)
				}
				if inputTotal.Cmp(outputTotal) != 0 {
					t.Fatal("private change failed to deduct public fee")
				}
			}
		})
	}
}
