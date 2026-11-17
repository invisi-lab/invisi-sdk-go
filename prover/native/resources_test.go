package native

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func hashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func fixtureStore(t *testing.T, mutate func(*resourceIndex)) (*resourceLoader, *atomic.Int32, string) {
	t.Helper()
	data := []byte(`{"bytecode":"tiny-test-only"}`)
	digest := hashBytes(data)
	idx := resourceIndex{Format: "invisi.prover-assets.v1", NoirVersion: "1.0.0", BBVersion: "6.0.0", Files: map[string]resourceFile{"circuits/test.json": {SHA256: digest, Size: int64(len(data))}}}
	if mutate != nil {
		mutate(&idx)
	}
	raw, _ := json.Marshal(idx)
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path == "/index.json" {
			w.Write(raw)
		} else if r.URL.Path == "/"+digest+".json" {
			w.Write(data)
		} else {
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	loader, e := newResourceLoader(ResourceOptions{CacheDir: t.TempDir(), Sources: []string{server.URL}}, hashBytes(raw))
	if e != nil {
		t.Fatal(e)
	}
	return loader, &hits, digest
}
func TestResourceLazyConcurrentCacheAndOffline(t *testing.T) {
	l, hits, digest := fixtureStore(t, nil)
	if hits.Load() != 0 {
		t.Fatal("construction fetched")
	}
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if _, e := l.read(context.Background(), "circuits/test.json"); e != nil {
				t.Error(e)
			}
		})
	}
	wg.Wait()
	if hits.Load() != 2 {
		t.Fatalf("not coalesced: %d", hits.Load())
	}
	l.options.Offline = true
	if _, e := l.read(context.Background(), "circuits/test.json"); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(l.options.CacheDir, digest+".json"), []byte("corrupt"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := l.read(context.Background(), "circuits/test.json"); e == nil {
		t.Fatal("corrupt offline cache accepted")
	}
	l.options.Offline = false
	if _, e := l.read(context.Background(), "circuits/test.json"); e != nil {
		t.Fatal(e)
	}
}
func TestResourceRejectsHashSizeAndVersion(t *testing.T) {
	for _, change := range []func(*resourceIndex){func(i *resourceIndex) {
		f := i.Files["circuits/test.json"]
		f.SHA256 = string(make([]byte, 64))
		i.Files["circuits/test.json"] = f
	}, func(i *resourceIndex) {
		f := i.Files["circuits/test.json"]
		f.Size--
		i.Files["circuits/test.json"] = f
	}, func(i *resourceIndex) { i.BBVersion = "6.1.0" }} {
		l, _, _ := fixtureStore(t, change)
		if _, e := l.read(context.Background(), "circuits/test.json"); e == nil {
			t.Fatal("untrusted resource accepted")
		}
	}
}
func TestResourceIndexPinCancellationAndMirrors(t *testing.T) {
	l, _, _ := fixtureStore(t, nil)
	l.pin = hashBytes([]byte("wrong"))
	if _, e := l.read(context.Background(), "circuits/test.json"); e == nil {
		t.Fatal("index pin ignored")
	}
	l, _, _ = fixtureStore(t, nil)
	l.options.Sources = append([]string{"http://127.0.0.1:1"}, l.options.Sources...)
	if _, e := l.read(context.Background(), "circuits/test.json"); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := l.read(ctx, "circuits/test.json"); e == nil {
		t.Fatal("cancellation ignored")
	}
}

func TestResourceRejectsDownloadedWrongBytesAndOversize(t *testing.T) {
	for _, body := range []string{"evil", "right-and-unbounded"} {
		t.Run(body, func(t *testing.T) {
			want := []byte("good")
			hash := hashBytes(want)
			raw, _ := json.Marshal(resourceIndex{Format: "invisi.prover-assets.v1", NoirVersion: "1.0.0", BBVersion: "6.0.0", Files: map[string]resourceFile{"circuits/test.json": {SHA256: hash, Size: int64(len(want))}}})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/index.json" {
					w.Write(raw)
				} else {
					w.Write([]byte(body))
				}
			}))
			defer server.Close()
			l, e := newResourceLoader(ResourceOptions{CacheDir: t.TempDir(), Sources: []string{server.URL}}, hashBytes(raw))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = l.read(context.Background(), "circuits/test.json"); e == nil {
				t.Fatal("downloaded untrusted bytes accepted")
			}
			if _, e = os.Stat(filepath.Join(l.options.CacheDir, hash+".json")); !os.IsNotExist(e) {
				t.Fatal("untrusted object persisted")
			}
		})
	}
}
func TestResourceOfflineDirectoryAndCorruptIndex(t *testing.T) {
	l, _, _ := fixtureStore(t, nil)
	if _, e := l.read(context.Background(), "circuits/test.json"); e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(l.options.CacheDir, l.pin+".index.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(l.options.CacheDir, "index.json"), raw, 0600); e != nil {
		t.Fatal(e)
	}
	offline, e := newResourceLoader(ResourceOptions{Directory: l.options.CacheDir, CacheDir: t.TempDir(), Offline: true}, l.pin)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = offline.read(context.Background(), "circuits/test.json"); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"index.json", l.pin + ".index.json"} {
		if e = os.WriteFile(filepath.Join(l.options.CacheDir, name), []byte("{}"), 0600); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = offline.read(context.Background(), "circuits/test.json"); e == nil {
		t.Fatal("changed snapshot index accepted")
	}
}
func TestResourceCancellationStopsInflightDownload(t *testing.T) {
	entered := make(chan struct{})
	ended := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-r.Context().Done(); close(ended) }))
	defer server.Close()
	l, e := newResourceLoader(ResourceOptions{CacheDir: t.TempDir(), Sources: []string{server.URL}}, hashBytes([]byte("pin")))
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	go func() { _, e := l.read(ctx, "circuits/test.json"); done <- e }()
	<-entered
	cancel()
	if e := <-done; !errors.Is(e, context.Canceled) {
		t.Fatal("cancellation lost", e)
	}
	select {
	case <-ended:
	case <-time.After(time.Second):
		t.Fatal("HTTP request not canceled")
	}
}

type observedContext struct {
	context.Context
	once    sync.Once
	waiting chan struct{}
}

func (c *observedContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}
func TestResourceLeaderCancellationDoesNotAbortActiveFollower(t *testing.T) {
	data := []byte("good")
	entered := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		select {
		case <-release:
			w.Write(data)
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	l, e := newResourceLoader(ResourceOptions{CacheDir: t.TempDir(), Sources: []string{server.URL}}, hashBytes([]byte("pin")))
	if e != nil {
		t.Fatal(e)
	}
	leader, cancel := context.WithCancel(context.Background())
	defer cancel()
	leaderDone := make(chan error, 1)
	load := func(ctx context.Context) error {
		_, e := l.object(ctx, hashBytes(data)+".json", hashBytes(data), int64(len(data)), int64(len(data)))
		return e
	}
	go func() { leaderDone <- load(leader) }()
	<-entered
	followerCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	follower := &observedContext{Context: followerCtx, waiting: make(chan struct{})}
	followerDone := make(chan error, 1)
	go func() { followerDone <- load(follower) }()
	<-follower.waiting
	cancel()
	if e := <-leaderDone; !errors.Is(e, context.Canceled) {
		t.Fatal("leader did not cancel promptly", e)
	}
	close(release)
	if e := <-followerDone; e != nil {
		t.Fatal("leader canceled another active caller", e)
	}
}
func TestResourceDefaultSourceUsesPinnedAssetsMirror(t *testing.T) {
	l, e := newResourceLoader(ResourceOptions{CacheDir: t.TempDir()}, hashBytes([]byte("pin")))
	if e != nil {
		t.Fatal(e)
	}
	want := "https://assets.invisipay.fi/prover/sets/" + l.pin + "/"
	if len(l.options.Sources) < 1 || l.options.Sources[0] != want {
		t.Fatal("default pinned assets mirror missing", l.options.Sources)
	}
}

func TestResourceFinalSubscriberCancelsSharedDownload(t *testing.T) {
	entered := make(chan struct{})
	aborted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-r.Context().Done(); close(aborted) }))
	defer server.Close()
	l, e := newResourceLoader(ResourceOptions{CacheDir: t.TempDir(), Sources: []string{server.URL}}, hashBytes([]byte("pin")))
	if e != nil {
		t.Fatal(e)
	}
	leader, cancelLeader := context.WithCancel(context.Background())
	defer cancelLeader()
	followerCtx, cancelFollower := context.WithCancel(context.Background())
	defer cancelFollower()
	leaderDone, followerDone := make(chan error, 1), make(chan error, 1)
	load := func(ctx context.Context) error { _, err := l.object(ctx, "index.json", l.pin, -1, 2<<20); return err }
	go func() { leaderDone <- load(leader) }()
	<-entered
	follower := &observedContext{Context: followerCtx, waiting: make(chan struct{})}
	go func() { followerDone <- load(follower) }()
	<-follower.waiting
	cancelLeader()
	if e := <-leaderDone; !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	select {
	case <-aborted:
		t.Fatal("download aborted with active follower")
	default:
	}
	cancelFollower()
	if e := <-followerDone; !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	select {
	case <-aborted:
	case <-time.After(time.Second):
		t.Fatal("last subscriber did not abort download")
	}
}

func TestSharedResourceCacheDefaultsAndPrecedence(t *testing.T) {
	t.Setenv("INVISI_PROVER_CACHE_DIR", "")
	base, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	loader, err := newResourceLoader(ResourceOptions{}, resourceIndexSHA256)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(base, "invisi", "prover-assets"); loader.options.CacheDir != want {
		t.Fatalf("cache = %q, want %q", loader.options.CacheDir, want)
	}
	t.Setenv("INVISI_PROVER_CACHE_DIR", filepath.Join(t.TempDir(), "environment"))
	loader, err = newResourceLoader(ResourceOptions{}, resourceIndexSHA256)
	if err != nil {
		t.Fatal(err)
	}
	if loader.options.CacheDir != os.Getenv("INVISI_PROVER_CACHE_DIR") {
		t.Fatal("environment override ignored")
	}
	explicit := filepath.Join(t.TempDir(), "explicit")
	loader, err = newResourceLoader(ResourceOptions{CacheDir: explicit}, resourceIndexSHA256)
	if err != nil || loader.options.CacheDir != explicit {
		t.Fatal("explicit override ignored")
	}
}
