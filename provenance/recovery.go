package provenance

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"math/big"
	"strings"
	"time"
)

func RecoverySubmissionKey(base string, sidecar ProofSidecar) (string, error) {
	if strings.HasPrefix(base, "internal/") {
		return "", errors.New("reserved idempotency key prefix")
	}
	if base == "" || len(base) > 63 {
		return "", errors.New("invalid recovery idempotency key")
	}
	for _, b := range []byte(base) {
		if b < 33 || b > 126 {
			return "", errors.New("invalid recovery idempotency key")
		}
	}
	raw, e := json.Marshal(sidecar.ProofBundle)
	if e != nil {
		return "", e
	}
	var value any
	if e = json.Unmarshal(raw, &value); e != nil {
		return "", e
	}
	canonical, e := json.Marshal(value)
	if e != nil {
		return "", e
	}
	hash := sha256.Sum256(canonical)
	return base + ":" + hex.EncodeToString(hash[:]), nil
}

type FrozenProvenanceIntent struct {
	OperationIndex string `json:"operationIndex"`
	Scope
	TokenAddress      string   `json:"tokenAddress"`
	OperationDigest   string   `json:"operationDigest"`
	InputNullifiers   []string `json:"inputNullifiers"`
	Calldata          string   `json:"calldata"`
	CalldataHash      string   `json:"calldataHash"`
	SignedRequestHash string   `json:"signedRequestHash"`
}
type RecoveryRecord struct {
	DeliveryMode       string                 `json:"deliveryMode"`
	TrackingID         string                 `json:"trackingId,omitempty"`
	ProofJobStatus     string                 `json:"proofJobStatus,omitempty"`
	ProofValid         bool                   `json:"proofValid,omitempty"`
	SourceState        string                 `json:"sourceState,omitempty"`
	ReasonCode         string                 `json:"reasonCode,omitempty"`
	CheckpointSequence string                 `json:"checkpointSequence,omitempty"`
	Format             string                 `json:"format"`
	Revision           uint64                 `json:"revision"`
	ID                 string                 `json:"id"`
	Intent             FrozenProvenanceIntent `json:"intent"`
	ManifestSHA256     string                 `json:"manifestSha256"`
	ChainGeneration    string                 `json:"chainGeneration"`
	Sidecar            ProofSidecar           `json:"sidecar"`
	IdempotencyKey     string                 `json:"idempotencyKey"`
	State              string                 `json:"state"`
	TransactionHash    string                 `json:"transactionHash,omitempty"`
	Locator            map[string]string      `json:"locator,omitempty"`
	JobID              string                 `json:"jobId,omitempty"`
}
type RecoveryCiphertext struct {
	Revision uint64
	Data     []byte
}

// AtomicRecoveryRecords adapters persist only ciphertext and atomically fence stale revisions.
type AtomicRecoveryRecords interface {
	Read(context.Context, string) (RecoveryCiphertext, bool, error)
	CompareAndSwap(context.Context, string, uint64, RecoveryCiphertext) (bool, error)
}
type RecoveryAtomicStore interface {
	Load(context.Context, string) (RecoveryRecord, bool, error)
	CompareAndSwap(context.Context, string, uint64, RecoveryRecord) (bool, error)
}
type EncryptedRecoveryStore struct {
	records AtomicRecoveryRecords
	aead    cipher.AEAD
}

func NewEncryptedRecoveryStore(records AtomicRecoveryRecords, key []byte) (*EncryptedRecoveryStore, error) {
	if records == nil || len(key) != 32 {
		return nil, errors.New("recovery encryption requires records and 32 key bytes")
	}
	b, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	a, e := cipher.NewGCM(b)
	return &EncryptedRecoveryStore{records, a}, e
}
func recoveryStorageKey(id string) (string, error) {
	if id == "" || len(id) > 256 {
		return "", errors.New("invalid recovery id")
	}
	return "Invisi.Provenance.Recovery.v1:" + id, nil
}
func recoveryAAD(key string, revision uint64) []byte {
	return []byte(fmt.Sprintf("Invisi.Provenance.RecoveryRecord.v1\x00%d\x00%s", revision, key))
}
func (s *EncryptedRecoveryStore) Load(ctx context.Context, id string) (RecoveryRecord, bool, error) {
	var out RecoveryRecord
	k, e := recoveryStorageKey(id)
	if e != nil {
		return out, false, e
	}
	r, ok, e := s.records.Read(ctx, k)
	if e != nil || !ok {
		return out, ok, e
	}
	if r.Revision == 0 || r.Revision > 9007199254740991 || len(r.Data) < 29 || r.Data[0] != 1 {
		return out, false, errors.New("invalid recovery ciphertext")
	}
	plain, e := s.aead.Open(nil, r.Data[1:13], r.Data[13:], recoveryAAD(k, r.Revision))
	if e != nil {
		return out, false, errors.New("recovery authentication failed")
	}
	defer clear(plain)
	dec := json.NewDecoder(strings.NewReader(string(plain)))
	dec.DisallowUnknownFields()
	if e = dec.Decode(&out); e != nil {
		return out, false, e
	}
	if out.Revision != r.Revision {
		return out, false, errors.New("recovery revision mismatch")
	}
	return out, true, nil
}
func (s *EncryptedRecoveryStore) CompareAndSwap(ctx context.Context, id string, old uint64, next RecoveryRecord) (bool, error) {
	for key := range next.Locator {
		if key != "blockHash" && key != "txHash" && key != "operationIndex" {
			return false, errors.New("unknown recovery locator field")
		}
	}
	if old >= 9007199254740991 || next.Revision != old+1 {
		return false, errors.New("invalid next recovery revision")
	}
	k, e := recoveryStorageKey(id)
	if e != nil {
		return false, e
	}
	plain, e := json.Marshal(next)
	if e != nil {
		return false, e
	}
	defer clear(plain)
	nonce := make([]byte, 12)
	if _, e = rand.Read(nonce); e != nil {
		return false, e
	}
	data := append([]byte{1}, nonce...)
	data = s.aead.Seal(data, nonce, plain, recoveryAAD(k, next.Revision))
	return s.records.CompareAndSwap(ctx, k, old, RecoveryCiphertext{next.Revision, data})
}

type RecoveryReconciliation struct {
	Status            string
	CalldataHash      string
	SignedRequestHash string
	ChainGeneration   string
	TransactionHash   string
	Locator           map[string]string
}

// Hosts authenticate the original transaction and independently verify local proofs, never receive private openings here.
type RecoveryHost interface {
	VerifyFrozenStatement(context.Context, FrozenProvenanceIntent, ProofSidecar) error
	ReconcileFrozenTransaction(context.Context, RecoveryRecord) (RecoveryReconciliation, error)
}
type LocalRecoveryReprover interface {
	LocalReprove(context.Context, RecoveryRecord, string) (ProofSidecar, error)
}
type RecoveryService interface {
	EvaluateTransaction(context.Context, any, []ProofSidecar) (Result, error)
	GetTracking(context.Context, string) (Result, error)
	SubmitProof(context.Context, map[string]any, ProofBundle, string) (Result, error)
	GetProofJob(context.Context, string) (Result, error)
	GetOperationRecognition(context.Context, map[string]any) (Result, error)
}
type RecoveryCoordinator struct {
	Store   RecoveryAtomicStore
	Trust   TrustContext
	Service RecoveryService
	Host    RecoveryHost
}

func recoveryClone(r RecoveryRecord) RecoveryRecord {
	b, _ := json.Marshal(r)
	var out RecoveryRecord
	json.Unmarshal(b, &out)
	return out
}
func calldataHash(data string) (string, error) {
	if !strings.HasPrefix(data, "0x") {
		return "", errors.New("invalid frozen calldata")
	}
	raw, e := hex.DecodeString(data[2:])
	if e != nil {
		return "", e
	}
	h := sha3.NewLegacyKeccak256()
	h.Write(raw)
	return "0x" + hex.EncodeToString(h.Sum(nil)), nil
}
func (c *RecoveryCoordinator) validate(ctx context.Context, r RecoveryRecord) error {
	if r.Format != "invisi-provenance-recovery-v1" || r.ID == "" || r.IdempotencyKey == "" || len(r.IdempotencyKey) > 63 || c.Store == nil || c.Host == nil || c.Service == nil {
		return errors.New("invalid recovery configuration")
	}
	if (r.DeliveryMode != "tracked" && r.DeliveryMode != "direct") || (r.Revision > 0 && r.DeliveryMode == "tracked" && r.TrackingID == "" && r.State != "needs_reproof") {
		return errors.New("recovery tracking mode/identity missing")
	}
	if _, e := RecoverySubmissionKey(r.IdempotencyKey, r.Sidecar); e != nil {
		return e
	}
	for _, p := range [][2]string{{r.Intent.ChainID, c.Trust.ChainID}, {r.Intent.ProtocolAddress, c.Trust.ProtocolAddress}} {
		if !equalNumber(p[0], p[1]) {
			return errors.New("recovery scope mismatch")
		}
	}
	h, e := calldataHash(r.Intent.Calldata)
	if e != nil || h != r.Intent.CalldataHash || r.ManifestSHA256 != c.Trust.ManifestSHA256 {
		return errors.New("frozen recovery intent mismatch")
	}
	if _, e := decodeFixed(r.Intent.SignedRequestHash, 32); e != nil {
		return e
	}
	if !equalNumber(r.Sidecar.OperationIndex, r.Intent.OperationIndex) {
		return errors.New("recovery operation index mismatch")
	}
	p := r.Sidecar.ProofBundle.PublicInputs
	if len(p) != 11 {
		return errors.New("recovery public ABI mismatch")
	}
	hi, e := Uint(p[7], 128)
	if e != nil {
		return e
	}
	lo, e := Uint(p[8], 128)
	if e != nil {
		return e
	}
	digest := new(big.Int).Or(new(big.Int).Lsh(hi, 128), lo)
	if !equalNumber(digest.String(), r.Intent.OperationDigest) || !equalNumber(p[2], r.Intent.TokenAddress) || !equalNumber(p[6], fmt.Sprint(len(r.Intent.InputNullifiers))) {
		return errors.New("recovery statement mismatch")
	}
	seen := map[string]bool{}
	for _, n := range r.Intent.InputNullifiers {
		v, e := word(n)
		if e != nil || seen[v] {
			return errors.New("invalid recovery nullifier")
		}
		seen[v] = true
	}
	inputHash, e := InputDigest(r.Intent.InputNullifiers)
	if e != nil || !equalNumber(inputHash, p[9]) {
		return errors.New("recovery nullifiers digest mismatch")
	}
	trust := c.Trust
	trust.ChainGeneration = r.ChainGeneration
	if e := ValidateProofBundle(r.Sidecar.ProofBundle, trust); e != nil {
		return e
	}
	detached := recoveryClone(r)
	return c.Host.VerifyFrozenStatement(ctx, detached.Intent, detached.Sidecar)
}
func (c *RecoveryCoordinator) save(ctx context.Context, r RecoveryRecord, previous uint64) (RecoveryRecord, error) {
	if previous >= 9007199254740991 {
		return r, errors.New("recovery revision exhausted")
	}
	r.Revision = previous + 1
	ok, e := c.Store.CompareAndSwap(ctx, r.ID, previous, recoveryClone(r))
	if e != nil {
		return r, e
	}
	if !ok {
		return r, errors.New("recovery CAS conflict")
	}
	return recoveryClone(r), nil
}
func (c *RecoveryCoordinator) Prepare(ctx context.Context, id string, intent FrozenProvenanceIntent, sidecar ProofSidecar, key string) (RecoveryRecord, error) {
	return c.prepareMode(ctx, id, intent, sidecar, key, "tracked")
}
func (c *RecoveryCoordinator) PrepareDirect(ctx context.Context, id string, intent FrozenProvenanceIntent, sidecar ProofSidecar, key string) (RecoveryRecord, error) {
	return c.prepareMode(ctx, id, intent, sidecar, key, "direct")
}
func (c *RecoveryCoordinator) prepareMode(ctx context.Context, id string, intent FrozenProvenanceIntent, sidecar ProofSidecar, key, mode string) (RecoveryRecord, error) {
	r := RecoveryRecord{DeliveryMode: mode, Format: "invisi-provenance-recovery-v1", ID: id, Intent: intent, ManifestSHA256: c.Trust.ManifestSHA256, ChainGeneration: c.Trust.ChainGeneration, Sidecar: sidecar, IdempotencyKey: key, State: "broadcast_possible"}
	r = recoveryClone(r)
	if e := c.validate(ctx, r); e != nil {
		return r, e
	}
	if _, ok, e := c.Store.Load(ctx, id); e != nil || ok {
		if e == nil {
			e = errors.New("recovery id already exists")
		}
		return r, e
	}
	if mode == "tracked" {
		evaluated, e := c.Service.EvaluateTransaction(ctx, r.Intent.Calldata, []ProofSidecar{recoveryClone(r).Sidecar})
		if e != nil {
			return r, e
		}
		tracking, e := PreflightTracking(evaluated, r.Intent.OperationIndex, r.Intent.OperationDigest)
		if e != nil {
			return r, e
		}
		r.TrackingID = tracking.ID
	}
	return c.save(ctx, r, 0)
}
func (c *RecoveryCoordinator) required(ctx context.Context, id string) (RecoveryRecord, error) {
	r, ok, e := c.Store.Load(ctx, id)
	if e != nil {
		return r, e
	}
	if !ok || r.ID != id {
		return r, errors.New("missing recovery record")
	}
	return r, c.validate(ctx, r)
}
func (c *RecoveryCoordinator) MarkBroadcast(ctx context.Context, id, tx string) (RecoveryRecord, error) {
	r, e := c.required(ctx, id)
	if e != nil {
		return r, e
	}
	tx, e = word(tx)
	if e != nil {
		return r, e
	}
	if r.TransactionHash != "" && r.TransactionHash != tx {
		return r, errors.New("frozen transaction hash mismatch")
	}
	r.TransactionHash = tx
	if r.State == "broadcast_possible" {
		r.State = "broadcast"
	}
	return c.save(ctx, r, r.Revision)
}
func (c *RecoveryCoordinator) Resume(ctx context.Context, id string, reprove bool) (RecoveryRecord, error) {
	if e := ctx.Err(); e != nil {
		return RecoveryRecord{}, e
	}
	r, e := c.required(ctx, id)
	if e != nil {
		return r, e
	}
	fact, e := c.Host.ReconcileFrozenTransaction(ctx, recoveryClone(r))
	if e != nil {
		return r, e
	}
	if fact.Locator != nil {
		copied := map[string]string{}
		for k, v := range fact.Locator {
			copied[k] = v
		}
		fact.Locator = copied
	}
	if e = ctx.Err(); e != nil {
		return r, e
	}
	if fact.CalldataHash != r.Intent.CalldataHash || fact.SignedRequestHash != r.Intent.SignedRequestHash || (r.TransactionHash != "" && fact.TransactionHash != r.TransactionHash) {
		return r, errors.New("reconciled frozen transaction mismatch")
	}
	if fact.Status != "unseen" && fact.Status != "pending" && fact.Status != "confirmed" && fact.Status != "reorged" {
		return r, errors.New("invalid reconciliation status")
	}
	if _, e = Uint(fact.ChainGeneration, 64); e != nil {
		return r, e
	}
	previous := r.Revision
	if r.ChainGeneration != fact.ChainGeneration || fact.Status == "reorged" || !equalNumber(fact.ChainGeneration, c.Trust.ChainGeneration) || r.State == "needs_reproof" || (reprove && r.State == "proof_rejected") {
		r.State = "needs_reproof"
		r.TrackingID = ""
		r.JobID = ""
		r.ProofValid = false
		r.ProofJobStatus = ""
		r.ReasonCode = ""
		r.SourceState = ""
		r.CheckpointSequence = ""
		r.Locator = nil
		local, ok := c.Host.(LocalRecoveryReprover)
		if !reprove || !ok || fact.Status != "confirmed" || !equalNumber(fact.ChainGeneration, c.Trust.ChainGeneration) {
			return c.save(ctx, r, previous)
		}
		next, e := local.LocalReprove(ctx, recoveryClone(r), fact.ChainGeneration)
		if e != nil {
			return r, e
		}
		r.Sidecar = recoveryClone(RecoveryRecord{Sidecar: next}).Sidecar
		r.ChainGeneration = fact.ChainGeneration
		check := recoveryClone(r)
		check.Revision = 0
		if e = c.validate(ctx, check); e != nil {
			return r, e
		}
		r.State = "pending"
		if r.DeliveryMode == "tracked" {
			evaluated, e := c.Service.EvaluateTransaction(ctx, r.Intent.Calldata, []ProofSidecar{recoveryClone(r).Sidecar})
			if e != nil {
				return r, e
			}
			tracking, e := PreflightTracking(evaluated, r.Intent.OperationIndex, r.Intent.OperationDigest)
			if e != nil {
				return r, e
			}
			r.TrackingID = tracking.ID
		}
	}
	if fact.Status == "confirmed" {
		if fact.TransactionHash == "" || len(fact.Locator) == 0 {
			return r, errors.New("missing authenticated mined locator")
		}
		for k := range fact.Locator {
			if k != "blockHash" && k != "txHash" && k != "operationIndex" {
				return r, errors.New("invalid authenticated locator")
			}
		}
		if _, e := decodeFixed(fact.TransactionHash, 32); e != nil {
			return r, e
		}
		if _, e := decodeFixed(fact.Locator["blockHash"], 32); e != nil {
			return r, e
		}
		if fact.Locator["txHash"] != fact.TransactionHash || !equalNumber(fact.Locator["operationIndex"], r.Intent.OperationIndex) {
			return r, errors.New("invalid authenticated locator")
		}
		r.TransactionHash = fact.TransactionHash
		r.Locator = fact.Locator
	}
	if r.DeliveryMode == "tracked" {
		if r.TrackingID == "" {
			return r, errors.New("recovery tracking id required")
		}
		result, e := c.Service.GetTracking(ctx, r.TrackingID)
		if errors.Is(e, ErrTrackingStaleGeneration) {
			r.State = "needs_reproof"
			return c.save(ctx, r, previous)
		}
		if e != nil {
			return r, e
		}
		tracking, e := TrackingFromResult(result, r.TrackingID)
		if e != nil {
			return r, e
		}
		if !equalNumber(tracking.OperationDigest, r.Intent.OperationDigest) {
			return r, errors.New("tracking statement mismatch")
		}
		r.ReasonCode = tracking.ReasonCode
		if tracking.Status == "expired" || tracking.Status == "superseded" {
			r.State = "tracking_" + tracking.Status
			r.ReasonCode = tracking.ReasonCode
			return c.save(ctx, r, previous)
		}
		if fact.Status != "confirmed" || tracking.Status == "pending_confirmation" {
			r.State = "pending"
			return c.save(ctx, r, previous)
		}
		if tracking.Matched == nil || fact.Locator == nil || tracking.Matched.TxHash != fact.TransactionHash || tracking.Matched.BlockHash != fact.Locator["blockHash"] || tracking.Matched.OperationIndex != fact.Locator["operationIndex"] {
			return r, errors.New("tracking matched another transaction")
		}
		if tracking.Status == "pending_source" {
			r.State = "pending_source"
			r.SourceState = "pending_source"
			r.TransactionHash = fact.TransactionHash
			r.Locator = fact.Locator
			return c.save(ctx, r, previous)
		}
	}
	if fact.Status != "confirmed" {
		return r, nil
	}
	if r.State == "recognized" || r.State == "proof_rejected" {
		return r, nil
	}
	if r.DeliveryMode == "direct" && r.JobID == "" {
		r, e = c.save(ctx, r, previous)
		if e != nil {
			return r, e
		}
		if e = ctx.Err(); e != nil {
			return r, e
		}
		q := map[string]any{}
		for k, v := range r.Locator {
			q[k] = v
		}
		submissionKey, e := RecoverySubmissionKey(r.IdempotencyKey, r.Sidecar)
		if e != nil {
			return r, e
		}
		result, e := c.Service.SubmitProof(ctx, q, r.Sidecar.ProofBundle, submissionKey)
		if e != nil {
			return r, e
		}
		if json.Unmarshal(result["jobId"], &r.JobID) != nil || r.JobID == "" {
			return r, errors.New("missing recovery job id")
		}
		r.State = "pending"
		r, e = c.save(ctx, r, r.Revision)
		if e != nil {
			return r, e
		}
	}
	if r.DeliveryMode == "direct" && !r.ProofValid {
		result, e := c.Service.GetProofJob(ctx, r.JobID)
		if e != nil {
			return r, e
		}
		var status string
		json.Unmarshal(result["status"], &status)
		r.ReasonCode = ""
		r.ProofJobStatus = status
		json.Unmarshal(result["reasonCode"], &r.ReasonCode)
		switch status {
		case "invalid":
			r.State = "proof_rejected"
		case "valid":
			r.ProofValid = true
			r.State = "pending_source"
		case "stale_generation":
			r.State = "needs_reproof"
		case "queued", "verifying", "retryable_error":
			r.State = "pending"
		default:
			return r, errors.New("unsupported proof job status")
		}

		r, e = c.save(ctx, r, r.Revision)
		if e != nil {
			return r, e
		}
	}
	if r.ProofValid || r.DeliveryMode == "tracked" {
		q := map[string]any{}
		for k, v := range r.Locator {
			q[k] = v
		}
		result, e := c.Service.GetOperationRecognition(ctx, q)
		if e != nil {
			return r, e
		}
		if json.Unmarshal(result["recognitionState"], &r.SourceState) != nil || r.SourceState == "" {
			return r, errors.New("missing operation recognition state")
		}
		var operation struct {
			ChainID         string   `json:"chainId"`
			ProtocolAddress string   `json:"protocolAddress"`
			TokenAddress    string   `json:"tokenAddress"`
			OperationDigest string   `json:"operationDigest"`
			BlockHash       string   `json:"blockHash"`
			TxHash          string   `json:"txHash"`
			OperationIndex  string   `json:"operationIndex"`
			Nullifiers      []string `json:"noteNullifiers"`
			Canonical       bool     `json:"canonical"`
			Success         bool     `json:"success"`
		}
		if json.Unmarshal(result["operation"], &operation) != nil || !operation.Canonical || !operation.Success || operation.BlockHash != r.Locator["blockHash"] || operation.TxHash != r.Locator["txHash"] || operation.OperationIndex != r.Locator["operationIndex"] {
			return r, errors.New("recognition locator mismatch")
		}
		for _, p := range [][2]string{{operation.ChainID, r.Intent.ChainID}, {operation.ProtocolAddress, r.Intent.ProtocolAddress}, {operation.TokenAddress, r.Intent.TokenAddress}, {operation.OperationDigest, r.Intent.OperationDigest}} {
			if !equalNumber(p[0], p[1]) {
				return r, errors.New("recognition statement mismatch")
			}
		}
		if len(operation.Nullifiers) != len(r.Intent.InputNullifiers) {
			return r, errors.New("recognition nullifiers mismatch")
		}
		for i, n := range operation.Nullifiers {
			if !equalNumber(n, r.Intent.InputNullifiers[i]) {
				return r, errors.New("recognition nullifiers mismatch")
			}
		}
		var details struct {
			ReasonCode         string `json:"reasonCode"`
			CheckpointSequence string `json:"checkpointSequence"`
		}
		json.Unmarshal(result["recognition"], &details)
		r.ReasonCode = details.ReasonCode
		r.CheckpointSequence = details.CheckpointSequence
		switch r.SourceState {
		case "recognized":
			if seq, e := Uint(r.CheckpointSequence, 64); e != nil || seq.Sign() == 0 {
				return r, errors.New("missing recognition checkpoint")
			}
			r.State = "recognized"
		case "source_rejected":
			r.State = "source_rejected"
		default:
			r.State = "pending_source"
		}
	}
	return c.save(ctx, r, r.Revision)
}
func (c *RecoveryCoordinator) Run(ctx context.Context, id string, reprove bool, poll time.Duration, maxPolls int) (RecoveryRecord, error) {
	if poll <= 0 || maxPolls <= 0 {
		return RecoveryRecord{}, errors.New("invalid recovery polling limits")
	}
	for i := 0; i < maxPolls; i++ {
		r, e := c.Resume(ctx, id, reprove)
		if e != nil || (r.State != "broadcast_possible" && r.State != "broadcast" && r.State != "pending" && r.State != "pending_source") {
			return r, e
		}
		timer := time.NewTimer(poll)
		select {
		case <-ctx.Done():
			timer.Stop()
			return r, ctx.Err()
		case <-timer.C:
		}
	}
	return RecoveryRecord{}, errors.New("recovery polling limit reached; resume remains available")
}

// FilterRecognizedCommitments performs local membership checks independently of protocol spendability.
func FilterRecognizedCommitments(scope Scope, tree *RecognizedTree, commitments []string) ([]string, error) {
	out := []string{}
	for _, commitment := range commitments {
		_, ok, e := tree.MembershipForCommitment(scope, commitment)
		if e != nil {
			return nil, e
		}
		if ok {
			out = append(out, commitment)
		}
	}
	return out, nil
}
