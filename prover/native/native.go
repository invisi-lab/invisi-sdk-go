// Package native executes local witness generation and Barretenberg 6.0.0.
package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/invisi-labs/invisi-sdk-go/prover"
)

// WitnessExecutor runs locally and returns the gzip-encoded ACVM witness.
// Its implementation must never send inputs to an external service.
type WitnessExecutor interface {
	Execute(context.Context, []byte, *prover.LocalWitness) ([]byte, error)
}
type NativeOptions struct {
	BBBinary        string
	NoirBinary      string
	Threads         int
	Timeout         time.Duration
	WitnessExecutor WitnessExecutor
	Resources       ResourceOptions
}
type NativeProver struct {
	options             NativeOptions
	resources           *resourceLoader
	provenanceOnce      sync.Once
	provenanceResources *resourceLoader
	provenanceError     error
}

func NewNativeProver(options NativeOptions) (*NativeProver, error) {
	if options.BBBinary == "" {
		options.BBBinary = "bb"
	}
	if options.Threads == 0 {
		options.Threads = 4
	}
	if options.Threads < 1 {
		return nil, errors.New("invalid proving thread count")
	}
	if options.Timeout == 0 {
		options.Timeout = 10 * time.Minute
	}
	if options.Timeout < 0 {
		return nil, errors.New("invalid proving timeout")
	}
	if options.NoirBinary == "" {
		options.NoirBinary = "noir-execute"
	}
	if options.WitnessExecutor == nil {
		options.WitnessExecutor = noirWitnessExecutor{binary: options.NoirBinary, timeout: options.Timeout}
	}
	if options.Resources.Sources != nil {
		options.Resources.Sources = append([]string{}, options.Resources.Sources...)
	}
	if options.Resources.ProvenanceSources != nil {
		options.Resources.ProvenanceSources = append([]string{}, options.Resources.ProvenanceSources...)
	}
	loader, err := newResourceLoader(options.Resources, coreResourcePin(options.Resources))
	if err != nil {
		return nil, err
	}
	return &NativeProver{options: options, resources: loader}, nil
}
func (p *NativeProver) run(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, p.options.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.options.BBBinary, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HARDWARE_CONCURRENCY=%d", p.options.Threads))
	// BB 6 fast MSM has an upstream Darwin arm64 failure on large recursive workloads.
	// Select its upstream stable MSM implementation without changing proof semantics.
	if _, configured := os.LookupEnv("BB_MSM_LEGACY"); !configured {
		cmd.Env = append(cmd.Env, "BB_MSM_LEGACY=1")
	}
	var diagnostic boundedOutput
	cmd.Stdout, cmd.Stderr = &diagnostic, &diagnostic
	e := cmd.Run()
	output := diagnostic.Bytes()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if e != nil {
		return nil, fmt.Errorf("local BB execution failed: %w", e)
	}
	if diagnostic.overflow {
		return nil, errors.New("BB diagnostic output exceeds limit")
	}
	return output, nil
}
func (p *NativeProver) version(ctx context.Context) error {
	out, e := p.run(ctx, "--version")
	if e != nil {
		return e
	}
	if !regexp.MustCompile(`(?:^|\s)6\.0\.0(?:\s|$)`).Match(out) {
		return errors.New("Barretenberg 6.0.0 is required")
	}
	return nil
}
func (p *NativeProver) assets(ctx context.Context, statement prover.ProofStatement) ([]byte, []byte, error) {
	expected, e := prover.Statement(statement.CircuitName, statement.Target, statement.ExpectedPublicInputs)
	if e != nil {
		return nil, nil, e
	}
	if expected.ArtifactSHA256 != statement.ArtifactSHA256 || expected.VKSHA256 != statement.VKSHA256 {
		return nil, nil, errors.New("untrusted circuit metadata")
	}
	loader := p.resources
	if p.options.Resources.LegacyCombined && statement.CircuitName == "create_credit_initial_payment" {
		loader, e = newResourceLoader(p.options.Resources, InitialPaymentResourceIndexSHA256)
		if e != nil {
			return nil, nil, e
		}
	}
	artifact, e := loader.read(ctx, "circuits/"+statement.CircuitName+".json")
	if e != nil {
		return nil, nil, e
	}
	key, e := loader.read(ctx, "circuits/"+statement.CircuitName+"."+statement.Target+".vk")
	if e != nil {
		return nil, nil, e
	}
	var compiled struct {
		NoirVersion string          `json:"noir_version"`
		ABI         json.RawMessage `json:"abi"`
		Bytecode    string          `json:"bytecode"`
	}
	if e = json.Unmarshal(artifact, &compiled); e != nil {
		return nil, nil, e
	}
	identity := struct {
		NoirVersion string          `json:"noirVersion"`
		ABI         json.RawMessage `json:"abi"`
		Bytecode    string          `json:"bytecode"`
	}{compiled.NoirVersion, compiled.ABI, compiled.Bytecode}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if e = encoder.Encode(identity); e != nil {
		return nil, nil, e
	}
	a, k := sha256.Sum256(bytes.TrimSuffix(buffer.Bytes(), []byte("\n"))), sha256.Sum256(key)
	if hex.EncodeToString(a[:]) != statement.ArtifactSHA256 || hex.EncodeToString(k[:]) != statement.VKSHA256 {
		return nil, nil, errors.New("circuit resource checksum mismatch")
	}
	return artifact, key, nil
}
func writePrivate(path string, data []byte) error { return os.WriteFile(path, data, 0600) }
func (p *NativeProver) Verify(ctx context.Context, bundle prover.ProofBundle, statement prover.ProofStatement, options prover.Options) (bool, error) {
	if e := prover.AssertProofBundle(bundle, statement); e != nil {
		return false, e
	}
	_, key, e := p.assets(ctx, statement)
	if e != nil {
		return false, e
	}
	if e = p.version(ctx); e != nil {
		return false, e
	}
	dir, e := os.MkdirTemp("", "invisi-verify-")
	if e != nil {
		return false, e
	}
	defer os.RemoveAll(dir)
	inputs := []byte{}
	for _, v := range bundle.PublicInputs {
		inputs = append(inputs, v[:]...)
	}
	for name, data := range map[string][]byte{"vk": key, "proof": bundle.Proof, "public_inputs": inputs} {
		if e = writePrivate(filepath.Join(dir, name), data); e != nil {
			return false, e
		}
	}
	out, e := p.run(ctx, "verify", "-t", bundle.Target, "-k", filepath.Join(dir, "vk"), "-p", filepath.Join(dir, "proof"), "-i", filepath.Join(dir, "public_inputs"))
	if e != nil {
		return false, e
	}
	return strings.Contains(string(out), "Proof verified successfully"), nil
}
func (p *NativeProver) Prove(ctx context.Context, request prover.ProvingRequest, options prover.Options) (prover.ProofBundle, error) {
	var bundle prover.ProofBundle
	if p.options.WitnessExecutor == nil {
		return bundle, errors.New("local Noir witness executor required")
	}
	artifact, key, e := p.assets(ctx, request.ProofStatement)
	if e != nil {
		return bundle, e
	}
	if e = p.version(ctx); e != nil {
		return bundle, e
	}
	progress := func(phase string) {
		if options.OnProgress != nil {
			options.OnProgress(prover.Progress{Phase: phase})
		}
	}
	progress("witness")
	witness, e := p.options.WitnessExecutor.Execute(ctx, artifact, request.Witness)
	if e != nil {
		return bundle, e
	}
	defer clear(witness)
	if len(witness) < 2 || witness[0] != 0x1f || witness[1] != 0x8b {
		return bundle, errors.New("witness executor must return gzip ACVM witness")
	}
	dir, e := os.MkdirTemp("", "invisi-proof-")
	if e != nil {
		return bundle, e
	}
	defer os.RemoveAll(dir)
	for name, data := range map[string][]byte{"circuit.json": artifact, "vk": key, "witness.gz": witness} {
		if e = writePrivate(filepath.Join(dir, name), data); e != nil {
			return bundle, e
		}
	}
	progress("proving")
	_, e = p.run(ctx, "prove", "-t", request.Target, "-b", filepath.Join(dir, "circuit.json"), "-w", filepath.Join(dir, "witness.gz"), "-k", filepath.Join(dir, "vk"), "-o", dir, "--output_format", "json")
	if e != nil {
		return bundle, e
	}
	var proof struct {
		BBVersion string   `json:"bb_version"`
		Scheme    string   `json:"scheme"`
		Proof     []string `json:"proof"`
	}
	raw, e := os.ReadFile(filepath.Join(dir, "proof.json"))
	if e != nil {
		return bundle, e
	}
	if e = json.Unmarshal(raw, &proof); e != nil {
		return bundle, e
	}
	if proof.BBVersion != "6.0.0" || proof.Scheme != "ultra_honk" || len(proof.Proof) == 0 {
		return bundle, errors.New("unsupported native proof output")
	}
	var public struct {
		PublicInputs []string `json:"public_inputs"`
	}
	raw, e = os.ReadFile(filepath.Join(dir, "public_inputs.json"))
	if e != nil {
		return bundle, e
	}
	if e = json.Unmarshal(raw, &public); e != nil {
		return bundle, e
	}
	bundle = prover.ProofBundle{FormatVersion: 1, CircuitName: request.CircuitName, Target: request.Target, BBVersion: "6.0.0", ArtifactSHA256: request.ArtifactSHA256, VKSHA256: request.VKSHA256}
	for _, value := range proof.Proof {
		v, e := word(value)
		if e != nil {
			return bundle, e
		}
		bundle.Proof = append(bundle.Proof, v[:]...)
	}
	for _, value := range public.PublicInputs {
		v, e := word(value)
		if e != nil {
			return bundle, e
		}
		bundle.PublicInputs = append(bundle.PublicInputs, v)
	}
	digest := sha256.Sum256(bundle.Proof)
	bundle.ProofSHA256 = hex.EncodeToString(digest[:])
	if e = prover.AssertProofBundle(bundle, request.ProofStatement); e != nil {
		return bundle, e
	}
	progress("verifying")
	valid, e := p.Verify(ctx, bundle, request.ProofStatement, options)
	if e != nil {
		return bundle, e
	}
	if !valid {
		return bundle, errors.New("local proof verification failed")
	}
	progress("complete")
	return bundle, nil
}
