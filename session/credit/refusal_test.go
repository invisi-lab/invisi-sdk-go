package session

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/rpc/broadcaster"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFirstCreditRefusalRollsBackButUncertainReplayRemainsImmutable(t *testing.T) {
	for _, code := range []int{-32024, -32025, -32026, -32000} {
		for _, mode := range []string{"first", "replay", "store-failure"} {
			replay := mode == "replay"
			t.Run(big.NewInt(int64(code)).String()+"/"+mode, func(t *testing.T) {
				ctx := context.Background()
				store := NewMemoryStore()
				managed, _ := New(Options{Store: store})
				_, g := testPlan(t)
				g.Role = credit.Owner
				managed.Track(ctx, g)
				p, e := credit.PrepareRevoke(credit.Context{ProtocolContext: g.Delivery.State.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}, credit.RevokeParams{Token: common.BigToAddress(big.NewInt(34)), Owner: common.BigToAddress(big.NewInt(44)), GrantID: big.NewInt(1), Deadline: big.NewInt(2000000300)})
				if e != nil {
					t.Fatal(e)
				}
				op, e := managed.Prepare(ctx, "invoice", p)
				if e != nil {
					t.Fatal(e)
				}
				// Exercise reservation retention even for this inexpensive signature-only transport fixture.
				if e = store.Update(ctx, func(s *State) error { v := s.Grants[g.key()]; v.ReservedBy = op.id; s.Grants[g.key()] = v; return nil }); e != nil {
					t.Fatal(e)
				}
				key, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
				signer := crypto.PubkeyToAddress(key.PublicKey)
				collector := common.HexToAddress("0x1234")
				server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var req struct {
						ID     uint64
						Method string
						Params []struct{ Data string }
					}
					json.NewDecoder(r.Body).Decode(&req)
					var result any
					switch req.Method {
					case "broadcaster_estimateQuote":
						result = broadcaster.Fee{FeeRecipient: collector, FeeAmount: "1", MaxGas: "2000000"}
					case "broadcaster_getQuote":
						data := common.FromHex(req.Params[0].Data)
						decoded, _ := broadcaster.DecodeIntent(data)
						hash, _ := broadcaster.RequestHash(data)
						q := broadcaster.Quote{Token: decoded.Token, FeeRecipient: collector, FeeAmount: "1", Deadline: decoded.Deadline.String(), MaxGas: "2000000", RequestHash: hash, Execution: "direct"}
						trust := broadcaster.Trust{ChainID: big.NewInt(1), Protocol: common.BigToAddress(big.NewInt(17)), Signer: signer}
						typed, _ := broadcaster.QuoteTypedData(q, trust)
						digest, _, _ := apitypes.TypedDataAndHash(typed)
						sig, _ := crypto.Sign(digest, key)
						result = broadcaster.SignedQuote{Quote: q, Signature: "0x" + common.Bytes2Hex(sig)}
					case "broadcaster_getStatus":
						result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "chainId": "1", "protocol": common.BigToAddress(big.NewInt(17)), "sender": signer, "feeRecipient": collector}
					case "broadcaster_submitTransaction":
						reason := map[int]string{-32024: "CREDIT_PERIOD_CHANGED", -32025: "INSUFFICIENT_EXECUTION_SLACK", -32026: "CREDIT_PERIOD_ENDING", -32000: "CREDIT_PERIOD_CHANGED"}[code]
						json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": code, "message": reason + ": refused"}})
						return
					default:
						t.Errorf("unexpected %s", req.Method)
					}
					json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
				}))
				defer server.Close()
				client, e := broadcaster.NewClient(broadcaster.Configuration{Endpoint: server.URL, QuoteSigner: signer, FeeRecipient: collector}, broadcaster.Options{HTTPClient: server.Client()})
				if e != nil {
					t.Fatal(e)
				}
				if e = client.Quote(ctx, op, broadcaster.QuoteOptions{MaxFeeAmount: big.NewInt(2)}); e != nil {
					t.Fatal(e)
				}
				if e = op.WithRevocationSignature(ctx, make([]byte, 65), func(context.Context, common.Address, apitypes.TypedData, []byte) (bool, error) { return true, nil }); e != nil {
					t.Fatal(e)
				}
				if replay {
					q, _ := op.ExportBroadcasterQuote(ctx)
					data, _, e := op.prepared.ExportBroadcasterSubmission(client)
					if e != nil {
						t.Fatal(e)
					}
					req, e := client.PrepareSubmission(data, q, nil, nil)
					if e != nil {
						t.Fatal(e)
					}
					store.Update(ctx, func(s *State) error {
						j := s.Operations[op.id]
						j.Externalized = true
						j.Status = "uncertain"
						j.Calldata = data
						j.BroadcasterRequest = &req
						s.Operations[op.id] = j
						return nil
					})
				}
				if mode == "store-failure" {
					managed.store = refusalPersistenceFailure{store}
				}
				_, e = client.Submit(ctx, op)
				if e == nil {
					t.Fatal("expected refusal")
				}
				var refused *broadcaster.SubmissionRefusedError
				if (code != -32000) != errors.As(e, &refused) {
					t.Fatalf("wrong error: %v", e)
				}
				state, _ := store.ReadSnapshot(ctx)
				j := state.Operations[op.id]
				wantExternal := replay || code == -32000 || mode == "store-failure"
				if j.Externalized != wantExternal || (j.BroadcasterRequest != nil) != wantExternal {
					t.Fatalf("unsafe refusal state: %+v", j)
				}
				if state.Grants[g.key()].ReservedBy != op.id {
					t.Fatal("reservation lost")
				}
				if !wantExternal && j.Status != "ready" {
					t.Fatalf("status=%s", j.Status)
				}
				if !wantExternal {
					if _, e = op.Calldata(ctx, common.Address{}); e == nil {
						t.Fatal("refused old calldata exported before refresh")
					}
				}
				managed.store = store
				managed.verifyRevocation = func(context.Context, common.Address, apitypes.TypedData, []byte) (bool, error) { return true, nil }
				resumed, e := managed.Resume(ctx, op.id, nil)
				if e != nil {
					t.Fatal(e)
				}
				if _, e = client.Submit(ctx, op); e == nil {
					t.Fatal("old worker not fenced")
				}
				if wantExternal {
					_, e = client.Submit(ctx, resumed)
					if e == nil {
						t.Fatal("expected replay refusal")
					}
					st, _ := store.ReadSnapshot(ctx)
					if !st.Operations[op.id].Externalized || st.Operations[op.id].BroadcasterRequest == nil {
						t.Fatal("replay cleared an unknown earlier attempt")
					}
				}
				if !wantExternal {
					if e = resumed.Abandon(ctx); e != nil {
						t.Fatal("known unsent operation cannot be abandoned", e)
					}
				}
			})
		}
	}
}

type refusalPersistenceFailure struct{ Store }

func (s refusalPersistenceFailure) Update(ctx context.Context, f func(*State) error) error {
	return s.Store.Update(ctx, func(state *State) error {
		if err := f(state); err != nil {
			return err
		}
		for _, j := range state.Operations {
			if j.SubmissionRefusal != "" && !j.Externalized {
				return errors.New("refusal persistence unavailable")
			}
		}
		return nil
	})
}
