// 04 charlie cancel.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
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
	if len(accounts) < 3 {
		log.Fatal("this example requires at least 3 initialized accounts")
	}
	charlie := accounts[2]

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
	signer, err := bind.NewKeyedTransactorWithChainID(charlie.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx

	// 3. Create the local prover; private witnesses stay on this machine.
	localProver, err := native.NewNativeProver(shared.Settings.Prover)
	if err != nil {
		log.Fatal(err)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client, Prover: localProver})
	var prepared *sdk.PreparedTransaction
	identity := sdk.TransactionIdentity{Wallet: charlie.PublicWallet, Protocol: pool}
	role := "owner"

	// 4. Discover authenticated Credit records for this role.
	candidates, err := shared.DiscoverCredits(ctx, provider, client, pool, charlie.PublicWallet, charlie.ViewingKey, token, role)
	if err != nil {
		log.Fatal(err)
	}
	eligible := []sdk.CreditInput{}
	for _, candidate := range candidates {
		eligible = append(eligible, candidate.Input)
	}
	if len(eligible) == 0 {
		log.Fatal("no eligible Credit for this account; run the corresponding creation example first")
	}
	selected := eligible[0]

	// 5. Prepare and freeze the transaction intent.
	prepared, err = invisi.PrepareCancelCredit(ctx, sdk.ExplicitCancelCreditPreparation{
		TransactionIdentity:     identity,
		CreditInputs:            []sdk.CreditInput{selected},
		CancelCreditPreparation: sdk.CancelCreditPreparation{TokenAddress: token, Credit: selected.Commitment},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Review frozen operation:", prepared.Preview())

	// 6. Export and sign the owner authorization without sharing the private key.
	request, err := prepared.ExportWalletAuthorization(ctx)
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
	signature, err := crypto.Sign(digest, charlie.Key)
	if err != nil {
		log.Fatal(err)
	}
	prepared, err = prepared.WithWalletSignature(ctx, signature)
	if err != nil {
		log.Fatal(err)
	}

	// 7. Generate and verify the transaction proof locally.
	prepared, err = prepared.Prove(ctx, sdk.TransactionProveOptions{Options: prover.Options{OnProgress: func(p prover.Progress) {
		fmt.Println("Proof phase:", p.Phase)
	}}, Execution: "single"})
	if err != nil {
		log.Fatal(err)
	}

	// 8. Export calldata for ordinary geth submission.
	call, err := prepared.ExportTransaction(ctx)
	if err != nil {
		log.Fatal(err)
	}
	transaction, err := bind.NewBoundContract(call.To, abi.ABI{}, provider, provider, provider).RawTransact(signer, call.Data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Submitted transaction:", transaction.Hash())
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
}
