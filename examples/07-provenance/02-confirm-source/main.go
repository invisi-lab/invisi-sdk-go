package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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
		fmt.Println("Pass --execute with TRANSACTION_HASH and the retained public record.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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
		ChainRPCURL      string
	}
	check(json.Unmarshal(must(os.ReadFile(os.Getenv("PROVENANCE_CONFIG"))), &config))
	var record struct {
		Data, BroadcastID, TrackingID, OperationDigest string
		Sidecar                                        provenance.ProofSidecar
	}
	check(json.Unmarshal(must(os.ReadFile(os.Getenv("PROVENANCE_RECORD"))), &record))
	trust := provenance.TrustContext{
		Scope: provenance.Scope{
			ChainID:         config.ChainID,
			ProtocolAddress: config.ProtocolAddress,
			PolicyID:        config.PolicyID,
		},
		ChainGeneration: config.Generation,
		ProviderID:      config.ProviderID,
		ManifestSHA256:  "f90f2ce43218de158c6392fa19e6517dda703c35823a4aa5bfc20b50788ad282",
		SigningKeys: map[string]ed25519.PublicKey{
			config.SigningKeyID: must(hex.DecodeString(strings.TrimPrefix(config.SigningPublicKey, "0x"))),
		},
	}
	check(provenance.ValidateProofBundle(record.Sidecar.ProofBundle, trust))
	primary := must(ethclient.DialContext(ctx, config.ChainRPCURL))
	defer primary.Close()
	if must(primary.ChainID(ctx)).Cmp(must(provenance.Uint(config.ChainID, 64))) != 0 {
		panic("Primary chain mismatch")
	}
	rawHash := os.Getenv("TRANSACTION_HASH")
	if len(rawHash) != 66 {
		panic("Set TRANSACTION_HASH")
	}
	hash := common.HexToHash(rawHash)
	receipt := must(primary.TransactionReceipt(ctx, hash))
	tx, pending, err := primary.TransactionByHash(ctx, hash)
	check(err)
	header := must(primary.HeaderByNumber(ctx, receipt.BlockNumber))
	if pending || receipt.Status != 1 ||
		header.Hash() != receipt.BlockHash || tx.Hash() != hash ||
		tx.To() == nil || *tx.To() != common.HexToAddress(config.ProtocolAddress) ||
		!strings.EqualFold("0x"+hex.EncodeToString(tx.Data()), record.Data) {
		panic("Successful canonical transaction mismatch")
	}
	source := &provenance.Client{Endpoint: config.Endpoint, Trust: trust}
	locator := map[string]any{"txHash": hash.Hex(), "blockHash": receipt.BlockHash.Hex(), "operationIndex": "0"}
	tracked := must(source.WaitTracking(ctx, record.TrackingID, 2*time.Second, 2*time.Minute))
	tracking := must(provenance.TrackingFromResult(tracked, record.TrackingID))
	if tracking.Status != "recognized" || tracking.Matched == nil || tracking.Matched.TxHash != hash.Hex() || tracking.Matched.BlockHash != receipt.BlockHash.Hex() || tracking.Matched.OperationIndex != "0" || must(provenance.Uint(tracking.OperationDigest, 256)).Cmp(must(provenance.Uint(record.OperationDigest, 256))) != 0 {
		panic("Tracking not recognized or matched another operation; retain original record")
	}
	operation := must(source.GetOperationRecognition(ctx, locator))
	if string(operation["recognitionState"]) != "\"recognized\"" {
		panic("Publication pending; retry the same tracking ID")
	}
	fmt.Println("Source service confirms operation recognition")
}
