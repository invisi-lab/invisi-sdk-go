// Schedule a Note transfer, resume its encrypted operation, then cancel and reconcile its inputs.
package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"flag"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsession "github.com/invisi-lab/invisi-sdk-go/session"
	walletsqlite "github.com/invisi-lab/invisi-sdk-go/session/sqlite"
	_ "modernc.org/sqlite"
)

func main() {
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient public long receiving address")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
	amountText := flag.String("amount", "10", "Note payment amount in token base units")
	feeText := flag.String("max-fee", "1", "maximum Broadcaster fee in token base units")
	delay := flag.Duration("delay", 10*time.Minute, "publication delay")
	flag.Parse()
	if !common.IsHexAddress(*tokenText) || *delay < time.Minute {
		log.Fatal("configure --token and --delay of at least one minute")
	}
	amount, ok := new(big.Int).SetString(*amountText, 10)
	if !ok || amount.Sign() <= 0 {
		log.Fatal("--amount must be a positive integer")
	}
	maxFee, ok := new(big.Int).SetString(*feeText, 10)
	if !ok || maxFee.Sign() <= 0 {
		log.Fatal("--max-fee must be a positive integer")
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
		Provider: provider, ProtocolRPC: rpc,
		Prover: backend, Provenance: source,
	})
	database, err := sql.Open("sqlite", "examples/wallet.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	storageKey, err := hex.DecodeString(strings.TrimPrefix(os.Getenv("INVISI_EXAMPLE_STORAGE_KEY"), "0x"))
	if err != nil {
		log.Fatal(err)
	}
	defer clear(storageKey)
	codec, err := walletsession.NewEncryptedWalletCodec(storageKey)
	if err != nil {
		log.Fatal(err)
	}
	store, err := walletsqlite.NewWalletStore(ctx, database, codec)
	if err != nil {
		log.Fatal(err)
	}
	managed, err := walletsession.Connect(ctx, walletsession.Options{
		Invisi: invisi, Wallet: publicWallet, Protocol: pool, Store: store,
		Namespace: "scheduled-broadcaster-example",
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
	if balance.AvailableKnown == nil || balance.AvailableKnown.Cmp(new(big.Int).Add(amount, maxFee)) < 0 {
		log.Fatal("fund and recognize Notes first; leave enough balance for the fee")
	}
	operators, err := pool.GetBroadcasters(ctx, sdk.BroadcasterOptions{TokenAddress: token})
	if err != nil {
		log.Fatal(err)
	}
	trust := broadcaster.Trust{ChainID: new(big.Int).SetUint64(pool.Context().ChainID), Protocol: pool.Context().ProtocolAddress}
	var relay *broadcaster.Client
	var capabilities broadcaster.SchedulingCapabilities
	for _, candidate := range operators {
		status, err := candidate.ServiceStatus(ctx, trust)
		if err == nil && status.Scheduling.Enabled && !status.SubmissionHalted {
			relay = candidate
			capabilities = status.Scheduling
			break
		}
	}
	if relay == nil {
		log.Fatal("no registered Broadcaster supports scheduling for this token")
	}
	window, err := strconv.ParseUint(capabilities.DefaultExecutionWindowSeconds, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	slack, err := strconv.ParseUint(capabilities.MinDeadlineSlackSeconds, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	after := uint64(time.Now().Add(*delay).Unix())
	before := after + window
	deadline := before + slack + 60
	output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
		ChainID: trust.ChainID, ProtocolAddress: trust.Protocol,
		TokenAddress: token, Amount: amount,
		Recipient: sdk.InvisiRecipient{Address: *recipient},
	})
	if err != nil {
		log.Fatal(err)
	}
	prepared, err := managed.PrepareTransfer(ctx, sdk.TransferPreparation{
		TokenAddress: token, Outputs: []sdk.NoteOutput{output}, Deadline: deadline,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = relay.Quote(ctx, prepared, sdk.BroadcasterQuoteOptions{
		MaxFeeAmount: maxFee,
		Schedule: &sdk.ScheduleOptions{
			PublishAfter:  strconv.FormatUint(after, 10),
			ExecuteBefore: strconv.FormatUint(before, 10),
		},
	})
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
	log.Printf("Verified scheduled quote: fee=%s publishAfter=%s executeBefore=%s", quote.Quote.FeeAmount, quote.Quote.Schedule.PublishAfter, quote.Quote.Schedule.ExecuteBefore)
	log.Printf("Encrypted operation=%s", ready.ID())
	operationID := ready.ID()
	submitted, err := relay.Submit(ctx, ready)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Submission: id=%s state=%s", submitted.ID, submitted.State)

	// Resume uses the local operation ID; remote status uses the ID returned by the Broadcaster.
	recovered, err := managed.Resume(ctx, operationID)
	if err != nil {
		log.Fatal(err)
	}
	status, err := relay.Status(ctx, submitted.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Recovered relay status: id=%s state=%s", status.ID, status.State)
	// Managed cancellation also refreshes canonical consumption before releasing input reservations.
	cancelled, err := relay.Cancel(ctx, recovered)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Cancellation: id=%s state=%s; canonical input state reconciled", cancelled.ID, cancelled.State)
}
