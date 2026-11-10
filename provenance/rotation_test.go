package provenance

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSharedRotationVectorAndHistoricalValidity(t *testing.T) {
	data, e := os.ReadFile("testdata/rotation-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		Hash             string         `json:"hash"`
		InitialPublicKey string         `json:"initialPublicKey"`
		Message          string         `json:"message"`
		Payload          string         `json:"payload"`
		ProviderID       string         `json:"providerId"`
		OldSeed          string         `json:"oldSeed"`
		NewSeed          string         `json:"newSeed"`
		Now              string         `json:"now"`
		Rotation         RotationRecord `json:"rotation"`
	}
	if e := json.Unmarshal(data, &v); e != nil {
		t.Fatal(e)
	}
	r := v.Rotation
	payload, e := RotationPayload(r)
	if e != nil || len(payload) != 169 || "0x"+hex.EncodeToString(payload) != v.Payload {
		t.Fatal("rotation payload mismatch", e)
	}
	message, e := RotationMessage(r)
	if e != nil || "0x"+hex.EncodeToString(message) != v.Message {
		t.Fatal("rotation message mismatch", e)
	}
	hash, e := RotationHash(r)
	if e != nil || hash != v.Hash {
		t.Fatal("rotation hash mismatch", e)
	}
	pub, _ := decodeFixed(v.InitialPublicKey, 32)
	initial := RotationCheckpoint{ProviderID: v.ProviderID, Keys: map[string]ed25519.PublicKey{r.OldKeyID: pub}, PreviousRotationHash: r.PreviousRotationHash, LastActivatedAt: "0", Now: v.Now}
	ring, e := VerifyRotationChain(initial, []RotationRecord{r})
	if e != nil {
		t.Fatal(e)
	}
	if "0x"+hex.EncodeToString(ring.Keys[r.NewKeyID]) != r.NewPublicKey || ring.Validity[r.OldKeyID].RetiredAt != "200" {
		t.Fatal("rotation state mismatch")
	}
	bad := initial
	bad.Keys = nil
	if _, e := VerifyRotationChain(bad, []RotationRecord{r}); e == nil {
		t.Fatal("TOFU rotation accepted")
	}
	bad = initial
	bad.Now = "199"
	if _, e := VerifyRotationChain(bad, []RotationRecord{r}); e == nil {
		t.Fatal("future rotation accepted")
	}
	if _, e := VerifyRotationChain(initial, []RotationRecord{r, r}); e == nil {
		t.Fatal("replayed rotation accepted")
	}
	changed := r
	changed.NewKeyID = r.OldKeyID
	if _, e := VerifyRotationChain(initial, []RotationRecord{changed}); e == nil {
		t.Fatal("wrong key hash accepted")
	}
	changed = r
	changed.Signature = "0x" + hex.EncodeToString(make([]byte, 64))
	if _, e := VerifyRotationChain(initial, []RotationRecord{changed}); e == nil {
		t.Fatal("bad signature accepted")
	}
	raw := vectors(t)["certificate"].(map[string]any)["fields"].(map[string]any)
	fields := map[string]string{}
	for k, value := range raw {
		fields[k] = num(value)
	}
	encoded, _ := json.Marshal(fields)
	var c RootCertificate
	json.Unmarshal(encoded, &c)
	c.ProviderID = v.ProviderID
	c.SigningKeyID = r.OldKeyID
	c.IssuedAt = "199"
	trust, e := WithRotationHistory(TrustContext{Scope: Scope{c.ChainID, c.ProtocolAddress}, ChainGeneration: c.ChainGeneration, ProviderID: c.ProviderID}, ring)
	if e != nil {
		t.Fatal(e)
	}
	sign := func(cert RootCertificate, seed string) RootCertificate {
		seedBytes, _ := decodeFixed(seed, 32)
		message, e := RootCertificateMessage(cert)
		if e != nil {
			t.Fatal(e)
		}
		cert.Signature = "0x" + hex.EncodeToString(ed25519.Sign(ed25519.NewKeyFromSeed(seedBytes), message))
		return cert
	}
	if e := AssertCertificateTrust(sign(c, v.OldSeed), trust); e != nil {
		t.Fatal("historical certificate rejected", e)
	}
	c.IssuedAt = "200"
	if AssertCertificateTrust(sign(c, v.OldSeed), trust) == nil {
		t.Fatal("retired key issued new certificate")
	}
	c.SigningKeyID = r.NewKeyID
	if e := AssertCertificateTrust(sign(c, v.NewSeed), trust); e != nil {
		t.Fatal("new active key rejected", e)
	}
	c.IssuedAt = "199"
	if AssertCertificateTrust(sign(c, v.NewSeed), trust) == nil {
		t.Fatal("preactivation certificate accepted")
	}
}
func TestProvenanceZeroSemantics(t *testing.T) {
	if _, e := RecognizedLeaf("1", "0"); e == nil {
		t.Fatal("zero commitment accepted")
	}
	if _, _, e := NoteOpening(Scope{"1", "1"}, "0", Opening{"1", "1", "1"}); e == nil {
		t.Fatal("zero token accepted")
	}
}

func TestPolicyHistoryDoesNotTrustReportedCurrentKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q map[string]any
		json.NewDecoder(r.Body).Decode(&q)
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q["id"], "result": map[string]any{"scope": Scope{"1", "2"}, "chainGeneration": "4", "providerId": "5", "signingKeyId": "7", "manifestSha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "keyRotations": []any{}}})
	}))
	defer server.Close()
	known, _ := word("6")
	unknown, _ := word("7")
	c := Client{Endpoint: server.URL, Trust: TrustContext{Scope: Scope{"1", "2"}, ChainGeneration: "4", ProviderID: "5", ManifestSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", SigningKeys: map[string]ed25519.PublicKey{known: make([]byte, 32)}}}
	if _, e := c.GetRules(context.Background()); e != nil {
		t.Fatal(e)
	}
	if c.Trust.SigningKeys[unknown] != nil {
		t.Fatal("untrusted reported key was accepted")
	}
	if _, e := c.GetStatus(context.Background()); e == nil {
		t.Fatal("unknown current signing key accepted outside history read")
	}
}
