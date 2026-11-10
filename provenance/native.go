package provenance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LocalAssets struct {
	Manifest        []byte
	Artifact        []byte
	VerificationKey []byte
}

// NativeProver executes trusted local programs without transmitting private witnesses to a service.
type NativeProver struct {
	NoirExecutePath       string
	BBPath                string
	TrustedManifestSHA256 string
	LoadAssets            func(int) (LocalAssets, error)
}

func (p NativeProver) Prove(ctx context.Context, identity ManifestIdentity, capacity int, w Witness, expected []string) (ProofResult, error) {
	if identity.SHA256 != p.TrustedManifestSHA256 || p.LoadAssets == nil {
		return ProofResult{}, errors.New("native prover trust configuration")
	}
	assets, e := p.LoadAssets(capacity)
	if e != nil {
		return ProofResult{}, e
	}
	authenticated, e := AuthenticateManifest(assets.Manifest, p.TrustedManifestSHA256)
	if e != nil {
		return ProofResult{}, e
	}
	var entry *ManifestEntry
	for i := range authenticated.Manifest.Circuits {
		if authenticated.Manifest.Circuits[i].Capacity == capacity {
			entry = &authenticated.Manifest.Circuits[i]
		}
	}
	if entry == nil {
		return ProofResult{}, errors.New("unsupported capacity")
	}
	for _, v := range []struct {
		b    []byte
		hash string
	}{{assets.Artifact, entry.ArtifactSHA256}, {assets.VerificationKey, entry.VKSHA256}} {
		d := sha256.Sum256(v.b)
		if hex.EncodeToString(d[:]) != v.hash {
			return ProofResult{}, errors.New("local artifact identity mismatch")
		}
	}
	dir, e := os.MkdirTemp("", "invisi-provenance-")
	if e != nil {
		return ProofResult{}, e
	}
	defer os.RemoveAll(dir)
	witness, e := json.Marshal(w)
	if e != nil {
		return ProofResult{}, e
	}
	defer clear(witness)
	for name, data := range map[string][]byte{"artifact.json": assets.Artifact, "vk": assets.VerificationKey, "witness.json": witness} {
		if e := os.WriteFile(filepath.Join(dir, name), data, 0600); e != nil {
			return ProofResult{}, e
		}
	}
	run := func(binary string, args ...string) error {
		if binary == "" {
			return errors.New("local prover executable required")
		}
		cmd := exec.CommandContext(ctx, binary, args...)
		if e := cmd.Run(); e != nil {
			return errors.New("local provenance proving command failed")
		}
		return nil
	}
	if e := run(p.NoirExecutePath, "-a", filepath.Join(dir, "artifact.json"), "-p", filepath.Join(dir, "witness.json"), "-o", dir, "-w", "witness"); e != nil {
		return ProofResult{}, e
	}
	if e := run(p.BBPath, "prove", "-t", "evm", "-b", filepath.Join(dir, "artifact.json"), "-w", filepath.Join(dir, "witness.gz"), "-k", filepath.Join(dir, "vk"), "-o", dir, "--output_format", "json"); e != nil {
		return ProofResult{}, e
	}
	data, e := os.ReadFile(filepath.Join(dir, "proof.json"))
	if e != nil {
		return ProofResult{}, e
	}
	var raw struct {
		Version string   `json:"bb_version"`
		Proof   []string `json:"proof"`
	}
	if e := json.Unmarshal(data, &raw); e != nil {
		return ProofResult{}, e
	}
	if raw.Version != "6.0.0" {
		return ProofResult{}, errors.New("unexpected BB runtime")
	}
	data, e = os.ReadFile(filepath.Join(dir, "public_inputs.json"))
	if e != nil {
		return ProofResult{}, e
	}
	var pub struct {
		Inputs []string `json:"public_inputs"`
	}
	if e := json.Unmarshal(data, &pub); e != nil {
		return ProofResult{}, e
	}
	if len(pub.Inputs) != 11 || len(expected) != 11 {
		return ProofResult{}, errors.New("prover public statement width")
	}
	publicBytes := []byte{}
	for i, v := range pub.Inputs {
		if !equalNumber(v, expected[i]) {
			return ProofResult{}, errors.New("prover public statement mismatch")
		}
		n, e := Field(v)
		if e != nil {
			return ProofResult{}, e
		}
		publicBytes = append(publicBytes, n.FillBytes(make([]byte, 32))...)
		pub.Inputs[i], _ = word(v)
	}
	proofBytes := []byte{}
	for _, v := range raw.Proof {
		n, e := Uint(v, 256)
		if e != nil {
			return ProofResult{}, e
		}
		proofBytes = append(proofBytes, n.FillBytes(make([]byte, 32))...)
	}
	if len(proofBytes) == 0 || len(proofBytes) > 65536 {
		return ProofResult{}, errors.New("proof size")
	}
	if e := os.WriteFile(filepath.Join(dir, "proof"), proofBytes, 0600); e != nil {
		return ProofResult{}, e
	}
	if e := os.WriteFile(filepath.Join(dir, "public_inputs"), publicBytes, 0600); e != nil {
		return ProofResult{}, e
	}
	if e := run(p.BBPath, "verify", "-t", "evm", "-k", filepath.Join(dir, "vk"), "-p", filepath.Join(dir, "proof"), "-i", filepath.Join(dir, "public_inputs")); e != nil {
		return ProofResult{}, e
	}
	return ProofResult{Proof: "0x" + strings.ToLower(hex.EncodeToString(proofBytes)), PublicInputs: pub.Inputs}, nil
}
