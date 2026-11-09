package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	creditsession "github.com/invisi-lab/invisi-sdk-go/credit/session"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

func syncGrantState() credit.State {
	return credit.State{AvailableCredit: core.ElementFromUint64(400), ProtocolContext: core.ProtocolContext{ChainID: core.ElementFromUint64(31337), ProtocolAddress: core.ElementFromUint64(17)}, TokenAddress: core.ElementFromUint64(34), OwnerWalletIdentity: core.ElementFromUint64(1), SpenderWalletIdentity: core.ElementFromUint64(2), RevocationShard: core.ElementFromUint64(1), Nonce: core.ElementFromUint64(3), Expiry: core.ElementFromUint64(2000000000), SharedKey: core.ElementFromUint64(4)}
}
func TestRefreshRestoresUnresolvedCreditReservationsForBothRoles(t *testing.T) {
	grant := syncGrantState()
	commitment, e := grant.Commitment()
	if e != nil {
		t.Fatal(e)
	}
	hash := common.Hash(commitment.Bytes())
	state := creditsession.EmptyState()
	for _, role := range []credit.Role{credit.Owner, credit.Spender} {
		state.Grants[creditsession.GrantKey(hash, role)] = creditsession.StoredState{Commitment: hash, Role: role}
	}
	state.Operations["pending"] = creditsession.Journal{ID: "pending", Externalized: true, Status: "broadcast", Intent: credit.Intent{Inputs: []credit.LocatedState{{State: grant}}}}
	if e = restoreCreditReservations(&state); e != nil {
		t.Fatal(e)
	}
	for _, stored := range state.Grants {
		if stored.ReservedBy != "pending" {
			t.Fatal("externalized authority reservation lost")
		}
	}
	state.Operations["other"] = creditsession.Journal{ID: "other", Status: "prepared", Intent: credit.Intent{Inputs: []credit.LocatedState{{State: grant}}}}
	if e = restoreCreditReservations(&state); e == nil {
		t.Fatal("conflicting restored reservation accepted")
	}
}
func TestCanonicalCreditCreationReceiptProducesStatesAndNoCashNotes(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t)
	original := session.options.Provider.(sessionProvider)
	snapshot, e := session.options.Store.ReadSnapshot(ctx, session.scope)
	if e != nil {
		t.Fatal(e)
	}
	state := syncGrantState()
	state.ProtocolContext = session.coreContext()
	state.TokenAddress = core.BigIntToElement(token.Big())
	identity := core.BigIntToElement(session.wallet.OwnerAddress().Big())
	state.OwnerWalletIdentity = identity
	state.SpenderWalletIdentity = identity
	state.Expiry = core.ElementFromUint64(original.head.Time + 1000)
	commitment, e := state.Commitment()
	if e != nil {
		t.Fatal(e)
	}
	leaf := common.Hash(commitment.Bytes())
	original.leaves.Store(uint64(2), leaf)
	keys := core.ReceivingKeys{ViewingPublicKey: session.wallet.ViewingPublicKey(), ClueKey: session.wallet.ClueKey()}
	delivery := credit.DeliveryState{State: state, OwnerReceivingKeys: keys, SpenderReceivingKeys: keys}
	deliveryContext := credit.DeliveryContext{ProtocolContext: session.coreContext(), TokenAddress: state.TokenAddress, Commitment: commitment, RecipientIdentity: identity}
	envelopes := []contracts.IInvisiProtocolDelivery{}
	for _, role := range []credit.Role{credit.Owner, credit.Spender} {
		cipher, e := credit.Encrypt(delivery, deliveryContext, role)
		if e != nil {
			t.Fatal(e)
		}
		envelopes = append(envelopes, contracts.IInvisiProtocolDelivery{Ciphertext: cipher, Clue: make([]byte, 68)})
	}
	contract, e := contracts.InvisiProtocolMetaData.GetAbi()
	if e != nil {
		t.Fatal(e)
	}
	tx := common.HexToHash("1234")
	log := func(name string, index uint, indexed []any, args ...any) *types.Log {
		t.Helper()
		event := contract.Events[name]
		data, e := event.Inputs.NonIndexed().Pack(args...)
		if e != nil {
			t.Fatal(e)
		}
		topics := []common.Hash{event.ID}
		for _, value := range indexed {
			v, e := abi.MakeTopics([]any{value})
			if e != nil {
				t.Fatal(e)
			}
			topics = append(topics, v[0][0])
		}
		return &types.Log{Address: session.scope.ProtocolAddress, Index: index, TxHash: tx, BlockNumber: 10, BlockHash: original.head.Hash(), Topics: topics, Data: data}
	}
	insertion := log("CommitmentsInserted", 1, []any{big.NewInt(2)}, uint32(0), [][32]byte{leaf})
	created := log("CreditCreated", 2, []any{token, session.wallet.OwnerAddress(), big.NewInt(2)}, uint64(1), uint16(0), [32]byte(leaf), uint32(0), envelopes, big.NewInt(int64(original.head.Time+100)), common.Address{}, big.NewInt(0))
	receipt := &types.Receipt{TxHash: tx, Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(10), BlockHash: original.head.Hash(), Logs: []*types.Log{insertion, created}}
	provider := receiptProvider{original, receipt}
	session.options.Provider = provider
	session.rpc = session.rpc.WithProvider(provider)
	notes, states, draws, e := readTransactionForTest(session, ctx, tx, *snapshot.ObservedThrough)
	if e != nil {
		t.Fatal(e)
	}
	if len(notes) != 0 || len(states) != 2 || len(draws) != 0 {
		t.Fatalf("mixed receipt projection notes=%d states=%d draws=%d", len(notes), len(states), len(draws))
	}
	if states[0].Role == states[1].Role || states[0].Commitment != leaf || states[1].Nullifier != states[0].Nullifier {
		t.Fatal("role envelope or state identity lost")
	}
	insertion.TxHash = common.HexToHash("99")
	if _, _, _, e = readTransactionForTest(session, ctx, tx, *snapshot.ObservedThrough); e == nil {
		t.Fatal("unauthenticated insertion accepted")
	}
}

type creditDrawSyncProvider struct {
	receiptProvider
	consumed common.Hash
}

func (p creditDrawSyncProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, height *big.Int) ([]byte, error) {
	a, e := contracts.InvisiProtocolMetaData.GetAbi()
	if e != nil {
		return nil, e
	}
	method, e := a.MethodById(msg.Data[:4])
	if e == nil && method.RawName == "isNullifierSpent" {
		values, e := method.Inputs.Unpack(msg.Data[4:])
		if e != nil {
			return nil, e
		}
		return method.Outputs.Pack(common.Hash(values[0].([32]byte)) == p.consumed)
	}
	return p.sessionProvider.CallContract(ctx, msg, height)
}
func TestRefreshKeepsAuthenticatedChildPayloadAndDropsUnknownRecoveredPayload(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(fmt.Sprint("corrupt=", corrupt), func(t *testing.T) {
			ctx := context.Background()
			session, token, _ := testSession(t, 2678401)
			base := session.options.Provider.(sessionProvider)
			dep := session.protocol.Deployment()
			block := protocolrpc.CanonicalBlock{Number: 10, Hash: base.head.Hash(), ParentHash: base.head.ParentHash, Timestamp: base.head.Time}
			state := syncGrantState()
			state.ProtocolContext = session.coreContext()
			state.TokenAddress = core.BigIntToElement(token.Big())
			identity := core.BigIntToElement(session.wallet.OwnerAddress().Big())
			state.OwnerWalletIdentity = identity
			state.SpenderWalletIdentity = identity
			state.Expiry = core.ElementFromUint64(base.head.Time + 10000)
			state.Payload = []byte{1, 2, 3}
			keys := core.ReceivingKeys{ViewingPublicKey: session.wallet.ViewingPublicKey(), ClueKey: session.wallet.ClueKey()}
			parent := credit.DeliveryState{State: state, OwnerReceivingKeys: keys, SpenderReceivingKeys: keys}
			oldLeaf, e := state.Commitment()
			if e != nil {
				t.Fatal(e)
			}
			oldNF, e := state.Nullifier()
			if e != nil {
				t.Fatal(e)
			}
			if e = session.options.Store.Update(ctx, session.scope, func(wallet *WalletState) error {
				for _, role := range []credit.Role{credit.Owner, credit.Spender} {
					wallet.PostpaidCredit.Grants[creditsession.GrantKey(common.Hash(oldLeaf.Bytes()), role)] = creditsession.StoredState{Delivery: parent, Commitment: common.Hash(oldLeaf.Bytes()), Nullifier: common.Hash(oldNF.Bytes()), TreeID: 1, Role: role}
				}
				return nil
			}); e != nil {
				t.Fatal(e)
			}
			draw, e := state.Draw(core.ElementFromUint64(7), base.head.Time)
			if e != nil {
				t.Fatal(e)
			}
			child := parent
			child.State = draw.State
			child.State.Payload = []byte{4, 5}
			note := common.HexToHash("8")
			cipherContext := credit.DeliveryContext{ProtocolContext: session.coreContext(), TokenAddress: state.TokenAddress, Commitment: draw.StateCommitment, RecipientIdentity: identity}
			deliveries := []contracts.IInvisiProtocolDelivery{}
			for _, role := range []credit.Role{credit.Owner, credit.Spender} {
				cipher, e := credit.Encrypt(child, cipherContext, role)
				if e != nil {
					t.Fatal(e)
				}
				if corrupt {
					cipher[len(cipher)-1] ^= 1
				}
				deliveries = append(deliveries, contracts.IInvisiProtocolDelivery{Ciphertext: cipher, Clue: make([]byte, 68)})
			}
			deliveries = append(deliveries, contracts.IInvisiProtocolDelivery{})
			a, e := contracts.InvisiProtocolMetaData.GetAbi()
			if e != nil {
				t.Fatal(e)
			}
			tx := common.HexToHash("5678")
			pack := func(name string, index uint, indexed []any, args ...any) *types.Log {
				t.Helper()
				event := a.Events[name]
				data, e := event.Inputs.NonIndexed().Pack(args...)
				if e != nil {
					t.Fatal(e)
				}
				topics := []common.Hash{event.ID}
				for _, value := range indexed {
					v, e := abi.MakeTopics([]any{value})
					if e != nil {
						t.Fatal(e)
					}
					topics = append(topics, v[0][0])
				}
				return &types.Log{Address: session.scope.ProtocolAddress, Index: index, TxHash: tx, BlockNumber: 10, BlockHash: block.Hash, Topics: topics, Data: data}
			}
			leaf := common.Hash(draw.StateCommitment.Bytes())
			nf := common.Hash(draw.Nullifier.Bytes())
			params := contracts.IInvisiProtocolCreditDrawParams{Token: token, Sources: []contracts.IInvisiProtocolInputSource{{TreeId: big.NewInt(1), MerkleRoot: [32]byte{9}}}, Payers: []contracts.IInvisiProtocolCreditPayer{{Payer: session.wallet.OwnerAddress(), Amount: big.NewInt(7)}}, GrantRefs: []contracts.IInvisiProtocolCreditGrantRef{{PayerIndex: big.NewInt(0), Shard: 1, Slot: 0}}, Nullifiers: [][32]byte{nf}, CreditCommitments: [][32]byte{leaf}, RecoveryAmounts: [][32]byte{draw.RecoveryAmount.Bytes()}, NoteCommitment: note, TotalAmount: big.NewInt(7), Deadline: new(big.Int).SetUint64(base.head.Time + 100), FeeAmount: big.NewInt(0)}
			insertion := pack("CommitmentsInserted", 1, []any{big.NewInt(2)}, uint32(0), [][32]byte{leaf, note})
			event := pack("CreditDraw", 2, []any{big.NewInt(2)}, params, []uint32{0, 1}, deliveries)
			receipt := &types.Receipt{TxHash: tx, Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(10), BlockHash: block.Hash, Logs: []*types.Log{insertion, event}}
			provider := creditDrawSyncProvider{receiptProvider{base, receipt}, nf}
			siblings := make([]common.Hash, 16)
			siblings[0] = note
			root := draw.StateCommitment
			for _, sibling := range siblings {
				root, e = core.Poseidon2Hash(root, core.BigIntToElement(sibling.Big()))
				if e != nil {
					t.Fatal(e)
				}
			}
			anchor := map[string]any{"treeId": 2, "root": common.Hash(root.Bytes()), "leafCount": 2, "block": block}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					ID     uint64
					Method string
				}
				json.NewDecoder(r.Body).Decode(&request)
				var result any
				switch request.Method {
				case "invisi_getStatus":
					result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": dep, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
				case "invisi_getNoteUpdates":
					result = map[string]any{"blockHash": block.Hash, "notes": []any{}}
				case "invisi_getCreditTransitions":
					result = []any{map[string]any{"block": block, "name": "CreditDraw", "transactionHash": tx, "logIndex": 2, "data": map[string]any{}}}
				case "invisi_getRoots":
					result = []any{anchor}
				case "invisi_getMembership":
					result = map[string]any{"anchor": anchor, "leafIndex": 0, "commitment": leaf, "siblings": siblings}
				default:
					t.Error("unexpected RPC", request.Method)
				}
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
			}))
			defer server.Close()
			client, e := protocolrpc.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &dep})
			if e != nil {
				t.Fatal(e)
			}
			session.rpc = client.WithProvider(provider)
			session.options.Provider = provider
			beforePending, _ := session.options.Store.ReadSnapshot(ctx, session.scope)
			pending, e := session.Refresh(ctx, RefreshOptions{FromBlock: 10})
			if e != nil {
				t.Fatal(e)
			}
			afterPending, _ := session.options.Store.ReadSnapshot(ctx, session.scope)
			if beforePending.Revision != afterPending.Revision {
				t.Fatal("pending discovery changed real store")
			}
			replies := make([]WalletRefreshResponse, len(pending.WalletRequests))
			good := -1
			for i, q := range pending.WalletRequests {
				response, executeErr := testExecutor(session.wallet).Execute(ctx, q)
				replies[i].RequestID = q.ID
				if executeErr == nil {
					replies[i].Response = &response
					good = i
				}
			}
			if good >= 0 {
				original := *replies[good].Response
				bad := original
				bad.ID = "substituted"
				replies[good].Response = &bad
				if _, err := pending.ApplyWalletResponses(ctx, replies); err == nil {
					t.Fatal("accepted substituted response")
				}
				afterBad, _ := session.options.Store.ReadSnapshot(ctx, session.scope)
				if afterBad.Revision != beforePending.Revision {
					t.Fatal("invalid response partially committed refresh")
				}
				replies[good].Response = &original
			}
			result, e := pending.ApplyWalletResponses(ctx, replies)
			if e != nil {
				t.Fatal(e)
			}
			if result.AddedNotes != 0 || result.AddedCredits != 2 {
				t.Fatal(result)
			}
			wallet, e := session.options.Store.ReadSnapshot(ctx, session.scope)
			if e != nil {
				t.Fatal(e)
			}
			for _, role := range []credit.Role{credit.Owner, credit.Spender} {
				got := wallet.PostpaidCredit.Grants[creditsession.GrantKey(leaf, role)]
				if core.ElementToBigInt(got.Delivery.State.TotalSpent).Uint64() != 7 || got.Delivery.State.AvailableCredit != core.ElementFromUint64(393) {
					t.Fatal("recovered usage lost")
				}
				if corrupt && len(got.Delivery.State.Payload) != 0 {
					t.Fatal("unknown payload was inherited")
				}
				if !corrupt && !bytes.Equal(got.Delivery.State.Payload, []byte{4, 5}) {
					t.Fatal("authenticated child payload lost")
				}
				if !wallet.PostpaidCredit.Grants[creditsession.GrantKey(common.Hash(oldLeaf.Bytes()), role)].Consumed {
					t.Fatal("parent consumption lost")
				}
			}
		})
	}
}

func TestDirectResetFencesExternalizedCreditJournalOnReorg(t *testing.T) {
	s, _ := refreshFixture(t, false, nil, nil)
	ctx := context.Background()
	if e := s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		copy := *state.ObservedThrough
		copy.Hash = common.HexToHash("dead")
		state.ObservedThrough = &copy
		state.PostpaidCredit.Operations["exposed"] = creditsession.Journal{ID: "exposed", Status: "confirmed", WorkerID: "old-worker", Externalized: true, Calldata: []byte{1, 2, 3}}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if _, e := refreshForTest(s, ctx, RefreshOptions{FromBlock: 1, Reset: true}); e != nil {
		t.Fatal(e)
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	journal := state.PostpaidCredit.Operations["exposed"]
	if journal.Status != "reorged" || journal.WorkerID != "" || !journal.Externalized || !bytes.Equal(journal.Calldata, []byte{1, 2, 3}) {
		t.Fatalf("direct reset released or hid exposed Credit journal: %+v", journal)
	}
}
