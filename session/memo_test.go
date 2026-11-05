package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestMemoPreparationCopiesAndPersists(t *testing.T) {
	session, token, _ := testSession(t)
	address, _ := session.wallet.GetAddress()
	memo := []byte{0xff, 0, 1}
	prepared, err := session.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30), Payload: memo}})})
	if err != nil {
		t.Fatal(err)
	}
	memo[0] = 0
	if !bytes.Equal(prepared.plan.Outputs[0].Note.Payload, []byte{0xff, 0, 1}) {
		t.Fatal("caller memo mutation changed prepared note")
	}
	resumed, err := session.Resume(context.Background(), prepared.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(resumed.plan.Outputs[0].Note.Payload, []byte{0xff, 0, 1}) {
		t.Fatal("memo lost on resume")
	}
}
func TestCreditPayloadsDoNotInheritConsumedMetadata(t *testing.T) {
	for _, kind := range []string{"draw", "topup"} {
		for _, override := range [][]byte{nil, {}, []byte("changed")} {
			t.Run(kind+string(override), func(t *testing.T) {
				session, token, _ := testSession(t)
				leaf := seedCredit(t, session, token)
				ctx := context.Background()
				if err := session.options.Store.Update(ctx, session.scope, func(state *WalletState) error {
					for key, credit := range state.Credits {
						credit.State.Note.Payload = []byte("parent")
						state.Credits[key] = credit
					}
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				var operation *Operation
				var err error
				if kind == "draw" {
					operation, err = session.PrepareDrawCredit(ctx, sdk.DrawCreditPreparation{TokenAddress: token, Credits: []sdk.CreditDrawInput{{Credit: leaf, Amount: big.NewInt(30), Payload: override}}})
				} else {
					operation, err = session.PrepareTopUpCredit(ctx, sdk.TopUpCreditPreparation{TokenAddress: token, Credit: leaf, AdditionalAmount: big.NewInt(10), Payload: override})
				}
				if err != nil {
					t.Fatal(err)
				}
				expected := override
				for i, role := range []core.CreditDeliveryRole{core.CreditOwner, core.CreditSpender} {
					commitment := leaf
					if kind == "draw" {
						commitment = operation.plan.DrawCredit.RemainderCommitments[0]
					} else {
						commitment = operation.plan.CreateCredit.CreditCommitment
					}
					decoded, err := session.viewing.DecryptCredit(operation.plan.Deliveries[i].Ciphertext, core.CreditNoteDecryptionContext{NoteDecryptionContext: core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: session.coreContext(), OutputCommitment: core.BigIntToElement(commitment.Big())}, TokenAddress: core.BigIntToElement(token.Big()), OwnerAddress: transaction.PrivateMaterial(session.wallet).OwnerAddress}, Role: role})
					if err != nil || !bytes.Equal(decoded.Note.Payload, expected) {
						t.Fatal("memo inheritance", err, decoded.Note.Payload, expected)
					}
				}
			})
		}
	}
}
func TestCreditCreationAndDisclosureMemoBoundary(t *testing.T) {
	session, token, _ := testSession(t)
	address, _ := session.wallet.GetAddress()
	memo := []byte("invoice")
	operation, err := session.PrepareCreateCredit(context.Background(), sdk.CreateCreditPreparation{TokenAddress: token, Spender: address, Amount: big.NewInt(70), Expiry: big.NewInt(1000), Payload: memo})
	if err != nil {
		t.Fatal(err)
	}
	memo[0] = 'x'
	note := operation.plan.CreditOutput.Note
	if string(note.Payload) != "invoice" {
		t.Fatal("credit memo alias")
	}
	opening, err := sdk.DisclosureFromCredit(note)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(opening.Export())
	if bytes.Contains(raw, []byte("invoice")) {
		t.Fatal("opening leaked memo")
	}
	output := sdk.DisclosureOutput{ProtocolContext: note.ProtocolContext, TokenAddress: note.TokenAddress, OutputCommitment: operation.plan.CreateCredit.CreditCommitment, Role: "owner", Ciphertext: operation.plan.Deliveries[0].Ciphertext}
	result, err := opening.Verify(output)
	if err != nil || len(result.Credit.Payload) != 0 {
		t.Fatal("opening claimed memo", err)
	}
	payload, err := session.viewing.DiscloseCredit(output.Ciphertext, core.CreditNoteDecryptionContext{NoteDecryptionContext: core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: core.BigIntToElement(output.OutputCommitment.Big())}, TokenAddress: note.TokenAddress, OwnerAddress: note.OwnerWalletIdentity}, Role: core.CreditOwner})
	if err != nil {
		t.Fatal(err)
	}
	result, err = payload.Verify(output)
	if err != nil || string(result.Credit.Payload) != "invoice" {
		t.Fatal("payload memo lost", err)
	}
}

func TestMemoStoreCloneAndFundsDisclosureBoundary(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t)
	snapshot, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	var stored StoredNote
	for _, current := range snapshot.Notes {
		stored = current
	}
	stored.Note.Payload = []byte{0xff, 0, 1}
	if err := session.options.Store.Update(ctx, session.scope, func(state *WalletState) error { state.Notes[stored.Commitment] = stored; return nil }); err != nil {
		t.Fatal(err)
	}
	stored.Note.Payload[0] = 1
	snapshot, err = session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Notes[stored.Commitment].Note.Payload[0] != 0xff {
		t.Fatal("store retained caller memo")
	}
	codec, err := NewEncryptedWalletCodec(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := codec.Seal(session.scope, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := codec.Open(session.scope, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Notes[stored.Commitment].Note.Payload[0] = 2
	note := restored.Notes[stored.Commitment].Note
	if note.Payload[0] != 0xff {
		t.Fatal("encrypted storage lost or aliased memo")
	}
	opening, err := sdk.DisclosureFromNote(note)
	if err != nil {
		t.Fatal(err)
	}
	output := sdk.DisclosureOutput{ProtocolContext: note.ProtocolContext, TokenAddress: core.BigIntToElement(token.Big()), OutputCommitment: stored.Commitment, Role: "funds"}
	result, err := opening.Verify(output)
	if err != nil || len(result.Note.Payload) != 0 || opening.GetScope().RevealsPayload {
		t.Fatal("opening claimed memo authenticity", err)
	}
	output.Ciphertext, err = core.EncryptPrivateTransferNote(session.viewing.GetPublicKey(), note, core.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: core.BigIntToElement(stored.Commitment.Big())})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := session.viewing.DiscloseNote(output.Ciphertext, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: core.BigIntToElement(stored.Commitment.Big())}, TokenAddress: note.TokenAddress})
	if err != nil {
		t.Fatal(err)
	}
	result, err = payload.Verify(output)
	if err != nil || !bytes.Equal(result.Note.Payload, note.Payload) || !payload.GetScope().RevealsPayload {
		t.Fatal("payload memo not authenticated", err)
	}
}

type unsupportedProtocolProvider struct{ sessionProvider }

func (p unsupportedProtocolProvider) CallContract(ctx context.Context, message ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if bytes.Equal(message.Data, crypto.Keccak256([]byte("VERSION()"))[:4]) {
		return common.LeftPadBytes([]byte{2}, 32), nil
	}
	return p.sessionProvider.CallContract(ctx, message, block)
}
func TestMemoProtocolGuardAcceptsVersionOneAndRejectsVersionTwo(t *testing.T) {
	base := sessionProvider{}
	pool, err := sdk.NewProtocol(protocolrpc.ProtocolDeployment{ChainID: "31337", Protocol: protocolrpc.ContractDeployment{Address: common.HexToAddress("0x1111111111111111111111111111111111111111"), RuntimeCodeHash: transaction.HashCode([]byte{0x60, 0}), FromBlock: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if err = pool.ValidateDeployment(context.Background(), base); err != nil {
		t.Fatal("Protocol VERSION 1 rejected", err)
	}
	if err = pool.ValidateDeployment(context.Background(), unsupportedProtocolProvider{base}); err == nil {
		t.Fatal("Protocol VERSION 2 accepted")
	}
}

func TestRefreshKeepsAuthenticatedChildMemoAndOmitsUnknownMemo(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(fmt.Sprint("corrupt=", corrupt), func(t *testing.T) {
			ctx := context.Background()
			session, token, _ := testSession(t)
			leaf := seedCredit(t, session, token)
			err := session.options.Store.Update(ctx, session.scope, func(state *WalletState) error {
				for key, credit := range state.Credits {
					credit.State.Note.Payload = []byte("parent")
					state.Credits[key] = credit
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			operation, err := session.PrepareDrawCredit(ctx, sdk.DrawCreditPreparation{TokenAddress: token, Credits: []sdk.CreditDrawInput{{Credit: leaf, Amount: big.NewInt(30), Payload: []byte("child")}}})
			if err != nil {
				t.Fatal(err)
			}
			base := session.options.Provider.(sessionProvider)
			childLeaf := common.Hash(operation.plan.DrawCredit.RemainderCommitments[0])
			base.leaves.Store(uint64(3), childLeaf)
			deliveries := operation.plan.Deliveries
			if corrupt {
				for i := range deliveries {
					deliveries[i].Ciphertext[len(deliveries[i].Ciphertext)-1] ^= 1
				}
			}
			contractABI, err := protocol.ProtocolABI()
			if err != nil {
				t.Fatal(err)
			}
			tx := common.HexToHash("0x1234")
			insertion := contractABI.Events["CommitmentsInserted"]
			insertionData, err := insertion.Inputs.NonIndexed().Pack(uint32(0), [][32]byte{childLeaf})
			if err != nil {
				t.Fatal(err)
			}
			draw := contractABI.Events["CreditDraw"]
			drawData, err := draw.Inputs.NonIndexed().Pack(*operation.plan.DrawCredit, []uint32{0}, deliveries)
			if err != nil {
				t.Fatal(err)
			}
			makeLog := func(topics []common.Hash, data []byte, index uint) *types.Log {
				return &types.Log{Address: session.scope.ProtocolAddress, Topics: topics, Data: data, Index: index, TxHash: tx, BlockHash: base.head.Hash(), BlockNumber: 10}
			}
			receipt := &types.Receipt{TxHash: tx, Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(10), BlockHash: base.head.Hash(), Logs: []*types.Log{
				makeLog([]common.Hash{insertion.ID, common.BigToHash(big.NewInt(3))}, insertionData, 0),
				makeLog([]common.Hash{draw.ID, common.BytesToHash(session.wallet.OwnerAddress().Bytes()), common.BigToHash(big.NewInt(3))}, drawData, 1),
			}}
			provider := receiptProvider{base, receipt}
			deployment := session.protocol.Deployment()
			block := protocolrpc.CanonicalBlock{Number: 10, Hash: base.head.Hash(), ParentHash: base.head.ParentHash, Timestamp: base.head.Time}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					ID     uint64
					Method string
					Params []json.RawMessage
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					return
				}
				var result any
				switch request.Method {
				case "invisi_getStatus":
					result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
				case "invisi_getUpdates":
					result = map[string]any{"blockHash": block.Hash, "creditTransitions": []any{map[string]any{"transactionHash": tx, "logIndex": 1, "block": block}}}
				case "invisi_getRoots", "invisi_getMembership":
					var tree uint64
					if err := json.Unmarshal(request.Params[0], &tree); err != nil {
						t.Error(err)
						return
					}
					stored, ok := base.leaves.Load(tree)
					if !ok {
						t.Error("unknown tree")
						return
					}
					commitment := stored.(common.Hash)
					value := core.BigIntToElement(commitment.Big())
					for range 16 {
						value, _ = core.Poseidon2Hash(value, core.ElementFromUint64(0))
					}
					anchor := map[string]any{"treeId": tree, "root": transaction.FieldHash(value), "leafCount": 1, "block": block}
					if request.Method == "invisi_getRoots" {
						result = []any{anchor}
					} else {
						result = map[string]any{"anchor": anchor, "leafIndex": 0, "commitment": commitment, "siblings": make([]common.Hash, 16)}
					}
				default:
					t.Error("unexpected method", request.Method)
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result}); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client, err := sdk.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &deployment})
			if err != nil {
				t.Fatal(err)
			}
			session.options.Provider = provider
			session.rpc = client.WithProvider(provider)
			if _, err = session.Refresh(ctx, RefreshOptions{FromBlock: 10}); err != nil {
				t.Fatal(err)
			}
			snapshot, err := session.options.Store.ReadSnapshot(ctx, session.scope)
			if err != nil {
				t.Fatal(err)
			}
			for _, role := range []string{"owner", "spender"} {
				child, ok := snapshot.Credits[transaction.CreditKey(childLeaf, role)]
				if !ok || child.State.Note.Amount != core.ElementFromUint64(70) {
					t.Fatal("financial recovery failed")
				}
				expected := "child"
				if corrupt {
					expected = ""
				}
				if string(child.State.Note.Payload) != expected {
					t.Fatal("wrong child memo", string(child.State.Note.Payload), expected)
				}
			}
		})
	}
}

type memoPreparationBarrier struct {
	sessionProvider
	entered, release chan struct{}
	once             sync.Once
}

func (p *memoPreparationBarrier) ChainID(ctx context.Context) (*big.Int, error) {
	p.once.Do(func() { close(p.entered) })
	select {
	case <-p.release:
		return p.sessionProvider.ChainID(ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func TestMemoPreparationSnapshotsBeforeNetworkWait(t *testing.T) {
	session, token, _ := testSession(t)
	address, _ := session.wallet.GetAddress()
	memo := []byte("original")
	provider := &memoPreparationBarrier{sessionProvider: session.options.Provider.(sessionProvider), entered: make(chan struct{}), release: make(chan struct{})}
	session.options.Provider = provider
	session.rpc = session.rpc.WithProvider(provider)
	type preparedResult struct {
		operation *Operation
		err       error
	}
	result := make(chan preparedResult, 1)
	go func() {
		operation, err := session.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30), Payload: memo}})})
		result <- preparedResult{operation, err}
	}()
	<-provider.entered
	memo[0] = 'X'
	close(provider.release)
	prepared := <-result
	if prepared.err != nil {
		t.Fatal(prepared.err)
	}
	if string(prepared.operation.plan.Outputs[0].Note.Payload) != "original" {
		t.Fatal("caller changed memo during network wait")
	}
}
