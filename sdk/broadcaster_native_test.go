package sdk

import (
	"context"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"math/big"
	"os"
	"testing"
	"time"
)

// The relay and chain reads are controlled HTTP fixtures. Witness execution,
// nonzero fee balance constraints, proof generation and verification are native.
func TestRealNativeBroadcasterFeeProofs(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine proving")
	}
	for _, kind := range []string{"transfer", "create-credit", "top-up-credit", "cancel-credit"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t)
			backend, e := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
			if e != nil {
				t.Fatal(e)
			}
			s.options.Prover = backend
			address, _ := s.wallet.GetAddress()
			deadline := uint64(time.Now().Unix() + 300)
			var op *Operation
			switch kind {
			case "transfer":
				op, e = s.PrepareTransfer(ctx, TransferPreparation{TokenAddress: token, Deadline: deadline, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
			case "create-credit":
				op, e = s.PrepareCreateCredit(ctx, CreateCreditPreparation{TokenAddress: token, Deadline: deadline, Spender: address, Amount: big.NewInt(30), Expiry: big.NewInt(1000)})
			case "top-up-credit":
				leaf := seedCredit(t, s, token)
				op, e = s.PrepareTopUpCredit(ctx, TopUpCreditPreparation{TokenAddress: token, Deadline: deadline, Credit: leaf, AdditionalAmount: big.NewInt(30), Expiry: big.NewInt(2000)})
			case "cancel-credit":
				leaf := seedCredit(t, s, token)
				op, e = s.PrepareCancelCredit(ctx, CancelCreditPreparation{TokenAddress: token, Deadline: deadline, Credit: leaf})
			}
			if e != nil {
				t.Fatal(e)
			}
			client, _ := testBroadcaster(t, op, []int64{5}, false)
			op, e = op.WithBroadcaster(ctx, client, WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
			if e != nil {
				t.Fatal(e)
			}
			op, e = op.Authorize(ctx)
			if e != nil {
				t.Fatal(e)
			}
			op, e = op.Prove(ctx)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = op.SubmitBroadcaster(ctx, client); e != nil {
				t.Fatal(e)
			}
		})
	}
}
