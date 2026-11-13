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

func main() {
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient receiving address")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
	feeText := flag.String("max-fee", "1", "maximum Broadcaster fee in token base units")
	finalityTimeout := flag.Duration("finality-timeout", 2*time.Minute, "independent native finality deadline")
	execute := flag.Bool("execute", false, "discover, prepare and prove")
	submit := flag.Bool("submit", false, "explicitly send the payment")
	flag.Parse()
	if !*execute {
		log.Print("Pass --execute; add --submit to send payment after proving")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	provider, err := ethclient.DialContext(ctx, *chainRPC)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	rpc, err := sdk.NewProtocolRpcClient(*protocolRPC)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, rpc, provider)
	if err != nil {
		log.Fatal(err)
	}
	publicWallet, err := sdk.WalletFromAddress(*address)
	if err != nil {
		log.Fatal(err)
	}
	backend, err := native.NewNativeProver(native.NativeOptions{
		BBBinary:   os.Getenv("BB_BINARY"),
		NoirBinary: os.Getenv("NOIR_EXECUTOR"),
	})
	if err != nil {
		log.Fatal(err)
	}
	source, err := shared.SourceCapability(pool)
	if err != nil {
		log.Fatal(err)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{
		Provider: provider, ProtocolRPC: rpc, Prover: backend, Provenance: source,
	})
	managed, err := walletsession.Connect(ctx, walletsession.Options{
		Invisi: invisi, Wallet: publicWallet, Protocol: pool,
		Store: walletsession.NewMemoryWalletStore(),
	})
	if err != nil {
		log.Fatal(err)
	}
	pending, err := managed.Refresh(ctx, walletsession.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock})
	if err != nil {
		log.Fatal(err)
	}
	_, err = pending.Decrypt(ctx, shared.AppDeliveryDecryptor{})
	if err != nil {
		log.Fatal(err)
	}
	token := common.HexToAddress(*tokenText)
	balance, err := managed.Balance(ctx, token)
	if err != nil {
		log.Fatal(err)
	}
	maxFee, ok := new(big.Int).SetString(*feeText, 10)
	if !ok || maxFee.Sign() < 0 {
		log.Fatal("--max-fee must be a nonnegative integer")
	}
	if balance.AvailableKnown == nil || balance.AvailableKnown.Cmp(big.NewInt(2)) <= 0 {
		log.Fatal("fund and recognize Notes first; leave funds for fees")
	}
	output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
		ChainID:         new(big.Int).SetUint64(pool.Context().ChainID),
		ProtocolAddress: pool.Context().ProtocolAddress, TokenAddress: token,
		Amount:    new(big.Int).Div(balance.AvailableKnown, big.NewInt(2)),
		Recipient: sdk.InvisiRecipient{Address: *recipient},
	})
	if err != nil {
		log.Fatal(err)
	}
	prepared, err := managed.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}})
	if err != nil {
		log.Fatal(err)
	}
	operators, err := pool.GetBroadcasters(ctx, sdk.BroadcasterOptions{TokenAddress: token})
	if err != nil {
		log.Fatal(err)
	}
	if len(operators) == 0 {
		log.Fatal("no registered Broadcaster supports this token")
	}
	relay := operators[0]
	err = relay.Quote(ctx, prepared, sdk.BroadcasterQuoteOptions{MaxFeeAmount: maxFee})
	if err != nil {
		log.Fatal(err)
	}
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
	ready, err := prepared.Prove(ctx)
	if err != nil {
		log.Fatal(err)
	}
	quote, err := ready.ExportBroadcasterQuote(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Verified quote: %+v", quote)
	if !*submit {
		log.Print("Ready; no payment sent without --submit")
		return
	}
	submitted, err := relay.Submit(ctx, ready)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Broadcaster result: %+v", submitted)
	// Inclusion and finality are independent; this reads the configured Chain RPC.
	if submitted.TransactionHash != "" {
		finality, err := shared.WaitNativeFinality(context.Background(), provider, common.HexToHash(submitted.TransactionHash), *finalityTimeout)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Native confirmation: %+v", finality)
	}
}
