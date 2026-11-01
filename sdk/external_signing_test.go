package sdk

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
	"testing"
)

func signFixtureRequest(request core.SigningRequest) ([]byte, error) {
	raw, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var copied core.SigningRequest
	if err = json.Unmarshal(raw, &copied); err != nil {
		return nil, err
	}
	typed, err := copied.TypedData()
	if err != nil {
		return nil, err
	}
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		return nil, err
	}
	key, err := crypto.HexToECDSA("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		return nil, err
	}
	return crypto.Sign(digest, key)
}
func externallyAuthorize(ctx context.Context, p *PreparedTransaction) (*PreparedTransaction, error) {
	request, err := p.ExportWalletAuthorization(ctx)
	if err != nil {
		return nil, err
	}
	signature, err := signFixtureRequest(request)
	if err != nil {
		return nil, err
	}
	return p.WithWalletSignature(ctx, signature)
}

func TestCompactFundingSignatureIsCanonicalized(t *testing.T) {
	ctx := context.Background()
	invisi, input, _ := explicitFixture(t)
	prepared, err := invisi.PrepareShield(ctx, ExplicitShieldPreparation{TransactionIdentity: input.TransactionIdentity, ShieldPreparation: ShieldPreparation{TokenAddress: input.TokenAddress, Outputs: input.Outputs, Funding: FundingSource{Payer: input.Wallet.identity, Amount: big.NewInt(30)}}})
	if err != nil {
		t.Fatal(err)
	}
	request, err := prepared.ExportFundingAuthorization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := signFixtureRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	compact := append([]byte(nil), signature[:64]...)
	compact[32] |= signature[64] << 7
	ready, err := prepared.WithFundingSignature(ctx, compact)
	if err != nil {
		t.Fatal(err)
	}
	if len(ready.operation.plan.Funding.Signature) != 65 || ready.operation.plan.Funding.Signature[64] < 27 {
		t.Fatal("contract signature is not canonical")
	}
}

func TestProofExecutionSelectedAfterExternalAuthorization(t *testing.T) {
	for _, execution := range []string{"single", "aggregate"} {
		t.Run(execution, func(t *testing.T) {
			ctx := context.Background()
			invisi, input, _ := explicitFixture(t)
			prepared, err := invisi.PrepareTransfer(ctx, input)
			if err != nil {
				t.Fatal(err)
			}
			authorized, err := externallyAuthorize(ctx, prepared)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = authorized.Prove(ctx, TransactionProveOptions{Execution: "noir-recursive"}); err == nil {
				t.Fatal("backend target accepted as public execution")
			}
			ready, err := authorized.Prove(ctx, TransactionProveOptions{Execution: execution})
			if err != nil {
				t.Fatal(err)
			}
			proof, err := ready.ExportProof(ctx)
			if err != nil {
				t.Fatal(err)
			}
			expected := "evm"
			if execution == "aggregate" {
				expected = "noir-recursive"
			}
			if proof.Target != expected {
				t.Fatalf("proof target %s", proof.Target)
			}
			_, err = ready.ExportTransaction(ctx)
			if (err != nil) != (execution == "aggregate") {
				t.Fatal("incorrect direct export permission", err)
			}
		})
	}
}

func TestManagedExplicitCommitmentsConsumeEntireSelectedSubset(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t)
	state, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	var first StoredNote
	for _, n := range state.Notes {
		first = n
	}
	second := first
	second.TreeID = 2
	second.Note.Nonce = core.ElementFromUint64(124)
	identity, _ := core.ComputeNoteIdentity(second.Note.WalletIdentity, second.Note.Nonce)
	leaf, err := core.ComputeNoteCommitment(second.Note.ChainID, identity, second.Note.TokenAddress, second.Note.Amount)
	if err != nil {
		t.Fatal(err)
	}
	second.Commitment = fieldHash(leaf)
	second.Nullifier, err = session.wallet.ExportNullifier().ForNote(second.Note)
	if err != nil {
		t.Fatal(err)
	}
	session.options.Provider.(sessionProvider).leaves.Store(uint64(2), second.Commitment)
	if err = session.options.Store.Update(ctx, session.scope, func(state *WalletState) error { state.Notes[second.Commitment] = second; return nil }); err != nil {
		t.Fatal(err)
	}
	address, _ := session.wallet.GetAddress()
	prepared, err := session.PrepareTransfer(ctx, TransferPreparation{TokenAddress: token, Commitments: []common.Hash{first.Commitment, second.Commitment}, Outputs: []PaymentOutput{{To: address, Amount: big.NewInt(30)}}})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Preview().InputCount != 2 {
		t.Fatal("selected input omitted")
	}
}

func TestOwnerFullPayoutRemainsUnifiedTransfer(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Outputs = nil
	input.Withdrawal = &Withdrawal{To: input.Wallet.OwnerAddress(), Amount: big.NewInt(100)}
	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Kind() != "transfer" {
		t.Fatal("legacy operation kind remains", prepared.Kind())
	}
	if prepared.Preview().NeedsWalletAuthorization {
		t.Fatal("full owner payout must not require arbitrary spending authority")
	}
}

type changedFundingNonceProvider struct{ Provider }

func (p changedFundingNonceProvider) CallContract(ctx context.Context, message ethereum.CallMsg, block *big.Int) ([]byte, error) {
	abi, err := protocol.ProtocolABI()
	if err != nil {
		return nil, err
	}
	method, err := abi.MethodById(message.Data[:4])
	if err == nil && method.RawName == "nonces" {
		return method.Outputs.Pack(big.NewInt(1))
	}
	return p.Provider.CallContract(ctx, message, block)
}
func TestFundingRequestDoesNotSilentlyRefreshPreparedNonce(t *testing.T) {
	ctx := context.Background()
	invisi, input, _ := explicitFixture(t)
	prepared, err := invisi.PrepareShield(ctx, ExplicitShieldPreparation{TransactionIdentity: input.TransactionIdentity, ShieldPreparation: ShieldPreparation{TokenAddress: input.TokenAddress, Outputs: input.Outputs, Funding: FundingSource{Payer: input.Wallet.identity, Amount: big.NewInt(30)}}})
	if err != nil {
		t.Fatal(err)
	}
	prepared.operation.session.options.Provider = changedFundingNonceProvider{Provider: prepared.operation.session.options.Provider}
	if _, err = prepared.ExportFundingAuthorization(ctx); err == nil {
		t.Fatal("changed on-chain nonce silently replaced the reviewed funding statement")
	}
}
