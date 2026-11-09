// Bob pays 10 units to Dave from discovered Notes and receives any excess as private change.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func main() {
	// Configure public actors and deployment; the native host retains account access.
	ctx := context.Background()
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC URL")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient public long receiving address")
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
	encoder := json.NewEncoder(os.Stdout)
	decoder := json.NewDecoder(os.Stdin)
	nativeWallet := shared.NativeWalletClient{Encoder: encoder, Decoder: decoder}
	output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
		ChainID:         new(big.Int).SetUint64(pool.Context().ChainID),
		ProtocolAddress: pool.Context().ProtocolAddress,
		TokenAddress:    token,
		Amount:          big.NewInt(10),
		Recipient:       sdk.InvisiRecipient{Address: *recipient},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Discover automatically through canonical receipts and externally decrypted, independently checked openings.
	candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, publicWallet, nativeWallet, token)
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
		selectedAmount.Add(selectedAmount, core.ElementToBigInt(candidate.Input.Note.Amount))
		if len(inputs) == 12 {
			break
		}
	}
	if len(inputs) == 0 {
		log.Fatal("no authenticated unspent notes; run a funding example first")
	}

	// Freeze recipients and amounts before authorization so the native response approves this exact intent.
	prepared, err := invisi.PrepareTransfer(ctx, sdk.ExplicitTransferPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: publicWallet, Protocol: pool},
		ExplicitFunds:       sdk.ExplicitFunds{Inputs: inputs, ChangeAddress: *address},
		TransferPreparation: sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}},
	})
	if err != nil {
		log.Fatal(err)
	}

	// The host approves only this frozen spending statement; no wallet key is passed to the SDK.
	request, err := prepared.ExportWalletRequest(ctx, "authorize-1")
	if err != nil {
		log.Fatal(err)
	}
	requests := []wallet.Request{request}
	if err = encoder.Encode(requests); err != nil {
		log.Fatal(err)
	}

	var responses []wallet.Response
	if err = decoder.Decode(&responses); err != nil {
		log.Fatal(err)
	}

	// Match by ID rather than arrival order; attachment independently verifies the signature and statement.
	ordered, err := wallet.ValidateResponses(requests, responses)
	if err != nil {
		log.Fatal(err)
	}
	prepared, err = prepared.WithWalletResponse(ctx, request, ordered[0])
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
	if err = encoder.Encode(call); err != nil {
		log.Fatal(err)
	}
}
