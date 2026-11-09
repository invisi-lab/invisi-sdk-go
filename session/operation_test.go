package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

type sessionProvider struct {
	head   *types.Header
	leaves *sync.Map
}

func (p sessionProvider) ChainID(context.Context) (*big.Int, error) { return big.NewInt(31337), nil }

func (p sessionProvider) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	return p.head, nil
}

func (p sessionProvider) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return []byte{0x60, 0}, nil
}

func (p sessionProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return nil, nil
}

func (p sessionProvider) CallContract(_ context.Context, message ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	a, e := protocol.ProtocolABI()
	if e != nil {
		return nil, e
	}
	m, e := a.MethodById(message.Data[:4])
	if e != nil {
		registry, x := contracts.InvisiAddressRegistryMetaData.GetAbi()
		if x != nil {
			return nil, x
		}
		m, e = registry.MethodById(message.Data[:4])
		if e != nil {
			return nil, e
		}
	}
	switch m.RawName {
	case "VERSION":
		return m.Outputs.Pack(big.NewInt(1))
	case "rootHistory":
		return m.Outputs.Pack(true)
	case "addressRegistry":
		return m.Outputs.Pack(common.HexToAddress("0x3333333333333333333333333333333333333333"))
	case "isRegistered":
		return m.Outputs.Pack(false)
	case "nonceOf":
		return m.Outputs.Pack(uint64(0))
	case "isNullifierSpent":
		return m.Outputs.Pack(false)
	}
	return m.Outputs.Pack(big.NewInt(0))
}

type testProver struct {
	alter    bool
	requests []prover.ProvingRequest
}

func (p *testProver) Prove(_ context.Context, r prover.ProvingRequest, _ prover.Options) (prover.ProofBundle, error) {
	p.requests = append(p.requests, r)
	proof := make([]byte, 32)
	if r.Target == "noir-recursive" {
		proof = make([]byte, 458*32)
	}
	proof[31] = 1
	hash := sha256.Sum256(proof)
	inputs := append([]common.Hash(nil), r.ExpectedPublicInputs...)
	if p.alter {
		inputs[0] = common.Hash{}
	}
	return prover.ProofBundle{FormatVersion: 1, CircuitName: r.CircuitName, Target: r.Target, BBVersion: "6.0.0", ArtifactSHA256: r.ArtifactSHA256, VKSHA256: r.VKSHA256, Proof: proof, ProofSHA256: hex.EncodeToString(hash[:]), PublicInputs: inputs}, nil
}

func (p *testProver) Verify(context.Context, prover.ProofBundle, prover.ProofStatement, prover.Options) (bool, error) {
	return true, nil
}

func testSession(t *testing.T, timestamps ...uint64) (*WalletSession, common.Address, *testProver) {
	timestamp := uint64(100)
	if len(timestamps) > 0 {
		timestamp = timestamps[0]
	}
	t.Helper()
	ctx := context.Background()
	wallet, e := localTestWallet("0x" + strings.Repeat("0", 63) + "1")
	if e != nil {
		t.Fatal(e)
	}
	provider := sessionProvider{head: &types.Header{Number: big.NewInt(10), Time: timestamp, ParentHash: common.HexToHash("0x09")}, leaves: &sync.Map{}}
	address := common.HexToAddress("0x1111111111111111111111111111111111111111")
	token := common.HexToAddress("0x2222222222222222222222222222222222222222")
	deployment := protocolrpc.ProtocolDeployment{ChainID: "31337", Protocol: protocolrpc.ContractDeployment{Address: address, RuntimeCodeHash: transaction.HashCode([]byte{0x60, 0}), FromBlock: 1}}
	block := protocolrpc.CanonicalBlock{Number: 10, Hash: provider.head.Hash(), ParentHash: provider.head.ParentHash, Timestamp: timestamp}
	note := core.Note{ProtocolContext: core.ProtocolContext{ChainID: core.ElementFromUint64(31337), ProtocolAddress: core.BigIntToElement(address.Big())}, OwnerAddress: core.BigIntToElement(wallet.OwnerAddress().Big()), TokenAddress: core.BigIntToElement(token.Big()), Amount: core.ElementFromUint64(100), Nonce: core.ElementFromUint64(123)}
	identity, _ := core.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
	leaf, _ := core.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
	nf, _ := core.ComputeNoteNullifier(note.ChainID, note.ProtocolAddress, note.Nonce, leaf)
	root := leaf
	for range 16 {
		root, _ = core.Poseidon2Hash(root, core.ElementFromUint64(0))
	}
	provider.leaves.Store(uint64(1), transaction.FieldHash(leaf))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		json.NewDecoder(r.Body).Decode(&request)
		var result any
		switch request.Method {
		case "invisi_getDeployment":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment}
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
		case "invisi_getRoots", "invisi_getMembership":
			var tree uint64
			if e := json.Unmarshal(request.Params[0], &tree); e != nil {
				t.Error(e)
			}
			stored, _ := provider.leaves.Load(tree)
			commitment := stored.(common.Hash)
			value, _ := core.HexToElement(commitment.Hex())
			for range 16 {
				value, _ = core.Poseidon2Hash(value, core.ElementFromUint64(0))
			}
			currentAnchor := map[string]any{"treeId": tree, "root": transaction.FieldHash(value), "leafCount": 1, "block": block}
			if request.Method == "invisi_getRoots" {
				result = []any{currentAnchor}
			} else {
				result = map[string]any{"anchor": currentAnchor, "leafIndex": 0, "commitment": commitment, "siblings": make([]common.Hash, 16)}
			}
		default:
			t.Error("unexpected RPC", request.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	rpc, e := sdk.NewProtocolRpcClient(server.URL + "/v1")
	if e != nil {
		t.Fatal(e)
	}
	pool, e := sdk.ProtocolFromRPC(ctx, rpc, provider)
	if e != nil {
		t.Fatal(e)
	}
	store := NewMemoryWalletStore()
	backend := &testProver{}
	session, e := Connect(ctx, Options{Invisi: sdk.NewInvisi(sdk.InvisiOptions{ProtocolRPC: rpc, Provider: provider, Prover: backend}), Store: store, Wallet: wallet, Protocol: pool, Namespace: "tests"})
	if e != nil {
		t.Fatal(e)
	}
	if e = store.Update(ctx, session.scope, func(state *WalletState) error {
		state.Notes[transaction.FieldHash(leaf)] = StoredNote{Note: note, Commitment: transaction.FieldHash(leaf), Nullifier: transaction.FieldHash(nf), TreeID: 1, LeafIndex: 0}
		state.ObservedThrough = &block
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	return session, token, backend
}

type missingTokenCodeProvider struct {
	sdk.Provider
	token common.Address
}

func (p missingTokenCodeProvider) CodeAt(ctx context.Context, address common.Address, block *big.Int) ([]byte, error) {
	if address == p.token {
		return nil, nil
	}
	return p.Provider.CodeAt(ctx, address, block)
}

func TestShieldAcceptsTokenWithoutRegistration(t *testing.T) {
	session, _, _ := testSession(t)
	address, err := session.wallet.GetAddress()
	if err != nil {
		t.Fatal(err)
	}
	asset := common.HexToAddress("0x3333333333333333333333333333333333333333")
	output, err := sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: new(big.Int).SetUint64(session.scope.ChainID), ProtocolAddress: session.scope.ProtocolAddress, TokenAddress: asset, Amount: big.NewInt(70), Recipient: sdk.InvisiRecipient{Address: address}})
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := session.PrepareShield(context.Background(), sdk.ShieldPreparation{
		TokenAddress: asset,
		Outputs:      []sdk.NoteOutput{output},
		Funding:      sdk.FundingSource{Payer: session.wallet.OwnerAddress(), Amount: big.NewInt(70)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Preview().OutputCount != 1 {
		t.Fatal("shield output missing")
	}
}

func TestPreparationRejectsTokenWithoutContractCode(t *testing.T) {
	session, token, _ := testSession(t)
	session.options.Provider = missingTokenCodeProvider{Provider: session.options.Provider, token: token}
	address, err := session.wallet.GetAddress()
	if err != nil {
		t.Fatal(err)
	}
	_, err = session.PrepareTransfer(context.Background(), sdk.TransferPreparation{
		TokenAddress: token,
		Outputs:      testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}}),
	})
	if err == nil || !strings.Contains(err.Error(), "token address must have contract code") {
		t.Fatalf("expected invalid token rejection, got %v", err)
	}
}

func TestTransferStagesReserveInputsAndAssembleCurrentWitness(t *testing.T) {
	ctx := context.Background()
	session, token, backend := testSession(t)
	address, _ := session.wallet.GetAddress()
	prepared, e := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	preview := prepared.Preview()
	if preview.InputCount != 1 || preview.OutputCount != 2 || !preview.NeedsWalletAuthorization {
		t.Fatal(preview)
	}
	if _, e = prepared.Prove(ctx); e == nil {
		t.Fatal("unsigned private transfer was proved")
	}
	balance, e := session.Balance(ctx, token)
	if e != nil || balance.AvailableKnown.Sign() != 0 || balance.Reserved.Cmp(big.NewInt(100)) != 0 {
		t.Fatal(balance, e)
	}
	if _, e = session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(1)}})}); e == nil {
		t.Fatal("reserved note selected twice")
	}
	authorized, e := authorizeForTest(prepared, ctx)
	if e != nil {
		t.Fatal(e)
	}
	proved, e := authorized.Prove(ctx)
	if e != nil || proved.Preview().Status != "ready" {
		t.Fatal(e)
	}
	request := backend.requests[0]
	if e = request.Witness.WithInputs(func(inputs map[string]any) error {
		if inputs["input_owner_address"] != localTestMaterial(session.wallet).OwnerAddress.String() {
			t.Fatal("wrong witness identity")
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if e = proved.Abandon(ctx); e == nil {
		t.Fatal("externalized proof released reserved notes")
	}
}

func TestProofForAnotherStatementIsRejected(t *testing.T) {
	ctx := context.Background()
	session, token, backend := testSession(t)
	backend.alter = true
	prepared, e := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Withdrawal: &sdk.Withdrawal{To: session.wallet.OwnerAddress(), Amount: big.NewInt(30)}})
	if e != nil {
		t.Fatal(e)
	}
	if prepared, e = authorizeForTest(prepared, ctx); e != nil {
		t.Fatal(e)
	}
	if _, e = prepared.Prove(ctx); e == nil {
		t.Fatal("proof for another statement accepted")
	}
	state, _ := session.options.Store.ReadSnapshot(ctx, session.scope)
	for _, note := range state.Notes {
		if note.ReservedBy == "" {
			t.Fatal("failed backend released possibly externalized proof inputs")
		}
	}
}

func TestViewOnlySweepUsesFullOwnerPayoutWithoutSignature(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t)
	prepared, e := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Withdrawal: &sdk.Withdrawal{To: session.wallet.OwnerAddress(), Amount: big.NewInt(100)}})
	if e != nil {
		t.Fatal(e)
	}
	preview := prepared.Preview()
	if preview.NeedsWalletAuthorization || preview.OutputCount != 0 || preview.WithdrawalAddress != session.wallet.OwnerAddress() || preview.WithdrawalAmount.Cmp(big.NewInt(100)) != 0 {
		t.Fatal(preview)
	}
	if _, e = prepared.Prove(ctx); e != nil {
		t.Fatal(e)
	}
}

func TestRegistryUsesOwnerSignatureAndIndependentNonce(t *testing.T) {
	ctx := context.Background()
	session, _, backend := testSession(t)
	registry := testRegistry(t, session, false)
	prepared, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Preview().Nonce != 0 {
		t.Fatal("wrong registry nonce")
	}
	if _, err = prepared.ExportTransaction(ctx); err == nil {
		t.Fatal("unsigned registration exported")
	}
	ready, err := prepared.WithSignature(ctx, signRegistry(t, prepared, session.wallet))
	if err != nil {
		t.Fatal(err)
	}
	if ready.Status() != "ready" {
		t.Fatal("authorization did not prepare calldata")
	}
	if _, err = ready.ExportTransaction(ctx); err != nil {
		t.Fatal(err)
	}
	if len(backend.requests) != 0 {
		t.Fatal("registration requested a circuit proof")
	}
}
