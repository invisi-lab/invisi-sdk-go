package session

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/provenance"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

// This prover exercises orchestration only; real Noir/BB acceptance is tested by the owned-chain harness.
type sourceFlowProver struct {
	calls     atomic.Int64
	expected  []string
	statement atomic.Value
	onProve   func()
}

func (p *sourceFlowProver) Prove(ctx context.Context, _ provenance.ManifestIdentity, k int, _ provenance.Witness, expected []string) (provenance.ProofResult, error) {
	p.calls.Add(1)
	if p.onProve != nil {
		p.onProve()
	}
	if e := ctx.Err(); e != nil {
		return provenance.ProofResult{}, e
	}
	p.expected = append([]string(nil), expected...)
	p.statement.Store(append([]string(nil), expected...))
	if k != 10 {
		panic("unexpected source capacity")
	}
	return provenance.ProofResult{Proof: "0x0102", PublicInputs: append([]string(nil), expected...)}, nil
}
func configureSourceFlow(t *testing.T, s *WalletSession, original *Operation, deny, badAnchor bool, historical ...bool) (*sourceFlowProver, *atomic.Int64) {
	t.Helper()
	return configureSourceFlowGeneration(t, s, original, deny, badAnchor, "1", historical...)
}
func configureSourceFlowGeneration(t *testing.T, s *WalletSession, original *Operation, deny, badAnchor bool, generation string, historical ...bool) (*sourceFlowProver, *atomic.Int64) {
	t.Helper()
	const pin = "6107789c5c0d24c4130bff05da6d389336feb5ccb247c4030eae5198ab05d762"
	bytes := provenance.ManifestBytes()
	manifest, err := provenance.AuthenticateManifest(bytes, pin)
	if err != nil {
		t.Fatal(err)
	}
	scope := provenance.Scope{ChainID: original.plan.Context.ChainID.String(), ProtocolAddress: original.plan.Context.ProtocolAddress.Hex()}
	tree := &provenance.RecognizedTree{}
	commitment := original.plan.Inputs[0].Commitment.Hex()
	leaf, err := provenance.RecognizedLeaf(mustScopeHash(t, scope), commitment)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tree.Append(leaf); err != nil {
		t.Fatal(err)
	}
	root, err := tree.Root()
	if err != nil {
		t.Fatal(err)
	}
	seed := sha256.Sum256([]byte("source-flow-signing-key-" + generation))
	key := ed25519.NewKeyFromSeed(seed[:])
	keyHash := sha256.Sum256(key.Public().(ed25519.PublicKey))
	keyID := "0x" + hex.EncodeToString(keyHash[:])
	head, err := s.options.Provider.HeaderByNumber(context.Background(), big.NewInt(10))
	if err != nil {
		t.Fatal(err)
	}
	cert := provenance.RootCertificate{CertificateVersion: "1", ProviderID: common.HexToHash("0x99").Hex(), SigningKeyID: keyID, ChainID: scope.ChainID, ProtocolAddress: scope.ProtocolAddress, ChainGeneration: generation, RootSequence: "1", TreeID: "1", LeafCount: "1", Root: root, ConfirmedBlockNumber: "10", ConfirmedBlockHash: head.Hash().Hex(), DecisionLogHash: common.Hash{}.Hex(), PreviousCheckpointHash: common.Hash{}.Hex(), IssuedAt: "1"}
	if badAnchor {
		cert.ConfirmedBlockHash = common.HexToHash("0x1234").Hex()
	}
	msg, err := provenance.RootCertificateMessage(cert)
	if err != nil {
		t.Fatal(err)
	}
	cert.Signature = "0x" + hex.EncodeToString(ed25519.Sign(key, msg))
	trust := provenance.TrustContext{Scope: scope, ChainGeneration: generation, ProviderID: cert.ProviderID, ManifestSHA256: pin, SigningKeys: map[string]ed25519.PublicKey{keyID: key.Public().(ed25519.PublicKey)}}
	evaluates := new(atomic.Int64)
	latest := cert
	if len(historical) > 0 && historical[0] {
		empty := &provenance.RecognizedTree{}
		latest.Root = mustTreeRoot(t, empty)
		latest.TreeID = "2"
		latest.LeafCount = "0"
		latest.RootSequence = "2"
		msg, e := provenance.RootCertificateMessage(latest)
		if e != nil {
			t.Fatal(e)
		}
		latest.Signature = "0x" + hex.EncodeToString(ed25519.Sign(key, msg))
	}
	tracking := provenance.Tracking{ID: strings.Repeat("a", 48), Status: "pending_confirmation", ExpiresAt: "4102444800"}
	prover := &sourceFlowProver{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		result := map[string]any{"scope": scope, "chainGeneration": generation, "providerId": cert.ProviderID, "signingKeyId": keyID, "manifestSha256": pin}
		switch req.Method {
		case "provenance_getStatus":
			result["checkpointSequence"] = latest.RootSequence
		case "provenance_getRoot":
			result["certificate"] = latest
			result["allowedForNewProof"] = true
			var query struct {
				RootID string `json:"rootId"`
			}
			json.Unmarshal(req.Params[0], &query)
			if query.RootID == generation+":1" {
				result["certificate"] = cert
				result["allowedForNewProof"] = len(historical) < 2 || !historical[1]
			}
		case "provenance_getRoots":
			result["items"] = []provenance.RootCertificate{cert, latest}
			result["nextCursor"] = ""
		case "provenance_getLeaves":
			result["rootId"] = generation + ":1"
			result["certificate"] = cert
			result["nextIndex"] = "1"
			result["items"] = []any{map[string]any{"scope": scope, "chainGeneration": generation, "treeId": "1", "leafIndex": "0", "noteCommitment": commitment, "leaf": leaf, "checkpointSequence": "1"}}
		case "provenance_evaluateTransaction", "provenance_submitSourceProof":
			evaluates.Add(1)
			var params struct {
				ProofSidecars []provenance.ProofSidecar `json:"proofSidecars"`
			}
			json.Unmarshal(req.Params[0], &params)
			if len(params.ProofSidecars) != 1 {
				t.Error("missing automatic source sidecar")
				return
			}
			pi := params.ProofSidecars[0].ProofBundle.PublicInputs
			hi, _ := provenance.Uint(pi[7], 128)
			lo, _ := provenance.Uint(pi[8], 128)
			tracking.OperationDigest = common.BigToHash(new(big.Int).Or(new(big.Int).Lsh(hi, 128), lo)).Hex()
			decision := "allow"
			if deny {
				decision = "deny"
			}
			result["broadcastDecision"] = decision
			result["operations"] = []any{map[string]any{"operationIndex": "0", "operationDigest": tracking.OperationDigest, "broadcastDecision": decision, "tracking": tracking}}
		case "provenance_getTracking":
			if raw := prover.statement.Load(); raw != nil {
				pi := raw.([]string)
				hi, _ := provenance.Uint(pi[7], 128)
				lo, _ := provenance.Uint(pi[8], 128)
				tracking.OperationDigest = common.BigToHash(new(big.Int).Or(new(big.Int).Lsh(hi, 128), lo)).Hex()
			}
			if deny {
				tracking.Status = "expired"
			}
			result["tracking"] = tracking
		default:
			t.Errorf("unexpected source request %s; must not expose per-note interest", req.Method)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	s.options.Provenance = &sdk.ProvenanceCapability{Client: &provenance.Client{Endpoint: server.URL, Trust: trust}, Manifest: manifest, Prover: prover, MaxRootPages: 1, MaxLeaves: 1}
	// Preparation already captured its runtime; this fixture installs the same explicit local capability.

	return prover, evaluates
}
func mustScopeHash(t *testing.T, s provenance.Scope) string {
	t.Helper()
	v, e := provenance.ScopeHash(s)
	if e != nil {
		t.Fatal(e)
	}
	return v
}

func mustTreeRoot(t *testing.T, tree *provenance.RecognizedTree) string {
	t.Helper()
	r, e := tree.Root()
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func readySourceOperation(t *testing.T, badAnchor bool) (*Operation, *sourceFlowProver) {
	t.Helper()
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	prover, _ := configureSourceFlow(t, s, original, false, badAnchor)
	next, e := authorizeForTest(original, ctx)
	if e != nil {
		t.Fatal(e)
	}
	next, e = next.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	return next, prover
}
func TestProvenanceProofRequiresReadyTransaction(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	prover, _ := configureSourceFlow(t, s, original, false, false)
	if _, e = original.ProveProvenance(ctx); e == nil {
		t.Fatal("unready Source proof accepted")
	}
	c, _ := testBroadcaster(t, original, []int64{5}, false, common.Hash{})
	if _, e = original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)}); e != nil {
		t.Fatal(e)
	}
	if prover.calls.Load() != 0 {
		t.Fatal("quotation generated Source proof")
	}
}
func TestProvenanceIndependentProofAndDetachedExport(t *testing.T) {
	ctx := context.Background()
	ready, prover := readySourceOperation(t, false)
	next, e := ready.ProveProvenance(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if next.plan.Broadcaster != nil || prover.calls.Load() != 1 {
		t.Fatal("Source proof depended on broadcaster")
	}
	proof, e := next.ExportProvenanceProof(ctx)
	if e != nil || len(proof.Sidecars) != 1 {
		t.Fatal(e)
	}
	if e = provenance.ValidateProvenanceProof(proof, next.session.options.Provenance.Client.Trust); e != nil {
		t.Fatal("public packet failed native validation", e)
	}
	mutated := proof
	mutated.PublicInputs = append([]string(nil), proof.PublicInputs...)
	mutated.PublicInputs[0] = common.Hash{}.Hex()
	if provenance.ValidateProvenanceProof(mutated, next.session.options.Provenance.Client.Trust) == nil {
		t.Fatal("mutated native public inputs accepted")
	}
	if _, e = next.SubmitProvenance(ctx); e != nil {
		t.Fatal("public Source proof requires transaction hash", e)
	}
	proof.Sidecars[0].OperationIndex = "99"
	again, e := next.ExportProvenanceProof(ctx)
	if e != nil || again.Sidecars[0].OperationIndex != "0" {
		t.Fatal("export aliases retained proof", e)
	}
	if _, e = next.session.Resume(ctx, next.ID()); e != nil {
		t.Fatal("independent proof recovery", e)
	}
}
func TestProvenanceAuthenticatesPrimaryAnchorBeforeProving(t *testing.T) {
	ready, prover := readySourceOperation(t, true)
	if _, e := ready.ProveProvenance(context.Background()); e == nil {
		t.Fatal("noncanonical Source anchor accepted")
	}
	if prover.calls.Load() != 0 || ready.Status() != "ready" {
		t.Fatal("failed Source proof mutated ready transaction")
	}
}

func TestBroadcasterAutomaticallyProvesOnlyFinalReadySource(t *testing.T) {
	for _, execution := range []string{"direct", "aggregate"} {
		t.Run(execution, func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t, uint64(time.Now().Unix()))
			address, _ := s.wallet.GetAddress()
			original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
			if e != nil {
				t.Fatal(e)
			}
			prover, _ := configureSourceFlow(t, s, original, false, false)
			c, fees := testBroadcasterExecution(t, original, []int64{5, 8, 8}, false, nil, execution, nil, common.Hash{})
			quoted, e := original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
			if e != nil {
				t.Fatal(e)
			}
			if fees.Load() != 3 || prover.calls.Load() != 0 {
				t.Fatal("quotation crossed Source proof boundary")
			}
			ready, e := authorizeForTest(quoted, ctx)
			if e != nil {
				t.Fatal(e)
			}
			ready, e = ready.Prove(ctx)
			if e != nil {
				t.Fatal(e)
			}
			provider := ready.session.options.Provider
			ready.session.options.Provider = nil
			_, missingErr := ready.SubmitBroadcaster(ctx, c)
			var missing *sdk.MissingCapabilityError
			if !errors.As(missingErr, &missing) {
				t.Fatal("first broadcast without primary must fail with capability error", missingErr)
			}
			if prover.calls.Load() != 0 {
				t.Fatal("missing primary invoked prover")
			}
			ready.session.options.Provider = provider
			status, e := ready.SubmitBroadcaster(ctx, c)
			if e != nil {
				t.Fatal(e)
			}
			if status.ID != "server-request-id" || prover.calls.Load() != 1 || ready.plan.Provenance == nil {
				t.Fatal("ready Source evidence or server ID missing")
			}
			if _, e = ready.SubmitBroadcaster(ctx, c); e != nil || prover.calls.Load() != 1 {
				t.Fatal("known submission regenerated Source evidence", e)
			}
			if ready.plan.Provenance.Data != "0x"+common.Bytes2Hex(ready.data) {
				t.Fatal("Source proof bound to different public envelope")
			}
			ready.session.options.Prover = nil
			ready.session.options.Provider = nil
			ready.session.options.Provenance.Prover = nil
			ready, e = ready.session.Resume(ctx, ready.ID())
			if e != nil {
				t.Fatal("cached source restore failed", e)
			}
			if _, e = ready.SubmitBroadcaster(ctx, c); e != nil || prover.calls.Load() != 1 {
				t.Fatal("offline retry regenerated source", e)
			}
		})
	}

}

func TestProvenanceReproofRefreshesTrustWithoutChangingPayment(t *testing.T) {
	ctx := context.Background()
	ready, initialProver := readySourceOperation(t, false)
	ready, err := ready.ProveProvenance(ctx)
	if err != nil {
		t.Fatal(err)
	}
	original, err := ready.ExportProvenanceProof(ctx)
	if err != nil {
		t.Fatal(err)
	}
	markSourceConfirmed(t, ready)
	mainProof := append([]byte(nil), ready.bundle.Proof...)
	refreshed, _ := configureSourceFlowGeneration(t, ready.session, ready, false, false, "2")
	if _, err = ready.ProveProvenance(ctx); err == nil {
		t.Fatal("stale cached Source evidence accepted")
	}
	// Failed explicit reproof must retain prior evidence and the original payment for retry.
	beforeFailure, _ := ready.session.options.Store.ReadSnapshot(ctx, ready.session.scope)
	cancelCtx, cancel := context.WithCancel(ctx)
	refreshed.onProve = cancel
	if _, err = ready.ProveProvenance(cancelCtx, sdk.ProvenanceProveOptions{Reprove: true}); err == nil {
		t.Fatal("cancelled Source reproof succeeded")
	}
	afterFailure, _ := ready.session.options.Store.ReadSnapshot(ctx, ready.session.scope)
	if !reflect.DeepEqual(beforeFailure.Operations[ready.ID()], afterFailure.Operations[ready.ID()]) {
		t.Fatal("failed reproof changed journal")
	}
	unchanged, err := ready.ExportProvenanceProof(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original, unchanged) {
		t.Fatal("failed Source reproof replaced prior evidence")
	}
	refreshed.onProve = nil
	ready, err = ready.ProveProvenance(ctx, sdk.ProvenanceProveOptions{Reprove: true})
	if err != nil {
		t.Fatal(err)
	}
	current, err := ready.ExportProvenanceProof(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ready.status != "confirmed" {
		t.Fatal("reproof changed terminal status")
	}
	if original.Data != current.Data || original.OperationDigest != current.OperationDigest || !reflect.DeepEqual(original.PublicInputs, current.PublicInputs) || !bytes.Equal(mainProof, ready.bundle.Proof) {
		t.Fatal("Source reproof changed the original payment")
	}
	cert := current.Sidecars[0].ProofBundle.RootCertificates[0]
	oldCert := original.Sidecars[0].ProofBundle.RootCertificates[0]
	if cert.ChainGeneration != "2" || cert.SigningKeyID == oldCert.SigningKeyID || initialProver.calls.Load() != 1 || refreshed.calls.Load() != 2 {
		t.Fatal("Source reproof did not refresh generation/signing identity")
	}
	if err = provenance.ValidateProvenanceProof(current, ready.session.options.Provenance.Client.Trust); err != nil {
		t.Fatal(err)
	}
}

func TestProvenanceDeniedHistoricalRootNeverProves(t *testing.T) {
	ready, _ := readySourceOperation(t, false)
	backend, _ := configureSourceFlow(t, ready.session, ready, false, false, true, true)
	if _, err := ready.ProveProvenance(context.Background()); err == nil {
		t.Fatal("denied historical root accepted")
	}
	if backend.calls.Load() != 0 {
		t.Fatal("prover called with denied historical root")
	}
}

func TestProvenanceCachedProofNeedsOnlyTrustedClient(t *testing.T) {
	ctx := context.Background()
	ready, backend := readySourceOperation(t, false)
	ready, err := ready.ProveProvenance(ctx)
	if err != nil {
		t.Fatal(err)
	}
	original, err := ready.ExportProvenanceProof(ctx)
	if err != nil {
		t.Fatal(err)
	}
	markSourceConfirmed(t, ready)
	mainProof := append([]byte(nil), ready.bundle.Proof...)
	capability := ready.session.options.Provenance
	originalHTTP := capability.Client.HTTP
	capability.Client.HTTP = &http.Client{Transport: offlineSourceTransport{t}}
	capability.Prover = nil
	capability.Manifest = provenance.ManifestIdentity{}
	ready.session.options.Provider = nil
	next, err := ready.ProveProvenance(ctx)
	if err != nil {
		t.Fatal("cached proof requires proving assets or network", err)
	}
	got, err := next.ExportProvenanceProof(ctx)
	if err != nil || !reflect.DeepEqual(original, got) || !bytes.Equal(mainProof, next.bundle.Proof) || next.status != "confirmed" || backend.calls.Load() != 1 {
		t.Fatal("cached proof changed", err)
	}
	capability.Client.HTTP = originalHTTP
	if _, err := next.SubmitProvenance(ctx); err != nil {
		t.Fatal("client-only cached submit failed", err)
	}
	if backend.calls.Load() != 1 {
		t.Fatal("cached submission invoked prover")
	}
}

type offlineSourceTransport struct{ t *testing.T }

func (o offlineSourceTransport) RoundTrip(*http.Request) (*http.Response, error) {
	o.t.Error("cached proof performed a Source RPC")
	return nil, context.Canceled
}

func markSourceConfirmed(t *testing.T, op *Operation) {
	t.Helper()
	// This fixture establishes terminal state and spent inputs; separate tests cover chain confirmation.
	err := op.session.options.Store.Update(context.Background(), op.session.scope, func(state *WalletState) error {
		record := state.Operations[op.ID()]
		record.Status = "confirmed"
		state.Operations[op.ID()] = record
		for key, note := range state.Notes {
			if note.ReservedBy == op.ID() {
				note.Spent = true
				note.ReservedBy = ""
				state.Notes[key] = note
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	op.status = "confirmed"
}

func TestProvenanceMissingPrimaryCapabilityPreservesJournal(t *testing.T) {
	for _, cached := range []bool{false, true} {
		t.Run(fmt.Sprint(cached), func(t *testing.T) {
			ctx := context.Background()
			ready, backend := readySourceOperation(t, false)
			var err error
			if cached {
				ready, err = ready.ProveProvenance(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}
			before, _ := ready.session.options.Store.ReadSnapshot(ctx, ready.session.scope)
			ready.session.options.Provider = nil
			_, err = ready.ProveProvenance(ctx, sdk.ProvenanceProveOptions{Reprove: cached})
			var missing *sdk.MissingCapabilityError
			if !errors.As(err, &missing) {
				t.Fatal("missing primary must return capability error", err)
			}
			after, _ := ready.session.options.Store.ReadSnapshot(ctx, ready.session.scope)
			if !reflect.DeepEqual(before.Operations[ready.ID()], after.Operations[ready.ID()]) {
				t.Fatal("failed proof changed journal")
			}
			expected := int64(0)
			if cached {
				expected = 1
			}
			if backend.calls.Load() != expected {
				t.Fatal("missing primary invoked prover")
			}
		})
	}
}
