package session

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

type preparationLatencyProvider struct {
	sdk.Provider
	delay  time.Duration
	calls  atomic.Int32
	active atomic.Int32
	peak   atomic.Int32
}

func (p *preparationLatencyProvider) wait(ctx context.Context) error {
	p.calls.Add(1)
	n := p.active.Add(1)
	defer p.active.Add(-1)
	for old := p.peak.Load(); n > old && !p.peak.CompareAndSwap(old, n); old = p.peak.Load() {
	}
	timer := time.NewTimer(p.delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func (p *preparationLatencyProvider) ChainID(ctx context.Context) (*big.Int, error) {
	if e := p.wait(ctx); e != nil {
		return nil, e
	}
	return p.Provider.ChainID(ctx)
}
func (p *preparationLatencyProvider) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	if e := p.wait(ctx); e != nil {
		return nil, e
	}
	return p.Provider.HeaderByNumber(ctx, n)
}
func (p *preparationLatencyProvider) CodeAt(ctx context.Context, a common.Address, n *big.Int) ([]byte, error) {
	if e := p.wait(ctx); e != nil {
		return nil, e
	}
	return p.Provider.CodeAt(ctx, a, n)
}
func (p *preparationLatencyProvider) CallContract(ctx context.Context, m ethereum.CallMsg, n *big.Int) ([]byte, error) {
	if e := p.wait(ctx); e != nil {
		return nil, e
	}
	return p.Provider.CallContract(ctx, m, n)
}

// Opt-in synthetic diagnostics: genuine local signing, stub proof, no broadcast.
func TestPreparationLatency10x10(t *testing.T) {
	if os.Getenv("INVISI_PREPARATION_PROFILE") == "" {
		t.Skip("opt-in preparation latency profile")
	}
	for round := 0; round < 3; round++ {
		ctx := context.Background()
		s, token, backend := testSession(t)
		provider := s.options.Provider.(sessionProvider)
		state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
		if e != nil {
			t.Fatal(e)
		}
		var original StoredNote
		for _, n := range state.Notes {
			original = n
			break
		}
		inputs := make([]sdk.NoteInput, 10)
		for i := range inputs {
			note := original.Note
			note.Nonce = primitives.ElementFromUint64(uint64(1000 + i))
			note.Amount = primitives.ElementFromUint64(100)
			identity, _ := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
			leaf, _ := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
			hash := transaction.FieldHash(leaf)
			tree := uint64(i + 1)
			provider.leaves.Store(tree, hash)
			inputs[i] = sdk.NoteInput{Note: note, Commitment: hash, TreeID: tree, LeafIndex: 0}
		}
		delayed := &preparationLatencyProvider{Provider: provider, delay: 100 * time.Millisecond}
		invisi := sdk.NewInvisi(sdk.InvisiOptions{ProtocolRPC: s.options.ProtocolRPC, Provider: delayed, Prover: backend})
		phases := map[string]int64{}
		mark := func(name string, start time.Time) { phases[name] = time.Since(start).Milliseconds() }
		outputs := make([]sdk.NoteOutput, 9)
		address, _ := s.wallet.GetAddress()
		start := time.Now()
		for i := range outputs {
			out, err := sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: big.NewInt(31337), ProtocolAddress: s.scope.ProtocolAddress, TokenAddress: token, Amount: big.NewInt(200), Recipient: sdk.InvisiRecipient{Address: address}})
			if err != nil {
				t.Fatal(err)
			}
			outputs[i] = out
		}
		mark("outputs", start)
		start = time.Now()
		prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{TransactionIdentity: sdk.TransactionIdentity{Wallet: s.wallet, Protocol: s.protocol}, ExplicitFunds: sdk.ExplicitFunds{Inputs: inputs}, TransferPreparation: sdk.TransferPreparation{TokenAddress: token, Outputs: outputs, Funding: &sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(1000), Authorization: "signature"}}})
		if err != nil {
			t.Fatal(err)
		}
		mark("prepare", start)
		if prepared.Preview().InputCount != 10 || prepared.Preview().OutputCount != 10 {
			t.Fatal("profile shape changed")
		}
		start = time.Now()
		_, err = prepared.EstimateGas()
		if err != nil {
			t.Fatal(err)
		}
		mark("estimate", start)
		start = time.Now()
		request, err := prepared.ExportWalletAuthorization(ctx)
		if err != nil {
			t.Fatal(err)
		}
		mark("wallet_export", start)
		start = time.Now()
		signature, err := signFixtureRequest(request)
		if err != nil {
			t.Fatal(err)
		}
		mark("wallet_sign", start)
		start = time.Now()
		authorized, err := prepared.WithWalletSignature(ctx, signature)
		if err != nil {
			t.Fatal(err)
		}
		mark("wallet_attach", start)
		start = time.Now()
		proved, err := authorized.Prove(ctx)
		if err != nil {
			t.Fatal(err)
		}
		mark("prove_stub", start)
		start = time.Now()
		funding, err := proved.ExportFundingAuthorization(ctx)
		if err != nil {
			t.Fatal(err)
		}
		mark("funding_export", start)
		start = time.Now()
		signature, err = signFixtureRequest(funding)
		if err != nil {
			t.Fatal(err)
		}
		mark("funding_sign", start)
		start = time.Now()
		ready, err := proved.WithFundingSignature(ctx, signature)
		if err != nil {
			t.Fatal(err)
		}
		mark("funding_attach", start)
		start = time.Now()
		_, err = ready.ExportTransaction(ctx)
		if err != nil {
			t.Fatal(err)
		}
		mark("transaction_export", start)
		phases["provider_calls"] = int64(delayed.calls.Load())
		phases["peak_concurrency"] = int64(delayed.peak.Load())
		raw, _ := json.Marshal(phases)
		t.Logf("PROFILE %s", raw)
	}
}
