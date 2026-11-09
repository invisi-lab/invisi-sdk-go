package provenance

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type liveTrust struct {
	TrackingID       string           `json:"trackingId"`
	Recovery         json.RawMessage  `json:"recovery,omitempty"`
	Format           string           `json:"format"`
	Endpoint         string           `json:"endpoint"`
	Scope            Scope            `json:"scope"`
	Generation       string           `json:"chainGeneration"`
	ProviderID       string           `json:"providerId"`
	InitialPublicKey string           `json:"initialSigningPublicKey"`
	InitialKeyID     string           `json:"initialSigningKeyId"`
	CurrentPublicKey string           `json:"currentSigningPublicKey"`
	CurrentKeyID     string           `json:"currentSigningKeyId"`
	Manifest         string           `json:"manifestSha256"`
	Rotations        []RotationRecord `json:"keyRotations"`
	CheckedAt        string           `json:"checkedAt"`
	RootID           string           `json:"rootId"`
	Commitment       string           `json:"commitment"`
	JobID            string           `json:"jobId"`
}

func TestProvenanceLiveService(t *testing.T) {
	if os.Getenv("INVISI_PROVENANCE_LIVE") != "1" {
		t.Skip("explicit INVISI_PROVENANCE_LIVE=1 and public operator trust file required")
	}
	path := os.Getenv("INVISI_PROVENANCE_TRUST")
	if path == "" {
		t.Fatal("missing INVISI_PROVENANCE_TRUST")
	}
	data, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var f liveTrust
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&f); e != nil {
		t.Fatal(e)
	}
	const pin = "f90f2ce43218de158c6392fa19e6517dda703c35823a4aa5bfc20b50788ad282"
	if f.Format != "invisi-provenance-sdk-live-v1" || f.Manifest != pin {
		t.Fatal("unreviewed live trust profile")
	}
	endpoint, e := url.Parse(f.Endpoint)
	if e != nil || (endpoint.Path != "" && endpoint.Path != "/") || (endpoint.Hostname() != "127.0.0.1" && endpoint.Hostname() != "localhost" && endpoint.Hostname() != "::1") {
		t.Fatal("live tests require explicit owned loopback endpoint")
	}
	initial, e := decodeFixed(f.InitialPublicKey, 32)
	if e != nil {
		t.Fatal(e)
	}
	current, e := decodeFixed(f.CurrentPublicKey, 32)
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range []struct {
		key []byte
		id  string
	}{{initial, f.InitialKeyID}, {current, f.CurrentKeyID}} {
		digest := sha256.Sum256(v.key)
		if "0x"+hex.EncodeToString(digest[:]) != v.id {
			t.Fatal("local key ID pin mismatch")
		}
	}
	ring, e := VerifyRotationChain(RotationCheckpoint{ProviderID: f.ProviderID, Keys: map[string]ed25519.PublicKey{f.InitialKeyID: initial}, PreviousRotationHash: "0x" + strings.Repeat("0", 64), LastActivatedAt: "0", Now: f.CheckedAt}, f.Rotations)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual([]byte(ring.Keys[f.CurrentKeyID]), current) {
		t.Fatal("local current key not authorized from pinned genesis")
	}
	trust, e := WithRotationHistory(TrustContext{Scope: f.Scope, ChainGeneration: f.Generation, ProviderID: f.ProviderID, ManifestSHA256: pin}, ring)
	if e != nil {
		t.Fatal(e)
	}
	client := Client{Endpoint: f.Endpoint, Trust: trust}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	text := func(r Result, key string) string {
		t.Helper()
		var v string
		if e := json.Unmarshal(r[key], &v); e != nil {
			t.Fatalf("missing/nonstring actual wire %s: %v", key, e)
		}
		return v
	}
	identity := func(r Result) {
		t.Helper()
		var scope Scope
		if e := json.Unmarshal(r["scope"], &scope); e != nil {
			t.Fatal(e)
		}
		if !equalNumber(scope.ChainID, f.Scope.ChainID) || !equalNumber(scope.ProtocolAddress, f.Scope.ProtocolAddress) || !equalNumber(scope.PolicyID, f.Scope.PolicyID) || text(r, "chainGeneration") != f.Generation || text(r, "providerId") != f.ProviderID || text(r, "signingKeyId") != f.CurrentKeyID || text(r, "manifestSha256") != pin {
			t.Fatal("live response identity does not match local operator pins")
		}
	}
	must := func(r Result, e error) Result {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
		identity(r)
		return r
	}
	version := must(client.GetVersion(ctx))
	if text(version, "apiVersion") != "v1" || text(version, "protocolVersion") != "1" || text(version, "publicInputCount") != "13" || text(version, "proofFormat") != "invisi-provenance-proof-v1" {
		t.Fatal("version wire mismatch")
	}
	status := must(client.GetStatus(ctx))
	if string(status["workerAvailable"]) != "true" {
		t.Fatal("real native worker unavailable")
	}
	policy := must(client.GetPolicy(ctx))
	if text(policy, "initialSigningPublicKey") != f.InitialPublicKey {
		t.Fatal("policy genesis differs from external pin")
	}
	var rotations []RotationRecord
	json.Unmarshal(policy["keyRotations"], &rotations)
	if len(rotations) != len(f.Rotations) || (len(rotations) > 0 && !reflect.DeepEqual(rotations, f.Rotations)) {
		t.Fatal("policy history differs from operator file")
	}
	var advertised map[string]string
	if json.Unmarshal(policy["providerKeys"], &advertised) != nil {
		t.Fatal("providerKeys wire")
	}
	for id, key := range ring.Keys {
		if advertised[id] != "0x"+hex.EncodeToString(key) {
			t.Fatal("advertised key differs from validated history")
		}
	}
	found := false
	cursor := ""
	for page := 0; page < 20; page++ {
		q := map[string]any{"limit": "1"}
		if cursor == "" {
			q["after"] = "0"
		} else {
			q["cursor"] = cursor
		}
		roots := must(client.GetRoots(ctx, q))
		var certs []RootCertificate
		if e := json.Unmarshal(roots["items"], &certs); e != nil {
			t.Fatal(e)
		}
		for _, c := range certs {
			if c.ChainGeneration+":"+c.RootSequence == f.RootID {
				found = true
			}
		}
		if len(roots["nextCursor"]) == 0 {
			break
		}
		cursor = text(roots, "nextCursor")
		if page == 19 {
			t.Fatal("unbounded root history")
		}
	}
	if !found {
		t.Fatal("pinned published root absent from actual root pages")
	}
	root := must(client.GetRoot(ctx, f.RootID))
	if text(root, "rootId") != f.RootID || string(root["allowedForNewProof"]) != "true" {
		t.Fatal("root not currently available")
	}
	var certificate RootCertificate
	if e := json.Unmarshal(root["certificate"], &certificate); e != nil {
		t.Fatal(e)
	}
	if e := AssertCertificateTrust(certificate, trust); e != nil {
		t.Fatal(e)
	}
	leaves := must(client.GetLeaves(ctx, f.RootID, "0", 1))
	if text(leaves, "rootId") != f.RootID || text(leaves, "nextIndex") != "1" {
		t.Fatal("leaf page wire mismatch")
	}
	tree, e := client.SyncRecognizedTree(ctx, f.RootID, certificate, 10000)
	if e != nil {
		t.Fatal(e)
	}
	local, ok, e := tree.MembershipForCommitment(f.Scope, f.Commitment)
	if e != nil || !ok {
		t.Fatal("local recognition membership missing", e)
	}
	remote := must(client.GetMembership(ctx, f.Commitment, f.RootID, true))
	var pathWords []string
	if e := json.Unmarshal(remote["path"], &pathWords); e != nil {
		t.Fatal(e)
	}
	if !equalNumber(text(remote, "leafIndex"), newNumberString(local.LeafIndex)) || len(pathWords) != len(local.Path) {
		t.Fatal("membership index/width differs")
	}
	for i, v := range pathWords {
		if !equalNumber(v, local.Path[i]) {
			t.Fatal("remote path differs from local replay")
		}
	}
	note := must(client.GetNoteRecognition(ctx, f.Commitment))
	if string(note["found"]) != "true" || text(note, "recognitionState") != "recognized" {
		t.Fatal("actual Note not recognized")
	}
	var publicNote struct {
		Commitment string `json:"noteCommitment"`
	}
	if json.Unmarshal(note["note"], &publicNote) != nil || publicNote.Commitment != f.Commitment {
		t.Fatal("Note recognition wire mismatch")
	}
	var noteCert RootCertificate
	if json.Unmarshal(note["certificate"], &noteCert) != nil || AssertCertificateTrust(noteCert, trust) != nil {
		t.Fatal("Note certificate not trusted")
	}
	tracked := must(client.GetTracking(ctx, f.TrackingID))
	identity(tracked)
	tracking, e := TrackingFromResult(tracked, f.TrackingID)
	if e != nil || tracking.Status != "recognized" {
		t.Fatal("tracking not recognized", e)
	}
	waited := must(client.WaitTracking(ctx, f.TrackingID, time.Millisecond, 3*time.Second))
	identity(waited)
	for name, change := range map[string]func(*TrustContext){"provider": func(t *TrustContext) { t.ProviderID = "0" }, "generation": func(t *TrustContext) { t.ChainGeneration = "0" }, "manifest": func(t *TrustContext) { t.ManifestSHA256 = strings.Repeat("0", 64) }, "key": func(t *TrustContext) { t.SigningKeys = map[string]ed25519.PublicKey{} }} {
		bad := trust
		change(&bad)
		untrusted := Client{Endpoint: f.Endpoint, Trust: bad}
		if _, e := untrusted.GetTracking(ctx, f.TrackingID); e == nil {
			t.Fatalf("actual response accepted with wrong %s pin", name)
		}
	}
	bad := trust
	bad.PolicyID = "0"
	untrusted := Client{Endpoint: f.Endpoint, Trust: bad}
	if _, e := untrusted.GetVersion(ctx); e == nil || !strings.Contains(e.Error(), "POLICY_UNKNOWN") {
		t.Fatal("actual response accepted wrong scope")
	}
}
func newNumberString(value uint64) string { return strconv.FormatUint(value, 10) }
