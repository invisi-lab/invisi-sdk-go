package main

import (
	"context"
	"flag"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsession "github.com/invisi-lab/invisi-sdk-go/session"
)

func must[T any](value T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return value
}
func main() {
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient receiving address")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
	execute := flag.Bool("execute", false, "discover, prepare and prove")
	submit := flag.Bool("submit", false, "explicitly send the payment")
	flag.Parse()
	if !*execute {
		log.Print("Pass --execute; add --submit to send payment after proving")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	provider := must(ethclient.DialContext(ctx, *chainRPC))
	defer provider.Close()
	rpc := must(sdk.NewProtocolRpcClient(*protocolRPC))
	pool := must(sdk.ProtocolFromRPC(ctx, rpc, provider))
	publicWallet := must(sdk.WalletFromAddress(*address))
	backend := must(native.NewNativeProver(native.NativeOptions{BBBinary: os.Getenv("BB_BINARY"), NoirBinary: os.Getenv("NOIR_EXECUTOR")}))
	source := must(shared.SourceCapability(pool))
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: rpc, Prover: backend, Provenance: source})
	managed := must(walletsession.Connect(ctx, walletsession.Options{Invisi: invisi, Wallet: publicWallet, Protocol: pool, Store: walletsession.NewMemoryWalletStore()}))
	pending := must(managed.Refresh(ctx, walletsession.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock}))
	_ = must(pending.Decrypt(ctx, shared.AppDeliveryDecryptor{}))
	token := common.HexToAddress(*tokenText)
	balance := must(managed.Balance(ctx, token))
	maxFee := big.NewInt(1)
	if balance.AvailableKnown == nil || balance.AvailableKnown.Cmp(big.NewInt(2)) <= 0 {
		log.Fatal("fund and recognize Notes first; leave funds for fees")
	}
	output := must(sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: new(big.Int).SetUint64(pool.Context().ChainID), ProtocolAddress: pool.Context().ProtocolAddress, TokenAddress: token, Amount: new(big.Int).Div(balance.AvailableKnown, big.NewInt(2)), Recipient: sdk.InvisiRecipient{Address: *recipient}}))
	prepared := must(managed.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}}))
	operators := must(pool.GetBroadcasters(ctx, sdk.BroadcasterOptions{TokenAddress: token}))
	if len(operators) == 0 {
		log.Fatal("no registered Broadcaster supports this token")
	}
	relay := operators[0]
	prepared = must(prepared.WithBroadcaster(ctx, relay, sdk.WithBroadcasterOptions{MaxFeeAmount: maxFee}))
	authorization := must(prepared.ExportWalletAuthorization(ctx))
	signature := must(shared.Sign(authorization, publicWallet.OwnerAddress(), "OWNER_PRIVATE_KEY"))
	prepared = must(prepared.WithWalletSignature(ctx, signature))
	ready := must(prepared.Prove(ctx))
	log.Printf("Verified quote: %+v", must(ready.ExportBroadcasterQuote(ctx)))
	if !*submit {
		log.Print("Ready; no payment sent without --submit")
		return
	}
	log.Printf("Broadcaster result: %+v", must(ready.SubmitBroadcaster(ctx, relay)))
}
