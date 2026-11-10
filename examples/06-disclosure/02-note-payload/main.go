// Discover an owned Note, verify its scoped opening and decode the authenticated application memo.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// Configure public actors and deployment.
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
	decryptor := shared.AppDeliveryDecryptor{}

	// Discover automatically through canonical receipts and externally decrypted, independently checked openings.
	candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, publicWallet, decryptor, token)
	if err != nil {
		log.Fatal(err)
	}
	if len(candidates) == 0 {
		log.Fatal("no authenticated unspent Note")
	}
	selected := candidates[0]
	disclosure, err := sdk.DisclosureFromNote(selected.Input.Note)
	if err != nil {
		log.Fatal(err)
	}
	verifier, err := sdk.NewDisclosureVerifier(sdk.DisclosureVerifierOptions{Protocol: pool, Provider: provider, TokenAddress: token, Confirmations: 1})
	if err != nil {
		log.Fatal(err)
	}
	if _, err = verifier.Verify(ctx, disclosure, sdk.DisclosureReference{
		TransactionHash: selected.Output.TransactionHash,
		TreeID:          selected.Input.TreeID,
		LeafIndex:       selected.Input.LeafIndex,
		Role:            "funds",
	}); err != nil {
		log.Fatal(err)
	}
	payload, err := core.FundsV1.Decode(selected.Input.Note.Payload)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Memo bytes:", len(payload.Memo))
	// This scoped opening is intended only for the selected auditor.
	log.Printf("%+v", disclosure.Export())
}
