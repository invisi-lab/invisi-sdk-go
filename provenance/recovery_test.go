package provenance

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type recoveryMemory struct{ values map[string]RecoveryCiphertext }

func (m *recoveryMemory) Read(_ context.Context, k string) (RecoveryCiphertext, bool, error) {
	r, ok := m.values[k]
	return r, ok, nil
}
func (m *recoveryMemory) CompareAndSwap(_ context.Context, k string, old uint64, next RecoveryCiphertext) (bool, error) {
	if m.values[k].Revision != old {
		return false, nil
	}
	m.values[k] = next
	return true, nil
}
func TestRecoveryEncryption(t *testing.T) {
	ctx := context.Background()
	kv := &recoveryMemory{map[string]RecoveryCiphertext{}}
	store, e := NewEncryptedRecoveryStore(kv, bytes.Repeat([]byte{7}, 32))
	if e != nil {
		t.Fatal(e)
	}
	r := RecoveryRecord{Revision: 1, Intent: FrozenProvenanceIntent{Calldata: "0xfeed"}}
	if ok, e := store.CompareAndSwap(ctx, "one", 0, r); !ok || e != nil {
		t.Fatal(e)
	}
	key := "Invisi.Provenance.Recovery.v1:one"
	if bytes.Contains(kv.values[key].Data, []byte("0xfeed")) {
		t.Fatal("plaintext stored")
	}
	if _, ok, e := store.Load(ctx, "one"); !ok || e != nil {
		t.Fatal(e)
	}
	if ok, _ := store.CompareAndSwap(ctx, "one", 0, r); ok {
		t.Fatal("stale CAS accepted")
	}
	wrong, _ := NewEncryptedRecoveryStore(kv, bytes.Repeat([]byte{8}, 32))
	if _, _, e := wrong.Load(ctx, "one"); e == nil {
		t.Fatal("wrong key accepted")
	}
	kv.values["Invisi.Provenance.Recovery.v1:two"] = kv.values[key]
	if _, _, e := store.Load(ctx, "two"); e == nil {
		t.Fatal("wrong AAD accepted")
	}
}

type recoveryFixture struct {
	IdempotencyLimits struct {
		MaxBaseSubmissionKey string `json:"maxBaseSubmissionKey"`
	} `json:"idempotencyLimits"`
	SubmissionKeys []string               `json:"submissionKeys"`
	Scope          Scope                  `json:"scope"`
	ProviderID     string                 `json:"providerId"`
	SigningKeyID   string                 `json:"signingKeyId"`
	PublicKey      string                 `json:"publicKey"`
	Manifest       string                 `json:"manifestSha256"`
	Intent         FrozenProvenanceIntent `json:"intent"`
	Sidecars       []ProofSidecar         `json:"sidecars"`
}
type recoveryHostFixture struct {
	pendingFirst bool
	fact         RecoveryReconciliation
	verifyIntent FrozenProvenanceIntent
	next         ProofSidecar
	reconciles   int
}

func (h *recoveryHostFixture) VerifyFrozenStatement(_ context.Context, intent FrozenProvenanceIntent, _ ProofSidecar) error {
	if !reflect.DeepEqual(intent, h.verifyIntent) {
		return errors.New("host statement mismatch")
	}
	return nil
}
func (h *recoveryHostFixture) ReconcileFrozenTransaction(_ context.Context, _ RecoveryRecord) (RecoveryReconciliation, error) {
	h.reconciles++
	if h.pendingFirst {
		h.pendingFirst = false
		fact := h.fact
		fact.Status = "pending"
		return fact, nil
	}
	return h.fact, nil
}
func (h *recoveryHostFixture) LocalReprove(_ context.Context, _ RecoveryRecord, _ string) (ProofSidecar, error) {
	return h.next, nil
}

type recoveryServiceFixture struct {
	jobStatus         string
	recognitionOutage bool
	intent            FrozenProvenanceIntent
	submits           int
	interrupt         bool
	recognition       string
	keys              []string
}

func (s *recoveryServiceFixture) SubmitProof(_ context.Context, _ map[string]any, _ ProofBundle, key string) (Result, error) {
	s.submits++
	s.keys = append(s.keys, key)
	if s.interrupt {
		s.interrupt = false
		return nil, errors.New("network interrupted after accepted")
	}
	return Result{"jobId": json.RawMessage(`"job1"`)}, nil
}
func (s *recoveryServiceFixture) GetProofJob(context.Context, string) (Result, error) {
	status := s.jobStatus
	if status == "" {
		status = "valid"
	}
	raw, _ := json.Marshal(status)
	return Result{"status": raw}, nil
}
func (s *recoveryServiceFixture) GetOperationRecognition(_ context.Context, locator map[string]any) (Result, error) {
	if s.recognitionOutage {
		s.recognitionOutage = false
		return nil, errors.New("provider outage")
	}
	state, _ := json.Marshal(s.recognition)
	operation := map[string]any{"chainId": s.intent.ChainID, "protocolAddress": s.intent.ProtocolAddress, "tokenAddress": s.intent.TokenAddress, "operationDigest": s.intent.OperationDigest, "noteNullifiers": s.intent.InputNullifiers, "canonical": true, "success": true}
	for k, v := range locator {
		operation[k] = v
	}
	op, _ := json.Marshal(operation)
	return Result{"operation": op, "recognitionState": state, "recognition": json.RawMessage(`{"reasonCode":"DEPENDENCIES_PENDING","checkpointSequence":"3"}`)}, nil
}
func setupRecovery(t *testing.T) (*RecoveryCoordinator, *recoveryHostFixture, *recoveryServiceFixture, recoveryFixture) {
	t.Helper()
	data, e := os.ReadFile("testdata/recovery-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var v recoveryFixture
	if e = json.Unmarshal(data, &v); e != nil {
		t.Fatal(e)
	}
	kv := &recoveryMemory{map[string]RecoveryCiphertext{}}
	store, _ := NewEncryptedRecoveryStore(kv, bytes.Repeat([]byte{7}, 32))
	pub, _ := hex.DecodeString(v.PublicKey[2:])
	tx := "0x" + strings.Repeat("ab", 32)
	host := &recoveryHostFixture{fact: RecoveryReconciliation{Status: "confirmed", CalldataHash: v.Intent.CalldataHash, SignedRequestHash: v.Intent.SignedRequestHash, ChainGeneration: "1", TransactionHash: tx, Locator: map[string]string{"txHash": tx, "blockHash": "0x" + strings.Repeat("cd", 32), "operationIndex": "0"}}, verifyIntent: v.Intent, next: v.Sidecars[1]}
	service := &recoveryServiceFixture{recognition: "pending_source", intent: v.Intent}
	return &RecoveryCoordinator{Store: store, Trust: TrustContext{Scope: v.Scope, ChainGeneration: "1", ProviderID: v.ProviderID, ManifestSHA256: v.Manifest, SigningKeys: map[string]ed25519.PublicKey{v.SigningKeyID: pub}}, Service: service, Host: host}, host, service, v
}
func TestRecoveryRestartAndRecognition(t *testing.T) {
	c, h, s, v := setupRecovery(t)
	ctx := context.Background()
	r, e := c.PrepareDirect(ctx, "send", v.Intent, v.Sidecars[0], "stable")
	if e != nil || r.State != "broadcast_possible" {
		t.Fatal(e)
	}
	if _, e = c.MarkBroadcast(ctx, "send", h.fact.TransactionHash); e != nil {
		t.Fatal(e)
	}
	s.interrupt = true
	if _, e = c.Resume(ctx, "send", false); e == nil {
		t.Fatal("network failure missing")
	}
	restart := *c
	r, e = restart.Resume(ctx, "send", false)
	if e != nil || r.State != "pending_source" || !r.ProofValid {
		t.Fatal(r.State, e)
	}
	if !reflect.DeepEqual(s.keys, []string{mustRecoveryKey(t, "stable", v.Sidecars[0]), mustRecoveryKey(t, "stable", v.Sidecars[0])}) {
		t.Fatal("idempotency changed")
	}
	s.recognition = "recognized"
	r, e = restart.Run(ctx, "send", false, time.Millisecond, 2)
	if e != nil || r.State != "recognized" || r.CheckpointSequence != "3" || s.submits != 2 {
		t.Fatal(r.State, e)
	}
	if !reflect.DeepEqual(r.Intent, v.Intent) {
		t.Fatal("frozen intent changed")
	}
}
func TestRecoveryCancellationGenerationAndMismatch(t *testing.T) {
	c, h, s, v := setupRecovery(t)
	ctx := context.Background()
	if _, e := c.PrepareDirect(ctx, "send", v.Intent, v.Sidecars[0], "stable"); e != nil {
		t.Fatal(e)
	}
	if _, e := c.MarkBroadcast(ctx, "send", h.fact.TransactionHash); e != nil {
		t.Fatal(e)
	}
	stop, cancel := context.WithCancel(ctx)
	cancel()
	if _, e := c.Resume(stop, "send", false); e == nil {
		t.Fatal("cancel missing")
	}
	r, _, _ := c.Store.Load(ctx, "send")
	if r.State != "broadcast" || s.submits != 0 {
		t.Fatal("cancellation affected broadcast")
	}
	h.fact.ChainGeneration = "2"
	r, e := c.Resume(ctx, "send", false)
	if e != nil || r.State != "needs_reproof" || h.reconciles != 1 || s.submits != 0 {
		t.Fatal(r.State, e)
	}
	c.Trust.ChainGeneration = "2"
	copyRecord := recoveryClone(RecoveryRecord{Sidecar: v.Sidecars[1]})
	badSidecar := copyRecord.Sidecar
	badSidecar.ProofBundle.PublicInputs[2], _ = word("3")
	h.next = badSidecar
	if _, e = c.Resume(ctx, "send", true); e == nil {
		t.Fatal("reprover changed frozen token")
	}
	unchanged, _, _ := c.Store.Load(ctx, "send")
	if unchanged.ChainGeneration != "1" || s.submits != 0 {
		t.Fatal("bad reprover escaped")
	}
	h.next = v.Sidecars[1]
	s.recognition = "recognized"
	r, e = c.Resume(ctx, "send", true)
	if e != nil || r.State != "recognized" || !reflect.DeepEqual(s.keys, []string{mustRecoveryKey(t, "stable", v.Sidecars[1])}) {
		t.Fatal(r.State, e)
	}
	bad := v.Intent
	bad.OperationDigest = "5"
	if _, e := c.PrepareDirect(ctx, "bad", bad, v.Sidecars[1], "bad"); e == nil {
		t.Fatal("mismatched statement accepted")
	}
}

func mustRecoveryKey(t *testing.T, base string, s ProofSidecar) string {
	t.Helper()
	key, e := RecoverySubmissionKey(base, s)
	if e != nil {
		t.Fatal(e)
	}
	return key
}

type rejectedRecoveryStore struct{ RecoveryAtomicStore }

func (s rejectedRecoveryStore) CompareAndSwap(context.Context, string, uint64, RecoveryRecord) (bool, error) {
	return false, nil
}
func TestRecoveryRefusesBroadcastOnPersistenceConflict(t *testing.T) {
	c, _, service, v := setupRecovery(t)
	original := c.Store
	c.Store = rejectedRecoveryStore{original}
	if _, e := c.PrepareDirect(context.Background(), "blocked", v.Intent, v.Sidecars[0], "stable"); e == nil {
		t.Fatal("broadcast allowed without atomic persistence")
	}
	if _, found, _ := original.Load(context.Background(), "blocked"); found || service.submits != 0 {
		t.Fatal("conflicted preparation escaped")
	}
}

func TestRecoveryCanonicalSubmissionKeys(t *testing.T) {
	_, _, _, v := setupRecovery(t)
	for i, s := range v.Sidecars {
		if mustRecoveryKey(t, "stable", s) != v.SubmissionKeys[i] {
			t.Fatal("cross-language canonical bundle key mismatch")
		}
	}
}

func TestRecoveryWaitsAfterPermitCrashWithoutMarkBroadcast(t *testing.T) {
	c, h, s, v := setupRecovery(t)
	if _, e := c.PrepareDirect(context.Background(), "possible", v.Intent, v.Sidecars[0], "stable"); e != nil {
		t.Fatal(e)
	}
	h.pendingFirst = true
	s.recognition = "recognized"
	r, e := c.Run(context.Background(), "possible", false, time.Millisecond, 2)
	if e != nil || r.State != "recognized" || h.reconciles != 2 || s.submits != 1 || r.TransactionHash != h.fact.TransactionHash {
		t.Fatal(r.State, e)
	}
}

func TestRecoverySourceOutageAndDecisionResume(t *testing.T) {
	c, _, s, v := setupRecovery(t)
	ctx := context.Background()
	if _, e := c.PrepareDirect(ctx, "pending", v.Intent, v.Sidecars[0], "stable"); e != nil {
		t.Fatal(e)
	}
	s.recognitionOutage = true
	if _, e := c.Resume(ctx, "pending", false); e == nil {
		t.Fatal("provider outage missing")
	}
	r, _, _ := c.Store.Load(ctx, "pending")
	if !r.ProofValid || r.State != "pending_source" {
		t.Fatal("proof-valid resumable state lost")
	}
	s.recognition = "source_rejected"
	if r, e := c.Resume(ctx, "pending", false); e != nil || r.State != "source_rejected" {
		t.Fatal(e)
	}
	s.recognition = "recognized"
	if r, e := c.Resume(ctx, "pending", false); e != nil || r.State != "recognized" || s.submits != 1 {
		t.Fatal(e)
	}
}

func TestRecoveryIdempotencyBoundaryBeforePermit(t *testing.T) {
	c, _, _, v := setupRecovery(t)
	max := strings.Repeat("a", 63)
	key, e := RecoverySubmissionKey(max, v.Sidecars[0])
	if e != nil || len(key) != 128 {
		t.Fatal("derived key exceeds service limit", e)
	}
	if _, e = c.PrepareDirect(context.Background(), "max", v.Intent, v.Sidecars[0], max); e != nil {
		t.Fatal(e)
	}
	for _, base := range []string{strings.Repeat("a", 64), strings.Repeat("a", 128), "é", " "} {
		if _, e = c.PrepareDirect(context.Background(), "blocked", v.Intent, v.Sidecars[0], base); e == nil {
			t.Fatal("invalid base escaped into broadcast permit")
		}
		if _, ok, _ := c.Store.Load(context.Background(), "blocked"); ok {
			t.Fatal("invalid base persisted")
		}
	}
}

func TestRecoverySameGenerationReplacementKey(t *testing.T) {
	_, _, _, v := setupRecovery(t)
	sidecar := recoveryClone(RecoveryRecord{Sidecar: v.Sidecars[0]}).Sidecar
	old := mustRecoveryKey(t, "stable", sidecar)
	sidecar.ProofBundle.Proof = "0x02"
	if old == mustRecoveryKey(t, "stable", sidecar) {
		t.Fatal("replacement bundle reused submission key")
	}
}

func TestRecoveryRealJobWireStates(t *testing.T) {
	ctx := context.Background()
	c, h, s, v := setupRecovery(t)
	if _, e := c.PrepareDirect(ctx, "invalid", v.Intent, v.Sidecars[0], "stable"); e != nil {
		t.Fatal(e)
	}
	s.jobStatus = "invalid"
	r, e := c.Resume(ctx, "invalid", false)
	if e != nil || r.State != "proof_rejected" {
		t.Fatal("invalid proof not rejected", r.State, e)
	}
	next := recoveryClone(RecoveryRecord{Sidecar: v.Sidecars[0]}).Sidecar
	next.ProofBundle.Proof = "0x02"
	h.next = next
	s.jobStatus = "valid"
	s.recognition = "recognized"
	r, e = c.Resume(ctx, "invalid", true)
	if e != nil || r.State != "recognized" || !reflect.DeepEqual(r.Intent, v.Intent) || len(s.keys) != 2 || s.keys[0] == s.keys[1] {
		t.Fatal("same-generation explicit reproof failed", e)
	}
	for _, status := range []string{"retryable_error", "stale_generation"} {
		p, _, service, v := setupRecovery(t)
		if _, e := p.PrepareDirect(ctx, status, v.Intent, v.Sidecars[0], "stable"); e != nil {
			t.Fatal(e)
		}
		service.jobStatus = status
		r, e := p.Resume(ctx, status, false)
		expected := "pending"
		if status == "stale_generation" {
			expected = "needs_reproof"
		}
		if e != nil || r.State != expected || r.ProofValid || service.submits != 1 {
			t.Fatal(status, r.State, e)
		}
	}
}

func (s *recoveryServiceFixture) EvaluateTransaction(context.Context, any, []ProofSidecar) (Result, error) {
	return nil, errors.New("direct fixture has no preflight")
}
func (s *recoveryServiceFixture) GetTracking(context.Context, string) (Result, error) {
	return nil, errors.New("direct fixture has no tracking")
}

type trackedRecoveryService struct {
	*recoveryServiceFixture
	tracking             Tracking
	evaluations, queries int
	missing              bool
}

func (s *trackedRecoveryService) EvaluateTransaction(_ context.Context, data any, sidecars []ProofSidecar) (Result, error) {
	s.evaluations++
	if data != s.intent.Calldata || len(sidecars) != 1 {
		return nil, errors.New("wrong frozen preflight")
	}
	if len(sidecars[0].ProofBundle.PublicInputs) == 11 {
		sidecars[0].ProofBundle.PublicInputs[2], _ = word("3")
	}
	digest, _ := json.Marshal(s.intent.OperationDigest)
	r := Result{"broadcastDecision": json.RawMessage(`"allow"`), "operationDigest": digest}
	if !s.missing {
		raw, _ := json.Marshal(s.tracking)
		r["tracking"] = raw
	}
	return r, nil
}
func (s *trackedRecoveryService) GetTracking(_ context.Context, id string) (Result, error) {
	s.queries++
	if id != s.tracking.ID {
		return nil, errors.New("wrong tracking id")
	}
	raw, _ := json.Marshal(s.tracking)
	r := Result{"tracking": raw}
	_, e := TrackingFromResult(r, id)
	return r, e
}
func TestDefaultRecoveryPersistsTrackingWithoutRedundantSubmit(t *testing.T) {
	c, h, s, v := setupRecovery(t)
	digest, _ := word(v.Intent.OperationDigest)
	tracking := Tracking{ID: strings.Repeat("a", 48), Status: "recognized", ExpiresAt: "1", OperationDigest: digest, Matched: &TrackingMatch{h.fact.TransactionHash, h.fact.Locator["blockHash"], "0"}}
	tracked := &trackedRecoveryService{recoveryServiceFixture: s, tracking: tracking}
	c.Service = tracked
	s.recognition = "recognized"
	r, e := c.Prepare(context.Background(), "tracked", v.Intent, v.Sidecars[0], "stable")
	if e != nil || r.TrackingID != tracking.ID || r.DeliveryMode != "tracked" {
		t.Fatal(e)
	}
	saved, _, _ := c.Store.Load(context.Background(), "tracked")
	if saved.TrackingID != tracking.ID {
		t.Fatal("tracking not durable before permit")
	}
	r, e = c.Run(context.Background(), "tracked", false, time.Millisecond, 2)
	if e != nil || r.State != "recognized" || s.submits != 0 || tracked.evaluations != 1 || tracked.queries != 1 {
		t.Fatal(r.State, e)
	}
	tracked.missing = true
	if _, e = c.Prepare(context.Background(), "missing", v.Intent, v.Sidecars[0], "stable"); e == nil {
		t.Fatal("silently downgraded missing tracking")
	}
	if _, ok, _ := c.Store.Load(context.Background(), "missing"); ok {
		t.Fatal("missing tracking got permit")
	}
}

func TestTerminalTrackingCannotReturnRecoveryPermit(t *testing.T) {
	for _, status := range []string{"expired", "superseded"} {
		c, _, s, v := setupRecovery(t)
		digest, _ := word(v.Intent.OperationDigest)
		service := &trackedRecoveryService{recoveryServiceFixture: s, tracking: Tracking{ID: strings.Repeat("a", 48), Status: status, ExpiresAt: "1", OperationDigest: digest}}
		c.Service = service
		if _, e := c.Prepare(context.Background(), "terminal", v.Intent, v.Sidecars[0], "stable"); e == nil {
			t.Fatal("terminal preflight got broadcast permit")
		}
		if _, ok, _ := c.Store.Load(context.Background(), "terminal"); ok {
			t.Fatal("terminal preflight persisted")
		}
		if s.submits != 0 {
			t.Fatal("terminal preflight submitted")
		}
	}
}

type reservedKeyRoundTripper func(*http.Request) (*http.Response, error)

func (f reservedKeyRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestReservedIdempotencyPrefix(t *testing.T) {
	t.Run("beforePermit", func(t *testing.T) {
		c, _, _, v := setupRecovery(t)
		if _, e := c.PrepareDirect(context.Background(), "reserved", v.Intent, v.Sidecars[0], "internal/test"); e == nil || !strings.Contains(e.Error(), "reserved") {
			t.Fatal("reserved base escaped permit validation", e)
		}
		if _, ok, _ := c.Store.Load(context.Background(), "reserved"); ok {
			t.Fatal("reserved base persisted")
		}
	})
	t.Run("beforeHTTP", func(t *testing.T) {
		c, _, _, v := setupRecovery(t)
		calls := 0
		client := Client{Endpoint: "https://example.test", Trust: c.Trust, HTTP: &http.Client{Transport: reservedKeyRoundTripper(func(*http.Request) (*http.Response, error) { calls++; return nil, errors.New("unexpected HTTP") })}}
		if _, e := client.SubmitProof(context.Background(), map[string]any{}, v.Sidecars[0].ProofBundle, "internal/test"); e == nil || !strings.Contains(e.Error(), "reserved") {
			t.Fatal("reserved key escaped HTTP validation", e)
		}
		if calls != 0 {
			t.Fatal("reserved key sent over HTTP")
		}
	})
}
