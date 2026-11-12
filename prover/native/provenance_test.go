package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	source "github.com/invisi-lab/invisi-sdk-go/prover/provenance"
	"os"
	"testing"
	"time"
)

func TestNativeProvenanceCapabilityAuthenticatesAllPackagedAssets(t *testing.T) {
	backend, err := NewNativeProver(NativeOptions{BBBinary: "configured-bb", NoirBinary: "configured-noir"})
	if err != nil {
		t.Fatal(err)
	}
	identity, offline, err := backend.GetProvenanceCapability()
	if err != nil {
		t.Fatal(err)
	}
	native, ok := offline.(source.NativeProver)
	if !ok || native.BBPath != "configured-bb" || native.NoirExecutePath != "configured-noir" || native.ExecuteWitness == nil {
		t.Fatal("local runtime configuration lost")
	}
	if len(identity.Manifest.Circuits) != 7 {
		t.Fatal("frozen seven circuit family changed")
	}
	for _, entry := range identity.Manifest.Circuits {
		assets, err := native.LoadAssets(entry.Capacity)
		if err != nil {
			t.Fatal(err)
		}
		for _, check := range []struct {
			data []byte
			hash string
		}{{assets.Artifact, entry.ArtifactSHA256}, {assets.VerificationKey, entry.VKSHA256}} {
			sum := sha256.Sum256(check.data)
			if hex.EncodeToString(sum[:]) != check.hash {
				t.Fatal("untrusted packaged asset", entry.Capacity)
			}
		}
	}
}

func TestRealNativeProvenanceK10(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("set BB_BINARY_PATH and NOIR_EXECUTE_PATH for genuine local proving")
	}
	backend, err := NewNativeProver(NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
	if err != nil {
		t.Fatal(err)
	}
	identity, offline, err := backend.GetProvenanceCapability()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("testdata/provenance-k10.json")
	if err != nil {
		t.Fatal(err)
	}
	var witness source.Witness
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err = decoder.Decode(&witness); err != nil {
		t.Fatal(err)
	}
	expected := []string{}
	for _, name := range source.PublicInputNames() {
		word, err := source.CanonicalWord(fmt.Sprint(witness[name]))
		if err != nil {
			t.Fatal(err)
		}
		expected = append(expected, word)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	result, err := offline.Prove(ctx, identity, 10, witness, expected)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Proof) <= 2 || len(result.PublicInputs) != 11 {
		t.Fatal("missing independently verified native Source proof")
	}
}
