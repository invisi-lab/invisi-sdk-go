package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	if !slices.Contains(os.Args, "--execute") {
		fmt.Println("Pass --execute to discover, prepare and prove; add --submit to broadcast. See README.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	var config struct {
		ChainID                                                                           string `json:"chainId"`
		ProtocolAddress, PolicyID, Generation, ProviderID, SigningKeyID, SigningPublicKey string
		Endpoint, AssetRoot, ChainRPCURL, ProtocolRPCURL                                  string
		WalletAddress, Recipient, TokenAddress, Amount, MaxFeeAmount                      string
		Broadcaster                                                                       struct{ Endpoint, QuoteSigner, FeeRecipient string }
	}
	check(json.Unmarshal(must(os.ReadFile(os.Getenv("PROVENANCE_CONFIG"))), &config))
	const pin = "f90f2ce43218de158c6392fa19e6517dda703c35823a4aa5bfc20b50788ad282"
	trust := provenance.TrustContext{Scope: provenance.Scope{ChainID: config.ChainID, ProtocolAddress: config.ProtocolAddress, PolicyID: config.PolicyID}, ChainGeneration: config.Generation, ProviderID: config.ProviderID, ManifestSHA256: pin, SigningKeys: map[string]ed25519.PublicKey{config.SigningKeyID: must(hex.DecodeString(strings.TrimPrefix(config.SigningPublicKey, "0x")))}}
	source := &provenance.Client{Endpoint: config.Endpoint, Trust: trust}
	manifest := must(os.ReadFile(filepath.Join(config.AssetRoot, "manifest.json")))
	identity := must(provenance.AuthenticateManifest(manifest, pin))
	sourceProver := provenance.NativeProver{NoirExecutePath: os.Getenv("NOIR_EXECUTOR"), BBPath: os.Getenv("BB_BINARY"), TrustedManifestSHA256: pin, LoadAssets: func(k int) (provenance.LocalAssets, error) {
		for _, entry := range identity.Manifest.Circuits {
			if entry.Capacity == k {
				a, e := os.ReadFile(filepath.Join(config.AssetRoot, entry.Artifact))
				if e != nil {
					return provenance.LocalAssets{}, e
				}
				vk, e := os.ReadFile(filepath.Join(config.AssetRoot, entry.VK))
				return provenance.LocalAssets{Manifest: manifest, Artifact: a, VerificationKey: vk}, e
			}
		}
		return provenance.LocalAssets{}, fmt.Errorf("unsupported capacity")
	}}
	primary := must(ethclient.DialContext(ctx, config.ChainRPCURL))
	defer primary.Close()
	rpc := must(sdk.NewProtocolRpcClient(config.ProtocolRPCURL))
	pool := must(sdk.ProtocolFromRPC(ctx, rpc, primary))
	if fmt.Sprint(pool.Context().ChainID) != config.ChainID || pool.Context().ProtocolAddress != common.HexToAddress(config.ProtocolAddress) {
		panic("configured deployment mismatch")
	}
	publicWallet := must(sdk.WalletFromAddress(config.WalletAddress))
	host := shared.NativeWalletClient{Encoder: json.NewEncoder(os.Stdout), Decoder: json.NewDecoder(os.Stdin)}
	backend := must(native.NewNativeProver(native.NativeOptions{BBBinary: os.Getenv("BB_BINARY"), NoirBinary: os.Getenv("NOIR_EXECUTOR")}))
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: primary, ProtocolRPC: rpc, Prover: backend, Provenance: &sdk.ProvenanceCapability{Client: source, Manifest: identity, Prover: sourceProver, MaxRootPages: 10, MaxLeaves: 100000}})
	amount := must(provenance.Uint(config.Amount, 128))
	maxFee := must(provenance.Uint(config.MaxFeeAmount, 128))
	token := common.HexToAddress(config.TokenAddress)
	if amount.Sign() == 0 {
		panic("positive payment required")
	}
	candidates := must(shared.DiscoverNotes(ctx, primary, rpc, pool, publicWallet, host, token))
	var inputs []sdk.NoteInput
	total := new(big.Int)
	required := new(big.Int).Add(new(big.Int).Set(amount), maxFee)
	for _, candidate := range candidates {
		if total.Cmp(required) >= 0 {
			break
		}
		inputs = append(inputs, candidate.Input)
		total.Add(total, core.ElementToBigInt(candidate.Input.Note.Amount))
	}
	if total.Cmp(required) < 0 {
		panic("insufficient authenticated private funds including fee budget")
	}
	output := must(sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: new(big.Int).SetUint64(pool.Context().ChainID), ProtocolAddress: pool.Context().ProtocolAddress, TokenAddress: token, Amount: amount, Recipient: sdk.InvisiRecipient{Address: config.Recipient}}))
	prepared := must(invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{TransactionIdentity: sdk.TransactionIdentity{Wallet: publicWallet, Protocol: pool}, ExplicitFunds: sdk.ExplicitFunds{Inputs: inputs, ChangeAddress: config.WalletAddress}, TransferPreparation: sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}}}))
	relay := must(broadcaster.NewProvenanceClient(broadcaster.Configuration{Endpoint: config.Broadcaster.Endpoint, QuoteSigner: common.HexToAddress(config.Broadcaster.QuoteSigner), FeeRecipient: common.HexToAddress(config.Broadcaster.FeeRecipient)}, common.HexToHash(config.PolicyID)))
	// Source roots, local membership and source proving use only the final fee-bearing plan.
	prepared = must(prepared.WithBroadcaster(ctx, relay, sdk.WithBroadcasterOptions{MaxFeeAmount: maxFee}))
	request := must(prepared.ExportWalletRequest(ctx, "authorize-final-plan"))
	responses := must(host.Execute(ctx, []wallet.Request{request}))
	prepared = must(prepared.WithWalletResponse(ctx, request, responses[0]))
	prepared = must(prepared.Prove(ctx))
	tracking := must(prepared.ExportProvenanceTracking(ctx))
	if len(tracking) != 1 {
		panic("missing durable source acknowledgement")
	}
	call := must(prepared.ExportTransaction(ctx))
	record := map[string]any{"data": "0x" + hex.EncodeToString(call.Data), "broadcastId": prepared.ID(), "trackingId": tracking[0].ID, "operationDigest": tracking[0].OperationDigest}
	// Exclusive creation prevents replacing an uncertain payment with a newly prepared one.
	path := os.Getenv("PROVENANCE_RECORD")
	if path == "" {
		panic("set PROVENANCE_RECORD to a new private path")
	}
	file := must(os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600))
	check(json.NewEncoder(file).Encode(record))
	check(file.Sync())
	check(file.Close())
	if slices.Contains(os.Args, "--submit") {
		status := must(prepared.SubmitBroadcaster(ctx, relay))
		fmt.Fprintln(os.Stderr, "Broadcaster status:", status.State, status.TransactionHash)
	} else {
		fmt.Fprintln(os.Stderr, "Prepared and proved; --submit was not supplied.")
	}
}
