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
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	execute := flag.Bool("execute", false, "authorize signing, proving and chain submission")
	signatureFunding := flag.Bool("signature-funding", false, "sign a public funding grant")
	count := flag.Int("outputs", 2, "number of transparent funded outputs, from 1 to 10")
	actor := flag.Int("actor", 0, "funding/spending account index")
	recipientIndex := flag.Int("recipient", 1, "receiving account index; set equal to actor for own notes")
	flag.Parse()
	if !*execute {
		log.Fatal("Review settings and code, then add --execute; no network request or transaction was made")
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
	signer, err := bind.NewKeyedTransactorWithChainID(alice.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx
	recipient, err := bob.PublicWallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}

	if *count < 1 || *count > 10 {
		log.Fatal("outputs must be 1..10")
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client})
	payload, err := core.FundsV1.Encode(core.PayloadDataV1{Memo: []byte("Transparent funded payment")})
	if err != nil {
		log.Fatal(err)
	}
	outputs := make([]sdk.NoteOutput, *count)
	for i := range outputs {
		outputs[i], err = sdk.NewNoteOutput(sdk.NoteOutputOptions{
			ChainID: chain, ProtocolAddress: pool.Context().ProtocolAddress, TokenAddress: token,
			Amount: big.NewInt(10), Recipient: sdk.InvisiRecipient{Address: recipient}, Payload: payload,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	amount := big.NewInt(int64(*count) * 10)
	authorization := "caller"
	if *signatureFunding {
		authorization = "signature"
	}
	// Token approval is public and separate from private wallet authorization.
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
	prepared, err := invisi.PrepareShield(ctx, sdk.ExplicitShieldPreparation{TransactionIdentity: sdk.TransactionIdentity{Wallet: alice.PublicWallet, Protocol: pool}, ShieldPreparation: sdk.ShieldPreparation{TokenAddress: token, Outputs: outputs, Funding: sdk.FundingSource{Payer: alice.PublicWallet.OwnerAddress(), Amount: amount, Authorization: authorization}}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Review frozen shield:", prepared.Preview())
	if !*signatureFunding {
		prepared, err = prepared.Ready(ctx)
		if err != nil {
			log.Fatal(err)
		}
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
	// Persist the transaction hash before waiting in a production application, then reconcile canonical state.

}
