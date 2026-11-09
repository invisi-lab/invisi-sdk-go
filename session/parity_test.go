package session

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

type failingLocalProver struct{ *testProver }

func (p failingLocalProver) Prove(context.Context, prover.ProvingRequest, prover.Options) (prover.ProofBundle, error) {
	return prover.ProofBundle{}, errors.New("simulated local proving failure")
}

func TestManagedStagesAreImmutableAndProofFailureIsRetryable(t *testing.T) {
	ctx := context.Background()
	session, token, backend := testSession(t)
	address, _ := session.wallet.GetAddress()
	prepared, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
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

func TestResumeRejectsModifiedCallAndProof(t *testing.T) {
	for _, change := range []string{"calldata", "proof", "owner"} {
		t.Run(change, func(t *testing.T) {
			ctx := context.Background()
			session, token, _ := testSession(t)
			address, _ := session.wallet.GetAddress()
			operation, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
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
					var plan transaction.Plan
					_ = json.Unmarshal(record.Plan, &plan)
					plan.Inputs[0].Note.OwnerAddress.SetUint64(123)
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
	session, _, _ := testSession(t)
	deployment := session.protocol.Deployment()
	deployment.Registry = &protocolrpc.ContractDeployment{Address: common.HexToAddress("0x3333333333333333333333333333333333333333"), RuntimeCodeHash: transaction.HashCode([]byte{0x60, 0}), FromBlock: 1}
	pool, err := sdk.NewProtocol(deployment)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := sdk.NewAddressRegistry(pool, fundingProvider{sessionProvider: session.options.Provider.(sessionProvider), payer: session.wallet.OwnerAddress()})
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
	if err != nil {
		t.Fatal(err)
	}
	signature := signRegistry(t, prepared, session.wallet)
	ready, err := prepared.WithSignature(ctx, signature)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.WithSignature(ctx, signature); err == nil {
		t.Fatal("stale registration authorized twice")
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
