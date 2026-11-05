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
	walletsession "github.com/invisi-lab/invisi-sdk-go/session"
	walletsqlite "github.com/invisi-lab/invisi-sdk-go/session/sqlite"
)

func main() {
	execute := flag.Bool("execute", false, "authorize signing, proving and chain submission")
	flag.Parse()
	if !*execute {
		log.Fatal("Review settings and code, then add --execute; no network request or transaction was made")
	}
	ctx := context.Background()

	// 1. Load local demonstration accounts.
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) < 2 {
		log.Fatal("this example requires at least 2 initialized accounts")
	}
	alice, bob := accounts[0], accounts[1]

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

	// 3. Create the local prover; private witnesses stay on this machine.
	localProver, err := native.NewNativeProver(shared.Settings.Prover)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Open the encrypted wallet inventory with an application-owned storage key.
	storageKey := crypto.Keccak256([]byte("Invisi.GoExamples.Storage.v1"), crypto.FromECDSA(alice.Key))
	defer clear(storageKey)
	codec, err := walletsession.NewEncryptedWalletCodec(storageKey)
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
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client, Prover: localProver})
	localWallet, err := sdk.WalletFromPrivateKey("0x" + hex.EncodeToString(crypto.FromECDSA(alice.Key)))
	if err != nil {
		log.Fatal(err)
	}

	// 5. Create the managed wallet session.
	session, err := walletsession.Connect(ctx, walletsession.Options{Invisi: invisi, Store: store, Wallet: localWallet, Protocol: pool, Signer: signer, Namespace: "go-examples"})
	if err != nil {
		log.Fatal(err)
	}

	// 6. Refresh authenticated wallet state and inspect the token balance.
	refreshed, err := session.Refresh(ctx, walletsession.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock})
	if err != nil {
		log.Fatal(err)
	}
	balance, err := session.Balance(ctx, token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Refreshed through block:", refreshed.Block.Number, "Available funds:", balance.AvailableKnown)

	// 7. Build the complete output with its deployment, token, amount and recipient.
	output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
		ChainID:         new(big.Int).SetUint64(pool.Context().ChainID),
		ProtocolAddress: pool.Context().ProtocolAddress,
		TokenAddress:    token,
		Amount:          big.NewInt(10),
		Recipient:       sdk.InvisiRecipient{Address: recipient},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 8. Prepare the managed transaction and reserve its inputs.
	prepared, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Review:", prepared.Preview())

	// 9. Authorize spending through the managed session.
	authorized, err := prepared.Authorize(ctx)
	if err != nil {
		log.Fatal(err)
	}
	ready, err := authorized.Prove(ctx, sdk.TransactionProveOptions{Options: prover.Options{OnProgress: func(p prover.Progress) {
		fmt.Println("Proof phase:", p.Phase)
	}}})
	if err != nil {
		log.Fatal(err)
	}
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
