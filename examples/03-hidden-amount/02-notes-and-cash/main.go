// 02 notes and cash.
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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
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
	if len(accounts) < 4 {
		log.Fatal("this example requires at least 4 initialized accounts")
	}
	bob, dave := accounts[1], accounts[3]

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
	signer, err := bind.NewKeyedTransactorWithChainID(bob.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx
	recipient, err := dave.PublicWallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}

	// 3. Create the local prover; private witnesses stay on this machine.
	localProver, err := native.NewNativeProver(shared.Settings.Prover)
	if err != nil {
		log.Fatal(err)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client, Prover: localProver})
	amount := big.NewInt(10)
	input := []sdk.NoteInput{}

	// 4. Discover authenticated notes and choose this transaction's inputs.
	candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, bob.PublicWallet, bob.ViewingKey, token)
	if err != nil {
		log.Fatal(err)
	}
	selectedAmount := new(big.Int)
	required := new(big.Int).Set(amount)
	if amount.Cmp(big.NewInt(1)) <= 0 {
		log.Fatal("payment amount must exceed the 1-unit public funding contribution")
	}
	required.Sub(required, big.NewInt(1))
	for _, candidate := range candidates {
		if selectedAmount.Cmp(required) >= 0 {
			break
		}
		input = append(input, candidate.Input)
		selectedAmount.Add(selectedAmount, core.ElementToBigInt(candidate.Input.Note.Amount))
	}
	if len(input) > 12 {
		log.Fatal("selected notes exceed this example's 12-input limit; consolidate or select a smaller set")
	}
	if len(input) == 0 {
		log.Fatal("no authenticated unspent notes; shield to this owner first")
	}
	payload, err := core.FundsV1.Encode(core.PayloadDataV1{Memo: []byte("Hidden amount payment")})
	if err != nil {
		log.Fatal(err)
	}

	// 5. Build the complete output with its deployment, token, amount and recipient.
	output, err := sdk.NewNoteOutput(sdk.NoteOutputOptions{
		ChainID:         chain,
		ProtocolAddress: pool.Context().ProtocolAddress,
		TokenAddress:    token,
		Amount:          amount,
		Recipient:       sdk.InvisiRecipient{Address: recipient},
		Payload:         payload,
	})
	if err != nil {
		log.Fatal(err)
	}
	params := sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}}
	authorization := "caller"
	fundingAmount := big.NewInt(1)
	params.Funding = &sdk.FundingSource{
		Payer:         bob.PublicWallet.OwnerAddress(),
		Amount:        fundingAmount,
		Authorization: authorization,
	}
	tokenABI, err := abi.JSON(strings.NewReader(`[{"type":"function","name":"approve","stateMutability":"nonpayable","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"type":"bool"}]}]`))
	if err != nil {
		log.Fatal(err)
	}
	fundingSigner, err := bind.NewKeyedTransactorWithChainID(bob.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	fundingSigner.Context = ctx
	approval, err := bind.NewBoundContract(token, tokenABI, provider, provider, provider).Transact(fundingSigner, "approve", pool.Context().ProtocolAddress, fundingAmount)
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

	// 6. Prepare and freeze the transaction intent.
	prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: bob.PublicWallet, Protocol: pool},
		TransferPreparation: params,
		ExplicitFunds:       sdk.ExplicitFunds{Inputs: input},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Review frozen operation:", prepared.Preview())

	// 7. Export and sign the owner authorization without sharing the private key.
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
	signature, err := crypto.Sign(digest, bob.Key)
	if err != nil {
		log.Fatal(err)
	}
	prepared, err = prepared.WithWalletSignature(ctx, signature)
	if err != nil {
		log.Fatal(err)
	}

	// 8. Generate and verify the transaction proof locally.
	prepared, err = prepared.Prove(ctx, sdk.TransactionProveOptions{Options: prover.Options{OnProgress: func(p prover.Progress) {
		fmt.Println("Proof phase:", p.Phase)
	}}, Execution: "single"})
	if err != nil {
		log.Fatal(err)
	}

	// 9. Export calldata for ordinary geth submission.
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
