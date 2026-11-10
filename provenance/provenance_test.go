package provenance

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func vectors(t *testing.T) map[string]any {
	t.Helper()
	b, e := os.ReadFile("testdata/golden-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var v map[string]any
	if e := json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func num(v any) string { return fmt.Sprint(v) }
func requestFromVector(t *testing.T, v map[string]any) WitnessRequest {
	t.Helper()
	r := WitnessRequest{Scope: Scope{num(v["chain_id"]), num(v["protocol_address"])}, Capacity: len(v["owners"].([]any)), TokenAddress: num(v["token_address"]), ChainGeneration: num(v["chain_generation"]), OperationDigest: num(v["operation_digest_lo"])}
	count, _ := Uint(num(v["input_count"]), 64)
	for i := 0; i < int(count.Int64()); i++ {
		input := WitnessInput{Opening: Opening{num(v["owners"].([]any)[i]), num(v["nonces"].([]any)[i]), num(v["amounts"].([]any)[i])}, SourceIndex: int(v["source_indices"].([]any)[i].(float64)), LeafIndex: uint64(v["leaf_indices"].([]any)[i].(float64))}
		for _, p := range v["paths"].([]any)[i].([]any) {
			input.Path = append(input.Path, num(p))
		}
		r.Inputs = append(r.Inputs, input)
	}
	sc, _ := Uint(num(v["source_count"]), 64)
	for i := 0; i < int(sc.Int64()); i++ {
		r.Sources = append(r.Sources, Source{num(v["tree_ids"].([]any)[i]), num(v["roots"].([]any)[i]), num(v["leaf_counts"].([]any)[i])})
	}
	return r
}
func TestSharedVectors(t *testing.T) {
	v := vectors(t)
	note := v["note"].(map[string]any)
	c := note["context"].(map[string]any)
	s := Scope{num(c["chainId"]), num(c["protocolAddress"])}
	scope, e := ScopeHash(s)
	if e != nil || scope != note["scope"] {
		t.Fatalf("scope %s %v", scope, e)
	}
	commit, n, e := NoteOpening(s, num(c["tokenAddress"]), Opening{num(c["owner"]), num(c["nonce"]), num(c["amount"])})
	if e != nil || commit != note["commitment"] || n != note["nullifier"] {
		t.Fatalf("note %v", e)
	}
	leaf, e := RecognizedLeaf(scope, commit)
	if e != nil || leaf != note["recognizedLeaf"] {
		t.Fatalf("leaf %v", e)
	}
	op := v["operation"].(map[string]any)
	pub := []string{}
	for _, p := range op["publicInputs"].([]any) {
		pub = append(pub, num(p))
	}
	digest, e := OperationDigest(Scope{num(op["chainId"]), num(op["protocolAddress"])}, int(op["mainCircuitCapacity"].(float64)), pub)
	if e != nil || digest != op["digest"] {
		t.Fatalf("operation %s %v", digest, e)
	}
	for _, name := range []string{"witnessK10SingleInput", "witnessK10CrossTree"} {
		raw := v[name].(map[string]any)
		w, e := BuildWitness(requestFromVector(t, raw))
		if e != nil {
			t.Fatal(e)
		}
		for _, k := range publicInputNames {
			if w[k] != num(raw[k]) {
				t.Fatalf("%s %s mismatch", name, k)
			}
		}
		for _, k := range []string{"owners", "nonces", "amounts", "source_indices", "leaf_indices", "tree_ids", "roots", "leaf_counts"} {
			expected := []string{}
			for _, x := range raw[k].([]any) {
				expected = append(expected, num(x))
			}
			if !reflect.DeepEqual(w[k], expected) {
				t.Fatalf("%s %s padding mismatch", name, k)
			}
		}
	}
}
func TestCertificateVectorAndMutation(t *testing.T) {
	v := vectors(t)["certificate"].(map[string]any)
	fields := map[string]string{}
	for k, v := range v["fields"].(map[string]any) {
		fields[k] = num(v)
	}
	fields["signature"] = "0x" + num(v["signatureHex"])
	b, _ := json.Marshal(fields)
	var c RootCertificate
	if e := json.Unmarshal(b, &c); e != nil {
		t.Fatal(e)
	}
	payload, e := RootCertificatePayload(c)
	if e != nil || len(payload) != 269 || hex.EncodeToString(payload) != v["payloadHex"] {
		t.Fatal("payload mismatch", e)
	}
	empty := c
	empty.LeafCount = "0"
	if payload, e := RootCertificatePayload(empty); e != nil || len(payload) != 269 {
		t.Fatal("empty certificate rejected", e)
	}
	message, e := RootCertificateMessage(c)
	if e != nil || hex.EncodeToString(message) != v["messageHex"] {
		t.Fatal("message mismatch", e)
	}
	checkpoint, e := CheckpointHash(c)
	if e != nil || checkpoint != v["checkpointHash"] {
		t.Fatal("checkpoint mismatch", e)
	}
	key, _ := hex.DecodeString(num(v["publicKeySpkiHex"]))
	if !VerifyRootCertificate(c, ed25519.PublicKey(key[len(key)-32:])) {
		t.Fatal("valid signature rejected")
	}
	c.Root = "11"
	if VerifyRootCertificate(c, ed25519.PublicKey(key[len(key)-32:])) {
		t.Fatal("mutation accepted")
	}
}
func TestLocalMembershipAndRejects(t *testing.T) {
	v := vectors(t)
	r := requestFromVector(t, v["witnessK10SingleInput"].(map[string]any))
	tree := RecognizedTree{}
	leaf := num(v["note"].(map[string]any)["recognizedLeaf"])
	if _, e := tree.Append(leaf); e != nil {
		t.Fatal(e)
	}
	membership, found, e := tree.MembershipForCommitment(r.Scope, num(v["note"].(map[string]any)["commitment"]))
	if e != nil || !found || membership.LeafIndex != 0 {
		t.Fatal("local commitment membership lookup failed", e)
	}
	path, root, e := tree.Path(0)
	if e != nil || root != r.Sources[0].Root || !reflect.DeepEqual(path, r.Inputs[0].Path) {
		t.Fatal("local replay mismatch", e)
	}
	r.Inputs[0].Path[0] = "1"
	if _, e := BuildWitness(r); e == nil {
		t.Fatal("corrupt path accepted")
	}
	for _, x := range []string{"-1", "01", "1.5", " 1"} {
		if _, e := Uint(x, 256); e == nil {
			t.Fatal("noncanonical integer accepted")
		}
	}
	if _, e := Field(modulus.String()); e == nil {
		t.Fatal("noncanonical field accepted")
	}
}
func TestPublicClientScopeAndGeneration(t *testing.T) {
	generation := "4"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chains/1" || r.Method != "POST" {
			t.Error("wrong transport")
		}
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		body, _ := json.Marshal(req)
		if bytesContainSecrets(string(body)) {
			t.Error("private query")
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req["id"], "result": map[string]any{"scope": Scope{"1", "2"}, "chainGeneration": generation, "providerId": "5", "signingKeyId": "6", "manifestSha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}})
	}))
	defer server.Close()
	c := Client{Endpoint: server.URL, Trust: TrustContext{Scope: Scope{"1", "2"}, ChainGeneration: "4", ProviderID: "5", ManifestSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", SigningKeys: map[string]ed25519.PublicKey{"0x0000000000000000000000000000000000000000000000000000000000000006": make([]byte, 32)}}}
	if _, e := c.EvaluateTransaction(context.Background(), map[string]string{"calldata": "0x12"}, []ProofSidecar{}); e != nil {
		t.Fatal(e)
	}
	generation = "5"
	if _, e := c.GetStatus(context.Background()); e == nil {
		t.Fatal("generation mismatch accepted")
	}
}
func bytesContainSecrets(s string) bool {
	for _, x := range []string{"witness", "nonce", "opening", "privateKey"} {
		for i := 0; i+len(x) <= len(s); i++ {
			if s[i:i+len(x)] == x {
				return true
			}
		}
	}
	return false
}
func TestNativeOfflineProof(t *testing.T) {
	root := os.Getenv("INVISI_PROTOCOL_ROOT")
	if root == "" {
		t.Skip("set INVISI_PROTOCOL_ROOT for genuine local Noir/BB proof")
	}
	manifest, e := os.ReadFile(filepath.Join(root, "circuits/provenance/manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	const pin = "6107789c5c0d24c4130bff05da6d389336feb5ccb247c4030eae5198ab05d762"
	identity, e := AuthenticateManifest(manifest, pin)
	if e != nil {
		t.Fatal(e)
	}
	fixture := "witnessK10SingleInput"
	if os.Getenv("INVISI_PROVENANCE_CAPACITY") == "10" {
		fixture = "witnessK10CrossTree"
	}
	r := requestFromVector(t, vectors(t)[fixture].(map[string]any))
	w, e := BuildWitness(r)
	if e != nil {
		t.Fatal(e)
	}
	expected := []string{}
	for _, k := range publicInputNames {
		v, _ := word(w[k].(string))
		expected = append(expected, v)
	}
	p := NativeProver{NoirExecutePath: filepath.Join(root, ".tools/noir/noir-execute"), BBPath: filepath.Join(root, ".tools/bb/bb"), TrustedManifestSHA256: pin, LoadAssets: func(k int) (LocalAssets, error) {
		for _, entry := range identity.Manifest.Circuits {
			if entry.Capacity == k {
				a, e := os.ReadFile(filepath.Join(root, entry.Artifact))
				if e != nil {
					return LocalAssets{}, e
				}
				vk, e := os.ReadFile(filepath.Join(root, entry.VK))
				return LocalAssets{manifest, a, vk}, e
			}
		}
		return LocalAssets{}, fmt.Errorf("capacity")
	}}
	result, e := p.Prove(context.Background(), identity, r.Capacity, w, expected)
	if e != nil {
		t.Fatal(e)
	}
	if len(result.PublicInputs) != 11 || len(result.Proof) < 1000 {
		t.Fatal("real proof absent")
	}
}

type localTestProver struct{ tamper bool }

func (p localTestProver) Prove(_ context.Context, _ ManifestIdentity, _ int, _ Witness, expected []string) (ProofResult, error) {
	pub := append([]string(nil), expected...)
	if p.tamper {
		pub[0], _ = word("1")
	}
	return ProofResult{"0x01", pub}, nil
}
func TestSidecarTrustBindingAndReplay(t *testing.T) {
	raw := vectors(t)
	r := requestFromVector(t, raw["witnessK10SingleInput"].(map[string]any))
	manifest, e := os.ReadFile("manifest.json")
	if e != nil {
		t.Fatal(e)
	}
	const pin = "6107789c5c0d24c4130bff05da6d389336feb5ccb247c4030eae5198ab05d762"
	identity, e := AuthenticateManifest(manifest, pin)
	if e != nil {
		t.Fatal(e)
	}
	fields := map[string]string{}
	for k, v := range raw["certificate"].(map[string]any)["fields"].(map[string]any) {
		fields[k] = num(v)
	}
	data, _ := json.Marshal(fields)
	var cert RootCertificate
	json.Unmarshal(data, &cert)
	cert.ChainID = r.ChainID
	cert.ProtocolAddress = r.ProtocolAddress
	cert.ChainGeneration = r.ChainGeneration
	cert.RootSequence = "1"
	cert.TreeID = "1"
	cert.LeafCount = "1"
	cert.Root = r.Sources[0].Root
	key := ed25519.NewKeyFromSeed(make([]byte, 32))
	message, e := RootCertificateMessage(cert)
	if e != nil {
		t.Fatal(e)
	}
	cert.Signature = "0x" + hex.EncodeToString(ed25519.Sign(key, message))
	id, _ := word(cert.SigningKeyID)
	trust := TrustContext{Scope: r.Scope, ChainGeneration: r.ChainGeneration, ProviderID: cert.ProviderID, ManifestSHA256: pin, SigningKeys: map[string]ed25519.PublicKey{id: key.Public().(ed25519.PublicKey)}}
	sidecar, e := ExportSidecar(context.Background(), r, []RootCertificate{cert}, trust, identity, localTestProver{}, 0)
	if e != nil {
		t.Fatal(e)
	}
	if e := ValidateProofBundle(sidecar.ProofBundle, trust); e != nil {
		t.Fatal(e)
	}
	data, _ = json.Marshal(sidecar)
	if bytesContainSecrets(string(data)) {
		t.Fatal("private witness in exported sidecar")
	}
	if sidecar.ProofBundle.CircuitID != "provenance_k10" || sidecar.ProofBundle.Format != "invisi-provenance-proof-v1" {
		t.Fatal("wire identity")
	}
	bad := sidecar.ProofBundle
	bad.PublicInputs = append([]string(nil), bad.PublicInputs...)
	bad.PublicInputs[10], _ = word("0")
	if ValidateProofBundle(bad, trust) == nil {
		t.Fatal("tampered binding accepted")
	}
	if _, e := ExportSidecar(context.Background(), r, []RootCertificate{cert}, trust, identity, localTestProver{true}, 0); e == nil {
		t.Fatal("host statement substitution accepted")
	}
	corrupt := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		json.NewDecoder(req.Body).Decode(&body)
		if body["method"] == "provenance_getMembership" {
			path := append([]string(nil), r.Inputs[0].Path...)
			if corrupt {
				path[0] = "1"
			}
			json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": body["id"], "result": map[string]any{"scope": r.Scope, "chainGeneration": r.ChainGeneration, "certificate": cert, "commitment": raw["note"].(map[string]any)["commitment"], "leafIndex": "0", "path": path}})
			return
		}
		leaf := num(raw["note"].(map[string]any)["recognizedLeaf"])
		if corrupt {
			leaf = "1"
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": body["id"], "result": map[string]any{"scope": r.Scope, "chainGeneration": r.ChainGeneration, "certificate": cert, "rootId": "1:1", "nextIndex": "1", "items": []any{map[string]any{"scope": r.Scope, "noteCommitment": raw["note"].(map[string]any)["commitment"], "leaf": leaf, "chainGeneration": "1", "treeId": "1", "leafIndex": "0", "checkpointSequence": "1"}}}})
	}))
	defer server.Close()
	c := Client{Endpoint: server.URL, Trust: trust}
	tree, e := c.SyncRecognizedTree(context.Background(), "1:1", cert, 100)
	if e != nil || tree.LeafCount() != 1 {
		t.Fatal("replay failed", e)
	}
	commitment := num(raw["note"].(map[string]any)["commitment"])
	if _, e := c.GetMembership(context.Background(), commitment, "1:1", false); e == nil {
		t.Fatal("membership query disclosed interest without opt-in")
	}
	if _, e := c.GetMembership(context.Background(), commitment, "1:1", true); e != nil {
		t.Fatal("membership proof rejected", e)
	}
	corrupt = true
	if _, e := c.GetMembership(context.Background(), commitment, "1:1", true); e == nil {
		t.Fatal("corrupt server membership path accepted")
	}
	if _, e := c.SyncRecognizedTree(context.Background(), "1:1", cert, 100); e == nil {
		t.Fatal("corrupt replay accepted")
	}
	if _, e := c.EvaluateTransaction(context.Background(), map[string]any{"witness": "secret"}, nil); e == nil {
		t.Fatal("private query accepted")
	}
}

func TestMetadataOmitsPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64           `json:"id"`
			Params []map[string]any `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Params) != 1 || req.Params[0]["policyId"] != nil {
			t.Error("metadata included policy")
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"scope": Scope{"1", "2"}, "chainGeneration": "4"}})
	}))
	defer server.Close()
	c := Client{Endpoint: server.URL, Trust: TrustContext{Scope: Scope{"1", "2"}, ChainGeneration: "4"}}
	if _, err := c.GetVersion(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWaitProofJobCurrentWireTerminalAndUnknown(t *testing.T) {
	for _, status := range []string{"stale_generation", "unknown_status"} {
		t.Run(status, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				var request map[string]any
				json.NewDecoder(r.Body).Decode(&request)
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request["id"], "result": map[string]any{"scope": Scope{"1", "2"}, "chainGeneration": "2", "providerId": "5", "signingKeyId": "6", "manifestSha256": strings.Repeat("a", 64), "status": status}})
			}))
			defer server.Close()
			client := &Client{Endpoint: server.URL, Trust: TrustContext{Scope: Scope{"1", "2"}, ChainGeneration: "2", ProviderID: "5", ManifestSHA256: strings.Repeat("a", 64), SigningKeys: map[string]ed25519.PublicKey{"0x" + strings.Repeat("0", 63) + "6": make([]byte, 32)}}}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			result, e := client.WaitProofJob(ctx, "job", time.Millisecond)
			if status == "stale_generation" {
				var got string
				json.Unmarshal(result["status"], &got)
				if e != nil || got != status {
					t.Fatal(got, e)
				}
			} else if e == nil || !strings.Contains(e.Error(), "status") {
				t.Fatal("unknown job state accepted", e)
			}
			if calls != 1 {
				t.Fatalf("made %d requests instead of one", calls)
			}
		})
	}
}

func TestSubmitSourceProofWithoutTransactionHash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64                       `json:"id"`
			Method string                       `json:"method"`
			Params []map[string]json.RawMessage `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.Method != "provenance_submitSourceProof" || len(request.Params) != 1 {
			t.Error("unexpected source request")
		}
		q := request.Params[0]
		if len(q) != 2 || len(q["data"]) == 0 || string(q["proofSidecars"]) != "[]" {
			t.Error("unexpected source payload")
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"scope": Scope{"1", "2"}, "chainGeneration": "3", "providerId": "4", "signingKeyId": "5", "manifestSha256": strings.Repeat("a", 64)}})
	}))
	defer server.Close()
	key, _ := word("5")
	c := Client{Endpoint: server.URL, Trust: TrustContext{Scope: Scope{"1", "2"}, ChainGeneration: "3", ProviderID: "4", ManifestSHA256: strings.Repeat("a", 64), SigningKeys: map[string]ed25519.PublicKey{key: make([]byte, 32)}}}
	if _, err := c.SubmitSourceProof(context.Background(), "0x12", nil); err != nil {
		t.Fatal(err)
	}
}
