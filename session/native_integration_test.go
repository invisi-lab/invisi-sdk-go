package session

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

// Chain and index reads use controlled fixtures. Witness execution, proof
// generation and independent verification use the genuine released binaries.
func TestRealNativeOperationProofs(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine proving")
	}
	for _, kind := range []string{"transfer", "withdraw", "sweep", "recursive"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			session, token, _ := testSession(t)
			backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
			if err != nil {
				t.Fatal(err)
			}
			session.options.Prover = backend
			address, err := session.wallet.GetAddress()
			if err != nil {
				t.Fatal(err)
			}
			var prepared *Operation
			switch kind {
			case "transfer", "recursive":
				prepared, err = session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
			case "withdraw":
				prepared, err = session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Withdrawal: &sdk.Withdrawal{To: session.wallet.OwnerAddress(), Amount: big.NewInt(30)}})
			case "sweep":
				prepared, err = session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Withdrawal: &sdk.Withdrawal{To: session.wallet.OwnerAddress(), Amount: big.NewInt(100)}})

			}
			if err != nil {
				t.Fatal(err)
			}
			if prepared.Preview().NeedsWalletAuthorization {
				if prepared, err = authorizeForTest(prepared, ctx); err != nil {
					t.Fatal(err)
				}
			}
			phases := []string{}
			if _, err = prepared.Prove(ctx, sdk.TransactionProveOptions{Options: prover.Options{OnProgress: func(event prover.Progress) { phases = append(phases, event.Phase) }}, Execution: map[bool]string{true: "aggregate", false: "single"}[kind == "recursive"]}); err != nil {
				t.Fatal(err)
			}
			if len(phases) < 4 || phases[0] != "witness" || phases[len(phases)-1] != "complete" {
				t.Fatal(phases)
			}
		})
	}
}

func TestRealNativeAggregation(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine proving")
	}
	ctx := context.Background()
	backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
	if err != nil {
		t.Fatal(err)
	}
	members := []prover.AggregateMember{}
	for i := 0; i < 4; i++ {
		session, token, _ := testSession(t)
		session.options.Prover = backend
		state, err := session.options.Store.ReadSnapshot(ctx, session.scope)
		if err != nil {
			t.Fatal(err)
		}
		var input StoredNote
		for _, note := range state.Notes {
			input = note
		}
		input.Note.Nonce = core.ElementFromUint64(uint64(123 + i))
		identity, err := core.ComputeNoteIdentity(input.Note.OwnerAddress, input.Note.Nonce)
		if err != nil {
			t.Fatal(err)
		}
		leaf, err := core.ComputeNoteCommitment(input.Note.ChainID, input.Note.ProtocolAddress, identity, input.Note.TokenAddress, input.Note.Amount)
		if err != nil {
			t.Fatal(err)
		}
		nullifier, err := core.ComputeNoteNullifier(input.Note.ChainID, input.Note.ProtocolAddress, input.Note.Nonce, leaf)
		if err != nil {
			t.Fatal(err)
		}
		input.Commitment, input.Nullifier = transaction.FieldHash(leaf), transaction.FieldHash(nullifier)
		session.options.Provider.(sessionProvider).leaves.Store(uint64(1), input.Commitment)
		if err = session.options.Store.Update(ctx, session.scope, func(state *WalletState) error {
			state.Notes = map[common.Hash]StoredNote{input.Commitment: input}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		address, err := session.wallet.GetAddress()
		if err != nil {
			t.Fatal(err)
		}
		prepared, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
		if err != nil {
			t.Fatal(err)
		}
		if prepared, err = authorizeForTest(prepared, ctx); err != nil {
			t.Fatal(err)
		}
		if prepared, err = prepared.Prove(ctx, sdk.TransactionProveOptions{Execution: "aggregate"}); err != nil {
			t.Fatal(err)
		}
		member, err := prepared.ExportAggregateMember(ctx)
		if err != nil {
			t.Fatal(err)
		}
		members = append(members, prover.AggregateMember{Bundle: member.Bundle, Statement: member.Statement})
	}
	aggregator, err := prover.NewAggregator(backend)
	if err != nil {
		t.Fatal(err)
	}
	result, err := aggregator.Aggregate(ctx, members)
	if err != nil {
		t.Fatal(err)
	}
	if result.CircuitName != "transact_aggregate_mixed_1_5_4" || len(result.PublicInputs) != 4+4*32+5 {
		t.Fatal("incorrect aggregate statement")
	}
}
