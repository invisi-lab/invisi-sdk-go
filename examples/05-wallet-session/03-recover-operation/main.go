package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsqlite "github.com/invisi-lab/invisi-sdk-go/sdk/sqlite"
)

func main() {
	recoverID := flag.String("recover", "", "resume a persisted operation ID")
	transactionHash := flag.String("transaction", "", "previously submitted transaction hash for recovery")
	flag.Parse()
	ctx := context.Background()

	// 1. Load local demonstration accounts.
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) < 2 {
		log.Fatal("this example requires at least 2 initialized accounts")
	}
	alice := accounts[0]

	// 2. Connect to the chain and pin the Protocol RPC deployment.
	provider, err := ethclient.DialContext(ctx, shared.Settings.ChainRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	chain, err := provider.ChainID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if chain.Uint64() != shared.Settings.ChainID {
		log.Fatal("chain does not match Settings.ChainID")
	}
	client, err := sdk.NewProtocolRpcClient(shared.Settings.ProtocolRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, client, provider)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Open the encrypted wallet inventory with an application-owned storage key.
	storageKey := crypto.Keccak256([]byte("Invisi.GoExamples.Storage.v1"), crypto.FromECDSA(alice.Key))
	defer clear(storageKey)
	codec, err := sdk.NewEncryptedWalletCodec(storageKey)
	if err != nil {
		log.Fatal(err)
	}
	database, err := sql.Open("sqlite", shared.Settings.DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	store, err := walletsqlite.NewWalletStore(ctx, database, codec)
	if err != nil {
		log.Fatal(err)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client, Store: store})
	localWallet, err := sdk.WalletFromPrivateKey("0x" + hex.EncodeToString(crypto.FromECDSA(alice.Key)))
	if err != nil {
		log.Fatal(err)
	}

	// 4. Create the managed wallet session.
	session, err := invisi.Connect(ctx, sdk.ConnectionOptions{Wallet: localWallet, Protocol: pool, Namespace: "go-examples"})
	if err != nil {
		log.Fatal(err)
	}
	if *recoverID == "" {
		log.Fatal("provide --recover with a persisted operation ID")
	}

	// 5. Resume the persisted operation and reconcile its transaction hash.
	operation, err := session.Resume(ctx, *recoverID)
	if err != nil {
		log.Fatal(err)
	}
	if *transactionHash != "" {
		if len(*transactionHash) != 66 {
			log.Fatal("transaction must be a 32-byte hex hash")
		}
		operation, err = operation.TrackTransaction(ctx, common.HexToHash(*transactionHash))
		if err != nil {
			log.Fatal(err)
		}
	}
	result, err := operation.Reconcile(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Recovered operation status:", result.Status())
}
