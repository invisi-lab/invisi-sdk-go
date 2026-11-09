// Discover an owned Credit state and export the grant-scoped audit opening.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// Configure public actors and deployment; the native host retains account access.
	ctx := context.Background()
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC URL")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	address := flag.String("wallet-address", "", "public long receiving address")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
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
	pool, err := sdk.ProtocolFromRPC(ctx, client, provider)
	if err != nil {
		log.Fatal(err)
	}
	token := common.HexToAddress(*tokenText)
	encoder := json.NewEncoder(os.Stdout)
	decoder := json.NewDecoder(os.Stdin)
	nativeWallet := shared.NativeWalletClient{Encoder: encoder, Decoder: decoder}

	// Discover automatically through canonical receipts and externally decrypted, independently checked openings.
	candidates, err := shared.DiscoverCredits(ctx, provider, client, pool, publicWallet, nativeWallet, token, "owner")
	if err != nil {
		log.Fatal(err)
	}
	if len(candidates) == 0 {
		log.Fatal("no authenticated active owner grant")
	}
	selected := candidates[0]
	disclosure, err := credit.Disclose(selected.Input.State)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = disclosure.Verify(selected.Commitment); err != nil {
		log.Fatal(err)
	}
	// This scoped opening is intended only for the selected auditor.
	if err = encoder.Encode(disclosure.Export()); err != nil {
		log.Fatal(err)
	}
}
