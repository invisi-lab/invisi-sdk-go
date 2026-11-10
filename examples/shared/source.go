package shared

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

const ProvenanceEndpoint = "https://provenance.invisipay.fi"
const ProvenanceManifestSHA256 = "6107789c5c0d24c4130bff05da6d389336feb5ccb247c4030eae5198ab05d762"

// SourceClient trust comes from authenticated operator settings, never the current service response.
func SourceClient(pool *sdk.Protocol) (*provenance.Client, error) {
	key, err := hex.DecodeString(strings.TrimPrefix(os.Getenv("PROVENANCE_SIGNING_PUBLIC_KEY"), "0x"))
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("set authenticated PROVENANCE_SIGNING_PUBLIC_KEY")
	}
	keyID := os.Getenv("PROVENANCE_SIGNING_KEY_ID")
	if len(keyID) != 66 {
		return nil, fmt.Errorf("set authenticated PROVENANCE_SIGNING_KEY_ID")
	}
	trust := provenance.TrustContext{Scope: provenance.Scope{ChainID: fmt.Sprint(pool.Context().ChainID), ProtocolAddress: pool.Context().ProtocolAddress.Hex()}, ChainGeneration: os.Getenv("PROVENANCE_GENERATION"), ProviderID: os.Getenv("PROVENANCE_PROVIDER_ID"), ManifestSHA256: ProvenanceManifestSHA256, SigningKeys: map[string]ed25519.PublicKey{strings.ToLower(keyID): key}}
	for _, value := range []string{trust.ChainGeneration, trust.ProviderID} {
		if _, err := provenance.Uint(value, 256); err != nil {
			return nil, fmt.Errorf("set authenticated Source generation/provider: %w", err)
		}
	}
	return &provenance.Client{Endpoint: ProvenanceEndpoint, Trust: trust}, nil
}

func SourceCapability(pool *sdk.Protocol) (*sdk.ProvenanceCapability, error) {
	client, err := SourceClient(pool)
	if err != nil {
		return nil, err
	}
	assetRoot := os.Getenv("PROVENANCE_ASSET_ROOT")
	if assetRoot == "" {
		assetRoot = ".cache/provenance"
	}
	manifest, err := os.ReadFile(filepath.Join(assetRoot, "manifest.json"))
	if err != nil {
		return nil, err
	}
	identity, err := provenance.AuthenticateManifest(manifest, ProvenanceManifestSHA256)
	if err != nil {
		return nil, err
	}
	local := provenance.NativeProver{NoirExecutePath: os.Getenv("NOIR_EXECUTOR"), BBPath: os.Getenv("BB_BINARY"), TrustedManifestSHA256: ProvenanceManifestSHA256, LoadAssets: func(k int) (provenance.LocalAssets, error) {
		for _, entry := range identity.Manifest.Circuits {
			if entry.Capacity == k {
				artifact, err := os.ReadFile(filepath.Join(assetRoot, entry.Artifact))
				if err != nil {
					return provenance.LocalAssets{}, err
				}
				vk, err := os.ReadFile(filepath.Join(assetRoot, entry.VK))
				if err != nil {
					return provenance.LocalAssets{}, err
				}
				return provenance.LocalAssets{Manifest: manifest, Artifact: artifact, VerificationKey: vk}, nil
			}
		}
		return provenance.LocalAssets{}, fmt.Errorf("unsupported Source capacity")
	}}
	return &sdk.ProvenanceCapability{Client: client, Manifest: identity, Prover: local, MaxRootPages: 10, MaxLeaves: 100000}, nil
}
