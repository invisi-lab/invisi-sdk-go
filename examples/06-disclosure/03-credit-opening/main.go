package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"log"
)

func main() {
	ctx := context.Background()
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) < 2 {
		log.Fatal("initialize local accounts first")
	}
	owner := accounts[1]
	provider, err := ethclient.DialContext(ctx, shared.Settings.ChainRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	client, err := sdk.NewProtocolRpcClient(shared.Settings.ProtocolRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, client, provider)
	if err != nil {
		log.Fatal(err)
	}
	candidates, err := shared.DiscoverCredits(ctx, provider, client, pool, owner.PublicWallet, owner.ViewingKey, common.HexToAddress(shared.Settings.TokenAddress), "owner")
	if err != nil {
		log.Fatal(err)
	}
	if len(candidates) == 0 {
		log.Fatal("no authenticated owner grant")
	}
	selected := candidates[0]
	disclosure, err := credit.Disclose(selected.Input.State)
	if err != nil {
		log.Fatal(err)
	}
	// The commitment anchor came from canonical receipt authentication during discovery.
	if _, err = disclosure.Verify(selected.Commitment); err != nil {
		log.Fatal(err)
	}
	encoded, err := json.MarshalIndent(disclosure.Export(), "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sensitive grant-scoped audit export; it permits following this grant's history:")
	fmt.Println(string(encoded))
}
