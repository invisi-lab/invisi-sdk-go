package native

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ResourceOptions selects public circuit assets. No witness or proof inputs are sent.
// Directory contains index.json and flat <sha256>.json/.vk objects, as published by Core.
// Sources overrides the default release URL and permits ordered mirrors; every source
// is authenticated using the SDK's fixed index pin. Offline disables all HTTP requests.
type ResourceOptions struct {
	// LegacyCombined explicitly selects historical combined catalogs; never a fallback.
	LegacyCombined      bool
	ProvenanceDirectory string
	ProvenanceSources   []string
	Directory           string
	CacheDir            string
	Sources             []string
	Offline             bool
	HTTPClient          *http.Client
}
type resourceFile struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type resourceIndex struct {
	Format      string                  `json:"format"`
	NoirVersion string                  `json:"noirVersion"`
	BBVersion   string                  `json:"bbVersion"`
	Files       map[string]resourceFile `json:"files"`
}
type resourceLoader struct {
	options ResourceOptions
	pin     string
	mu      sync.Mutex
	flights map[string]*resourceFlight
}

// A shared fetch is owned by its active subscribers, rather than its first caller.
type resourceFlight struct {
	done     chan struct{}
	cancel   context.CancelFunc
	waiters  int
	finished bool
	data     []byte
	err      error
}

func (l *resourceLoader) shared(ctx context.Context, key string, load func(context.Context) ([]byte, error)) ([]byte, error) {
	l.mu.Lock()
	if l.flights == nil {
		l.flights = make(map[string]*resourceFlight)
	}
	f := l.flights[key]
	if f == nil {
		sharedCtx, cancel := context.WithCancel(context.Background())
		f = &resourceFlight{done: make(chan struct{}), cancel: cancel}
		l.flights[key] = f
		go func() {
			data, err := load(sharedCtx)
			l.mu.Lock()
			f.data, f.err, f.finished = data, err, true
			if l.flights[key] == f {
				delete(l.flights, key)
			}
			close(f.done)
			l.mu.Unlock()
			cancel()
		}()
	}
	f.waiters++
	l.mu.Unlock()
	defer func() {
		l.mu.Lock()
		f.waiters--
		if f.waiters == 0 && !f.finished {
			f.cancel()
			if l.flights[key] == f {
				delete(l.flights, key)
			}
		}
		l.mu.Unlock()
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.done:
		return f.data, f.err
	}
}

const maxResourceSize int64 = 512 << 20

var resourceDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

func newResourceLoader(options ResourceOptions, pin string) (*resourceLoader, error) {
	if !resourceDigest.MatchString(pin) {
		return nil, errors.New("invalid pinned resource index")
	}
	if options.CacheDir == "" {
		options.CacheDir = os.Getenv("INVISI_PROVER_CACHE_DIR")
	}
	if options.CacheDir == "" {
		base, e := os.UserCacheDir()
		if e != nil {
			return nil, e
		}
		options.CacheDir = filepath.Join(base, "invisi", "prover-assets")
	}
	if options.Sources == nil {
		options.Sources = []string{"https://assets.invisipay.fi/prover/sets/" + pin + "/", "https://github.com/invisi-labs/" + resourceRepository(pin) + "/releases/download/prover-assets-" + pin + "/"}
	}
	options.Sources = append([]string(nil), options.Sources...)
	for _, source := range options.Sources {
		u, e := url.Parse(source)
		if e != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
			return nil, errors.New("invalid resource source URL")
		}
	}
	if options.HTTPClient == nil {
		options.HTTPClient = &http.Client{Timeout: 2 * time.Minute}
	}
	return &resourceLoader{options: options, pin: pin}, nil
}
func resourceValid(data []byte, hash string, size int64) bool {
	sum := sha256.Sum256(data)
	return (size < 0 || int64(len(data)) == size) && hex.EncodeToString(sum[:]) == hash
}
func readBounded(r io.Reader, limit int64) ([]byte, error) {
	b, e := io.ReadAll(io.LimitReader(r, limit+1))
	if e == nil && int64(len(b)) > limit {
		e = errors.New("resource exceeds expected size")
	}
	return b, e
}
func (l *resourceLoader) object(ctx context.Context, name, hash string, size, limit int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cacheName := name
	if name == "index.json" {
		cacheName = hash + ".index.json"
	}
	return l.shared(ctx, cacheName, func(ctx context.Context) ([]byte, error) {
		// Recheck disk on every operation, including previously successful reads.
		for _, dir := range []string{l.options.CacheDir, l.options.Directory} {
			if dir == "" {
				continue
			}
			f, e := os.Open(filepath.Join(dir, cacheName))
			if e != nil && name == "index.json" && dir == l.options.Directory {
				f, e = os.Open(filepath.Join(dir, name))
			}
			if e == nil {
				b, err := readBounded(f, limit)
				f.Close()
				if err == nil && resourceValid(b, hash, size) {
					return b, nil
				}
			}
		}
		if l.options.Offline {
			return nil, fmt.Errorf("resource %s unavailable or corrupt in offline mode", name)
		}
		var last error
		for _, base := range l.options.Sources {
			requestCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			req, e := http.NewRequestWithContext(requestCtx, http.MethodGet, strings.TrimRight(base, "/")+"/"+name, nil)
			if e != nil {
				cancel()
				return nil, e
			}
			res, e := l.options.HTTPClient.Do(req)
			if e != nil {
				cancel()
				last = e
				continue
			}
			var data []byte
			if res.StatusCode != http.StatusOK {
				e = fmt.Errorf("resource HTTP status %d", res.StatusCode)
			} else {
				data, e = readBounded(res.Body, limit)
			}
			res.Body.Close()
			cancel()
			if e == nil && !resourceValid(data, hash, size) {
				e = errors.New("resource checksum or size mismatch")
			}
			if e != nil {
				last = e
				continue
			}
			if e = atomicResourceWrite(l.options.CacheDir, cacheName, data); e != nil {
				return nil, e
			}
			return data, nil
		}
		if last == nil {
			last = errors.New("no resource sources configured")
		}
		return nil, fmt.Errorf("load %s: %w", name, last)
	})
}
func atomicResourceWrite(dir, name string, data []byte) error {
	if e := os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(dir, ".resource-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(data); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(f.Name(), filepath.Join(dir, name))
}
func (l *resourceLoader) index(ctx context.Context) (resourceIndex, error) {
	var idx resourceIndex
	b, e := l.object(ctx, "index.json", l.pin, -1, 2<<20)
	if e != nil {
		return idx, e
	}
	if e = json.Unmarshal(b, &idx); e != nil {
		return idx, e
	}
	if idx.Format != "invisi.prover-assets.v1" || idx.NoirVersion != "1.0.0" || idx.BBVersion != "6.0.0" || len(idx.Files) == 0 {
		return idx, errors.New("unsupported resource index version")
	}
	for name, f := range idx.Files {
		if path.Clean(name) != name || strings.Contains(name, "\\") || strings.HasPrefix(name, "/") || (!strings.HasPrefix(name, "circuits/") && !strings.HasPrefix(name, "provenance/")) || (path.Ext(name) != ".json" && path.Ext(name) != ".vk" && !strings.HasSuffix(name, "/vk")) || !resourceDigest.MatchString(f.SHA256) || f.Size <= 0 || f.Size > maxResourceSize {
			return idx, errors.New("invalid resource index entry")
		}
	}
	return idx, nil
}
func (l *resourceLoader) read(ctx context.Context, name string) ([]byte, error) {
	idx, e := l.index(ctx)
	if e != nil {
		return nil, e
	}
	f, ok := idx.Files[name]
	if !ok {
		return nil, fmt.Errorf("resource absent from pinned index: %s", name)
	}
	return l.object(ctx, f.SHA256+resourceExtension(name), f.SHA256, f.Size, f.Size)
}

func resourceExtension(name string) string {
	if strings.HasSuffix(name, "/vk") {
		return ".vk"
	}
	return path.Ext(name)
}

// ResourceIndexSHA256 identifies the immutable public resource snapshot pinned by this SDK.
const ResourceIndexSHA256 = resourceIndexSHA256

// PrefetchResources verifies the pinned index and makes selected resources available
// before proving, without requiring Noir or BB binaries. An empty selection prepares
// the complete snapshot. Local directory/cache hits are verified and reused; missing
// files are downloaded into CacheDir unless Offline is enabled.
func PrefetchResources(ctx context.Context, options ResourceOptions, logicalPaths []string) error {
	if len(logicalPaths) == 0 {
		loader, e := newResourceLoader(options, coreResourcePin(options))
		if e != nil {
			return e
		}
		return loader.prefetch(ctx, nil)
	}
	type group struct {
		pin        string
		names      []string
		provenance bool
	}
	groups := map[string]*group{}
	for _, name := range logicalPaths {
		pin := coreResourcePin(options)
		provenance := strings.HasPrefix(name, "provenance/")
		if provenance {
			pin = provenanceResourcePin(options)
		} else if strings.HasPrefix(name, "circuits/create_credit_initial_payment.") {
			pin = circuitResourcePin("create_credit_initial_payment", options)
		}
		key := pin
		if provenance {
			key = "provenance/" + key
		}
		if groups[key] == nil {
			groups[key] = &group{pin: pin, provenance: provenance}
		}
		groups[key].names = append(groups[key].names, name)
	}
	for _, g := range groups {
		selected := options
		if g.provenance {
			selected = provenanceOptions(options)
		}
		loader, e := newResourceLoader(selected, g.pin)
		if e != nil {
			return e
		}
		if e = loader.prefetch(ctx, g.names); e != nil {
			return e
		}
	}
	return nil
}

// PrefetchProvenanceResources prepares only Source assets, without Core or proving runtimes.
func PrefetchProvenanceResources(ctx context.Context, options ResourceOptions) error {
	loader, e := newProvenanceResourceLoader(options)
	if e != nil {
		return e
	}
	index, e := loader.index(ctx)
	if e != nil {
		return e
	}
	var names []string
	for name := range index.Files {
		if strings.HasPrefix(name, "provenance/") {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return errors.New("Provenance resources absent from pinned release")
	}
	return loader.prefetch(ctx, names)
}

func coreResourcePin(options ResourceOptions) string {
	if options.LegacyCombined {
		return LegacyResourceIndexSHA256
	}
	return resourceIndexSHA256
}
func provenanceResourcePin(options ResourceOptions) string {
	if options.LegacyCombined {
		return LegacyResourceIndexSHA256
	}
	return ProvenanceResourceIndexSHA256
}
func circuitResourcePin(name string, options ResourceOptions) string {
	if options.LegacyCombined && name == "create_credit_initial_payment" {
		return InitialPaymentResourceIndexSHA256
	}
	return coreResourcePin(options)
}
func resourceRepository(pin string) string {
	if pin == ProvenanceResourceIndexSHA256 {
		return "invisi-periphery"
	}
	return "invisi-core"
}
func provenanceOptions(options ResourceOptions) ResourceOptions {
	selected := options
	if options.ProvenanceDirectory != "" {
		selected.Directory = options.ProvenanceDirectory
	}
	if options.ProvenanceSources != nil {
		selected.Sources = options.ProvenanceSources
	} else if !options.LegacyCombined {
		selected.Sources = nil
	}
	return selected
}
func newProvenanceResourceLoader(options ResourceOptions) (*resourceLoader, error) {
	return newResourceLoader(provenanceOptions(options), provenanceResourcePin(options))
}

func (l *resourceLoader) prefetch(ctx context.Context, logicalPaths []string) error {
	index, err := l.index(ctx)
	if err != nil {
		return err
	}
	names := append([]string(nil), logicalPaths...)
	if len(names) == 0 {
		for name := range index.Files {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		if _, ok := index.Files[name]; !ok {
			return fmt.Errorf("resource absent from pinned index: %s", name)
		}
	}
	for _, name := range names {
		file := index.Files[name]
		if _, err = l.object(ctx, file.SHA256+resourceExtension(name), file.SHA256, file.Size, file.Size); err != nil {
			return err
		}
	}
	return nil
}
