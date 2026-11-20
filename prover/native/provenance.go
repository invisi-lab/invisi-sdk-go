package native

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	protocol "github.com/invisi-labs/invisi-sdk-go/protocol/provenance"
	"github.com/invisi-labs/invisi-sdk-go/prover"
	source "github.com/invisi-labs/invisi-sdk-go/prover/provenance"
)

// GetProvenanceCapability reuses the local native executables with authenticated Source assets.
func (p *NativeProver) GetProvenanceCapability() (source.ManifestIdentity, source.OfflineProver, error) {
	raw := protocol.ManifestBytes()
	digest := sha256.Sum256(raw)
	pin := hex.EncodeToString(digest[:])
	identity, err := source.AuthenticateManifest(raw, pin)
	if err != nil {
		return source.ManifestIdentity{}, nil, err
	}
	loader, err := p.provenanceResourceLoader()
	if err != nil {
		return source.ManifestIdentity{}, nil, err
	}
	load := func(ctx context.Context, capacity int) (source.LocalAssets, error) {
		var artifactPath, keyPath string
		for _, entry := range identity.Manifest.Circuits {
			if entry.Capacity == capacity {
				artifactPath, keyPath = entry.Artifact, entry.VK
			}
		}
		if artifactPath == "" {
			return source.LocalAssets{}, fmt.Errorf("unsupported provenance capacity %d", capacity)
		}
		manifest, e := loader.read(ctx, "provenance/manifest.json")
		if e != nil {
			return source.LocalAssets{}, e
		}
		if _, e = source.AuthenticateManifest(manifest, pin); e != nil {
			return source.LocalAssets{}, e
		}
		artifact, e := loader.read(ctx, "provenance/"+artifactPath)
		if e != nil {
			return source.LocalAssets{}, e
		}
		key, e := loader.read(ctx, "provenance/"+keyPath)
		if e != nil {
			return source.LocalAssets{}, e
		}
		return source.LocalAssets{Manifest: manifest, Artifact: artifact, VerificationKey: key}, nil
	}
	backend := source.NativeProver{LoadAssetsContext: load, NoirExecutePath: p.options.NoirBinary, BBPath: p.options.BBBinary, TrustedManifestSHA256: pin, Run: func(ctx context.Context, args ...string) error { _, err := p.run(ctx, args...); return err }, ExecuteWitness: func(ctx context.Context, artifact, witness []byte) ([]byte, error) {
		if err := p.version(ctx); err != nil {
			return nil, err
		}
		var inputs map[string]any
		if err := json.Unmarshal(witness, &inputs); err != nil {
			return nil, err
		}
		local, err := prover.NewLocalWitness(inputs)
		if err != nil {
			return nil, err
		}
		return p.options.WitnessExecutor.Execute(ctx, artifact, local)
	}, LoadAssets: func(capacity int) (source.LocalAssets, error) { return load(context.Background(), capacity) }}
	return identity, backend, nil
}

func (p *NativeProver) provenanceResourceLoader() (*resourceLoader, error) {
	p.provenanceOnce.Do(func() { p.provenanceResources, p.provenanceError = newProvenanceResourceLoader(p.options.Resources) })
	return p.provenanceResources, p.provenanceError
}
