package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"

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
	mode := flag.String("mode", "create", "create, draw, top-up or cancel")
	actor := flag.Int("actor", -1, "owner account index (default Bob); draw defaults to Alice")
	amountText := flag.String("amount", "10", "amount in token base units")
	flag.Parse()
	if !*execute {
		log.Fatal("Review settings and code, then add --execute; no network request or transaction was made")
	}

	ctx := context.Background()
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if *actor == -1 {
		if *mode == "draw" {
			*actor = 0
		} else {
			*actor = 1
		}
	}
	if *actor < 0 || *actor >= len(accounts) {
		log.Fatal("actor index out of range")
	}
	actorAccount, spender := accounts[*actor], accounts[0]
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
	signer, err := bind.NewKeyedTransactorWithChainID(actorAccount.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx
	recipient, err := spender.PublicWallet.GetAddress()
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
	head, err := provider.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	payload, err := core.CreditV1.Encode(core.PayloadDataV1{Memo: []byte("Credit grant")})
	if err != nil {
		log.Fatal(err)
	}
	var prepared *sdk.PreparedTransaction
	identity := sdk.TransactionIdentity{Wallet: actorAccount.PublicWallet, Protocol: pool}
	if *mode == "create" {
		candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, actorAccount.PublicWallet, actorAccount.ViewingKey, token)
		if err != nil {
			log.Fatal(err)
		}
		inputs := []sdk.NoteInput{}
		selectedAmount := new(big.Int)
		for _, candidate := range candidates {
			if selectedAmount.Cmp(amount) >= 0 {
				break
			}
			inputs = append(inputs, candidate.Input)
			selectedAmount.Add(selectedAmount, core.ElementToBigInt(candidate.Input.Note.Amount))
		}
		if len(inputs) > 10 {
			log.Fatal("Credit creation supports at most 10 input notes")
		}
		if len(inputs) == 0 {
			log.Fatal("owner needs authenticated token notes before creating Credit")
		}
		prepared, err = invisi.PrepareCreateCredit(ctx, sdk.ExplicitCreateCreditPreparation{TransactionIdentity: identity, ExplicitFunds: sdk.ExplicitFunds{Inputs: inputs}, CreateCreditPreparation: sdk.CreateCreditPreparation{TokenAddress: token, Spender: recipient, Amount: amount, Expiry: new(big.Int).SetUint64(head.Time + 86400), SingleLimit: new(big.Int).Set(amount), DailyLimit: new(big.Int).Set(amount), MonthlyLimit: new(big.Int).Set(amount), Payload: payload}})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		role := "owner"
		if *mode == "draw" {
			role = "spender"
		}
		candidates, err := shared.DiscoverCredits(ctx, provider, client, pool, actorAccount.PublicWallet, actorAccount.ViewingKey, token, role)
		if err != nil {
			log.Fatal(err)
		}
		eligible := []sdk.CreditInput{}
		for _, candidate := range candidates {
			expiry := core.ElementToBigInt(candidate.Input.State.Note.Expiry)
			expired := expiry.Cmp(new(big.Int).SetUint64(head.Time)) <= 0
			if *mode == "cancel" || !expired {
				eligible = append(eligible, candidate.Input)
			}
		}
		if len(eligible) == 0 {
			log.Fatal("no authenticated Credit matches this owner/spender and expiry; create Credit or select the appropriate actor")
		}
		switch *mode {
		case "draw":
			// Alice consumes one authenticated grant from each of two distinct owners in the same draw.
			// Create grants as Bob (--actor 1) and Charlie (--actor 2) before running this mode.
			owners := map[string]bool{}
			selected := []sdk.CreditInput{}
			for _, credit := range eligible {
				owner := core.ElementToBigInt(credit.State.Note.OwnerWalletIdentity).String()
				if owners[owner] {
					continue
				}
				owners[owner] = true
				selected = append(selected, credit)
				if len(selected) == 2 {
					break
				}
			}
			if len(selected) != 2 {
				log.Fatal("batch draw requires two authenticated unspent grants from different owners, such as Bob and Charlie")
			}
			eligible = selected
			draws := []sdk.CreditDrawInput{}
			for _, credit := range eligible {
				draws = append(draws, sdk.CreditDrawInput{Credit: credit.Commitment, Amount: new(big.Int).Set(amount)})
			}
			prepared, err = invisi.PrepareDrawCredit(ctx, sdk.ExplicitDrawCreditPreparation{TransactionIdentity: identity, CreditInputs: eligible, DrawCreditPreparation: sdk.DrawCreditPreparation{TokenAddress: token, Credits: draws}})
		case "top-up":
			selected := eligible[0]
			notes, err := shared.DiscoverNotes(ctx, provider, client, pool, actorAccount.PublicWallet, actorAccount.ViewingKey, token)
			if err != nil {
				log.Fatal(err)
			}
			inputs := []sdk.NoteInput{}
			selectedAmount := new(big.Int)
			for _, candidate := range notes {
				if selectedAmount.Cmp(amount) >= 0 {
					break
				}
				inputs = append(inputs, candidate.Input)
				selectedAmount.Add(selectedAmount, core.ElementToBigInt(candidate.Input.Note.Amount))
			}
			if len(inputs) > 10 {
				log.Fatal("Credit top-up supports at most 10 input notes")
			}
			expiry := core.ElementToBigInt(selected.State.Note.Expiry)
			expiry.Add(expiry, big.NewInt(86400))
			prepared, err = invisi.PrepareTopUpCredit(ctx, sdk.ExplicitTopUpCreditPreparation{TransactionIdentity: identity, ExplicitFunds: sdk.ExplicitFunds{Inputs: inputs}, CreditInputs: []sdk.CreditInput{selected}, TopUpCreditPreparation: sdk.TopUpCreditPreparation{TokenAddress: token, Credit: selected.Commitment, AdditionalAmount: amount, Expiry: expiry}})
			// Payload and policy nil preserve the existing raw application bytes and per-grant limits.
		case "cancel":
			selected := eligible[0]
			prepared, err = invisi.PrepareCancelCredit(ctx, sdk.ExplicitCancelCreditPreparation{TransactionIdentity: identity, CreditInputs: []sdk.CreditInput{selected}, CancelCreditPreparation: sdk.CancelCreditPreparation{TokenAddress: token, Credit: selected.Commitment}})
		default:
			log.Fatal("unknown Credit mode")
		}
		if err != nil {
			log.Fatal(err)
		}
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
		signature, err := crypto.Sign(digest, actorAccount.Key)
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
