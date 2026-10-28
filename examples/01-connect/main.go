// Select a Protocol through independent Chain RPC and Protocol RPC services.
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"log"
	"os"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
func run(ctx context.Context) error {
	chainURL, protocolURL := os.Getenv("CHAIN_RPC_URL"), os.Getenv("PROTOCOL_RPC_URL")
	if chainURL == "" || protocolURL == "" {
		return errors.New("set CHAIN_RPC_URL and PROTOCOL_RPC_URL")
	}
	provider, err := ethclient.DialContext(ctx, chainURL)
	if err != nil {
		return err
	}
	defer provider.Close()
	protocolRPC, err := sdk.NewProtocolRpcClient(protocolURL)
	if err != nil {
		return err
	}
	// This verifies the deployment code and VERSION against the selected chain.
	pool, err := sdk.ProtocolFromRPC(ctx, protocolRPC, provider)
	if err != nil {
		return err
	}
	fmt.Println("Verified protocol:", pool.Context().ProtocolAddress)
	fmt.Println("Chain ID:", pool.Context().ChainID)
	// No wallet, Store, prover or history refresh is needed to connect.
	return nil
}
