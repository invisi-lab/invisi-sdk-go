package provenance

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type recoveryLiveData struct {
	TrackingID      string       `json:"trackingId"`
	PrimaryRPC      string       `json:"primaryRPC"`
	ProtocolRoot    string       `json:"protocolRoot"`
	TransactionHash string       `json:"transactionHash"`
	BlockHash       string       `json:"blockHash"`
	BlockNumber     string       `json:"blockNumber"`
	Calldata        string       `json:"calldata"`
	OperationIndex  string       `json:"operationIndex"`
	OperationDigest string       `json:"operationDigest"`
	TokenAddress    string       `json:"tokenAddress"`
	InputNullifiers []string     `json:"inputNullifiers"`
	Sidecar         ProofSidecar `json:"sidecar"`
	InvalidJobID    string       `json:"invalidJobId"`
}
type liveRecoveryHost struct {
	primary             *ethclient.Client
	fixture             recoveryLiveData
	trust               liveTrust
	intent              FrozenProvenanceIntent
	bb, vk              string
	nativeVerifications int
	primaryCalls        int
}

func (h *liveRecoveryHost) ReconcileFrozenTransaction(ctx context.Context, _ RecoveryRecord) (RecoveryReconciliation, error) {
	var out RecoveryReconciliation
	h.primaryCalls += 4
	chain, e := h.primary.ChainID(ctx)
	if e != nil || chain.String() != h.trust.Scope.ChainID {
		return out, errors.New("primary chain mismatch")
	}
	tx, pending, e := h.primary.TransactionByHash(ctx, common.HexToHash(h.fixture.TransactionHash))
	if e != nil || pending || tx == nil {
		return out, fmt.Errorf("primary transaction unavailable: %v", e)
	}
	protocol, _ := Uint(h.trust.Scope.ProtocolAddress, 160)
	if tx.To() == nil || *tx.To() != common.BigToAddress(protocol) || "0x"+hex.EncodeToString(tx.Data()) != h.fixture.Calldata || tx.Hash().Hex() != h.fixture.TransactionHash {
		return out, errors.New("primary frozen transaction mismatch")
	}
	receipt, e := h.primary.TransactionReceipt(ctx, tx.Hash())
	if e != nil || receipt.Status != 1 || receipt.BlockHash.Hex() != h.fixture.BlockHash || receipt.TxHash != tx.Hash() {
		return out, errors.New("primary receipt mismatch")
	}
	header, e := h.primary.HeaderByNumber(ctx, receipt.BlockNumber)
	if e != nil || header.Hash() != receipt.BlockHash {
		return out, errors.New("primary canonical block mismatch")
	}
	hash, _ := calldataHash("0x" + hex.EncodeToString(tx.Data()))
	return RecoveryReconciliation{Status: "confirmed", CalldataHash: hash, SignedRequestHash: tx.Hash().Hex(), TransactionHash: tx.Hash().Hex(), ChainGeneration: h.trust.Generation, Locator: map[string]string{"blockHash": header.Hash().Hex(), "txHash": tx.Hash().Hex(), "operationIndex": h.fixture.OperationIndex}}, nil
}
func (h *liveRecoveryHost) VerifyFrozenStatement(ctx context.Context, intent FrozenProvenanceIntent, sidecar ProofSidecar) error {
	if !reflect.DeepEqual(intent, h.intent) {
		return errors.New("frozen intent changed")
	}
	if _, e := h.ReconcileFrozenTransaction(ctx, RecoveryRecord{}); e != nil {
		return e
	}
	dir, e := os.MkdirTemp("", "invisi-recovery-live-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(dir)
	proof, e := hex.DecodeString(strings.TrimPrefix(sidecar.ProofBundle.Proof, "0x"))
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(dir, "proof"), proof, 0600); e != nil {
		return e
	}
	var inputs []byte
	for _, p := range sidecar.ProofBundle.PublicInputs {
		v, e := decodeFixed(p, 32)
		if e != nil {
			return e
		}
		inputs = append(inputs, v...)
	}
	if e = os.WriteFile(filepath.Join(dir, "inputs"), inputs, 0600); e != nil {
		return e
	}
	cmd := exec.CommandContext(ctx, h.bb, "verify", "-t", "evm", "-k", h.vk, "-p", filepath.Join(dir, "proof"), "-i", filepath.Join(dir, "inputs"))
	cmd.Env = append(os.Environ(), "HARDWARE_CONCURRENCY=4")
	data, e := cmd.CombinedOutput()
	if e != nil || !bytes.Contains(data, []byte("Proof verified successfully")) {
		return errors.New("native proof verification rejected")
	}
	h.nativeVerifications++
	return nil
}
func (h *liveRecoveryHost) LocalReprove(context.Context, RecoveryRecord, string) (ProofSidecar, error) {
	return recoveryClone(RecoveryRecord{Sidecar: h.fixture.Sidecar}).Sidecar, nil
}
func TestProvenanceRecoveryLive(t *testing.T) {
	if os.Getenv("INVISI_PROVENANCE_RECOVERY_LIVE") != "1" {
		t.Skip("explicit owned recovery live opt-in required")
	}
	path := os.Getenv("INVISI_PROVENANCE_TRUST")
	data, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var f liveTrust
	if e = json.Unmarshal(data, &f); e != nil {
		t.Fatal(e)
	}
	var r recoveryLiveData
	if e = json.Unmarshal(f.Recovery, &r); e != nil {
		t.Fatal(e)
	}
	const pin = "6107789c5c0d24c4130bff05da6d389336feb5ccb247c4030eae5198ab05d762"
	if f.Manifest != pin || f.Format != "invisi-provenance-sdk-live-v1" {
		t.Fatal("untrusted manifest")
	}
	for _, endpoint := range []string{f.Endpoint, r.PrimaryRPC} {
		u, e := url.Parse(endpoint)
		if e != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" && u.Hostname() != "::1") {
			t.Fatal("owned loopback fixture required")
		}
	}
	initial, e := decodeFixed(f.InitialPublicKey, 32)
	if e != nil {
		t.Fatal(e)
	}
	hash := sha256.Sum256(initial)
	if "0x"+hex.EncodeToString(hash[:]) != f.InitialKeyID {
		t.Fatal("genesis pin mismatch")
	}
	ring, e := VerifyRotationChain(RotationCheckpoint{ProviderID: f.ProviderID, Keys: map[string]ed25519.PublicKey{f.InitialKeyID: initial}, PreviousRotationHash: "0x" + strings.Repeat("0", 64), LastActivatedAt: "0", Now: f.CheckedAt}, f.Rotations)
	if e != nil {
		t.Fatal(e)
	}
	if "0x"+hex.EncodeToString(ring.Keys[f.CurrentKeyID]) != f.CurrentPublicKey {
		t.Fatal("current key not genesis-authorized")
	}
	trust, e := WithRotationHistory(TrustContext{Scope: f.Scope, ChainGeneration: f.Generation, ProviderID: f.ProviderID, ManifestSHA256: pin}, ring)
	if e != nil {
		t.Fatal(e)
	}
	client := &Client{Endpoint: f.Endpoint, Trust: trust}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	if _, e = client.GetVersion(ctx); e != nil {
		t.Fatal(e)
	}
	if _, e = client.GetStatus(ctx); e != nil {
		t.Fatal(e)
	}
	primary, e := ethclient.DialContext(ctx, r.PrimaryRPC)
	if e != nil {
		t.Fatal(e)
	}
	defer primary.Close()
	calldataDigest, e := calldataHash(r.Calldata)
	if e != nil {
		t.Fatal(e)
	}
	intent := FrozenProvenanceIntent{Scope: f.Scope, OperationIndex: r.OperationIndex, TokenAddress: r.TokenAddress, OperationDigest: r.OperationDigest, InputNullifiers: r.InputNullifiers, Calldata: r.Calldata, CalldataHash: calldataDigest, SignedRequestHash: r.TransactionHash}
	manifest, e := os.ReadFile(filepath.Join(r.ProtocolRoot, "circuits/provenance/manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	identity, e := AuthenticateManifest(manifest, pin)
	if e != nil {
		t.Fatal(e)
	}
	var entry ManifestEntry
	for _, candidate := range identity.Manifest.Circuits {
		if fmt.Sprintf("provenance_k%d", candidate.Capacity) == r.Sidecar.ProofBundle.CircuitID {
			entry = candidate
		}
	}
	vk := filepath.Join(r.ProtocolRoot, entry.VK)
	vkBytes, e := os.ReadFile(vk)
	if e != nil {
		t.Fatal(e)
	}
	vkDigest := sha256.Sum256(vkBytes)
	if hex.EncodeToString(vkDigest[:]) != entry.VKSHA256 {
		t.Fatal("untrusted VK bytes")
	}
	bb := filepath.Join(r.ProtocolRoot, ".tools/bb/bb")
	version, e := exec.CommandContext(ctx, bb, "--version").Output()
	if e != nil || strings.TrimSpace(string(version)) != "6.0.0" {
		t.Fatal("BB version mismatch")
	}
	host := &liveRecoveryHost{primary: primary, fixture: r, trust: f, intent: intent, bb: bb, vk: vk}
	db, e := sqlOpenRecovery(t)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	records, e := NewSQLiteRecoveryRecords(ctx, db)
	if e != nil {
		t.Fatal(e)
	}
	key := bytes.Repeat([]byte{19}, 32)
	store, _ := NewEncryptedRecoveryStore(records, key)
	c := &RecoveryCoordinator{Store: store, Trust: trust, Service: client, Host: host}
	bad := recoveryClone(RecoveryRecord{Sidecar: r.Sidecar}).Sidecar
	raw, _ := hex.DecodeString(bad.ProofBundle.Proof[2:])
	raw[len(raw)-1] ^= 1
	bad.ProofBundle.Proof = "0x" + hex.EncodeToString(raw)
	if _, e = c.PrepareDirect(ctx, "bad-local", intent, bad, "go-bad-local"); e == nil {
		t.Fatal("invalid native proof accepted locally")
	}
	if _, ok, _ := store.Load(ctx, "bad-local"); ok {
		t.Fatal("bad proof persisted")
	}
	id := "go-recovery-live"
	base := fmt.Sprintf("go-recovery-%d", time.Now().UnixNano())
	if _, e = c.PrepareDirect(ctx, id, intent, r.Sidecar, base); e != nil {
		t.Fatal(e)
	}
	if r.InvalidJobID != "" {
		saved, _, e := store.Load(ctx, id)
		if e != nil {
			t.Fatal(e)
		}
		fact, e := host.ReconcileFrozenTransaction(ctx, saved)
		if e != nil {
			t.Fatal(e)
		}
		old := saved.Revision
		saved.Revision++
		saved.JobID = r.InvalidJobID
		saved.State = "pending"
		saved.Locator = fact.Locator
		saved.TransactionHash = fact.TransactionHash
		if ok, e := store.CompareAndSwap(ctx, id, old, saved); e != nil || !ok {
			t.Fatal(e)
		}
		invalid, e := c.Resume(ctx, id, false)
		if e != nil || invalid.State != "proof_rejected" || invalid.ProofJobStatus != "invalid" || invalid.ReasonCode != "PROOF_INVALID" {
			t.Fatal(invalid.State, e)
		}
	}
	store, _ = NewEncryptedRecoveryStore(records, key)
	restart := &RecoveryCoordinator{Store: store, Trust: trust, Service: client, Host: host}
	result, e := restart.Run(ctx, id, r.InvalidJobID != "", 100*time.Millisecond, 300)
	if e != nil || result.State != "recognized" || !result.ProofValid || !reflect.DeepEqual(result.Intent, intent) || result.CheckpointSequence == "" || result.ReasonCode == "PROOF_INVALID" {
		t.Fatal(result.State, e)
	}
	automatic := &RecoveryCoordinator{Store: store, Trust: trust, Service: trackingOnlyLiveService{client}, Host: host}
	trackedPermit, e := automatic.Prepare(ctx, "go-tracked-live", intent, r.Sidecar, fmt.Sprintf("go-tracked-%d", time.Now().UnixNano()))
	if e != nil || trackedPermit.TrackingID != r.TrackingID {
		t.Fatal("tracking not coalesced", e)
	}
	trackedResult, e := automatic.Run(ctx, "go-tracked-live", false, 100*time.Millisecond, 300)
	if e != nil || trackedResult.State != "recognized" || !reflect.DeepEqual(trackedResult.Intent, intent) {
		t.Fatal(trackedResult.State, e)
	}
	report, _ := json.MarshalIndent(map[string]any{"sdk": "go", "manifestSha256": pin, "realPrimaryHTTP": true, "primaryCalls": host.primaryCalls, "realNativeVerification": true, "nativeVerifications": host.nativeVerifications, "encryptedRestart": true, "trackedRecovery": true, "trackingId": trackedResult.TrackingID, "badLocalProofRejected": true, "reasonCode": result.ReasonCode, "invalidJobImported": r.InvalidJobID != "", "jobStatus": result.ProofJobStatus, "state": result.State, "jobId": result.JobID, "checkpointSequence": result.CheckpointSequence, "frozenTransactionHash": result.TransactionHash, "mocks": false}, "", "  ")
	if e = os.WriteFile(filepath.Join(filepath.Dir(path), "sdk-recovery-go-validation.json"), append(report, '\n'), 0600); e != nil {
		t.Fatal(e)
	}
}

func sqlOpenRecovery(t *testing.T) (*sql.DB, error) {
	return sql.Open("sqlite", filepath.Join(t.TempDir(), "recovery.sqlite"))
}

type trackingOnlyLiveService struct{ *Client }

func (s trackingOnlyLiveService) SubmitProof(context.Context, map[string]any, ProofBundle, string) (Result, error) {
	return nil, errors.New("tracked recovery must not submit proof")
}
func (s trackingOnlyLiveService) GetProofJob(context.Context, string) (Result, error) {
	return nil, errors.New("tracked recovery must not fetch proof job")
}
