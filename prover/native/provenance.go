package native

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	protocol "github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	source "github.com/invisi-lab/invisi-sdk-go/prover/provenance"
)

//go:embed provenance-assets/*
var provenanceResources embed.FS

// GetProvenanceCapability reuses the local native executables with authenticated Source assets.
func (p *NativeProver) GetProvenanceCapability() (source.ManifestIdentity, source.OfflineProver, error) {
	raw := protocol.ManifestBytes()
	digest := sha256.Sum256(raw)
	pin := hex.EncodeToString(digest[:])
	identity, err := source.AuthenticateManifest(raw, pin)
	if err != nil {
		return source.ManifestIdentity{}, nil, err
	}
	backend := source.NativeProver{NoirExecutePath: p.options.NoirBinary, BBPath: p.options.BBBinary, TrustedManifestSHA256: pin, Run: func(ctx context.Context, args ...string) error { _, err := p.run(ctx, args...); return err }, ExecuteWitness: func(ctx context.Context, artifact, witness []byte) ([]byte, error) {
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
	}, LoadAssets: func(capacity int) (source.LocalAssets, error) {
		artifact, e := provenanceResources.ReadFile(fmt.Sprintf("provenance-assets/k%d.json", capacity))
		if e != nil {
			return source.LocalAssets{}, e
		}
		key, e := provenanceResources.ReadFile(fmt.Sprintf("provenance-assets/k%d.vk", capacity))
		if e != nil {
			return source.LocalAssets{}, e
		}
		return source.LocalAssets{Manifest: raw, Artifact: artifact, VerificationKey: key}, nil
	}}
	return identity, backend, nil
}
