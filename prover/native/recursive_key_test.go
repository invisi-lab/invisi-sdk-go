package native

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/invisi-labs/invisi-sdk-go/prover"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRecursiveKeyLoaderUsesBackendOfflineDirectory(t *testing.T) {
	dir := t.TempDir()
	raw, err := os.ReadFile("testdata/core.index.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "index.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	p, err := NewNativeProver(NativeOptions{Resources: ResourceOptions{Directory: dir, CacheDir: t.TempDir(), Offline: true}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.LoadVerificationKey(context.Background(), "transact_user_slots1", "noir-recursive", prover.Options{})
	if err == nil || !strings.Contains(err.Error(), "offline") {
		t.Fatalf("expected missing offline key, got %v", err)
	}
}

func TestRecursiveKeyLoaderLazyAuthenticatedCache(t *testing.T) {
	raw, err := os.ReadFile("testdata/core.index.json")
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile("../testdata/recursive/transact_user_slots2.vk")
	if err != nil {
		t.Fatal(err)
	}
	var index resourceIndex
	if err = json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	keyFile := index.Files["circuits/transact_user_slots2.noir-recursive.vk"]
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case "/index.json":
			w.Write(raw)
		case "/" + keyFile.SHA256 + ".vk":
			w.Write(binary)
		default:
			t.Errorf("unexpected resource requested %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	cache := filepath.Join(t.TempDir(), "cache")
	p, err := NewNativeProver(NativeOptions{Resources: ResourceOptions{CacheDir: cache, Sources: []string{server.URL}}})
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 0 {
		t.Fatal("constructor fetched resources")
	}
	if _, err = os.Stat(cache); !os.IsNotExist(err) {
		t.Fatal("constructor wrote resource cache")
	}
	key, err := prover.LoadRecursiveVerificationKey(context.Background(), "transact_user_slots2", p, prover.Options{})
	if err != nil || len(key.Fields) != 115 {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("expected only index and key downloads, got %d", hits.Load())
	}
	offline, err := NewNativeProver(NativeOptions{Resources: ResourceOptions{CacheDir: cache, Offline: true, Sources: []string{server.URL}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prover.LoadRecursiveVerificationKey(context.Background(), "transact_user_slots2", offline, prover.Options{}); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatal("offline read fetched resources")
	}
	if err = os.WriteFile(filepath.Join(cache, keyFile.SHA256+".vk"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = prover.LoadRecursiveVerificationKey(context.Background(), "transact_user_slots2", offline, prover.Options{}); err == nil {
		t.Fatal("corrupt cached key accepted after previous success")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.LoadVerificationKey(ctx, "transact_user_slots2", "noir-recursive", prover.Options{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatal("cancelled loader fetched resources")
	}
}
