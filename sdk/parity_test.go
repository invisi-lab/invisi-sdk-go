package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type failingLocalProver struct{ *testProver }

func (p failingLocalProver) Prove(context.Context, prover.ProvingRequest, prover.Options) (prover.ProofBundle, error) {
	return prover.ProofBundle{}, errors.New("simulated local proving failure")
}

func TestManagedStagesAreImmutableAndProofFailureIsRetryable(t *testing.T) {
	ctx := context.Background()
	session, token, backend := testSession(t)
	address, _ := session.wallet.GetAddress()
	prepared, err := session.PrepareTransfer(ctx, TransferPreparation{TokenAddress: token, Outputs: []PaymentOutput{{To: address, Amount: big.NewInt(30)}}})
	if err != nil {
		t.Fatal(err)
	}
	authorized, err := prepared.Authorize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if authorized == prepared {
		t.Fatal("stage reused the original handle")
	}
	if _, err = prepared.Prove(ctx); err == nil {
		t.Fatal("stale handle accepted")
	}
	authorized.session.options.Prover = failingLocalProver{backend}
	if _, err = authorized.Prove(ctx); err == nil {
		t.Fatal("failed prover accepted")
	}
	state, _ := session.options.Store.ReadSnapshot(ctx, session.scope)
	record := state.Operations[prepared.ID()]
	if record.Externalized || record.Status != "authorized" {
		t.Fatalf("local failure externalized operation: %+v", record)
	}
	recovered, err := session.Resume(ctx, prepared.ID())
	if err != nil {
		t.Fatal(err)
	}
	recovered, err = recovered.Authorize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = recovered.Prove(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestCreditPreviewIncludesActualPayoutAndInputs(t *testing.T) {
	session, token, _ := testSession(t)
	leaf := seedCredit(t, session, token)
	operation, err := session.PrepareDrawCredit(context.Background(), DrawCreditPreparation{TokenAddress: token, Credits: []CreditDrawInput{{Credit: leaf, Amount: big.NewInt(30)}}})
	if err != nil {
		t.Fatal(err)
	}
	preview := operation.Preview()
	if preview.Kind != "draw-credit" || preview.InputCount != 1 || preview.WithdrawalAmount.Cmp(big.NewInt(30)) != 0 || preview.WithdrawalAddress != session.wallet.identity {
		t.Fatalf("incorrect review: %+v", preview)
	}
}

func TestResumeRejectsModifiedCallAndProof(t *testing.T) {
	for _, change := range []string{"calldata", "proof", "owner"} {
		t.Run(change, func(t *testing.T) {
			ctx := context.Background()
			session, token, _ := testSession(t)
			address, _ := session.wallet.GetAddress()
			operation, err := session.PrepareTransfer(ctx, TransferPreparation{TokenAddress: token, Outputs: []PaymentOutput{{To: address, Amount: big.NewInt(30)}}})
			if err != nil {
				t.Fatal(err)
			}
			operation, err = operation.Authorize(ctx)
			if err != nil {
				t.Fatal(err)
			}
			operation, err = operation.Prove(ctx)
			if err != nil {
				t.Fatal(err)
			}
			err = session.options.Store.Update(ctx, session.scope, func(state *WalletState) error {
				record := state.Operations[operation.ID()]
				switch change {
				case "calldata":
					record.CallData = []byte{1, 2, 3}
				case "proof":
					var bundle prover.ProofBundle
					_ = json.Unmarshal(record.Proof, &bundle)
					bundle.CircuitName = "credit_cancel"
					record.Proof, _ = json.Marshal(bundle)
				case "owner":
					var plan operationPlan
					_ = json.Unmarshal(record.Plan, &plan)
					plan.Inputs[0].Note.WalletIdentity.SetUint64(123)
					record.Plan, _ = json.Marshal(plan)
				}
				state.Operations[operation.ID()] = record
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = session.Resume(ctx, operation.ID()); err == nil {
				t.Fatal("modified operation accepted")
			}
		})
	}
}

func TestRegistryStagesAndExportsFencePreviousHandles(t *testing.T) {
	ctx := context.Background()
	session, _, backend := testSession(t)
	deployment := session.protocol.Deployment()
	deployment.Registry = &protocolrpc.ContractDeployment{Address: common.HexToAddress("0x3333333333333333333333333333333333333333"), RuntimeCodeHash: hashCode([]byte{0x60, 0}), FromBlock: 1}
	pool, err := NewProtocol(deployment)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewAddressRegistry(pool, session.options.Provider)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := registry.PrepareRegistration(ctx, RegistrationPreparation{Wallet: session.wallet, Prover: backend})
	if err != nil {
		t.Fatal(err)
	}
	authorized, err := prepared.Authorize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.Authorize(ctx); err == nil {
		t.Fatal("stale registration authorized twice")
	}
	ready, err := authorized.Prove(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ready.ExportTransaction(ctx); err != nil {
		t.Fatal(err)
	}
	tracked, err := ready.TrackTransaction(ctx, common.Hash{1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ready.ExportTransaction(ctx); err == nil {
		t.Fatal("stale registration exported")
	}
	current, err := tracked.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if current == tracked || current.Status() != "uncertain" {
		t.Fatal("registry reconciliation reused mutable handle")
	}
	if _, err = tracked.TrackTransaction(ctx, common.Hash{2}); err == nil {
		t.Fatal("stale registry tracked twice")
	}
}
