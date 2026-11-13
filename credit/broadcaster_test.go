package credit

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

type creditRoundTrip func(*http.Request) (*http.Response, error)

func (f creditRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func creditBroadcaster(t *testing.T, c Context, signer *ecdsa.PrivateKey, collector common.Address, policies ...common.Hash) *broadcaster.Client {
	t.Helper()
	transport := creditRoundTrip(func(r *http.Request) (*http.Response, error) {
		var request struct {
			ID     json.RawMessage
			Method string
			Params []map[string]any
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		var result any
		switch request.Method {
		case "broadcaster_estimateQuote":
			result = broadcaster.Fee{FeeRecipient: collector, FeeAmount: "1", MaxGas: "2000000"}
		case "broadcaster_getQuote":
			data := common.FromHex(request.Params[0]["data"].(string))
			intent, err := broadcaster.DecodeIntent(data)
			if err != nil {
				t.Fatal(err)
			}
			hash, err := broadcaster.RequestHash(data)
			if err != nil {
				t.Fatal(err)
			}
			if len(policies) > 0 {
				bound, e := provenance.ProvenanceIntentHash(hash.Hex())
				if e != nil {
					t.Fatal(e)
				}
				hash = common.HexToHash(bound)
			}
			var schedule *broadcaster.Schedule
			if value := request.Params[0]["schedule"]; value != nil {
				raw, _ := json.Marshal(value)
				schedule = &broadcaster.Schedule{}
				if err := json.Unmarshal(raw, schedule); err != nil {
					t.Fatal(err)
				}
			}
			q := broadcaster.Quote{Schedule: schedule, Token: intent.Token, FeeRecipient: intent.FeeRecipient, FeeAmount: intent.FeeAmount.String(), Deadline: intent.Deadline.String(), MaxGas: "2000000", RequestHash: hash, Execution: "direct"}
			typed, err := broadcaster.QuoteTypedData(q, broadcaster.Trust{ChainID: primitives.ElementToBigInt(c.ChainID), Protocol: common.BigToAddress(primitives.ElementToBigInt(c.ProtocolAddress)), Signer: crypto.PubkeyToAddress(signer.PublicKey)})
			if err != nil {
				t.Fatal(err)
			}
			digest, _, err := apitypes.TypedDataAndHash(typed)
			if err != nil {
				t.Fatal(err)
			}
			signature, err := crypto.Sign(digest, signer)
			if err != nil {
				t.Fatal(err)
			}
			signed := broadcaster.SignedQuote{Quote: q, Signature: "0x" + common.Bytes2Hex(signature)}
			if len(policies) > 0 {
				signed.ProvenanceRequired = true
				signed.RequestHashFormat = broadcaster.ProvenanceRequestHashFormat
			}
			result = signed
		case "broadcaster_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "chainId": primitives.ElementToBigInt(c.ChainID).String(), "protocol": common.BigToAddress(primitives.ElementToBigInt(c.ProtocolAddress)).Hex(), "sender": crypto.PubkeyToAddress(signer.PublicKey).Hex(), "feeRecipient": collector.Hex(), "submissionHalted": false, "scheduling": map[string]any{"enabled": true, "maxDelaySeconds": "1209600", "defaultExecutionWindowSeconds": "300", "maxExecutionWindowSeconds": "3600", "minDeadlineSlackSeconds": "30"}, "provenance": map[string]any{"enabled": len(policies) > 0, "requestHashFormat": broadcaster.ProvenanceRequestHashFormat, "trackingRequired": true, "trackingSupported": true}}
		case "broadcaster_getTransactionStatus":
			result = broadcaster.Status{ID: request.Params[0]["id"].(string), State: "uncertain"}
		case "broadcaster_submitTransaction":
			result = broadcaster.Status{ID: "server-request-id", State: "submitted", TransactionHash: common.HexToHash("0x1234").Hex()}
		default:
			t.Fatalf("unexpected method %s", request.Method)
		}
		raw, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
	})
	client, err := broadcaster.NewClient(broadcaster.Configuration{Endpoint: "https://operator.example", QuoteSigner: crypto.PubkeyToAddress(signer.PublicKey), FeeRecipient: collector}, broadcaster.Options{HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
func TestCreditQuoteFreezesFeeBeforeAuthorization(t *testing.T) {
	s, f := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	commitment, _ := s.Commitment()
	p := CreateParams{Token: c.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}
	d := f["delivery"].(map[string]any)
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: bytesHex(t, d["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, d["spenderCiphertext"].(string)), Clue: clue}}
	prepared, err := PrepareCreate(c, p, ds, s)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := prepared.ExportBroadcasterIntent()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := broadcaster.DecodeIntent(intent)
	if err != nil || decoded.Method != "createCredit" || decoded.FeeAmount.Sign() != 0 {
		t.Fatal("unsigned Credit intent", err)
	}
	key, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	collector := common.HexToAddress("0x1234")
	client := creditBroadcaster(t, c, key, collector)
	if err = client.Quote(context.Background(), prepared, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	quoted, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if quoted.Create.FeeRecipient != collector || quoted.Create.FeeAmount.Cmp(big.NewInt(1)) != 0 || quoted.Create.CreditCommitment != p.CreditCommitment || quoted.CreatedState.CreditSharedKey != s.CreditSharedKey {
		t.Fatal("quote changed grant or omitted fee")
	}
}

func TestCreditRequiredPolicyQuoteWithoutSourceProver(t *testing.T) {
	s, f := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	commitment, _ := s.Commitment()
	p := CreateParams{Token: c.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}
	d := f["delivery"].(map[string]any)
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: bytesHex(t, d["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, d["spenderCiphertext"].(string)), Clue: clue}}
	prepared, err := PrepareCreate(c, p, ds, s)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := prepared.ExportBroadcasterIntent()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := broadcaster.DecodeIntent(intent)
	if err != nil || decoded.Method != "createCredit" || decoded.FeeAmount.Sign() != 0 {
		t.Fatal("unsigned Credit intent", err)
	}
	key, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	collector := common.HexToAddress("0x1234")
	client := creditBroadcaster(t, c, key, collector, common.HexToHash("0x77"))
	if err = client.Quote(context.Background(), prepared, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	quoted, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if quoted.Create.FeeRecipient != collector || quoted.Create.FeeAmount.Cmp(big.NewInt(1)) != 0 || quoted.Create.CreditCommitment != p.CreditCommitment || quoted.CreatedState.CreditSharedKey != s.CreditSharedKey {
		t.Fatal("quote changed grant or omitted fee")
	}
}

func TestCreditKnownSubmissionQueriesWithoutFreshProof(t *testing.T) {
	s, _ := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	clue := make([]byte, 68)
	clue[64] = 24
	commitment, _ := s.Commitment()
	prepared, err := PrepareCreate(c, CreateParams{Token: c.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}, []protocol.Delivery{{Ciphertext: []byte{0x82}, Clue: clue}, {Ciphertext: []byte{0x82}, Clue: clue}}, s)
	if err != nil {
		t.Fatal(err)
	}
	key, _ := crypto.GenerateKey()
	client := creditBroadcaster(t, c, key, common.HexToAddress("0x1234"), common.HexToHash("0x77"))
	if err = client.Quote(context.Background(), prepared, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	prepared.submission = Submission{RequestID: "original", State: "uncertain"}
	status, err := client.Submit(context.Background(), prepared)
	if err != nil || status.State != "uncertain" {
		t.Fatal("known externalized request needed a new proof/send", status, err)
	}
}

func TestCreditScheduledQuotesCoverEveryMethod(t *testing.T) {
	s, f := fixture(t)
	raw := common.LeftPadBytes([]byte{1}, 32)
	key, _ := crypto.ToECDSA(raw)
	owner := crypto.PubkeyToAddress(key.PublicKey)
	s.OwnerWalletIdentity = primitives.BigIntToElement(owner.Big())
	s.SpenderWalletIdentity = s.OwnerWalletIdentity
	commitment, _ := s.Commitment()
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	d := f["delivery"].(map[string]any)
	clue := make([]byte, 68)
	clue[64] = 24
	deliveries := []protocol.Delivery{{Ciphertext: bytesHex(t, d["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, d["spenderCiphertext"].(string)), Clue: clue}}
	p, err := PrepareCreate(c, CreateParams{Token: c.TokenAddress, Owner: owner, RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}, deliveries, s)
	if err != nil {
		t.Fatal(err)
	}
	input := LocatedState{State: s, TreeID: 1, Debit: primitives.ElementFromUint64(7)}
	root, err := input.root()
	if err != nil {
		t.Fatal(err)
	}
	drawn, err := s.Draw(primitives.ElementFromUint64(7), c.Timestamp)
	if err != nil {
		t.Fatal(err)
	}
	identity := primitives.ElementFromUint64(1)
	noteCommitment, err := primitives.ComputeNoteCommitment(s.ChainID, s.ProtocolAddress, identity, s.TokenAddress, primitives.ElementFromUint64(7))
	if err != nil {
		t.Fatal(err)
	}
	draw, err := PrepareDraw(c, DrawParams{Token: c.TokenAddress, Sources: []protocol.InputSource{{TreeId: big.NewInt(1), MerkleRoot: root}}, Payers: []Payer{{Payer: owner, Amount: big.NewInt(7)}}, GrantRefs: []GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(drawn.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(drawn.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(drawn.RecoveryAmount.Bytes())}, NoteCommitment: common.Hash(noteCommitment.Bytes()), TotalAmount: big.NewInt(7), Deadline: big.NewInt(2000000300)}, append(deliveries, protocol.Delivery{}), []LocatedState{input}, identity)
	if err != nil {
		t.Fatal(err)
	}
	change := Change{Policy: s.Policy, Expiry: s.Expiry, IncreaseAvailableCredit: primitives.ElementFromUint64(5)}
	updated, err := s.Update(change, 2000000300)
	if err != nil {
		t.Fatal(err)
	}
	update, err := PrepareUpdate(c, UpdateParams{Token: c.TokenAddress, Owner: owner, Source: protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: root}, Nullifier: common.Hash(updated.Nullifier.Bytes()), CreditCommitment: common.Hash(updated.StateCommitment.Bytes()), RevocationShard: 1, Deadline: big.NewInt(2000000300)}, deliveries, input, change)
	if err != nil {
		t.Fatal(err)
	}

	revoke, err := PrepareRevoke(c, RevokeParams{Token: c.TokenAddress, Owner: owner, Shard: 1, Deadline: big.NewInt(2000000300)})
	if err != nil {
		t.Fatal(err)
	}
	quoteKey, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	client := creditBroadcaster(t, c, quoteKey, common.HexToAddress("0x1234"))
	seen := map[string]bool{}
	for _, prepared := range []*Prepared{p, draw, update, revoke} {
		after := strconv.FormatInt(time.Now().Unix()+60, 10)
		if err = client.Quote(context.Background(), prepared, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2), Schedule: &broadcaster.ScheduleOptions{PublishAfter: after}}); err != nil {
			t.Fatal(err)
		}
		intent, err := prepared.ExportIntent()
		if err != nil {
			t.Fatal(err)
		}
		binding := intent.Broadcaster
		if binding == nil || binding.CancelToken == "" || binding.Quote.Quote.Schedule == nil {
			t.Fatal("missing cancellation binding")
		}
		schedule := binding.Quote.Quote.Schedule
		if schedule.PublishAfter != after || crypto.Keccak256Hash(common.FromHex(binding.CancelToken)).Hex() != schedule.CancelTokenHash || seen[binding.CancelToken] {
			t.Fatal("schedule secret reused or unbound")
		}
		seen[binding.CancelToken] = true
		if strings.Contains(fmt.Sprintf("%+v %#v", binding, binding), binding.CancelToken) {
			t.Fatal("cancellation secret leaked in formatting")
		}
		end, _ := strconv.ParseUint(schedule.ExecuteBefore, 10, 64)
		start, _ := strconv.ParseUint(after, 10, 64)
		if end-start != 300 {
			t.Fatal("default execution window ignored")
		}
		if _, err = prepared.ExportBroadcasterQuote(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCreditFeeAttemptBoundsAndDefault(t *testing.T) {
	for _, attempts := range []int{0, 1, 16, 17, -1} {
		t.Run(strconv.Itoa(attempts), func(t *testing.T) {
			state, f := fixture(t)
			c := Context{ProtocolContext: state.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
			commitment, _ := state.Commitment()
			d := f["delivery"].(map[string]any)
			clue := make([]byte, 68)
			clue[64] = 24
			prepared, err := PrepareCreate(c, CreateParams{Token: c.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}, []protocol.Delivery{{Ciphertext: bytesHex(t, d["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, d["spenderCiphertext"].(string)), Clue: clue}}, state)
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			collector := common.HexToAddress("0x1234")
			transport := creditRoundTrip(func(r *http.Request) (*http.Response, error) {
				calls++
				var req struct {
					ID     uint64
					Method string
				}
				json.NewDecoder(r.Body).Decode(&req)
				if req.Method != "broadcaster_estimateQuote" {
					t.Fatal(req.Method)
				}
				raw, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": broadcaster.Fee{FeeRecipient: collector, FeeAmount: strconv.Itoa(calls), MaxGas: "2000000"}})
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
			})
			client, err := broadcaster.NewClient(broadcaster.Configuration{Endpoint: "https://operator.example", QuoteSigner: common.HexToAddress("0x11"), FeeRecipient: collector}, broadcaster.Options{HTTPClient: &http.Client{Transport: transport}})
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Quote(context.Background(), prepared, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(100), MaxAttempts: attempts}); err == nil {
				t.Fatal("nonconverging quotation accepted")
			}
			expected := attempts
			if attempts == 0 {
				expected = 4
			}
			if attempts < 0 || attempts > 16 {
				expected = 0
			}
			if calls != expected {
				t.Fatal("fee attempt semantics changed", calls, expected)
			}
			intent, err := prepared.ExportIntent()
			if err != nil || intent.Broadcaster != nil || intent.Create.FeeAmount != nil && intent.Create.FeeAmount.Sign() != 0 {
				t.Fatal("failed quote mutated preparation", err)
			}
		})
	}
}

func TestDetachedCreditUnknownIDReplaysFrozenTransportWithoutRebuildingProof(t *testing.T) {
	state, _ := fixture(t)
	c := Context{ProtocolContext: state.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	p, err := PrepareRevoke(c, RevokeParams{Token: c.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), Shard: 1, Deadline: big.NewInt(2000000300)})
	if err != nil {
		t.Fatal(err)
	}
	key, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	client := creditBroadcaster(t, c, key, common.HexToAddress("0x1234"))
	// 传输日志已存在时，只验证精确重放，不声称此合成请求可通过链上验证。
	p.intent.Broadcaster = &BroadcasterBinding{Configuration: client.Configuration()}
	p.submission = Submission{State: "uncertain"}
	p.request = &broadcaster.Submission{Data: "0x010203", Signature: "frozen", Quote: broadcaster.Quote{Execution: "direct", Deadline: "1"}}
	if status, err := client.Submit(context.Background(), p); err != nil || status.ID != "server-request-id" {
		t.Fatal(status, err)
	}
}

func TestCreditScheduleRejectsDeadlineDelayAndWindowViolations(t *testing.T) {
	state, _ := fixture(t)
	now := time.Now().Unix()
	c := Context{ProtocolContext: state.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: uint64(now)}
	key, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	client := creditBroadcaster(t, c, key, common.HexToAddress("0x1234"))
	for _, test := range []struct {
		name                    string
		deadline, after, before int64
	}{
		{"deadline-slack", now + 380, now + 60, now + 360},
		{"delay", now + 2000000, now + 1210200, now + 1210500},
		{"window-order", now + 1000, now + 60, now + 59},
		{"window-limit", now + 5000, now + 60, now + 3661},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, err := PrepareRevoke(c, RevokeParams{Token: c.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), Shard: 1, Deadline: big.NewInt(test.deadline)})
			if err != nil {
				t.Fatal(err)
			}
			options := BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2), Schedule: &broadcaster.ScheduleOptions{PublishAfter: strconv.FormatInt(test.after, 10), ExecuteBefore: strconv.FormatInt(test.before, 10)}}
			if err = client.Quote(context.Background(), p, options); err == nil {
				t.Fatal("invalid schedule accepted")
			}
			intent, err := p.ExportIntent()
			if err != nil || intent.Broadcaster != nil {
				t.Fatal("failed schedule mutated preparation", err)
			}
		})
	}
}
