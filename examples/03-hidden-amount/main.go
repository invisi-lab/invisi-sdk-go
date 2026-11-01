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
	mode := flag.String("mode", "notes", "notes, cash, mixed, withdraw or sweep")
	signatureFunding := flag.Bool("signature-funding", false, "sign public funding authorization")
	amountText := flag.String("amount", "10", "payment or withdrawal in token base units")
	actor := flag.Int("actor", 0, "funding/spending account index")
	recipientIndex := flag.Int("recipient", 1, "receiving account index; set equal to actor for own notes")
	submitterIndex := flag.Int("submitter", -1, "separate transaction sender index, e.g. Frank=5; defaults to actor")
	flag.Parse()
	if !*execute {
		log.Fatal("Review settings and code, then add --execute; no network request or transaction was made")
	}

	if *signatureFunding && *mode != "cash" && *mode != "mixed" {
		log.Fatal("signature funding requires cash or mixed mode")
	}
	ctx := context.Background()
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if *actor < 0 || *actor >= len(accounts) || *recipientIndex < 0 || *recipientIndex >= len(accounts) {
		log.Fatal("account index out of range")
	}
	alice, bob := accounts[*actor], accounts[*recipientIndex]
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
	if *submitterIndex == -1 {
		*submitterIndex = *actor
	}
	if *submitterIndex < 0 || *submitterIndex >= len(accounts) {
		log.Fatal("submitter index out of range")
	}
	if *submitterIndex != *actor && (*mode == "cash" || *mode == "mixed") && !*signatureFunding {
		log.Fatal("a separate funding submitter requires --signature-funding")
	}
	submitter := accounts[*submitterIndex]
	signer, err := bind.NewKeyedTransactorWithChainID(submitter.Key, chain)
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
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client, Prover: localProver})

	amount, ok := new(big.Int).SetString(*amountText, 10)
	if !ok || amount.Sign() <= 0 {
		log.Fatal("positive decimal amount required")
	}
	input := []sdk.NoteInput{}
	if *mode != "cash" {
		candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, alice.PublicWallet, alice.ViewingKey, token)
		if err != nil {
			log.Fatal(err)
		}
		selectedAmount := new(big.Int)
		required := new(big.Int).Set(amount)
		if *mode == "mixed" {
			if amount.Cmp(big.NewInt(1)) <= 0 {
				log.Fatal("mixed mode amount must exceed its 1-unit public funding contribution")
			}
			required.Sub(required, big.NewInt(1))
		}
		for _, candidate := range candidates {
			if *mode != "sweep" && selectedAmount.Cmp(required) >= 0 {
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
	}
	payload, err := core.FundsV1.Encode(core.PayloadDataV1{Memo: []byte("Hidden amount payment")})
	if err != nil {
		log.Fatal(err)
	}
	params := sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.PaymentOutput{{To: recipient, Amount: amount, Payload: payload}}}
	switch *mode {
	case "notes":
	case "cash", "mixed":
		authorization := "caller"
		if *signatureFunding {
			authorization = "signature"
		}
		fundingAmount := new(big.Int).Set(amount)
		if *mode == "mixed" {
			fundingAmount = big.NewInt(1)
		}
		params.Funding = &sdk.FundingSource{Payer: alice.PublicWallet.OwnerAddress(), Amount: fundingAmount, Authorization: authorization}
		tokenABI, err := abi.JSON(strings.NewReader(`[{"type":"function","name":"approve","stateMutability":"nonpayable","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"type":"bool"}]}]`))
		if err != nil {
			log.Fatal(err)
		}
		fundingSigner, err := bind.NewKeyedTransactorWithChainID(alice.Key, chain)
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
	case "withdraw":
		params.Outputs = nil
		params.Withdrawal = &sdk.Withdrawal{To: alice.PublicWallet.OwnerAddress(), Amount: amount}
	case "sweep":
		total := new(big.Int)
		for _, note := range input {
			total.Add(total, core.ElementToBigInt(note.Note.Amount))
		}
		params.Outputs = nil
		params.Withdrawal = &sdk.Withdrawal{To: alice.PublicWallet.OwnerAddress(), Amount: total}
	default:
		log.Fatal("unknown hidden amount mode")
	}
	prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{TransactionIdentity: sdk.TransactionIdentity{Wallet: alice.PublicWallet, Protocol: pool}, TransferPreparation: params, ExplicitFunds: sdk.ExplicitFunds{Inputs: input}})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Review frozen operation:", prepared.Preview())
	if prepared.Preview().NeedsWalletAuthorization {
		request, err := prepared.ExportWalletAuthorization(ctx)
		if err != nil {
			log.Fatal(err)
		}
		// A hardware wallet or another process can receive this JSON-safe EIP-712 request instead.
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
		prepared, err = prepared.WithWalletSignature(ctx, signature)
		if err != nil {
			log.Fatal(err)
		}
	}

	prepared, err = prepared.Prove(ctx, sdk.TransactionProveOptions{Options: prover.Options{OnProgress: func(p prover.Progress) { fmt.Println("Proof phase:", p.Phase) }}, Execution: "single"})
	if err != nil {
		log.Fatal(err)
	}

	if *signatureFunding {
		// Funding authorization follows proof generation and binds its final proof bytes.
		request, err := prepared.ExportFundingAuthorization(ctx)
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
		prepared, err = prepared.WithFundingSignature(ctx, signature)
		if err != nil {
			log.Fatal(err)
		}
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

}
