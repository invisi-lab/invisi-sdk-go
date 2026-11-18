package native

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/invisi-labs/invisi-sdk-go/prover"
)

func TestExecutorMetadataDoesNotChangePublishedCircuit(t *testing.T) {
	original, err := json.Marshal(map[string]any{"noir_version": "1.0.0+275d3ab1a5c916c6ea73e8d6b943f685f574500c", "abi": map[string]any{"parameters": []any{}}, "bytecode": "tiny-test-only"})
	if err != nil {
		t.Fatal(err)
	}
	adapted, err := executableArtifact(original)
	if err != nil {
		t.Fatal(err)
	}
	var before, after map[string]json.RawMessage
	if err = json.Unmarshal(original, &before); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(adapted, &after); err != nil {
		t.Fatal(err)
	}
	for key, value := range before {
		var a, b bytes.Buffer
		if err = json.Compact(&a, value); err != nil {
			t.Fatal(err)
		}
		if err = json.Compact(&b, after[key]); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a.Bytes(), b.Bytes()) {
			t.Fatalf("published circuit field %s changed", key)
		}
	}
	if string(after["hash"]) != "0" {
		t.Fatal("official executor requires the non-semantic hash metadata field")
	}
	var encoded string
	if err = json.Unmarshal(after["debug_symbols"], &encoded); err != nil {
		t.Fatal(err)
	}
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reader := flate.NewReader(bytes.NewReader(compressed))
	defer reader.Close()
	decoded, err := io.ReadAll(reader)
	if err != nil || string(decoded) != `{"debug_infos":[]}` {
		t.Fatal(string(decoded), err)
	}
	if string(after["file_map"]) != "{}" {
		t.Fatal("source map should be empty")
	}
}

func TestExecutorRejectsArtifactMismatchAndHonorsCancellation(t *testing.T) {
	witness, err := prover.NewLocalWitness(map[string]any{"secret": "must-stay-local"})
	if err != nil {
		t.Fatal(err)
	}
	executor := noirWitnessExecutor{binary: "must-not-execute", timeout: time.Minute}
	if _, err = executor.Execute(context.Background(), []byte(`{"noir_version":"invalid-version"}`), witness); err == nil {
		t.Fatal("mismatched Noir artifact accepted")
	}
	artifact, err := json.Marshal(map[string]any{"noir_version": "1.0.0+275d3ab1a5c916c6ea73e8d6b943f685f574500c", "abi": map[string]any{"parameters": []any{}}, "bytecode": "tiny-test-only"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = executor.Execute(ctx, artifact, witness); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation ignored", err)
	}
}

func TestLocalDiagnosticBufferIsBounded(t *testing.T) {
	var output boundedOutput
	input := bytes.Repeat([]byte{1}, 128<<10)
	n, err := output.Write(input)
	if err != nil || n != len(input) || output.Len() != 64<<10 || !output.overflow {
		t.Fatal("unbounded local tool output")
	}
	if _, err = io.Copy(&output, bytes.NewReader(input)); err != nil || output.Len() != 64<<10 {
		t.Fatal("diagnostic limit failed")
	}
}

func TestExecutorPreservesExistingMetadataHash(t *testing.T) {
	raw := []byte(`{"hash":123,"abi":{},"bytecode":"test"}`)
	adapted, err := executableArtifact(raw)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(adapted, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["hash"]) != "123" {
		t.Fatal("overwrote existing compiler metadata")
	}
}
