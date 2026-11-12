package session

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

type offlineProvider struct {
	sdk.Provider
	disabled atomic.Bool
	calls    atomic.Int32
}

func (p *offlineProvider) check() error {
	p.calls.Add(1)
	if p.disabled.Load() {
		return errors.New("network disabled")
	}
	return nil
}
func (p *offlineProvider) ChainID(c context.Context) (*big.Int, error) {
	if e := p.check(); e != nil {
		return nil, e
	}
	return p.Provider.ChainID(c)
}
func (p *offlineProvider) HeaderByNumber(c context.Context, n *big.Int) (*types.Header, error) {
	if e := p.check(); e != nil {
		return nil, e
	}
	return p.Provider.HeaderByNumber(c, n)
}
func (p *offlineProvider) CodeAt(c context.Context, a common.Address, n *big.Int) ([]byte, error) {
	if e := p.check(); e != nil {
		return nil, e
	}
	return p.Provider.CodeAt(c, a, n)
}
func (p *offlineProvider) CallContract(c context.Context, m ethereum.CallMsg, n *big.Int) ([]byte, error) {
	if e := p.check(); e != nil {
		return nil, e
	}
	return p.Provider.CallContract(c, m, n)
}
func TestSynchronizedPreparationReopensOfflineAndChecksBeforeSending(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, 2000000000)
	transport := &offlineProvider{Provider: s.options.Provider}
	s.options.Provider = transport
	s.rpc = s.rpc.WithProvider(transport)
	if e := s.RefreshState(ctx); e != nil {
		t.Fatal(e)
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	if state.PreparationSnapshot == nil {
		t.Fatal("sync did not retain preparation data")
	}
	transport.disabled.Store(true)
	transport.calls.Store(0)
	deadRPC, _ := sdk.NewProtocolRpcClient("http://127.0.0.1:1/v1")
	restored, e := Connect(ctx, Options{Invisi: sdk.NewInvisi(sdk.InvisiOptions{Provider: transport, ProtocolRPC: deadRPC, Prover: &testProver{}}), Wallet: s.wallet, Protocol: s.protocol, Store: s.options.Store, Namespace: s.scope.Namespace, PublicHistory: s.publicHistory})
	if e != nil {
		t.Fatal(e)
	}
	address, _ := s.wallet.GetAddress()
	op, e := restored.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(restored, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	if op.Preview().SubmissionValidation != "required" {
		t.Fatal("missing pending validation marker")
	}
	request, e := op.ExportWalletAuthorization(ctx)
	if e != nil {
		t.Fatal(e)
	}
	sig, e := signFixtureRequest(request)
	if e != nil {
		t.Fatal(e)
	}
	authorized, e := op.WithWalletSignature(ctx, sig)
	if e != nil {
		t.Fatal(e)
	}
	ready, e := authorized.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ready.ExportTransaction(ctx); e != nil {
		t.Fatal(e)
	}
	if n := transport.calls.Load(); n != 0 {
		t.Fatalf("offline path attempted %d reads", n)
	}
	if e = ready.ValidateSubmission(ctx); e == nil {
		t.Fatal("offline operation bypassed live submission gate")
	}
}

// Old owned inventory also needs a path when a refresh fixture only emits new deliveries.
func retainedNoteFixture(base sessionProvider, block protocolrpc.CanonicalBlock, method string, params []json.RawMessage) any {
	var tree uint64
	_ = json.Unmarshal(params[0], &tree)
	raw, _ := base.leaves.Load(tree)
	commitment := raw.(common.Hash)
	root, _ := primitives.HexToElement(commitment.Hex())
	for range 16 {
		root, _ = primitives.Poseidon2Hash(root, primitives.ElementFromUint64(0))
	}
	anchor := map[string]any{"treeId": tree, "root": common.Hash(root.Bytes()), "leafCount": 1, "block": block}
	if method == "invisi_getRoots" {
		return []any{anchor}
	}
	return map[string]any{"anchor": anchor, "leafIndex": 0, "commitment": commitment, "siblings": make([]common.Hash, 16)}
}

type submissionProvider struct {
	sessionProvider
	spent, unknownRoot        bool
	nonce, balance, allowance *big.Int
}

func (p *submissionProvider) CallContract(ctx context.Context, m ethereum.CallMsg, b *big.Int) ([]byte, error) {
	a, _ := protocol.ProtocolABI()
	if method, e := a.MethodById(m.Data[:4]); e == nil {
		switch method.RawName {
		case "isNullifierSpent":
			return method.Outputs.Pack(p.spent)
		case "rootHistory":
			return method.Outputs.Pack(!p.unknownRoot)
		case "nonces":
			v := new(big.Int)
			if p.nonce != nil {
				v = p.nonce
			}
			return method.Outputs.Pack(v)
		}
	}
	a, _ = contracts.IERC20MetaData.GetAbi()
	if method, e := a.MethodById(m.Data[:4]); e == nil {
		v := big.NewInt(1000)
		switch method.RawName {
		case "balanceOf":
			if p.balance != nil {
				v = p.balance
			}
			return method.Outputs.Pack(v)
		case "allowance":
			if p.allowance != nil {
				v = p.allowance
			}
			return method.Outputs.Pack(v)
		}
	}
	return p.sessionProvider.CallContract(ctx, m, b)
}
func TestOfflineNoteSubmissionRequiresCanonicalAnchorAndLiveInputs(t *testing.T) {
	for _, kind := range []string{"spent", "root", "anchor", "expiry"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t, 2000000000)
			live := &submissionProvider{sessionProvider: s.options.Provider.(sessionProvider)}
			s.options.Provider = live
			s.rpc = s.rpc.WithProvider(live)
			if e := s.RefreshState(ctx); e != nil {
				t.Fatal(e)
			}
			address, _ := s.wallet.GetAddress()
			op, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
			if e != nil {
				t.Fatal(e)
			}
			if e = op.ValidateSubmission(ctx); e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "spent":
				live.spent = true
			case "root":
				live.unknownRoot = true
			case "anchor":
				live.head = types.CopyHeader(live.head)
				live.head.Extra = []byte("reorg")
			case "expiry":
				live.head = types.CopyHeader(live.head)
				live.head.Time += 400
			}
			if e = op.ValidateSubmission(ctx); e == nil {
				t.Fatal("accepted stale " + kind)
			}
		})
	}
}
func TestOfflineShieldFundingAndLiveBalanceAllowance(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, 2000000000)
	live := &submissionProvider{sessionProvider: s.options.Provider.(sessionProvider)}
	transport := &offlineProvider{Provider: live}
	s.options.Provider = transport
	s.rpc = s.rpc.WithProvider(transport)
	if e := s.RefreshState(ctx); e != nil {
		t.Fatal(e)
	}
	address, _ := s.wallet.GetAddress()
	outputs := testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})
	transport.disabled.Store(true)
	transport.calls.Store(0)
	op, e := s.PrepareShield(ctx, sdk.ShieldPreparation{TokenAddress: token, Outputs: outputs, Funding: sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(30)}})
	if e != nil {
		t.Fatal(e)
	}
	request, e := op.ExportFundingAuthorization(ctx)
	if e != nil {
		t.Fatal(e)
	}
	sig, e := signFixtureRequest(request)
	if e != nil {
		t.Fatal(e)
	}
	ready, e := op.WithFundingSignature(ctx, sig)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ready.ExportTransaction(ctx); e != nil {
		t.Fatal(e)
	}
	if n := transport.calls.Load(); n != 0 {
		t.Fatalf("offline Shield attempted %d reads", n)
	}
	transport.disabled.Store(false)
	if e = ready.ValidateSubmission(ctx); e != nil {
		t.Fatal(e)
	}
	for _, field := range []string{"nonce", "balance", "allowance"} {
		live.nonce = nil
		live.balance = nil
		live.allowance = nil
		switch field {
		case "nonce":
			live.nonce = big.NewInt(1)
		case "balance":
			live.balance = big.NewInt(0)
		case "allowance":
			live.allowance = big.NewInt(0)
		}
		if e = ready.ValidateSubmission(ctx); e == nil {
			t.Fatal("accepted changed funding " + field)
		}
	}
}

func TestPreparationSnapshotEncryptedRoundtripAndLiveFallback(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t)
	if e := s.RefreshState(ctx); e != nil {
		t.Fatal(e)
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	codec, e := NewEncryptedWalletCodec(make([]byte, 32))
	if e != nil {
		t.Fatal(e)
	}
	sealed, e := codec.Seal(s.scope, state)
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := codec.Open(s.scope, sealed)
	if e != nil {
		t.Fatal(e)
	}
	snap := decoded.PreparationSnapshot
	if snap == nil || !snap.Match(s.scope, s.protocol.Deployment()) || len(snap.Memberships) != 1 {
		t.Fatal("encrypted snapshot lost its authenticated facts")
	}
	// Changing construction time must never rewrite the canonical historical header.
	address, _ := s.wallet.GetAddress()
	op, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	if op.plan.Block.Timestamp != 100 || op.plan.Context.Now.Uint64() <= 100 || op.plan.Deadline <= 100 {
		t.Fatal("construction time replaced the canonical anchor")
	}
	if e = op.Abandon(ctx); e != nil {
		t.Fatal(e)
	}
	if e = s.options.Store.Update(ctx, s.scope, func(v *WalletState) error { v.PreparationSnapshot.Code = map[common.Address][]byte{}; return nil }); e != nil {
		t.Fatal(e)
	}
	fallback, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	if fallback.Preview().SubmissionValidation != "" {
		t.Fatal("incomplete facts did not fall back online")
	}
}
func TestRefreshReusesAuthenticatedImmutablePaths(t *testing.T) {
	ctx := context.Background()
	s, _, _ := testSession(t)
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	var note StoredNote
	for _, v := range state.Notes {
		note = v
		break
	}
	head := s.options.Provider.(sessionProvider).head
	block := protocolrpc.CanonicalBlock{Number: 10, Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	m, e := s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: note.TreeID, LeafIndex: note.LeafIndex, Commitment: note.Commitment}, block)
	if e != nil {
		t.Fatal(e)
	}
	s.publicHistory.preparationCapture.record(m)
	original := s.rpc
	dead, _ := sdk.NewProtocolRpcClient("http://127.0.0.1:1/v1")
	s.rpc = dead.WithProvider(s.options.Provider)
	if e = s.RefreshState(ctx); e != nil {
		t.Fatal("normal sync proof was fetched again", e)
	}
	s.publicHistory.preparationCapture = &preparationCapture{paths: map[common.Hash]protocolrpc.Membership{}}
	if e = s.RefreshState(ctx); e != nil {
		t.Fatal("unchanged immutable path was fetched again", e)
	}
	// A stale root cannot be reused merely because its path is still stored.
	s.rpc = original
	live := &submissionProvider{sessionProvider: s.options.Provider.(sessionProvider), unknownRoot: true}
	live.head = types.CopyHeader(live.head)
	live.head.Extra = []byte("new canonical head")
	s.options.Provider = live
	s.rpc = s.rpc.WithProvider(live)
	if e = s.RefreshState(ctx); e == nil {
		t.Fatal("accepted an expired membership root")
	}
}

func TestOfflineReopenedSessionRetriesMissingPreparationFactsOnline(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t)
	if e := s.RefreshState(ctx); e != nil {
		t.Fatal(e)
	}
	if e := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		v.PreparationSnapshot.Memberships = map[common.Hash]protocolrpc.Membership{}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	restored, e := Connect(ctx, Options{Invisi: sdk.NewInvisi(s.options.InvisiOptions), Wallet: s.wallet, Protocol: s.protocol, Store: s.options.Store, Namespace: s.scope.Namespace, PublicHistory: s.publicHistory})
	if e != nil {
		t.Fatal(e)
	}
	address, _ := s.wallet.GetAddress()
	op, e := restored.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(restored, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	if op.Preview().SubmissionValidation != "" {
		t.Fatal("reopened session reused incomplete facts instead of live fallback")
	}
}

func TestSnapshotRootLeaderIncludesNewHigherLeaf(t *testing.T) {
	ctx := context.Background()
	s, _, _ := testSession(t)
	if err := s.RefreshState(ctx); err != nil {
		t.Fatal(err)
	}
	state, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	var first StoredNote
	for _, n := range state.Notes {
		first = n
		break
	}
	next := first
	next.Note.Nonce = primitives.ElementFromUint64(124)
	next.LeafIndex = 1
	identity, _ := primitives.ComputeNoteIdentity(next.Note.OwnerAddress, next.Note.Nonce)
	leaf, _ := primitives.ComputeNoteCommitment(next.Note.ChainID, next.Note.ProtocolAddress, identity, next.Note.TokenAddress, next.Note.Amount)
	next.Commitment = common.Hash(leaf.Bytes())
	next.Nullifier, _ = s.wallet.ExportNullifier().ForNote(next.Note)
	root, _ := primitives.Poseidon2Hash(primitives.BigIntToElement(first.Commitment.Big()), leaf)
	for range 15 {
		root, _ = primitives.Poseidon2Hash(root, primitives.ElementFromUint64(0))
	}
	rootHash := common.Hash(root.Bytes())
	block := state.PreparationSnapshot.Block
	oldRoot := state.PreparationSnapshot.Memberships[first.Commitment].Root
	path := make([]common.Hash, 16)
	path[0] = first.Commitment
	s.publicHistory.preparationCapture.record(protocolrpc.Membership{TreeID: 1, LeafIndex: 1, Leaf: next.Commitment, Root: rootHash, Path: path, Block: block})
	deployment := s.protocol.Deployment()
	oldAnchor := map[string]any{"treeId": 1, "root": oldRoot, "leafCount": 1, "block": block}
	anchor := map[string]any{"treeId": 1, "root": rootHash, "leafCount": 2, "block": block}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		var result any
		switch req.Method {
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": block}
		case "invisi_getRoots":
			result = []any{oldAnchor, anchor}
		case "invisi_getMembership":
			var index uint32
			_ = json.Unmarshal(req.Params[1], &index)
			proof := make([]common.Hash, 16)
			commitment := first.Commitment
			proof[0] = next.Commitment
			if index == 1 {
				commitment = next.Commitment
				proof[0] = first.Commitment
			}
			result = map[string]any{"anchor": anchor, "leafIndex": index, "commitment": commitment, "siblings": proof}
		default:
			t.Error("unexpected snapshot RPC", req.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	client, err := sdk.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &deployment})
	if err != nil {
		t.Fatal(err)
	}
	s.rpc = client.WithProvider(s.options.Provider)
	for range 12 {
		if err = s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
			v.Notes[next.Commitment] = next
			v.PreparationSnapshot = state.PreparationSnapshot
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err = s.RefreshState(ctx); err != nil {
			t.Fatal("older root led before the higher leaf", err)
		}
		got, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
		for _, n := range []StoredNote{first, next} {
			if got.PreparationSnapshot.Memberships[n.Commitment].Root != rootHash {
				t.Fatal("tree paths do not share the root containing every leaf")
			}
		}
	}
}
