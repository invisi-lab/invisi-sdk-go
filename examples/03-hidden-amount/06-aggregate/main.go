// Authorize four independent transfers and submit their aggregate.
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
	"github.com/invisi-lab/invisi-sdk-go/protocol"
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
	signer, err := bind.NewKeyedTransactorWithChainID(alice.Key, chain)
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
	if len(accounts) < 3 {
		log.Fatal("aggregate example requires Alice, Bob and Charlie")
	}
	head, err := provider.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	deadline := head.Time + 3600
	token := common.HexToAddress(shared.Settings.TokenAddress)
	members := []prover.AggregateMember{}
	requests := []protocol.TransactRequest{}

	// 4. Bob and Charlie each sign and prove two distinct transfers.
	// Each proof has one owner; the aggregate combines four independent members.
	for _, owner := range accounts[1:3] {
		candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, owner.Wallet, owner.ViewingKey, token)
		if err != nil {
			log.Fatal(err)
		}
		selected := []sdk.NoteInput{}
		for _, candidate := range candidates {
			if core.ElementToBigInt(candidate.Input.Note.Amount).Cmp(big.NewInt(2)) >= 0 {
				selected = append(selected, candidate.Input)
			}
			if len(selected) == 2 {
				break
			}
		}
		if len(selected) != 2 {
			log.Fatal(owner.Name, " needs two distinct authenticated unspent notes of at least 2 units")
		}
		destination, err := alice.Wallet.GetAddress()
		if err != nil {
			log.Fatal(err)
		}
		for _, input := range selected {
			amount := new(big.Int).Div(core.ElementToBigInt(input.Note.Amount), big.NewInt(2))
			output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
				ChainID:         chain,
				ProtocolAddress: pool.Context().ProtocolAddress,
				TokenAddress:    token,
				Amount:          amount,
				Recipient:       sdk.InvisiRecipient{Address: destination},
			})
			if err != nil {
				log.Fatal(err)
			}
			prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
				TransactionIdentity: sdk.TransactionIdentity{Wallet: owner.Wallet, Protocol: pool},
				ExplicitFunds:       sdk.ExplicitFunds{Inputs: []sdk.NoteInput{input}},
				TransferPreparation: sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}, Deadline: deadline},
			})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(owner.Name, " review:", prepared.Preview())
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
			signature, err := crypto.Sign(digest, owner.Key)
			if err != nil {
				log.Fatal(err)
			}
			authorized, err := prepared.WithWalletSignature(ctx, signature)
			if err != nil {
				log.Fatal(err)
			}
			ready, err := authorized.Prove(ctx, sdk.TransactionProveOptions{Options: prover.Options{OnProgress: func(p prover.Progress) {
				fmt.Println("Member proof phase:", p.Phase)
			}}, Execution: "aggregate"})
			if err != nil {
				log.Fatal(err)
			}
			member, err := ready.ExportAggregateMember(ctx)
			if err != nil {
				log.Fatal(err)
			}
			members = append(members, prover.AggregateMember{Bundle: member.Bundle, Statement: member.Statement})
			requests = append(requests, member.Request)
		}
	}
	aggregator, err := prover.NewAggregator(localProver)
	if err != nil {
		log.Fatal(err)
	}
	bundle, err := aggregator.Aggregate(ctx, members, prover.Options{OnProgress: func(p prover.Progress) {
		fmt.Println("Aggregate phase:", p.Phase)
	}})
	if err != nil {
		log.Fatal(err)
	}
	call, err := pool.Aggregate(protocol.AggregateRequest{Proof: bundle.Proof, Requests: requests})
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
