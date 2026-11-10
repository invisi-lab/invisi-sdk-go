// Withdraw 10 units to Bob and retain any excess from selected Notes as private change.
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
	// Select only the required private portion; any selected excess returns to this owner.
	var inputs []sdk.NoteInput
	selectedAmount := new(big.Int)
	for _, candidate := range candidates {
		if selectedAmount.Cmp(big.NewInt(10)) >= 0 {
			break
		}
		inputs = append(inputs, candidate.Input)
		selectedAmount.Add(selectedAmount, primitives.ElementToBigInt(candidate.Input.Note.Amount))
		if len(inputs) == 12 {
			break
		}
	}
	if len(inputs) == 0 {
		log.Fatal("no authenticated unspent notes; run a funding example first")
	}

	// Freeze recipients and amounts before authorization.
	prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: publicWallet, Protocol: pool},
		ExplicitFunds:       sdk.ExplicitFunds{Inputs: inputs, ChangeAddress: *address},
		TransferPreparation: sdk.TransferPreparation{
			TokenAddress: token,
			Withdrawal:   &sdk.Withdrawal{To: publicWallet.OwnerAddress(), Amount: big.NewInt(10)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Authorize the frozen statement with an ordinary geth EIP-712 signature.
	authorization, err := prepared.ExportWalletAuthorization(ctx)
	if err != nil {
		log.Fatal(err)
	}
	signature, err := shared.Sign(authorization, publicWallet.OwnerAddress(), "OWNER_PRIVATE_KEY")
	if err != nil {
		log.Fatal(err)
	}
	prepared, err = prepared.WithWalletSignature(ctx, signature)
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
