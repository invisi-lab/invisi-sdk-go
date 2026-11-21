package prover

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"os"
	"strconv"
	"strings"
	"testing"
)

type fixtureKeyLoader struct{}

func (fixtureKeyLoader) LoadVerificationKey(ctx context.Context, name, target string, _ Options) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.ReadFile("testdata/recursive/" + name + ".vk")
}
func TestEmbeddedRecursiveMetadataOnly(t *testing.T) {
	if strings.Contains(string(recursiveKeyJSON), "verificationKeyFields") {
		t.Fatal("full recursive keys remain embedded")
	}
	var pins map[string]map[string]string
	if err := json.Unmarshal(recursiveKeyJSON, &pins); err != nil || len(pins) != 10 {
		t.Fatal(err)
	}
}
func TestRecursiveKeysAuthenticateRealBinaries(t *testing.T) {
	for c := 1; c <= 10; c++ {
		name := fmtCircuit(c)
		key, err := LoadRecursiveVerificationKey(context.Background(), name, fixtureKeyLoader{}, Options{})
		if err != nil || len(key.Fields) != 115 {
			t.Fatalf("%s: %v", name, err)
		}
	}
}
func fmtCircuit(c int) string { return "transact_user_slots" + strconv.Itoa(c) }

func TestRecursiveKeyRejectsCorruptionWidthNoncanonicalAndWrongCommitment(t *testing.T) {
	data, err := (fixtureKeyLoader{}).LoadVerificationKey(context.Background(), fmtCircuit(1), "noir-recursive", Options{})
	if err != nil {
		t.Fatal(err)
	}
	var pins map[string]recursiveKeyPin
	if err = json.Unmarshal(recursiveKeyJSON, &pins); err != nil {
		t.Fatal(err)
	}
	digest := func(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
	for _, test := range []struct {
		name   string
		data   []byte
		digest string
		pin    recursiveKeyPin
		want   string
	}{
		{"corrupt", append([]byte{255}, data[1:]...), digest(data), pins["1"], "checksum"},
		{"short", data[:len(data)-1], digest(data[:len(data)-1]), pins["1"], "width"},
		{"long", append(append([]byte{}, data...), 0), digest(append(append([]byte{}, data...), 0)), pins["1"], "width"},
		{"commitment", data, digest(data), recursiveKeyPin{pins["1"].KeyHash, common.Hash{}.Hex()}, "commitment"},
		{"hash", data, digest(data), recursiveKeyPin{common.Hash{}.Hex(), pins["1"].InnerVKCommitment}, "commitment"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := authenticateRecursiveKey(context.Background(), test.data, test.digest, test.pin)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v", err)
			}
		})
	}
	noncanonical := append([]byte{}, data...)
	primitives.FieldModulus().FillBytes(noncanonical[:32])
	if _, err = authenticateRecursiveKey(context.Background(), noncanonical, digest(noncanonical), pins["1"]); err == nil || !strings.Contains(err.Error(), "noncanonical") {
		t.Fatal(err)
	}
}

type keyLoaderFunc func(context.Context, string, string, Options) ([]byte, error)

func (f keyLoaderFunc) LoadVerificationKey(ctx context.Context, name, target string, opt Options) ([]byte, error) {
	return f(ctx, name, target, opt)
}
func TestRecursiveKeyCancellationAndFreshOwnership(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	loader := keyLoaderFunc(func(ctx context.Context, name, target string, opt Options) ([]byte, error) {
		called = true
		return (fixtureKeyLoader{}).LoadVerificationKey(ctx, name, target, opt)
	})
	if _, err := LoadRecursiveVerificationKey(ctx, fmtCircuit(1), loader, Options{}); !errors.Is(err, context.Canceled) || called {
		t.Fatal("cancelled load started", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	loader = keyLoaderFunc(func(context.Context, string, string, Options) ([]byte, error) {
		data, err := os.ReadFile("testdata/recursive/transact_user_slots1.vk")
		cancel()
		return data, err
	})
	if _, err := LoadRecursiveVerificationKey(ctx, fmtCircuit(1), loader, Options{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	key, err := LoadRecursiveVerificationKey(context.Background(), fmtCircuit(1), fixtureKeyLoader{}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	want := key.Fields[0]
	key.Fields[0] = "bad"
	key, err = LoadRecursiveVerificationKey(context.Background(), fmtCircuit(1), fixtureKeyLoader{}, Options{})
	if err != nil || key.Fields[0] != want {
		t.Fatal("retained mutable decoded key", err)
	}
}

type customBackend struct{ Prover }

func TestAggregatorLoaderPrecedenceFallbackAndNoConstructorIO(t *testing.T) {
	backend := &aggregateBackend{}
	called := false
	fallback := keyLoaderFunc(func(ctx context.Context, name, target string, opt Options) ([]byte, error) {
		called = true
		return (fixtureKeyLoader{}).LoadVerificationKey(ctx, name, target, opt)
	})
	a, err := NewAggregator(backend, AggregatorOptions{VerificationKeyLoader: fallback})
	if err != nil || called {
		t.Fatal("constructor performed key I/O", err)
	}
	if _, err = a.Aggregate(context.Background(), []AggregateMember{aggregateMember(t, 1, 100, 34)}); err != nil || called {
		t.Fatal("backend capability must take precedence", err)
	}
	a, err = NewAggregator(customBackend{backend}, AggregatorOptions{VerificationKeyLoader: fallback})
	if err != nil || called {
		t.Fatal(err)
	}
	if _, err = a.Aggregate(context.Background(), []AggregateMember{aggregateMember(t, 1, 200, 34)}); err != nil || !called {
		t.Fatal("explicit custom loader unavailable", err)
	}
	a, err = NewAggregator(customBackend{backend})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Aggregate(context.Background(), []AggregateMember{aggregateMember(t, 1, 300, 34)}); err == nil || !strings.Contains(err.Error(), "VerificationKeyLoader") {
		t.Fatal("missing custom loader must fail clearly", err)
	}
}
