package provenance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	protocol "github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

type rewriteTransport struct{ target string }

func (r rewriteTransport) RoundTrip(q *http.Request) (*http.Response, error) {
	if q.URL.Scheme != "https" || q.URL.Host != "provenance.invisipay.fi" {
		panic("official identity changed")
	}
	q = q.Clone(q.Context())
	q.URL.Scheme = "http"
	q.URL.Host = strings.TrimPrefix(r.target, "http://")
	return http.DefaultTransport.RoundTrip(q)
}
func TestOfficialBootstrapPinsScopeAndRetriesFailure(t *testing.T) {
	var calls atomic.Int32
	manifest := sha256.Sum256(protocol.ManifestBytes())
	key := make([]byte, 32)
	id := sha256.Sum256(key)
	word, _ := CanonicalWord("1")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if r.URL.Path != "/v1/chains/1" {
			t.Error(r.URL.Path)
		}
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["method"] != "provenance_getRules" {
			t.Error(req)
		}
		scope := Scope{ChainID: "1", ProtocolAddress: "1"}
		if n == 1 {
			scope.ChainID = "2"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": map[string]any{"scope": scope, "manifestSha256": hex.EncodeToString(manifest[:]), "chainGeneration": "1", "providerId": word, "initialSigningPublicKey": "0x" + hex.EncodeToString(key), "signingKeyId": "0x" + hex.EncodeToString(id[:]), "keyRotations": []any{}}})
	}))
	defer server.Close()
	clients := OfficialClients{HTTP: &http.Client{Transport: rewriteTransport{server.URL}}}
	scope := Scope{ChainID: "1", ProtocolAddress: "1"}
	if _, e := clients.Get(context.Background(), scope); e == nil {
		t.Fatal("wrong deployment accepted")
	}
	first, e := clients.Get(context.Background(), scope)
	if e != nil {
		t.Fatal(e)
	}
	second, e := clients.Get(context.Background(), Scope{ChainID: "0x1", ProtocolAddress: word})
	if e != nil || first != second || calls.Load() != 2 {
		t.Fatal("not scoped cached", e, calls.Load())
	}
	if first.Trust.SigningKeys["0x"+hex.EncodeToString(id[:])] == nil {
		t.Fatal("missing key pin")
	}
}

func TestOfficialBootstrapRejectsUntrustedRules(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"manifest", func(r map[string]any) { r["manifestSha256"] = "wrong" }},
		{"generation", func(r map[string]any) { r["chainGeneration"] = "0" }},
		{"active key", func(r map[string]any) { r["signingKeyId"] = "1" }},
		{"rotations missing", func(r map[string]any) { delete(r, "keyRotations") }},
		{"rotations bounded", func(r map[string]any) { r["keyRotations"] = make([]RotationRecord, 129) }},
		{"initial key", func(r map[string]any) { r["initialSigningPublicKey"] = "0x01" }},
		{"unverified rotation", func(r map[string]any) { r["keyRotations"] = []RotationRecord{{Version: "1"}} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			manifest := sha256.Sum256(protocol.ManifestBytes())
			key := make([]byte, 32)
			id := sha256.Sum256(key)
			provider, _ := word("1")
			rules := map[string]any{"scope": Scope{ChainID: "1", ProtocolAddress: "1"}, "manifestSha256": hex.EncodeToString(manifest[:]), "chainGeneration": "1", "providerId": provider, "initialSigningPublicKey": "0x" + hex.EncodeToString(key), "signingKeyId": "0x" + hex.EncodeToString(id[:]), "keyRotations": []any{}}
			test.mutate(rules)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": rules})
			}))
			defer server.Close()
			if _, err := (&OfficialClients{HTTP: &http.Client{Transport: rewriteTransport{server.URL}}}).Get(context.Background(), Scope{ChainID: "1", ProtocolAddress: "1"}); err == nil {
				t.Fatal("untrusted bootstrap accepted")
			}
		})
	}
}
func TestOfficialBootstrapRejectsRedirectAndOversize(t *testing.T) {
	for _, oversize := range []bool{false, true} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if oversize {
				_, _ = w.Write([]byte(strings.Repeat(" ", 256*1024+1)))
			} else {
				http.Redirect(w, r, "https://other.invalid/", http.StatusFound)
			}
		}))
		_, err := (&OfficialClients{HTTP: &http.Client{Transport: rewriteTransport{server.URL}}}).Get(context.Background(), Scope{ChainID: "1", ProtocolAddress: "1"})
		server.Close()
		if err == nil {
			t.Fatal("redirect or oversized bootstrap accepted")
		}
	}
}
