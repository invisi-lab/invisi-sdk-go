package session

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func testBroadcaster(t *testing.T, p *Operation, fees []int64, failQuote bool, policies ...common.Hash) (*broadcaster.Client, *atomic.Int64) {
	t.Helper()
	return testBroadcasterQuote(t, p, fees, failQuote, nil, policies...)
}
func testBroadcasterQuote(t *testing.T, p *Operation, fees []int64, failQuote bool, transform func(*broadcaster.SignedQuote), policies ...common.Hash) (*broadcaster.Client, *atomic.Int64) {
	t.Helper()
	return testBroadcasterExecution(t, p, fees, failQuote, transform, "direct", nil, policies...)
}
func testBroadcasterExecution(t *testing.T, p *Operation, fees []int64, failQuote bool, transform func(*broadcaster.SignedQuote), execution string, observe func(string, json.RawMessage), policies ...common.Hash) (*broadcaster.Client, *atomic.Int64) {
	t.Helper()
	key, _ := crypto.GenerateKey()
	recipient := common.HexToAddress("0x8877")
	calls := new(atomic.Int64)
	cancelCalls := new(atomic.Int64)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Params) != 1 && req.Method != "broadcaster_getStatus" {
			t.Error("positional params required")
		}
		var param struct {
			Data           string
			Schedule       *broadcaster.Schedule `json:"schedule"`
			Quote          broadcaster.Quote     `json:"quote"`
			IdempotencyKey string
			ID             string
			ProofSidecars  []provenance.ProofSidecar
		}
		if len(req.Params) > 0 {
			json.Unmarshal(req.Params[0], &param)
		}
		if observe != nil && len(req.Params) > 0 {
			observe(req.Method, req.Params[0])
		}
		var result any
		switch req.Method {
		case "broadcaster_estimateQuote":
			j := int(calls.Add(1)) - 1
			if j >= len(fees) {
				j = len(fees) - 1
			}
			result = broadcaster.Fee{FeeRecipient: recipient, FeeAmount: big.NewInt(fees[j]).String(), MaxGas: "1000000"}
		case "broadcaster_getQuote":
			if failQuote {
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": -1, "message": "quote unavailable"}})
				return
			}
			data := common.FromHex(param.Data)
			i, e := broadcaster.DecodeIntent(data)
			if e != nil {
				t.Error(e)
				return
			}
			h, e := broadcaster.RequestHash(data)
			if e != nil {
				t.Error(e)
				return
			}
			if len(policies) > 0 {
				bound, err := provenance.ProvenanceIntentHash(h.Hex())
				if err != nil {
					t.Error(err)
					return
				}
				h = common.HexToHash(bound)
			}
			q := broadcaster.Quote{Schedule: param.Schedule, Token: i.Token, FeeRecipient: i.FeeRecipient, FeeAmount: i.FeeAmount.String(), Deadline: i.Deadline.String(), MaxGas: "1000000", RequestHash: h, Execution: execution}
			trust := broadcaster.Trust{ChainID: p.plan.Context.ChainID, Protocol: p.plan.Context.ProtocolAddress, Signer: crypto.PubkeyToAddress(key.PublicKey)}
			td, _ := broadcaster.QuoteTypedData(q, trust)
			digest, _, _ := apitypes.TypedDataAndHash(td)
			sig, _ := crypto.Sign(digest, key)
			signed := broadcaster.SignedQuote{Quote: q, Signature: "0x" + common.Bytes2Hex(sig)}
			if len(policies) > 0 {
				signed.ProvenanceRequired = true
				signed.RequestHashFormat = "invisi-broadcaster-provenance-intent-v1"
			}
			result = signed
			if transform != nil {
				transform(&signed)
			}
			result = signed
		case "broadcaster_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "chainId": p.plan.Context.ChainID.String(), "protocol": p.plan.Context.ProtocolAddress.Hex(), "sender": crypto.PubkeyToAddress(key.PublicKey).Hex(), "feeRecipient": recipient.Hex(), "submissionHalted": false, "scheduling": map[string]any{"enabled": true, "maxDelaySeconds": "1209600", "defaultExecutionWindowSeconds": "300", "maxExecutionWindowSeconds": "3600", "minDeadlineSlackSeconds": "30"}, "provenance": map[string]any{"enabled": len(policies) > 0, "requestHashFormat": broadcaster.ProvenanceRequestHashFormat, "trackingRequired": true, "trackingSupported": true, "aggregateSupported": true}}
		case "broadcaster_cancelTransaction":
			state := "cancelled"
			if cancelCalls.Add(1) == 1 {
				state = "scheduled"
			}
			result = broadcaster.Status{ID: param.ID, State: state}
		case "broadcaster_getTransactionStatus":
			if param.ID == "not-found" {
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": -32001, "message": "not found"}})
				return
			}
			state := "uncertain"
			switch param.ID {
			case "queued", "proving", "rejected", "expired":
				state = param.ID
			}
			result = broadcaster.Status{ID: param.ID, State: state}
		case "broadcaster_submitTransaction":
			if len(policies) > 0 {
				if len(param.ProofSidecars) != 1 || param.ProofSidecars[0].OperationIndex != "0" {
					t.Error("submit did not automatically attach retained source proof")
				}
			}
			result = broadcaster.Status{ID: "server-request-id", State: "uncertain"}
			if param.Quote.Schedule != nil {
				result = broadcaster.Status{ID: "opaque-server-job/17", State: "scheduled"}
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	c, e := broadcaster.NewClient(broadcaster.Configuration{Endpoint: server.URL, QuoteSigner: crypto.PubkeyToAddress(key.PublicKey), FeeRecipient: recipient}, broadcaster.Options{HTTPClient: server.Client()})
	if e != nil {
		t.Fatal(e)
	}
	return c, calls
}

func TestBroadcasterPrivateChangeConvergesAndFreezesBusinessOutput(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	before, e := frozen(original.plan)
	if e != nil {
		t.Fatal(e)
	}
	business := before.Outputs[0]
	c, calls := testBroadcaster(t, original, []int64{5, 8, 8}, false)
	next, e := original, c.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	if calls.Load() != 3 || next.Preview().FeeAmount.Int64() != 8 || next.Preview().FundingAmount.Sign() != 0 || next.plan.Outputs[1].Amount.Int64() != 62 {
		t.Fatal(next.Preview(), calls.Load())
	}
	if business.Commitment != next.plan.Outputs[0].Commitment || string(business.Delivery.Ciphertext) != string(next.plan.Outputs[0].Delivery.Ciphertext) || before.Outputs[1].Amount.Int64() != 70 {
		t.Fatal("business output changed during quotation")
	}
	quote, e := next.ExportBroadcasterQuote(ctx)
	if e != nil || quote.Context == nil {
		t.Fatal(e)
	}
	intent, e := next.ExportBroadcasterIntent(ctx)
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := broadcaster.DecodeIntent(intent)
	if e != nil || decoded.FeeAmount.Int64() != 8 {
		t.Fatal(decoded, e)
	}
	if e = c.Quote(ctx, next, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)}); e == nil {
		t.Fatal("quote replacement accepted")
	}
}
func TestBroadcasterQuoteFailurePreservesManagedReservations(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, original, []int64{5}, true)
	if e = c.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)}); e == nil {
		t.Fatal("failed quote accepted")
	}
	if e = original.checkLease(ctx); e != nil {
		t.Fatal("original lease lost", e)
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	for _, n := range state.Notes {
		if n.ReservedBy != original.ID() {
			t.Fatal("reservation released")
		}
	}
	if original.plan.Broadcaster != nil || original.plan.Outputs[1].Amount.Int64() != 70 {
		t.Fatal("original mutated")
	}
}
func TestBroadcasterShieldGrossFundingAndDecodedCurrentCalldata(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareShield(ctx, sdk.ShieldPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}}), Funding: sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(30), Authorization: "caller"}})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, original, []int64{5}, false)
	next, e := original, c.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	if next.plan.FundingMode != "signature" || next.plan.Funding.Amount.Int64() != 35 || next.plan.Outputs[0].Amount.Int64() != 30 {
		t.Fatal(next.Preview())
	}
	a, _ := protocol.ProtocolABI()
	data, _ := next.ExportBroadcasterIntent(ctx)
	m, e := a.MethodById(data[:4])
	if e != nil || m.RawName != "shield" {
		t.Fatal(e)
	}
	args, e := m.Inputs.Unpack(data[4:])
	if e != nil {
		t.Fatal(e)
	}
	if len(args) != 4 {
		t.Fatal(args)
	}
}

func TestBroadcasterReadySubmitKeepsUncertainReservations(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, original, []int64{5}, false)
	next, e := original, c.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.Submit(ctx, next); e == nil {
		t.Fatal("sent unready operation")
	}
	next, e = authorizeForTest(next, ctx)
	if e != nil {
		t.Fatal(e)
	}
	next, e = next.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	status, e := c.Submit(ctx, next)
	if e != nil || status.State != "uncertain" || next.Status() != "uncertain" {
		t.Fatal(status, e)
	}
	snapshot, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	record := snapshot.Operations[next.ID()]
	if !record.Externalized || record.Status != "uncertain" {
		t.Fatal(record)
	}
	for _, note := range snapshot.Notes {
		if note.ReservedBy != next.ID() {
			t.Fatal("uncertain send lost reservation")
		}
	}
}

func TestBroadcasterPrivateShortfallSelectsAndAtomicallyReservesAnotherNote(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(100)}})})
	if e != nil {
		t.Fatal(e)
	}
	note := original.plan.Inputs[0].Note
	note.Amount = primitives.ElementFromUint64(50)
	note.Nonce = primitives.ElementFromUint64(456)
	identity, _ := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
	commit, _ := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
	nf, _ := primitives.ComputeNoteNullifier(note.ChainID, note.ProtocolAddress, note.Nonce, commit)
	leaf := transaction.FieldHash(commit)
	s.options.Provider.(sessionProvider).leaves.Store(uint64(2), leaf)
	if e = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		state.Notes[leaf] = StoredNote{Note: note, Commitment: leaf, Nullifier: transaction.FieldHash(nf), TreeID: 2}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	bad, _ := testBroadcaster(t, original, []int64{5}, true)
	if e = bad.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)}); e == nil {
		t.Fatal("failed quote accepted")
	}
	snapshot, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if snapshot.Notes[leaf].ReservedBy != "" {
		t.Fatal("failed quote reserved additional input")
	}
	good, _ := testBroadcaster(t, original, []int64{5}, false)
	next, e := original, good.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	if len(next.plan.Inputs) != 2 || next.plan.Outputs[1].Amount.Int64() != 45 {
		t.Fatal(next.Preview())
	}
	snapshot, _ = s.options.Store.ReadSnapshot(ctx, s.scope)
	if snapshot.Notes[leaf].ReservedBy != original.ID() || len(snapshot.Operations[original.ID()].Inputs) != 2 {
		t.Fatal("additional input not atomically reserved")
	}
	next, e = authorizeForTest(next, ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = next.Prove(ctx); e != nil {
		t.Fatal(e)
	}
}

func TestBroadcasterDetachedHandleAndPermitCallbackFinalGross(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	op, e := s.PrepareShield(ctx, sdk.ShieldPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}}), Funding: sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(30), Authorization: "caller", Permit: &protocol.TokenPermit{Deadline: big.NewInt(time.Now().Unix() + 1000)}}})
	if e != nil {
		t.Fatal(e)
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	prepared := detachedForTest(t, op, state)
	failing, _ := testBroadcaster(t, op, []int64{5}, true)
	if err := failing.Quote(ctx, prepared, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10), FundingPermit: func(context.Context, sdk.FundingPermitRequest) (*protocol.TokenPermit, error) {
		return &protocol.TokenPermit{Deadline: big.NewInt(time.Now().Unix() + 1000)}, nil
	}}); err == nil {
		t.Fatal("detached failed quote accepted")
	}
	if !transaction.PreparedIsCurrent(prepared) || prepared.Preview().FeeAmount.Sign() != 0 {
		t.Fatal("detached failed quote changed original handle")
	}
	c, _ := testBroadcaster(t, op, []int64{5, 8, 8, 8}, false)
	called := 0
	stale := *prepared
	quoted, e := prepared, c.Quote(ctx, prepared, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10), FundingPermit: func(ctx context.Context, r sdk.FundingPermitRequest) (*protocol.TokenPermit, error) {
		called++
		if r.Amount.Int64() != 38 || r.Payer != s.wallet.OwnerAddress() || r.TokenAddress != token || r.ProtocolAddress != s.scope.ProtocolAddress {
			t.Fatal(r)
		}
		return &protocol.TokenPermit{Deadline: big.NewInt(time.Now().Unix() + 1000)}, nil
	}})
	if e != nil {
		t.Fatal(e)
	}
	if called != 1 || quoted.Preview().FundingAmount.Int64() != 38 {
		t.Fatal(called, quoted.Preview())
	}
	if _, e = stale.ExportBroadcasterQuote(ctx); e == nil {
		t.Fatal("stale prepared handle accepted")
	}
	q, e := quoted.ExportBroadcasterQuote(ctx)
	if e != nil || q.Context == nil {
		t.Fatal(e)
	}
}

type broadcasterFailTransport struct{ base http.RoundTripper }

func (f broadcasterFailTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	body, e := io.ReadAll(r.Body)
	if e != nil {
		return nil, e
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	var req struct{ Method string }
	json.Unmarshal(body, &req)
	if req.Method == "broadcaster_submitTransaction" {
		return nil, errors.New("simulated uncertain network failure")
	}
	if f.base != nil {
		return f.base.RoundTrip(r)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return transport.RoundTrip(r)
}
func TestBroadcasterTransportFailureStaysExternalizedAndResumeRetainsQuote(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	op, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, op, []int64{5}, false)
	e = c.Quote(ctx, op, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	op, e = s.Resume(ctx, op.ID())
	if e != nil {
		t.Fatal("durable quote lost during resume", e)
	}
	quote, e := op.ExportBroadcasterQuote(ctx)
	if e != nil || quote.Context == nil {
		t.Fatal(e)
	}
	op, e = authorizeForTest(op, ctx)
	if e != nil {
		t.Fatal(e)
	}
	op, e = op.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	failing, e := broadcaster.NewClient(c.Configuration(), broadcaster.Options{HTTPClient: &http.Client{Transport: broadcasterFailTransport{}}})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = failing.Submit(ctx, op); e == nil {
		t.Fatal("network error swallowed")
	}
	snapshot, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	if !snapshot.Operations[op.ID()].Externalized || op.Status() != "uncertain" {
		t.Fatal("uncertain network call not retained")
	}
}

func TestBroadcasterRejectsAggregateBeforeBackendAndPreservesHandle(t *testing.T) {
	ctx := context.Background()
	s, token, backend := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	op, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, op, []int64{5}, false)
	e = c.Quote(ctx, op, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	op, e = authorizeForTest(op, ctx)
	if e != nil {
		t.Fatal(e)
	}
	before := len(backend.requests)
	if _, e = op.Prove(ctx, sdk.TransactionProveOptions{Execution: "aggregate"}); e == nil {
		t.Fatal("quoted aggregate accepted")
	}
	if len(backend.requests) != before || op.plan.Target == "noir-recursive" {
		t.Fatal("failed aggregate invoked prover or mutated handle")
	}
	if _, e = op.Prove(ctx); e != nil {
		t.Fatal("original handle invalidated", e)
	}
}

func TestBroadcasterDetachedFailedSubmitRetainsInspectableUncertainty(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	op, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	snapshot, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	prepared := detachedForTest(t, op, snapshot)
	c, _ := testBroadcaster(t, op, []int64{5}, false)
	e = c.Quote(ctx, prepared, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	prepared, e = externallyAuthorize(ctx, prepared)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = prepared.Prove(ctx, sdk.TransactionProveOptions{Execution: "aggregate"}); e == nil {
		t.Fatal("detached quoted aggregate accepted")
	}
	prepared, e = prepared.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	failing, _ := broadcaster.NewClient(c.Configuration(), broadcaster.Options{HTTPClient: &http.Client{Transport: broadcasterFailTransport{}}})
	if _, e = failing.Submit(ctx, prepared); e == nil {
		t.Fatal("transport failure swallowed")
	}
	if prepared.Preview().Status != "uncertain" {
		t.Fatal(prepared.Preview())
	}
	if _, e = prepared.ExportBroadcasterQuote(ctx); e != nil {
		t.Fatal("uncertain handle cannot inspect saved quote", e)
	}
	if status, err := c.Submit(ctx, prepared); err != nil || status.State != "uncertain" {
		t.Fatal("retained request retry failed", status, err)
	}
}

func TestScheduledCancellationRefreshesAndReleasesReservations(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{Deadline: uint64(time.Now().Add(9 * 24 * time.Hour).Unix()), TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, original, []int64{5}, false)
	after := time.Now().Add(8 * 24 * time.Hour).Unix()
	next, e := original, c.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10), Schedule: &sdk.ScheduleOptions{PublishAfter: strconv.FormatInt(after, 10)}})
	if e != nil {
		t.Fatal(e)
	}
	saved := next.plan.Broadcaster
	if saved.CancelToken == "" || saved.Signed.Quote.Schedule == nil {
		t.Fatal("missing local cancellation handle")
	}
	a, _ := strconv.ParseUint(saved.Signed.Quote.Schedule.ExecuteBefore, 10, 64)
	if a != uint64(after+300) {
		t.Fatal("incorrect default execution window")
	}
	if saved.RequestID != "" {
		t.Fatal("SDK invented a broadcaster ID before submission")
	}
	if _, err := c.Cancel(ctx, next); err == nil {
		t.Fatal("unsubmitted cancellation created a remote job")
	}
	next, e = authorizeForTest(next, ctx)
	if e != nil {
		t.Fatal(e)
	}
	next, e = next.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	admitted, e := c.Submit(ctx, next)
	if e != nil || admitted.ID != "opaque-server-job/17" {
		t.Fatal("opaque server ID rejected", admitted, e)
	}
	persisted, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	var admittedPlan transaction.Plan
	if err := json.Unmarshal(persisted.Operations[next.ID()].Plan, &admittedPlan); err != nil || admittedPlan.Broadcaster.RequestID != admitted.ID {
		t.Fatal("server ID not durable before response", err)
	}

	status, e := c.Cancel(ctx, next)
	if e == nil || status.State != "scheduled" {
		t.Fatal(status, e)
	}
	retained, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	for _, input := range next.plan.Inputs {
		if retained.Notes[input.Commitment].ReservedBy != next.ID() {
			t.Fatal("nonterminal cancellation released input")
		}
	}
	status, e = c.Cancel(ctx, next)
	if e != nil || status.State != "cancelled" {
		t.Fatal(status, e)
	}
	state, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	if state.Operations[next.ID()].Status != "cancelled" {
		t.Fatal("cancellation not persisted")
	}
	for _, note := range state.Notes {
		if note.ReservedBy == next.ID() {
			t.Fatal("reservation retained after canonical cancellation")
		}
	}
	if e = restoreReservations(&state); e != nil {
		t.Fatal(e)
	}
	for _, note := range state.Notes {
		if note.ReservedBy == next.ID() {
			t.Fatal("cancellation lost during restore")
		}
	}
}

func TestScheduledDeadlineMustCoverWindowAndSlack(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, original, []int64{5}, false)
	after := time.Now().Add(8 * 24 * time.Hour).Unix()
	if e = c.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10), Schedule: &sdk.ScheduleOptions{PublishAfter: strconv.FormatInt(after, 10)}}); e == nil {
		t.Fatal("accepted native deadline before execution window")
	}
}

type failedCancellationRefresh struct{ sdk.Provider }

func (failedCancellationRefresh) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	return nil, errors.New("temporary canonical refresh unavailable")
}

type rejectCancellationTransport struct{}

func (rejectCancellationTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("broadcaster offline after acknowledged cancellation")
}

type dropScheduledSubmitResponse struct{}

func (dropScheduledSubmitResponse) RoundTrip(r *http.Request) (*http.Response, error) {
	raw, e := io.ReadAll(r.Body)
	if e != nil {
		return nil, e
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	var req struct{ Method string }
	json.Unmarshal(raw, &req)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response, e := transport.RoundTrip(r)
	if req.Method == "broadcaster_submitTransaction" && e == nil {
		response.Body.Close()
		return nil, errors.New("lost response after server admission")
	}
	return response, e
}
func TestScheduledLostSubmitRetainsScopedHandleAndCancellationAcknowledgement(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{Deadline: uint64(time.Now().Add(9 * 24 * time.Hour).Unix()), TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	var payloadMu sync.Mutex
	var payloads [][]byte
	var cancellationCalls, statusCalls atomic.Int64
	c, _ := testBroadcasterExecution(t, original, []int64{5}, false, nil, "direct", func(method string, raw json.RawMessage) {
		if method == "broadcaster_submitTransaction" {
			snapshot, err := s.options.Store.ReadSnapshot(ctx, s.scope)
			if err != nil {
				t.Error(err)
			} else {
				record := snapshot.Operations[original.ID()]
				var plan transaction.Plan
				if err = json.Unmarshal(record.Plan, &plan); err != nil {
					t.Error(err)
				} else if plan.Broadcaster == nil || plan.Broadcaster.Request == nil || plan.Broadcaster.CancelToken == "" || plan.Broadcaster.Signed.Signature == "" || plan.Broadcaster.RequestID != "" || !record.Externalized {
					t.Error("secret, signed quote and unknown-ID frozen request were not durable before network send")
				}
			}
			payloadMu.Lock()
			payloads = append(payloads, append([]byte(nil), raw...))
			payloadMu.Unlock()
		}
		if method == "broadcaster_cancelTransaction" {
			cancellationCalls.Add(1)
		}
		if method == "broadcaster_getTransactionStatus" {
			statusCalls.Add(1)
		}
	})
	after := time.Now().Add(8 * 24 * time.Hour).Unix()
	op, e := original, c.Quote(ctx, original, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10), Schedule: &sdk.ScheduleOptions{PublishAfter: strconv.FormatInt(after, 10)}})
	if e != nil {
		t.Fatal(e)
	}
	requestID := op.plan.Broadcaster.RequestID
	if requestID != "" || op.plan.Broadcaster.CancelToken == "" {
		t.Fatal("ID must be opaque and unknown before submit; secret must be held locally")
	}
	op, e = authorizeForTest(op, ctx)
	if e != nil {
		t.Fatal(e)
	}
	op, e = op.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	failing, e := broadcaster.NewClient(c.Configuration(), broadcaster.Options{HTTPClient: &http.Client{Transport: dropScheduledSubmitResponse{}}})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = failing.Submit(ctx, op); e == nil {
		t.Fatal("lost submit response swallowed")
	}
	if op.plan.Broadcaster.RequestID != requestID || op.status != "uncertain" {
		t.Fatal("lost scoped cancellation handle")
	}
	if op.plan.Broadcaster.Request == nil {
		t.Fatal("lost response discarded frozen submission")
	}
	if _, e = c.Cancel(ctx, op); e == nil {
		t.Fatal("unknown ID cancellation did not require explicit recovery")
	}
	if cancellationCalls.Load() != 0 || statusCalls.Load() != 0 {
		t.Fatal("SDK guessed opaque ID or silently admitted during cancellation")
	}
	recovered, e := c.Submit(ctx, op)
	if e != nil || recovered.ID != "opaque-server-job/17" || op.plan.Broadcaster.RequestID != recovered.ID {
		t.Fatal("frozen replay did not recover server ID", recovered, e)
	}
	payloadMu.Lock()
	same := len(payloads) == 2 && bytes.Equal(payloads[0], payloads[1])
	payloadMu.Unlock()
	if !same {
		t.Fatal("recovery changed frozen submission")
	}
	if status, err := c.Cancel(ctx, op); err == nil || status.State != "scheduled" {
		t.Fatal(status, err)
	}
	provider := op.session.options.Provider
	op.session.options.Provider = failedCancellationRefresh{provider}
	result, e := c.Cancel(ctx, op)
	if e == nil || result.State != "cancelled" {
		t.Fatal("refresh outage not surfaced", result, e)
	}
	snapshot, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	var persisted transaction.Plan
	if e = json.Unmarshal(snapshot.Operations[op.ID()].Plan, &persisted); e != nil {
		t.Fatal(e)
	}
	if !persisted.Broadcaster.CancellationConfirmed {
		t.Fatal("acknowledgement lost before canonical refresh")
	}
	for _, input := range op.plan.Inputs {
		if snapshot.Notes[input.Commitment].ReservedBy != op.ID() {
			t.Fatal("refresh outage released input")
		}
	}
	op.session.options.Provider = provider
	op, e = s.Resume(ctx, op.ID())
	if e != nil {
		t.Fatal(e)
	}
	offline, e := broadcaster.NewClient(c.Configuration(), broadcaster.Options{HTTPClient: &http.Client{Transport: rejectCancellationTransport{}}})
	if e != nil {
		t.Fatal(e)
	}
	if result, e = offline.Cancel(ctx, op); e != nil || result.State != "cancelled" {
		t.Fatal("persisted acknowledgement could not reconcile offline", result, e)
	}
	snapshot, _ = s.options.Store.ReadSnapshot(ctx, s.scope)
	for _, input := range op.plan.Inputs {
		if snapshot.Notes[input.Commitment].ReservedBy == op.ID() {
			t.Fatal("confirmed cancellation retained reservation")
		}
	}
}
