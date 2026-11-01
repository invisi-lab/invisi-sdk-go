package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsqlite "github.com/invisi-lab/invisi-sdk-go/sdk/sqlite"
)

func main() {
	execute := flag.Bool("execute", false, "authorize signing, proving and chain submission")
	recoverID := flag.String("recover", "", "resume a persisted operation ID")
	transactionHash := flag.String("transaction", "", "previously submitted transaction hash for recovery")
	flag.Parse()
	ctx := context.Background()
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	alice, bob := accounts[0], accounts[1]
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
	token := common.HexToAddress(shared.Settings.TokenAddress)
	signer, err := bind.NewKeyedTransactorWithChainID(alice.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx
	recipient, err := bob.PublicWallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}

	localProver, err := native.NewNativeProver(shared.Settings.Prover)
	if err != nil {
		log.Fatal(err)
	}
	// This example derives a separate application storage key locally; production applications can use a key manager.
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
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client, Prover: localProver, Store: store})
	localWallet, err := sdk.WalletFromPrivateKey("0x" + hex.EncodeToString(crypto.FromECDSA(alice.Key)))
	if err != nil {
		log.Fatal(err)
	}
	// Private capabilities and the geth sender are injected only into the managed connection.
	session, err := invisi.Connect(ctx, sdk.ConnectionOptions{Wallet: localWallet, Protocol: pool, Signer: signer, Namespace: "go-examples"})
	if err != nil {
		log.Fatal(err)
	}
	if *recoverID != "" {
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
		return
	}
	refreshed, err := session.Refresh(ctx, sdk.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock})
	if err != nil {
		log.Fatal(err)
	}
	balance, err := session.Balance(ctx, token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Refreshed through block:", refreshed.Block.Number, "Available funds:", balance.AvailableKnown)
	if !*execute {
		fmt.Println("Read-only restore complete. Add --execute to prepare and submit a transfer.")
		return
	}
	prepared, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.PaymentOutput{{To: recipient, Amount: big.NewInt(10)}}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Review:", prepared.Preview())
	authorized, err := prepared.Authorize(ctx)
	if err != nil {
		log.Fatal(err)
	}
	ready, err := authorized.Prove(ctx, sdk.TransactionProveOptions{Options: prover.Options{OnProgress: func(p prover.Progress) { fmt.Println("Proof phase:", p.Phase) }}})
	if err != nil {
		log.Fatal(err)
	}
	// A session reserves notes and persists the operation before its geth delegate sends.
	transaction, err := session.SendTransaction(ctx, ready)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Persisted operation ID:", ready.ID(), "Transaction:", transaction.Hash())
	receipt, err := bind.WaitMined(ctx, provider, transaction)
	if err != nil {
		log.Fatal(err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Fatal("transaction reverted; reconcile the persisted operation")
	}
	tracked, err := session.Resume(ctx, ready.ID())
	if err != nil {
		log.Fatal(err)
	}
	confirmed, err := tracked.Reconcile(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Canonical status:", confirmed.Status())

}
