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
		Data, BroadcastID string
		Sidecar           provenance.ProofSidecar
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
	submitted := must(source.SubmitProof(ctx, locator, record.Sidecar.ProofBundle, record.BroadcastID+"-source"))
	var jobID string
	check(json.Unmarshal(submitted["jobId"], &jobID))
	job := must(source.WaitProofJob(ctx, jobID, time.Second))
	if string(job["status"]) != "\"valid\"" {
		panic("Proof not valid; retain transaction and record for retry/reproof")
	}
	// Proof verification is not publication. Wait for this exact operation's recognized state.
	for attempt := 0; attempt < 60; attempt++ {
		operation := must(source.GetOperationRecognition(ctx, locator))
		if string(operation["recognitionState"]) == "\"recognized\"" {
			fmt.Println("Source service confirms operation recognition")
			return
		}
		select {
		case <-ctx.Done():
			panic(ctx.Err())
		case <-time.After(2 * time.Second):
		}
	}
	panic("Publication pending; rerun with the same transaction and record")
}
