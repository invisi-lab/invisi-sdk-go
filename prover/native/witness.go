package native

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/invisi-lab/invisi-sdk-go/prover"
)

const noirVersion = "1.0.0"

// noirWitnessExecutor uses the official artifact executor; no compiler or source tree is required.
type noirWitnessExecutor struct {
	binary  string
	timeout time.Duration
}

func (n noirWitnessExecutor) Execute(ctx context.Context, artifact []byte, witness *prover.LocalWitness) ([]byte, error) {
	if witness == nil {
		return nil, errors.New("local witness required")
	}
	var identity struct {
		NoirVersion string `json:"noir_version"`
	}
	if err := json.Unmarshal(artifact, &identity); err != nil {
		return nil, errors.New("invalid Noir artifact")
	}
	if identity.NoirVersion != noirVersion+"+275d3ab1a5c916c6ea73e8d6b943f685f574500c" {
		return nil, errors.New("Noir artifact version does not match the released executor")
	}
	ctx, cancel := context.WithTimeout(ctx, n.timeout)
	defer cancel()
	versionCommand := exec.CommandContext(ctx, n.binary, "--version")
	var versionOutput boundedOutput
	versionCommand.Stdout = &versionOutput
	versionCommand.Stderr = io.Discard
	err := versionCommand.Run()
	version := versionOutput.Bytes()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if versionOutput.overflow {
		return nil, errors.New("invalid noir-execute version response")
	}
	if err != nil {
		return nil, fmt.Errorf("local noir-execute %s is required: %w", noirVersion, err)
	}
	if strings.TrimSpace(string(version)) != "noir-execute version = "+noirVersion {
		return nil, errors.New("noir-execute 1.0.0 is required")
	}
	dir, err := os.MkdirTemp("", "invisi-witness-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	artifactPath := filepath.Join(dir, "circuit.json")
	inputPath := filepath.Join(dir, "inputs.json")
	artifact, err = executableArtifact(artifact)
	if err != nil {
		return nil, err
	}
	if err = writePrivate(artifactPath, artifact); err != nil {
		return nil, err
	}
	err = witness.WithInputs(func(inputs map[string]any) error {
		encoded, err := json.Marshal(inputs)
		if err != nil {
			return err
		}
		defer clear(encoded)
		return writePrivate(inputPath, encoded)
	})
	if err != nil {
		return nil, err
	}
	// Never forward executor output: Noir assertions/print calls can contain private inputs.
	command := exec.CommandContext(ctx, n.binary, "--artifact-path", artifactPath, "--prover-file", inputPath, "--output-dir", dir, "--witness-name", "witness")
	command.Stdout, command.Stderr = io.Discard, io.Discard
	if err = command.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("local Noir witness execution failed: %w", err)
	}
	path := filepath.Join(dir, "witness.gz")
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() < 2 || info.Size() > 512<<20 {
		return nil, errors.New("invalid local witness output")
	}
	if err = os.Chmod(path, 0600); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// Release assets omit source/debug data. The official CLI requires these metadata
// fields even though they do not participate in execution or circuit identity.
func executableArtifact(artifact []byte) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(artifact, &fields); err != nil {
		return nil, err
	}
	if _, ok := fields["debug_symbols"]; !ok {
		var compressed bytes.Buffer
		writer, err := flate.NewWriter(&compressed, flate.DefaultCompression)
		if err != nil {
			return nil, err
		}
		if _, err = writer.Write([]byte(`{"debug_infos":[]}`)); err != nil {
			return nil, err
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		fields["debug_symbols"], err = json.Marshal(base64.StdEncoding.EncodeToString(compressed.Bytes()))
		if err != nil {
			return nil, err
		}
	}
	if _, ok := fields["file_map"]; !ok {
		fields["file_map"] = json.RawMessage(`{}`)
	}
	return json.Marshal(fields)
}
