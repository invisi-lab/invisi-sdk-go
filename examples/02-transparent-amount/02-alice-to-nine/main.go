// 02 alice to nine.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
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
	if len(accounts) < 10 {
		log.Fatal("this example requires at least 10 initialized accounts")
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
	token := common.HexToAddress(shared.Settings.TokenAddress)
	signer, err := bind.NewKeyedTransactorWithChainID(alice.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client})

	// 3. Build one complete output for each recipient.
	outputs := make([]sdk.NoteOutput, 9)
	for i := range outputs {
		recipient, err := accounts[i+1].PublicWallet.GetAddress()
		if err != nil {
			log.Fatal(err)
		}
		outputs[i], err = sdk.NewNoteOutput(sdk.NoteOutputOptions{
			ChainID:         chain,
			ProtocolAddress: pool.Context().ProtocolAddress,
			TokenAddress:    token,
			Amount:          big.NewInt(10),
			Recipient:       sdk.InvisiRecipient{Address: recipient},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	amount := big.NewInt(90)
	authorization := "caller"
	tokenABI, err := abi.JSON(strings.NewReader(`[{"type":"function","name":"approve","stateMutability":"nonpayable","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"type":"bool"}]}]`))
	if err != nil {
		log.Fatal(err)
	}
	approval, err := bind.NewBoundContract(token, tokenABI, provider, provider, provider).Transact(signer, "approve", pool.Context().ProtocolAddress, amount)
	if err != nil {
		log.Fatal(err)
	}
	receipt, err := bind.WaitMined(ctx, provider, approval)
	if err != nil {
		log.Fatal(err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Fatal("approval reverted")
	}

	// 4. Prepare and freeze the transaction intent.
	prepared, err := invisi.PrepareShield(ctx, sdk.ExplicitShieldPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: alice.PublicWallet, Protocol: pool},
		ShieldPreparation: sdk.ShieldPreparation{
			TokenAddress: token,
			Outputs:      outputs,
			Funding: sdk.FundingSource{
				Payer:         alice.PublicWallet.OwnerAddress(),
				Amount:        amount,
				Authorization: authorization,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Review frozen shield:", prepared.Preview())
	prepared, err = prepared.Ready(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 5. Export calldata for ordinary geth submission.
	call, err := prepared.ExportTransaction(ctx)
	if err != nil {
		log.Fatal(err)
	}
	transaction, err := bind.NewBoundContract(call.To, abi.ABI{}, provider, provider, provider).RawTransact(signer, call.Data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Submitted transaction:", transaction.Hash())
	receipt, err = bind.WaitMined(ctx, provider, transaction)
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
