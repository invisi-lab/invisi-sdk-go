package shared

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

const ProvenanceEndpoint = "https://provenance.invisipay.fi"
const ProvenanceManifestSHA256 = "6107789c5c0d24c4130bff05da6d389336feb5ccb247c4030eae5198ab05d762"

// SourceClient defaults to the official HTTPS bootstrap; explicit operator pins select private trust.
func SourceClient(pool *sdk.Protocol) (*provenance.Client, error) {
	if os.Getenv("PROVENANCE_SIGNING_PUBLIC_KEY") == "" && os.Getenv("PROVENANCE_SIGNING_KEY_ID") == "" && os.Getenv("PROVENANCE_GENERATION") == "" && os.Getenv("PROVENANCE_PROVIDER_ID") == "" {
		if os.Getenv("PROVENANCE_ENDPOINT") != "" {
			return nil, fmt.Errorf("private endpoint requires authenticated Source trust pins")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return (&provenance.OfficialClients{}).Get(ctx, provenance.Scope{ChainID: fmt.Sprint(pool.Context().ChainID), ProtocolAddress: pool.Context().ProtocolAddress.Hex()})
	}
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
	endpoint := os.Getenv("PROVENANCE_ENDPOINT")
	if endpoint == "" {
		endpoint = ProvenanceEndpoint
	}
	return &provenance.Client{Endpoint: endpoint, Trust: trust}, nil
}

func SourceCapability(pool *sdk.Protocol) (*sdk.ProvenanceCapability, error) {
	client, err := SourceClient(pool)
	if err != nil {
		return nil, err
	}
	assetRoot := os.Getenv("PROVENANCE_ASSET_ROOT")
	if assetRoot == "" {
		backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: os.Getenv("BB_BINARY"), NoirBinary: os.Getenv("NOIR_EXECUTOR")})
		if err != nil {
			return nil, err
		}
		identity, local, err := backend.GetProvenanceCapability()
		if err != nil {
			return nil, err
		}
		return &sdk.ProvenanceCapability{Client: client, Manifest: identity, Prover: local, MaxRootPages: 10, MaxLeaves: 100000}, nil
	}
	manifest, err := os.ReadFile(filepath.Join(assetRoot, "manifest.json"))
	if err != nil {
		return nil, err
	}
	identity, err := provenance.AuthenticateManifest(manifest, ProvenanceManifestSHA256)
	if err != nil {
		return nil, err
	}
	backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: os.Getenv("BB_BINARY"), NoirBinary: os.Getenv("NOIR_EXECUTOR")})
	if err != nil {
		return nil, err
	}
	_, offline, err := backend.GetProvenanceCapability()
	if err != nil {
		return nil, err
	}
	local := offline.(provenance.NativeProver)
	local.LoadAssets = func(k int) (provenance.LocalAssets, error) {
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
	}
	return &sdk.ProvenanceCapability{Client: client, Manifest: identity, Prover: local, MaxRootPages: 10, MaxLeaves: 100000}, nil
}

// WaitNativeFinality observes only the original hash with an independent finite deadline.
func WaitNativeFinality(ctx context.Context, provider protocolrpc.TransactionConfirmationProvider, hash common.Hash, timeout time.Duration) (protocolrpc.TransactionConfirmation, error) {
	if timeout <= 0 {
		return protocolrpc.TransactionConfirmation{}, fmt.Errorf("positive finality timeout required")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		confirmation := protocolrpc.ReadTransactionConfirmation(ctx, provider, hash)
		if confirmation.Level == "finalized" && confirmation.Status == protocolrpc.FinalityAvailable {
			receipt, err := provider.TransactionReceipt(ctx, hash)
			if err != nil {
				return confirmation, err
			}
			if receipt == nil || receipt.Status != 1 || confirmation.Block == nil || receipt.BlockHash != confirmation.Block.Hash {
				return confirmation, fmt.Errorf("original payment failed or changed after finality observation")
			}
			return confirmation, nil
		}
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return confirmation, ctx.Err()
		case <-timer.C:
		}
	}
}
