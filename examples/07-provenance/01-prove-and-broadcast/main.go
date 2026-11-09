package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if !slices.Contains(os.Args, "--execute") {
		fmt.Println("Pass --execute to prove/quote; add --submit to broadcast. See README.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	var config struct {
		ChainID          string `json:"chainId"`
		ProtocolAddress  string
		PolicyID         string
		Generation       string
		ProviderID       string
		SigningKeyID     string
		SigningPublicKey string
		Endpoint         string
		AssetRoot        string
		ChainRPCURL      string
		Broadcaster      struct {
			Endpoint     string
			QuoteSigner  string
			FeeRecipient string
		}
	}
	check(json.Unmarshal(must(os.ReadFile(os.Getenv("PROVENANCE_CONFIG"))), &config))
	const pin = "f90f2ce43218de158c6392fa19e6517dda703c35823a4aa5bfc20b50788ad282"
	trust := provenance.TrustContext{
		Scope: provenance.Scope{
			ChainID:         config.ChainID,
			ProtocolAddress: config.ProtocolAddress,
			PolicyID:        config.PolicyID,
		},
		ChainGeneration: config.Generation,
		ProviderID:      config.ProviderID,
		ManifestSHA256:  pin,
		SigningKeys: map[string]ed25519.PublicKey{
			config.SigningKeyID: must(hex.DecodeString(strings.TrimPrefix(config.SigningPublicKey, "0x"))),
		},
	}
	source := &provenance.Client{Endpoint: config.Endpoint, Trust: trust}
	must(source.GetVersion(ctx))

	// This is host-local prepared material. Never upload or print the opening or witness.
	var prepared struct {
		RootID                                  string `json:"rootId"`
		TokenAddress, MainCircuitCapacity, Data string
		PublicInputs                            []string
		Opening                                 provenance.Opening
	}
	check(json.Unmarshal(must(io.ReadAll(io.LimitReader(os.Stdin, 1<<20))), &prepared))
	root := must(source.GetRoot(ctx, prepared.RootID))
	if string(root["allowedForNewProof"]) != "true" {
		panic("Root unavailable for new proofs")
	}
	var cert provenance.RootCertificate
	check(json.Unmarshal(root["certificate"], &cert))
	primary := must(ethclient.DialContext(ctx, config.ChainRPCURL))
	defer primary.Close()
	if must(primary.ChainID(ctx)).Cmp(must(provenance.Uint(config.ChainID, 64))) != 0 {
		panic("Primary chain mismatch")
	}
	anchor := must(primary.HeaderByNumber(ctx, must(provenance.Uint(cert.ConfirmedBlockNumber, 64))))
	if anchor.Hash() != common.HexToHash(cert.ConfirmedBlockHash) {
		panic("Checkpoint anchor is not canonical")
	}
	tree := must(source.SyncRecognizedTree(ctx, prepared.RootID, cert, 100000))
	commitment, nullifier, err := provenance.NoteOpening(trust.Scope, prepared.TokenAddress, prepared.Opening)
	check(err)
	membership, found, err := tree.MembershipForCommitment(trust.Scope, commitment)
	check(err)
	if !found {
		panic("Input Note is not in this certified recognized prefix")
	}
	capacity := must(provenance.Uint(prepared.MainCircuitCapacity, 32)).Int64()
	if capacity < 1 || capacity > 190 || len(prepared.PublicInputs) != 17+3*int(capacity) {
		panic("Native statement width")
	}
	main := prepared.PublicInputs
	eq := func(a, b string) bool { return must(provenance.Uint(a, 256)).Cmp(must(provenance.Uint(b, 256))) == 0 }
	if !eq(main[0], config.ChainID) ||
		!eq(main[1], config.ProtocolAddress) ||
		!eq(main[2], prepared.TokenAddress) ||
		!eq(main[3], "1") ||
		!eq(main[6+2*int(capacity)], nullifier) {
		panic("Main statement does not spend this exact one-input Note")
	}
	digest := must(provenance.OperationDigestDecimal(trust.Scope, prepared.MainCircuitCapacity, prepared.PublicInputs))
	manifest := must(os.ReadFile(filepath.Join(config.AssetRoot, "manifest.json")))
	identity := must(provenance.AuthenticateManifest(manifest, pin))
	prover := provenance.NativeProver{
		NoirExecutePath:       os.Getenv("NOIR_EXECUTOR"),
		BBPath:                os.Getenv("BB_BINARY"),
		TrustedManifestSHA256: pin,
		LoadAssets: func(capacity int) (provenance.LocalAssets, error) {
			for _, entry := range identity.Manifest.Circuits {
				if entry.Capacity == capacity {
					artifact, err := os.ReadFile(filepath.Join(config.AssetRoot, entry.Artifact))
					if err != nil {
						return provenance.LocalAssets{}, err
					}
					vk, err := os.ReadFile(filepath.Join(config.AssetRoot, entry.VK))
					return provenance.LocalAssets{Manifest: manifest, Artifact: artifact, VerificationKey: vk}, err
				}
			}
			return provenance.LocalAssets{}, fmt.Errorf("unsupported capacity")
		}}
	// One source input uses K1, independently of the main circuit's total input/output slots.
	// ExportSidecar constructs the witness and calls the real local prover, including independent verify.
	sidecar := must(provenance.ExportSidecar(ctx, provenance.WitnessRequest{
		Scope:           trust.Scope,
		Capacity:        1,
		TokenAddress:    prepared.TokenAddress,
		ChainGeneration: trust.ChainGeneration,
		OperationDigest: digest,
		Inputs:          []provenance.WitnessInput{{Opening: prepared.Opening, SourceIndex: 0, LeafIndex: membership.LeafIndex, Path: membership.Path}},
		Sources:         []provenance.Source{{TreeID: cert.TreeID, Root: cert.Root, LeafCount: cert.LeafCount}},
	}, []provenance.RootCertificate{cert}, trust, identity, prover, 0))
	evaluated := must(source.EvaluateTransaction(ctx, prepared.Data, []provenance.ProofSidecar{sidecar}))
	tracking := must(provenance.PreflightTracking(evaluated, "0", digest))
	relay := must(broadcaster.NewProvenanceClient(broadcaster.Configuration{
		Endpoint:     config.Broadcaster.Endpoint,
		QuoteSigner:  common.HexToAddress(config.Broadcaster.QuoteSigner),
		FeeRecipient: common.HexToAddress(config.Broadcaster.FeeRecipient),
	}, common.HexToHash(config.PolicyID)))
	data := must(hex.DecodeString(strings.TrimPrefix(prepared.Data, "0x")))
	chain := must(provenance.Uint(config.ChainID, 64))
	// Main calldata already includes the settled fee. Changing it requires fresh main/source proofs.
	quote := must(relay.Quote(ctx, data, broadcaster.Trust{
		ChainID:  new(big.Int).Set(chain),
		Protocol: common.HexToAddress(config.ProtocolAddress),
	}, []provenance.ProofSidecar{sidecar}))
	id := os.Getenv("BROADCAST_ID")
	if len(id) == 0 || len(id) > 63 {
		panic("Set stable BROADCAST_ID, at most 63 ASCII characters")
	}
	for _, c := range []byte(id) {
		if c < 33 || c > 126 {
			panic("Invalid BROADCAST_ID")
		}
	}
	record := map[string]any{"data": prepared.Data, "sidecar": sidecar, "broadcastId": id, "quote": quote, "operationDigest": digest, "trackingId": tracking.ID}
	// Save public evidence before the external side effect. Never overwrite an uncertain submission.
	file := must(os.OpenFile(os.Getenv("PROVENANCE_RECORD"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600))
	check(json.NewEncoder(file).Encode(record))
	check(file.Sync())
	check(file.Close())
	fmt.Println("Public sidecar and verified quote saved; retain the record for confirmation/recovery.")
	if slices.Contains(os.Args, "--submit") {
		fmt.Printf("Broadcaster status: %+v\n", must(relay.Submit(ctx, data, quote, id, []provenance.ProofSidecar{sidecar})))
	}
	// On uncertain HTTP results, recover this ID; do not create another payment.
}
