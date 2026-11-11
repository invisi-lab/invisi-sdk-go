// Bob and Charlie each authorize two half-note transfers; combine their four recursive proofs.
package main

import (
	"context"
	"flag"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// Configure public actors and deployment.
	ctx := context.Background()
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC URL")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient public long receiving address")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
	secondAddress := flag.String("second-wallet-address", "", "second owner public long receiving address")
	flag.Parse()

	publicWallet, err := sdk.WalletFromAddress(*address)
	if err != nil {
		log.Fatal(err)
	}

	// Authenticate the Protocol RPC deployment against Chain RPC before preparing financial intent.
	provider, err := ethclient.DialContext(ctx, *chainRPC)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	client, err := sdk.NewProtocolRpcClient(*protocolRPC)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, client, provider)
	if err != nil {
		log.Fatal(err)
	}

	// Proving runs locally; this backend never executes wallet account keys.
	backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: "bb", NoirBinary: "noir-execute"})
	if err != nil {
		log.Fatal(err)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client, Prover: backend})
	token := common.HexToAddress(*tokenText)
	decryptor := shared.AppDeliveryDecryptor{}
	secondWallet, err := sdk.WalletFromAddress(*secondAddress)
	if err != nil {
		log.Fatal(err)
	}
	if secondWallet.OwnerAddress() == publicWallet.OwnerAddress() {
		log.Fatal("two distinct public owners required")
	}
	members := []prover.AggregateMember{}
	transacts := []protocol.TransactRequest{}

	// Each owner signs separately; aggregates may contain proofs from different owners.
	for ownerIndex, ownerWallet := range []*sdk.Wallet{publicWallet, secondWallet} {
		ownerAddress := ownerWallet.OwnerAddress().Hex()

		// Discover automatically through canonical receipts and externally decrypted, independently checked openings.
		candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, ownerWallet, decryptor, token)
		if err != nil {
			log.Fatal(err)
		}
		var inputs []sdk.NoteInput
		for _, candidate := range candidates {
			if primitives.ElementToBigInt(candidate.Input.Note.Amount).Cmp(big.NewInt(2)) < 0 {
				continue
			}
			inputs = append(inputs, candidate.Input)
			if len(inputs) == 2 {
				break
			}
		}
		if len(inputs) != 2 {
			log.Fatal("two distinct selected unspent notes required")
		}
		preparedMembers := make([]*sdk.PreparedTransaction, 2)
		for i, input := range inputs {
			amount := new(big.Int).Div(primitives.ElementToBigInt(input.Note.Amount), big.NewInt(2))
			if amount.Sign() == 0 {
				log.Fatal("each note must contain at least two units")
			}
			output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
				ChainID:         new(big.Int).SetUint64(pool.Context().ChainID),
				ProtocolAddress: pool.Context().ProtocolAddress,
				TokenAddress:    token,
				Amount:          amount,
				Recipient:       sdk.InvisiRecipient{Address: *recipient},
			})
			if err != nil {
				log.Fatal(err)
			}

			// Freeze recipients and amounts before authorization.
			prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
				TransactionIdentity: sdk.TransactionIdentity{Wallet: ownerWallet, Protocol: pool},
				ExplicitFunds:       sdk.ExplicitFunds{Inputs: []sdk.NoteInput{input}, ChangeAddress: ownerAddress},
				TransferPreparation: sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}},
			})
			if err != nil {
				log.Fatal(err)
			}

			preparedMembers[i] = prepared
		}

		for _, prepared := range preparedMembers {
			authorization, err := prepared.ExportWalletAuthorization(ctx)
			if err != nil {
				log.Fatal(err)
			}
			environment := "OWNER_PRIVATE_KEY"
			if ownerIndex == 1 {
				environment = "SECOND_OWNER_PRIVATE_KEY"
			}
			signature, err := shared.Sign(authorization, ownerWallet.OwnerAddress(), environment)
			if err != nil {
				log.Fatal(err)
			}
			authorized, err := prepared.WithWalletSignature(ctx, signature)
			if err != nil {
				log.Fatal(err)
			}

			// Recursive proofs are verified separately before entering the aggregate.
			ready, err := authorized.Prove(ctx, sdk.TransactionProveOptions{Execution: "aggregate"})
			if err != nil {
				log.Fatal(err)
			}
			member, err := ready.ExportAggregateMember(ctx)
			if err != nil {
				log.Fatal(err)
			}
			members = append(members, prover.AggregateMember{Bundle: member.Bundle, Statement: member.Statement})
			transacts = append(transacts, member.Request)
		}
	}

	// Aggregate only independently verified recursive members; owners may differ between members.
	aggregator, err := prover.NewAggregator(backend)
	if err != nil {
		log.Fatal(err)
	}
	bundle, err := aggregator.Aggregate(ctx, members, prover.Options{})
	if err != nil {
		log.Fatal(err)
	}
	call, err := pool.Aggregate(protocol.AggregateRequest{Proof: bundle.Proof, Requests: transacts})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", call)
}
