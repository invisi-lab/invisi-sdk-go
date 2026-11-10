// Register Bob's public receiving keys; Bob submits the exported direct-owner call.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// Configure public actors and deployment.
	ctx := context.Background()
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC URL")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	address := flag.String("wallet-address", "", "public long receiving address")

	flag.Parse()

	publicWallet, err := sdk.WalletFromAddress(*address)
	if err != nil {
		log.Fatal(err)
	}

	// Authenticate the Protocol RPC deployment against Chain RPC before preparing financial intent.
	provider, err := ethclient.DialContext(ctx, *chainRPC)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	client, err := sdk.NewProtocolRpcClient(*protocolRPC)
	if err != nil {
		log.Fatal(err)
	}

	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client})

	// Freeze recipients and amounts before authorization.
	prepared, err := invisi.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: publicWallet, Update: false})
	if err != nil {
		log.Fatal(err)
	}
	prepared, err = prepared.AuthorizeDirect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Export unsigned calldata for the external signer; submission and canonical reconciliation remain explicit.
	call, err := prepared.ExportTransaction(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", call)
}
