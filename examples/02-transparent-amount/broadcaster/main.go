// Quote a Shield payment, sign its final public funding, and delegate submission.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	execute := flag.Bool("execute", false, "authorize approval, funding signature and broadcaster submission")
	maxFeeText := flag.String("max-fee", "1000", "maximum broadcaster fee in token base units")
	flag.Parse()
	if !*execute {
		log.Fatal("Review the code and settings, then add --execute; no network request was made")
	}
	maxFee, ok := new(big.Int).SetString(*maxFeeText, 10)
	if !ok || maxFee.Sign() <= 0 {
		log.Fatal("max-fee must be a positive integer")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	alice, bob := accounts[0], accounts[1]

	// 1. Discover and validate one Protocol. The SDK owns no Alice signing key.
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
		log.Fatal("chain mismatch")
	}
	rpc, err := sdk.NewProtocolRpcClient(shared.Settings.ProtocolRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, rpc, provider)
	if err != nil {
		log.Fatal(err)
	}
	tokenAddress := common.HexToAddress(shared.Settings.TokenAddress)
	bobAddress, err := bob.PublicWallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}
	output, err := sdk.NewNoteOutput(sdk.NoteOutputOptions{
		ChainID: chain, ProtocolAddress: pool.Context().ProtocolAddress,
		TokenAddress: tokenAddress, Amount: big.NewInt(10),
		Recipient: sdk.InvisiRecipient{Address: bobAddress},
	})
	if err != nil {
		log.Fatal(err)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: rpc})

	// 2. Prepare the business intent. The fee is not known yet.
	prepared, err := invisi.PrepareShield(ctx, sdk.ExplicitShieldPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: alice.PublicWallet, Protocol: pool},
		ShieldPreparation: sdk.ShieldPreparation{
			TokenAddress: tokenAddress, Outputs: []sdk.NoteOutput{output},
			Funding: sdk.FundingSource{Payer: alice.PublicWallet.OwnerAddress(), Amount: big.NewInt(10), Authorization: "signature"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 3. Obtain registered services that currently support this token.
	clients, err := pool.GetBroadcasters(ctx, sdk.BroadcasterOptions{TokenAddress: tokenAddress})
	if err != nil {
		log.Fatal(err)
	}
	if len(clients) == 0 {
		log.Fatal("no compatible broadcaster")
	}
	// Applications may request quotes from several clients and choose the cheapest.
	client := clients[0]
	quoted, err := prepared.WithBroadcaster(ctx, client, sdk.WithBroadcasterOptions{MaxFeeAmount: maxFee})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Broadcaster fee (base units):", quoted.Preview().FeeAmount)
	// The same attachment flow supports Transfer, Credit creation, top-up and cancellation.
	// Credit draw and registration do not support broadcasters. Private change can pay
	// fees; otherwise existing public funding increases. Cancellation uses its refund.

	// 4. Alice approves the final gross funding amount, including the fixed fee.
	// Approval and the EIP-712 funding authorization are separate permissions.
	signer, err := bind.NewKeyedTransactorWithChainID(alice.Key, chain)
	if err != nil {
		log.Fatal(err)
	}
	signer.Context = ctx
	token, err := contracts.NewIERC20(tokenAddress, provider)
	if err != nil {
		log.Fatal(err)
	}
	gross := quoted.Preview().FundingAmount
	allowance, err := token.Allowance(&bind.CallOpts{Context: ctx}, signer.From, pool.Context().ProtocolAddress)
	if err != nil {
		log.Fatal(err)
	}
	if allowance.Cmp(gross) < 0 {
		if allowance.Sign() != 0 {
			reset, err := token.Approve(signer, pool.Context().ProtocolAddress, new(big.Int))
			if err != nil {
				log.Fatal(err)
			}
			receipt, err := bind.WaitMined(ctx, provider, reset)
			if err != nil {
				log.Fatal(err)
			}
			if receipt.Status != types.ReceiptStatusSuccessful {
				log.Fatal("allowance reset reverted")
			}
		}
		approval, err := token.Approve(signer, pool.Context().ProtocolAddress, gross)
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
	}

	// 5. Sign the final request externally. Shield does not need a ZK proof.
	// For private operations, owner authorization and proving precede funding signing.
	request, err := quoted.ExportFundingAuthorization(ctx)
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
	ready, err := quoted.WithFundingSignature(ctx, signature)
	if err != nil {
		log.Fatal(err)
	}

	// 6. Submit the bound transaction. Keep the ID even if a request times out:
	// query client.Status(ctx, id) rather than constructing another spend.
	fmt.Println("Keep this request ID before submitting:", ready.ID())
	status, err := ready.SubmitBroadcaster(ctx, client)
	if err != nil {
		log.Fatal("Preserve the prepared transaction ID and recover its status: ", err)
	}
	fmt.Println("Broadcaster request ID:", status.ID)
	fmt.Println("Submission state:", status.State)
	if status.TransactionHash != "" {
		fmt.Println("Transaction hash:", status.TransactionHash)
	}
}
