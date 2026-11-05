package sdk

import (
	"context"
	"errors"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
)

type forbiddenStore struct{}

func (forbiddenStore) ReadSnapshot(context.Context, WalletScope) (WalletState, error) {
	panic("explicit preparation accessed store")
}
func (forbiddenStore) Update(context.Context, WalletScope, func(*WalletState) error) error {
	panic("explicit preparation accessed store")
}

func explicitFixture(t *testing.T) (*Invisi, ExplicitTransferPreparation, *testProver) {
	t.Helper()
	s, token, backend := testSession(t)
	state, err := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	if err != nil {
		t.Fatal(err)
	}
	inputs := []NoteInput{}
	for _, n := range state.Notes {
		inputs = append(inputs, NoteInput{Note: n.Note, Commitment: n.Commitment, TreeID: n.TreeID, LeafIndex: n.LeafIndex})
	}
	address, _ := s.wallet.GetAddress()
	opts := s.options.InvisiOptions
	opts.Store = forbiddenStore{}
	return NewInvisi(opts), ExplicitTransferPreparation{TransactionIdentity: TransactionIdentity{Wallet: s.wallet.PublicWallet, Protocol: s.protocol}, TransferPreparation: TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})}, ExplicitFunds: ExplicitFunds{Inputs: inputs}}, backend
}
func TestExplicitTransferNeverUsesStore(t *testing.T) {
	ctx := context.Background()
	invisi, input, backend := explicitFixture(t)
	prepared, err := invisi.PrepareTransfer(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	input.Outputs[0].Amount().SetInt64(99)
	if p := prepared.Preview(); p.InputCount != 1 || p.OutputCount != 2 || p.Outputs[0].Amount.Int64() != 30 {
		t.Fatal(p)
	}
	authorized, err := externallyAuthorize(ctx, prepared)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = externallyAuthorize(ctx, prepared); err == nil {
		t.Fatal("stale handle accepted")
	}
	proved, err := authorized.Prove(ctx)
	if err != nil {
		t.Fatal(err)
	}
	call, err := proved.ExportTransaction(ctx)
	if err != nil || len(call.Data) < 4 {
		t.Fatal(call, err)
	}
	call.Data[0] ^= 1
	again, err := proved.ExportTransaction(ctx)
	if err != nil || again.Data[0] == call.Data[0] {
		t.Fatal("export aliases internal calldata", err)
	}
	if len(backend.requests) != 1 {
		t.Fatal("proof missing")
	}
}
func TestExplicitRejectsDuplicateAndTamperedInputs(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Inputs = append(input.Inputs, input.Inputs[0])
	if _, err := invisi.PrepareTransfer(context.Background(), input); err == nil {
		t.Fatal("duplicate accepted")
	}
	input.Inputs = input.Inputs[:1]
	input.Inputs[0].Commitment = common.Hash{}
	if _, err := invisi.PrepareTransfer(context.Background(), input); err == nil {
		t.Fatal("tampered commitment accepted")
	}
}
func TestExplicitFailedProofCanRetry(t *testing.T) {
	ctx := context.Background()
	invisi, input, backend := explicitFixture(t)
	prepared, err := invisi.PrepareTransfer(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	authorized, err := externallyAuthorize(ctx, prepared)
	if err != nil {
		t.Fatal(err)
	}
	backend.alter = true
	if _, err = authorized.Prove(ctx); err == nil {
		t.Fatal("altered proof accepted")
	}
	backend.alter = false
	if _, err = authorized.Prove(ctx); err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = invisi.PrepareTransfer(canceled, input); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestExplicitConsumesSelectedInputsWithPublicFunding(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Funding = &FundingSource{Payer: input.Wallet.identity, Amount: big.NewInt(30), Authorization: "signature"}
	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if p := prepared.Preview(); p.InputCount != 1 || len(p.Outputs) != 2 || p.Outputs[1].Amount.Int64() != 100 {
		t.Fatal(p)
	}
}
func TestExplicitOtherFundsRoutes(t *testing.T) {
	ctx := context.Background()
	for _, kind := range []string{"withdraw", "sweep", "shield", "create-credit"} {
		t.Run(kind, func(t *testing.T) {
			invisi, input, _ := explicitFixture(t)
			var p *PreparedTransaction
			var err error
			switch kind {
			case "withdraw":
				p, err = invisi.PrepareTransfer(ctx, ExplicitTransferPreparation{TransactionIdentity: input.TransactionIdentity, ExplicitFunds: input.ExplicitFunds, TransferPreparation: TransferPreparation{TokenAddress: input.TokenAddress, Withdrawal: &Withdrawal{To: input.Wallet.identity, Amount: big.NewInt(30)}}})
			case "sweep":
				input.Wallet, err = PublicWalletFromAddress(input.Outputs[0].To())
				if err != nil {
					t.Fatal(err)
				}
				p, err = invisi.PrepareTransfer(ctx, ExplicitTransferPreparation{TransactionIdentity: input.TransactionIdentity, ExplicitFunds: input.ExplicitFunds, TransferPreparation: TransferPreparation{TokenAddress: input.TokenAddress, Withdrawal: &Withdrawal{To: input.Wallet.identity, Amount: big.NewInt(100)}}})
			case "shield":
				p, err = invisi.PrepareShield(ctx, ExplicitShieldPreparation{TransactionIdentity: input.TransactionIdentity, ShieldPreparation: ShieldPreparation{TokenAddress: input.TokenAddress, Outputs: input.Outputs, Funding: FundingSource{Payer: input.Wallet.identity, Amount: big.NewInt(30), Authorization: "signature"}}})
			case "create-credit":
				p, err = invisi.PrepareCreateCredit(ctx, ExplicitCreateCreditPreparation{TransactionIdentity: input.TransactionIdentity, ExplicitFunds: input.ExplicitFunds, CreateCreditPreparation: CreateCreditPreparation{TokenAddress: input.TokenAddress, Spender: input.Outputs[0].To(), Amount: big.NewInt(70), Expiry: big.NewInt(1000)}})
			}
			if err != nil {
				t.Fatal(err)
			}
			if kind == "shield" {
				if p.Preview().InputCount != 0 {
					t.Fatal("shield has inputs")
				}
				return
			}
			if p.Preview().NeedsWalletAuthorization {
				p, err = externallyAuthorize(ctx, p)
				if err != nil {
					t.Fatal(err)
				}
			}
			p, err = p.Prove(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.ExportTransaction(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestExplicitCreditLifecycle(t *testing.T) {
	ctx := context.Background()
	for _, kind := range []string{"draw", "cancel", "top-up"} {
		t.Run(kind, func(t *testing.T) {
			session, token, _ := testSession(t)
			leaf := seedCredit(t, session, token)
			state, err := session.options.Store.ReadSnapshot(ctx, session.scope)
			if err != nil {
				t.Fatal(err)
			}
			role := "owner"
			if kind == "draw" {
				role = "spender"
			}
			stored := state.Credits[creditKey(leaf, role)]
			credits := []CreditInput{{State: stored.State, Commitment: stored.Commitment, TreeID: stored.TreeID, LeafIndex: stored.LeafIndex, Role: role}}
			funds := ExplicitFunds{}
			for _, n := range state.Notes {
				funds.Inputs = append(funds.Inputs, NoteInput{Note: n.Note, Commitment: n.Commitment, TreeID: n.TreeID, LeafIndex: n.LeafIndex})
			}
			opts := session.options.InvisiOptions
			opts.Store = forbiddenStore{}
			invisi := NewInvisi(opts)
			identity := TransactionIdentity{Wallet: session.wallet.PublicWallet, Protocol: session.protocol}
			var p *PreparedTransaction
			switch kind {
			case "draw":
				p, err = invisi.PrepareDrawCredit(ctx, ExplicitDrawCreditPreparation{TransactionIdentity: identity, CreditInputs: credits, DrawCreditPreparation: DrawCreditPreparation{TokenAddress: token, Credits: []CreditDrawInput{{Credit: leaf, Amount: big.NewInt(30)}}}})
			case "cancel":
				p, err = invisi.PrepareCancelCredit(ctx, ExplicitCancelCreditPreparation{TransactionIdentity: identity, CreditInputs: credits, CancelCreditPreparation: CancelCreditPreparation{TokenAddress: token, Credit: leaf}})
			case "top-up":
				p, err = invisi.PrepareTopUpCredit(ctx, ExplicitTopUpCreditPreparation{TransactionIdentity: identity, ExplicitFunds: funds, CreditInputs: credits, TopUpCreditPreparation: TopUpCreditPreparation{TokenAddress: token, Credit: leaf, AdditionalAmount: big.NewInt(30), Expiry: big.NewInt(2000)}})
			}
			if err != nil {
				t.Fatal(err)
			}
			p, err = externallyAuthorize(ctx, p)
			if err != nil {
				t.Fatal(err)
			}
			p, err = p.Prove(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.ExportTransaction(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestExplicitSignedChangeCanUseAnotherOwner(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	other, err := WalletFromPrivateKey("0x" + strings.Repeat("0", 63) + "2")
	if err != nil {
		t.Fatal(err)
	}
	input.ChangeAddress, _ = other.GetAddress()
	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Preview().Outputs[1].To != input.ChangeAddress {
		t.Fatal("change recipient ignored")
	}
}

func TestExplicitRejectsWrongTokenOwnerAndPosition(t *testing.T) {
	for _, kind := range []string{"token", "owner", "position"} {
		t.Run(kind, func(t *testing.T) {
			invisi, input, _ := explicitFixture(t)
			switch kind {
			case "token":
				input.TokenAddress = common.HexToAddress("0x3333333333333333333333333333333333333333")
			case "owner":
				input.Inputs[0].Note.OwnerAddress = core.ElementFromUint64(7)
			case "position":
				input.Inputs[0].LeafIndex = 1
			}
			if _, err := invisi.PrepareTransfer(context.Background(), input); err == nil {
				t.Fatal("invalid explicit input accepted")
			}
		})
	}
}

func TestExplicitConcurrentProofIsFenced(t *testing.T) {
	ctx := context.Background()
	invisi, input, backend := explicitFixture(t)
	prepared, err := invisi.PrepareTransfer(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	authorized, err := externallyAuthorize(ctx, prepared)
	if err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	blocking := &blockingExplicitProver{testProver: backend, started: started, release: release}
	authorized.operation.session.options.Prover = blocking
	result := make(chan error, 1)
	go func() { _, err := authorized.Prove(ctx); result <- err }()
	<-started
	if _, err := authorized.Prove(ctx); err == nil {
		t.Fatal("concurrent proof accepted")
	}
	close(release)
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if _, err := authorized.ExportProof(ctx); err == nil {
		t.Fatal("stale proof handle accepted")
	}
}

type blockingExplicitProver struct {
	*testProver
	started, release chan struct{}
}

func (p *blockingExplicitProver) Prove(ctx context.Context, request prover.ProvingRequest, options prover.Options) (prover.ProofBundle, error) {
	close(p.started)
	select {
	case <-p.release:
	case <-ctx.Done():
		return prover.ProofBundle{}, ctx.Err()
	}
	return p.testProver.Prove(ctx, request, options)
}

func TestRealNativeExplicitTransfer(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine proving")
	}
	invisi, input, _ := explicitFixture(t)
	backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
	if err != nil {
		t.Fatal(err)
	}
	invisi.options.Prover = backend
	p, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	p, err = externallyAuthorize(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	p, err = p.Prove(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.ExportTransaction(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestExplicitCallerFundedShieldNeedsNoProof(t *testing.T) {
	ctx := context.Background()
	invisi, input, backend := explicitFixture(t)
	p, err := invisi.PrepareShield(ctx, ExplicitShieldPreparation{TransactionIdentity: input.TransactionIdentity, ShieldPreparation: ShieldPreparation{TokenAddress: input.TokenAddress, Outputs: input.Outputs, Funding: FundingSource{Payer: input.Wallet.identity, Amount: big.NewInt(30), Authorization: "caller"}}})
	if err != nil {
		t.Fatal(err)
	}
	ready, err := p.Ready(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ready.ExportTransaction(ctx); err != nil {
		t.Fatal(err)
	}
	if len(backend.requests) != 0 {
		t.Fatal("shield generated a private proof")
	}
}
func TestExplicitFundingCannotMutateFrozenStatement(t *testing.T) {
	ctx := context.Background()
	invisi, input, _ := explicitFixture(t)
	key, err := crypto.HexToECDSA(strings.Repeat("0", 63) + "1")
	if err != nil {
		t.Fatal(err)
	}
	p, err := invisi.PrepareShield(ctx, ExplicitShieldPreparation{TransactionIdentity: input.TransactionIdentity, ShieldPreparation: ShieldPreparation{TokenAddress: input.TokenAddress, Outputs: input.Outputs, Funding: FundingSource{Payer: input.Wallet.identity, Amount: big.NewInt(30)}}})
	if err != nil {
		t.Fatal(err)
	}
	request, err := p.ExportFundingAuthorization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	typed, err := request.TypedData()
	if err != nil {
		t.Fatal(err)
	}
	typed.Domain.VerifyingContract = "0x3333333333333333333333333333333333333333"
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := crypto.Sign(digest, key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.WithFundingSignature(ctx, signature); err == nil {
		t.Fatal("mutated funding statement accepted")
	}
}
