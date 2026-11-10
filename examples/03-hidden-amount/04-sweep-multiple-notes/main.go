// Withdraw every selected Note to its original owner without private spending authorization.
package main

import (
	"context"
	"flag"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// Configure public actors and deployment.
	ctx := context.Background()
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC URL")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	address := flag.String("wallet-address", "", "public long receiving address")

	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
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

	// Discover automatically through canonical receipts and externally decrypted, independently checked openings.
	candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, publicWallet, decryptor, token)
	if err != nil {
		log.Fatal(err)
	}
	var inputs []sdk.NoteInput
	for _, candidate := range candidates {
		inputs = append(inputs, candidate.Input)
		if len(inputs) == 12 {
			break
		}
	}
	if len(inputs) == 0 {
		log.Fatal("no authenticated unspent notes; run a funding example first")
	}

	amount := new(big.Int)
	for _, input := range inputs {
		amount.Add(amount, core.ElementToBigInt(input.Note.Amount))
	}
	withdrawal := &sdk.Withdrawal{To: common.BigToAddress(core.ElementToBigInt(inputs[0].Note.OwnerAddress)), Amount: amount}

	// A full owner-bound sweep needs no spending signature; funds cannot be redirected to the gas payer.
	prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: publicWallet, Protocol: pool},
		ExplicitFunds:       sdk.ExplicitFunds{Inputs: inputs, ChangeAddress: *address},
		TransferPreparation: sdk.TransferPreparation{TokenAddress: token, Withdrawal: withdrawal},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Generate the local proof and independently verify it before exposing calldata.
	prepared, err = prepared.Prove(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Export unsigned calldata for the external signer; submission and canonical reconciliation remain explicit.
	call, err := prepared.ExportTransaction(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", call)
}
