package session

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
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
	calls    atomic.Int64
	expected []string
	onProve  func()
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
	if k != 1 {
		panic("unexpected source capacity")
	}
	return provenance.ProofResult{Proof: "0x0102", PublicInputs: append([]string(nil), expected...)}, nil
}
func configureSourceFlow(t *testing.T, s *WalletSession, original *Operation, deny, badAnchor bool, historical ...bool) (*sourceFlowProver, *atomic.Int64) {
	t.Helper()
	const pin = "f90f2ce43218de158c6392fa19e6517dda703c35823a4aa5bfc20b50788ad282"
	bytes, err := os.ReadFile("../provenance/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := provenance.AuthenticateManifest(bytes, pin)
	if err != nil {
		t.Fatal(err)
	}
	scope := provenance.Scope{ChainID: original.plan.Context.ChainID.String(), ProtocolAddress: original.plan.Context.ProtocolAddress.Hex(), PolicyID: common.HexToHash("0x77").Hex()}
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
	key := ed25519.NewKeyFromSeed(make([]byte, 32))
	keyHash := sha256.Sum256(key.Public().(ed25519.PublicKey))
	keyID := "0x" + hex.EncodeToString(keyHash[:])
	head, err := s.options.Provider.HeaderByNumber(context.Background(), big.NewInt(10))
	if err != nil {
		t.Fatal(err)
	}
	cert := provenance.RootCertificate{CertificateVersion: "1", ProviderID: common.HexToHash("0x99").Hex(), SigningKeyID: keyID, PolicyID: scope.PolicyID, ChainID: scope.ChainID, ProtocolAddress: scope.ProtocolAddress, ChainGeneration: "1", RootSequence: "1", TreeID: "1", LeafCount: "1", Root: root, ConfirmedBlockNumber: "10", ConfirmedBlockHash: head.Hash().Hex(), DecisionLogHash: common.Hash{}.Hex(), PreviousCheckpointHash: common.Hash{}.Hex(), IssuedAt: "1"}
	if badAnchor {
		cert.ConfirmedBlockHash = common.HexToHash("0x1234").Hex()
	}
	msg, err := provenance.RootCertificateMessage(cert)
	if err != nil {
		t.Fatal(err)
	}
	cert.Signature = "0x" + hex.EncodeToString(ed25519.Sign(key, msg))
	trust := provenance.TrustContext{Scope: scope, ChainGeneration: "1", ProviderID: cert.ProviderID, ManifestSHA256: pin, SigningKeys: map[string]ed25519.PublicKey{keyID: key.Public().(ed25519.PublicKey)}}
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
		result := map[string]any{"scope": scope, "chainGeneration": "1", "providerId": cert.ProviderID, "signingKeyId": keyID, "manifestSha256": pin}
		switch req.Method {
		case "provenance_getStatus":
			result["checkpointSequence"] = latest.RootSequence
		case "provenance_getRoot":
			result["certificate"] = latest
			result["allowedForNewProof"] = true
		case "provenance_getRoots":
			result["items"] = []provenance.RootCertificate{cert, latest}
			result["nextCursor"] = ""
		case "provenance_getLeaves":
			result["rootId"] = "1:1"
			result["certificate"] = cert
			result["nextIndex"] = "1"
			result["items"] = []any{map[string]any{"scope": scope, "chainGeneration": "1", "treeId": "1", "leafIndex": "0", "noteCommitment": commitment, "leaf": leaf, "checkpointSequence": "1"}}
		case "provenance_evaluateTransaction":
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
			hi, _ := provenance.Uint(pi[9], 128)
			lo, _ := provenance.Uint(pi[10], 128)
			tracking.OperationDigest = common.BigToHash(new(big.Int).Or(new(big.Int).Lsh(hi, 128), lo)).Hex()
			decision := "allow"
			if deny {
				decision = "deny"
			}
			result["broadcastDecision"] = decision
			result["operations"] = []any{map[string]any{"operationIndex": "0", "operationDigest": tracking.OperationDigest, "broadcastDecision": decision, "tracking": tracking}}
		case "provenance_getTracking":
			result["tracking"] = tracking
		default:
			t.Errorf("unexpected source request %s; must not expose per-note interest", req.Method)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	prover := &sourceFlowProver{}
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

func TestProvenanceFinalFeeAndAutomaticAttachments(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	prover, evaluates := configureSourceFlow(t, s, original, false, false)
	c, fees := testBroadcaster(t, original, []int64{5, 8, 8}, false, common.HexToHash("0x77"))
	next, e := original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	if fees.Load() != 3 || prover.calls.Load() != 1 || evaluates.Load() != 1 {
		t.Fatal("source proof ran before final fee", fees.Load(), prover.calls.Load(), evaluates.Load())
	}
	words, e := next.execution().Words()
	if e != nil {
		t.Fatal(e)
	}
	values := make([]string, len(words))
	for i, w := range words {
		values[i] = common.Hash(w).Hex()
	}
	digest, e := provenance.OperationDigest(s.options.Provenance.Client.Trust.Scope, next.plan.Capacity, values)
	if e != nil {
		t.Fatal(e)
	}
	hi, _ := provenance.Uint(prover.expected[9], 128)
	lo, _ := provenance.Uint(prover.expected[10], 128)
	if common.BigToHash(new(big.Int).Or(new(big.Int).Lsh(hi, 128), lo)).Hex() != digest {
		t.Fatal("source proof bound an intermediate fee plan")
	}
	if next.plan.Broadcaster.Provenance == nil || len(next.plan.Broadcaster.Provenance.Sidecars) != 1 {
		t.Fatal("source evidence not persisted")
	}
	next, e = authorizeForTest(next, ctx)
	if e != nil {
		t.Fatal(e)
	}
	next, e = next.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = next.SubmitBroadcaster(ctx, c); e != nil {
		t.Fatal(e)
	}
	if prover.calls.Load() != 1 || evaluates.Load() != 1 {
		t.Fatal("submission regenerated source evidence")
	}
	s.options.Provenance = nil
	s.options.Provider.(sessionProvider).head.Time += 86400
	again, e := next.SubmitBroadcaster(ctx, c)
	if e != nil || again.State != "uncertain" {
		t.Fatal("known externalized status was gated by new-send freshness", e)
	}

}
func TestProvenanceDeniedAndNoncanonicalAnchorPreservePrepared(t *testing.T) {
	for _, badAnchor := range []bool{false, true} {
		t.Run(map[bool]string{false: "denied", true: "primary-anchor"}[badAnchor], func(t *testing.T) {
			ctx := context.Background()
			s, token, _ := testSession(t, uint64(time.Now().Unix()))
			address, _ := s.wallet.GetAddress()
			original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
			if e != nil {
				t.Fatal(e)
			}
			prover, _ := configureSourceFlow(t, s, original, true, badAnchor)
			c, _ := testBroadcaster(t, original, []int64{5}, false, common.HexToHash("0x77"))
			if _, e = original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)}); e == nil {
				t.Fatal("source rejection accepted")
			}
			if original.Status() != "prepared" || original.plan.Broadcaster != nil {
				t.Fatal("failed source preparation mutated handle")
			}
			if badAnchor && prover.calls.Load() != 0 {
				t.Fatal("proved before primary anchor authentication")
			}
			if e = original.checkLease(ctx); e != nil {
				t.Fatal("reservation lost", e)
			}
		})
	}
}

func TestProvenanceStoredAttachmentBindingAndRecovery(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	configureSourceFlow(t, s, original, false, false)
	c, _ := testBroadcaster(t, original, []int64{5}, false, common.HexToHash("0x77"))
	next, e := original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Resume(ctx, next.ID()); e != nil {
		t.Fatal("public evidence could not survive managed recovery", e)
	}
	source := next.plan.Broadcaster.Provenance
	for _, which := range []string{"index", "token", "nullifiers", "tracking"} {
		t.Run(which, func(t *testing.T) {
			oldIndex := source.Sidecars[0].OperationIndex
			oldPI := append([]string(nil), source.Sidecars[0].ProofBundle.PublicInputs...)
			oldTracking := source.Tracking[0]
			defer func() {
				source.Sidecars[0].OperationIndex = oldIndex
				source.Sidecars[0].ProofBundle.PublicInputs = oldPI
				source.Tracking[0] = oldTracking
			}()
			bundle := &source.Sidecars[0].ProofBundle
			switch which {
			case "index":
				source.Sidecars[0].OperationIndex = "1"
			case "token":
				bundle.PublicInputs[2] = common.HexToHash("0x42").Hex()
			case "nullifiers":
				bundle.PublicInputs[11] = common.HexToHash("0x42").Hex()
			case "tracking":
				source.Tracking[0].OperationDigest = common.HexToHash("0x42").Hex()
			}
			if which == "token" || which == "nullifiers" {
				pi := bundle.PublicInputs
				scope, e := provenance.ScopeHash(s.options.Provenance.Client.Trust.Scope)
				if e != nil {
					t.Fatal(e)
				}
				hi, _ := provenance.Uint(pi[9], 128)
				lo, _ := provenance.Uint(pi[10], 128)
				digest := new(big.Int).Or(new(big.Int).Lsh(hi, 128), lo)
				binding, e := provenance.BindingDigest(scope, pi[2], pi[5], pi[6], pi[7], pi[8], digest.String(), pi[11])
				if e != nil {
					t.Fatal(e)
				}
				n, _ := provenance.Field(binding)
				pi[12] = common.BigToHash(n).Hex()
			}
			if e := next.execution().ValidateBroadcasterPlan(); e == nil {
				t.Fatal("mutated retained source attachment accepted")
			}
		})
	}
}

func TestProvenanceTrackingExportIsDetached(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	configureSourceFlow(t, s, original, false, false)
	c, _ := testBroadcaster(t, original, []int64{5}, false, common.HexToHash("0x77"))
	next, e := original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	tracking, e := next.ExportProvenanceTracking(ctx)
	if e != nil || len(tracking) != 1 {
		t.Fatal(tracking, e)
	}
	tracking[0].OperationDigest = "changed"
	again, e := next.ExportProvenanceTracking(ctx)
	if e != nil || again[0].OperationDigest == "changed" {
		t.Fatal("tracking export aliases retained evidence", e)
	}
}

func TestProvenancePolicyShieldNeedsNoSourceCapability(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareShield(ctx, sdk.ShieldPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}}), Funding: sdk.FundingSource{Payer: s.wallet.OwnerAddress(), Amount: big.NewInt(30), Authorization: "caller"}})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := testBroadcaster(t, original, []int64{5}, false, common.HexToHash("0x77"))
	next, e := original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	if next.plan.Broadcaster.Provenance == nil || len(next.plan.Broadcaster.Provenance.Sidecars) != 0 {
		t.Fatal("no-proof operation unexpectedly required source proof")
	}
}

func mustTreeRoot(t *testing.T, tree *provenance.RecognizedTree) string {
	t.Helper()
	root, e := tree.Root()
	if e != nil {
		t.Fatal(e)
	}
	return root
}
func TestProvenanceHistoricalTreeAndCanceledLocalProver(t *testing.T) {
	for _, cancelProof := range []bool{false, true} {
		t.Run(map[bool]string{false: "historical-tree", true: "cancel"}[cancelProof], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			s, token, _ := testSession(t, uint64(time.Now().Unix()))
			address, _ := s.wallet.GetAddress()
			original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
			if e != nil {
				t.Fatal(e)
			}
			prover, _ := configureSourceFlow(t, s, original, false, false, true)
			if cancelProof {
				prover.onProve = cancel
			}
			c, _ := testBroadcaster(t, original, []int64{5}, false, common.HexToHash("0x77"))
			next, e := original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
			if cancelProof {
				if e == nil || original.Status() != "prepared" {
					t.Fatal("canceled source preparation advanced handle", e)
				}
				if e := original.checkLease(context.Background()); e != nil {
					t.Fatal(e)
				}
			} else if e != nil || next.plan.Broadcaster.Provenance.Sidecars[0].ProofBundle.RootCertificates[0].TreeID != "1" {
				t.Fatal("historical recognized tree unavailable", e)
			}
		})
	}
}

func TestProvenanceSubmitRechecksPrimaryCheckpoint(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	address, _ := s.wallet.GetAddress()
	original, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	configureSourceFlow(t, s, original, false, false)
	c, _ := testBroadcaster(t, original, []int64{5}, false, common.HexToHash("0x77"))
	next, e := original.WithBroadcaster(ctx, c, sdk.WithBroadcasterOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	next, e = authorizeForTest(next, ctx)
	if e != nil {
		t.Fatal(e)
	}
	next, e = next.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	s.options.Provider.(sessionProvider).head.Extra = []byte("different canonical block")
	if _, e = next.SubmitBroadcaster(ctx, c); e == nil || !strings.Contains(e.Error(), "root anchor") {
		t.Fatal("submission trusted stale source anchor", e)
	}
	if next.Status() != "ready" {
		t.Fatal("rejected submission advanced journal")
	}
}
