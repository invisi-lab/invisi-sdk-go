package native

import (
	"context"
	"encoding/json"
	protocol "github.com/invisi-labs/invisi-sdk-go/protocol/provenance"
	source "github.com/invisi-labs/invisi-sdk-go/prover/provenance"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSplitCapabilityResources(t *testing.T) {
	coreDir, provDir := t.TempDir(), t.TempDir()
	put := func(dir, filename string, b []byte) {
		t.Helper()
		if e := os.WriteFile(filepath.Join(dir, filename), b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	for name, dir := range map[string]string{"core": coreDir, "provenance": provDir} {
		b, e := os.ReadFile("testdata/" + name + ".index.json")
		if e != nil {
			t.Fatal(e)
		}
		put(dir, "index.json", b)
	}
	b, _ := os.ReadFile("testdata/core.index.json")
	var index resourceIndex
	if e := json.Unmarshal(b, &index); e != nil {
		t.Fatal(e)
	}
	vk, e := os.ReadFile("testdata/core.vk")
	if e != nil {
		t.Fatal(e)
	}
	put(coreDir, index.Files["circuits/transact_user_slots2.evm.vk"].SHA256+".vk", vk)
	manifest := protocol.ManifestBytes()
	put(provDir, hashBytes(manifest)+".json", manifest)
	options := ResourceOptions{Directory: coreDir, ProvenanceDirectory: provDir, CacheDir: t.TempDir(), Offline: true}
	core, e := newResourceLoader(options, coreResourcePin(options))
	if e != nil {
		t.Fatal(e)
	}
	if core.pin != "85a5d1dffe40fc965ab4345a28b632f463c3ad861c44d5845850fbc7447fd8f6" {
		t.Fatal(core.pin)
	}
	if _, e = core.read(context.Background(), "circuits/transact_user_slots2.evm.vk"); e != nil {
		t.Fatal(e)
	}
	prov, e := newProvenanceResourceLoader(options)
	if e != nil {
		t.Fatal(e)
	}
	if prov.pin != "8d5a3e60e3e5cbe315150511207d7aff67f4e2924148d396b67d82017ecd1f1c" {
		t.Fatal(prov.pin)
	}
	if _, e = prov.read(context.Background(), "provenance/manifest.json"); e != nil {
		t.Fatal(e)
	}
	if _, e = core.read(context.Background(), "provenance/manifest.json"); e == nil {
		t.Fatal("Core consumed Provenance")
	}
	options.LegacyCombined = true
	if coreResourcePin(options) != LegacyResourceIndexSHA256 || provenanceResourcePin(options) != LegacyResourceIndexSHA256 {
		t.Fatal("legacy must be explicit")
	}
	if circuitResourcePin("create_credit_initial_payment", ResourceOptions{}) != resourceIndexSHA256 {
		t.Fatal("initial-payment must use Core set")
	}
}
func TestProvenanceSourceIsIsolatedWithoutLegacyDowngrade(t *testing.T) {
	var coreCalls, provCalls int
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { coreCalls++; http.NotFound(w, r) }))
	defer core.Close()
	prov := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provCalls++
		if strings.Contains(r.URL.Path, LegacyResourceIndexSHA256) {
			t.Error("legacy downgrade")
		}
		http.NotFound(w, r)
	}))
	defer prov.Close()
	l, e := newProvenanceResourceLoader(ResourceOptions{Sources: []string{core.URL}, ProvenanceSources: []string{prov.URL}, CacheDir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = l.read(context.Background(), "provenance/manifest.json"); e == nil {
		t.Fatal("expected missing catalog error")
	}
	if coreCalls != 0 || provCalls == 0 {
		t.Fatal(coreCalls, provCalls)
	}
}
func TestExplicitEmptyNativeSourcesRemainDisabled(t *testing.T) {
	p, e := NewNativeProver(NativeOptions{Resources: ResourceOptions{Sources: []string{}, ProvenanceSources: []string{}, CacheDir: t.TempDir()}})
	if e != nil {
		t.Fatal(e)
	}
	if len(p.resources.options.Sources) != 0 {
		t.Fatal("explicitly disabled Core sources were restored")
	}
	l, e := newProvenanceResourceLoader(p.options.Resources)
	if e != nil {
		t.Fatal(e)
	}
	if len(l.options.Sources) != 0 {
		t.Fatal("explicitly disabled Provenance sources were restored")
	}
}
func TestRetainedLegacyCatalogsRemainExplicitlyLoadable(t *testing.T) {
	dir := os.Getenv("INVISI_TEST_LEGACY_RESOURCES")
	if dir == "" {
		t.Skip("retained authenticated legacy catalogs required")
	}
	options := ResourceOptions{Directory: dir, CacheDir: t.TempDir(), Offline: true, LegacyCombined: true}
	for _, name := range []string{"circuits/transact_user_slots2.evm.vk", "circuits/create_credit_initial_payment.evm.vk", "provenance/manifest.json"} {
		var l *resourceLoader
		var e error
		if strings.HasPrefix(name, "provenance/") {
			l, e = newProvenanceResourceLoader(options)
		} else {
			circuit := "transact_user_slots2"
			if strings.Contains(name, "initial_payment") {
				circuit = "create_credit_initial_payment"
			}
			l, e = newResourceLoader(options, circuitResourcePin(circuit, options))
		}
		if e != nil {
			t.Fatal(e)
		}
		if _, e = l.read(context.Background(), name); e != nil {
			t.Fatal(name, e)
		}
	}
}

func TestCapabilitiesShareProvenanceColdDownloads(t *testing.T) {
	index, e := os.ReadFile("testdata/provenance.index.json")
	if e != nil {
		t.Fatal(e)
	}
	var calls atomic.Int32
	started, second, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "index.json") {
			n := calls.Add(1)
			if n == 1 {
				close(started)
			}
			if n == 2 {
				close(second)
			}
			<-release
			_, _ = w.Write(index)
			return
		}
		if strings.Contains(r.URL.Path, hashBytes(protocol.ManifestBytes())) {
			_, _ = w.Write(protocol.ManifestBytes())
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	defer releaseOnce.Do(func() { close(release) })
	p, e := NewNativeProver(NativeOptions{Resources: ResourceOptions{Sources: []string{}, ProvenanceSources: []string{server.URL}, CacheDir: t.TempDir()}})
	if e != nil {
		t.Fatal(e)
	}
	_, a, e := p.GetProvenanceCapability()
	if e != nil {
		t.Fatal(e)
	}
	_, b, e := p.GetProvenanceCapability()
	if e != nil {
		t.Fatal(e)
	}
	done := make(chan error, 2)
	for _, capability := range []source.OfflineProver{a, b} {
		go func(c source.OfflineProver) {
			_, e := c.(source.NativeProver).LoadAssetsContext(context.Background(), 10)
			done <- e
		}(capability)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("no Provenance request")
	}
	select {
	case <-second:
	case <-time.After(200 * time.Millisecond):
	}
	releaseOnce.Do(func() { close(release) })
	for i := 0; i < 2; i++ {
		if e := <-done; e == nil {
			t.Fatal("missing artifact unexpectedly loaded")
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("separate capabilities downloaded the same index %d times", calls.Load())
	}
}
