package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	execute := flag.Bool("execute", false, "authorize signing, proving and chain submission")
	update := flag.Bool("update", false, "replace existing receiving keys")
	actor := flag.Int("actor", 1, "account index to register or update (default Bob)")
	flag.Parse()
	if !*execute {
		log.Fatal("Review settings and code, then add --execute; no network request or transaction was made")
	}

	ctx := context.Background()
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if *actor < 0 || *actor >= len(accounts) {
		log.Fatal("actor index out of range")
	}
	alice := accounts[*actor]
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
	signer, err := bind.NewKeyedTransactorWithChainID(alice.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client})
	prepared, err := invisi.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: alice.PublicWallet, Update: *update})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Review registry operation:", prepared.Preview())
	request, err := prepared.ExportRegistrationAuthorization(ctx)
	if err != nil {
		log.Fatal(err)
	}
	typed, err := request.TypedData()
	if err != nil {
		log.Fatal(err)
	}
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		log.Fatal(err)
	}
	signature, err := crypto.Sign(digest, alice.Key)
	if err != nil {
		log.Fatal(err)
	}
	prepared, err = prepared.WithSignature(ctx, signature)
	if err != nil {
		log.Fatal(err)
	}

	call, err := prepared.ExportTransaction(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// Ordinary geth owns gas, nonce selection and submission. This branch requires --execute.
	transaction, err := bind.NewBoundContract(call.To, abi.ABI{}, provider, provider, provider).RawTransact(signer, call.Data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Submitted transaction:", transaction.Hash())
	tracked, err := prepared.TrackTransaction(ctx, transaction.Hash())
	if err != nil {
		log.Fatal(err)
	}
	receipt, err := bind.WaitMined(ctx, provider, transaction)
	if err != nil {
		log.Fatal(err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Fatal("transaction reverted")
	}
	canonical, err := provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if err != nil {
		log.Fatal(err)
	}
	if canonical == nil || canonical.Hash() != receipt.BlockHash {
		log.Fatal("transaction receipt is no longer canonical; preserve the transaction hash and reconcile")
	}
	fmt.Println("Mined block:", receipt.BlockNumber)
	// Persist the transaction hash before waiting in a production application, then reconcile canonical state.

	confirmed, err := tracked.Reconcile(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Canonical registration status:", confirmed.Status())
	registry, err := sdk.NewAddressRegistry(pool, provider)
	if err != nil {
		log.Fatal(err)
	}
	address, err := registry.Resolve(ctx, alice.PublicWallet.OwnerAddress().Hex())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Registered receiving address:", address)

}
