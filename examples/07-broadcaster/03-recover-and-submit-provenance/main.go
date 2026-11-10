package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsession "github.com/invisi-lab/invisi-sdk-go/session"
	walletsqlite "github.com/invisi-lab/invisi-sdk-go/session/sqlite"
)

func must[T any](value T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return value
}
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient receiving address")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
	execute := flag.Bool("execute", false, "run this phase")
	reprove := flag.Bool("reprove", false, "regenerate Source evidence with updated trusted roots/generation")
	resume := flag.Bool("resume", false, "resume the original saved operation")
	operationID := flag.String("operation-id", "", "original operation ID")
	transactionHash := flag.String("transaction-hash", "", "original hash if a crash prevented recording it")
	publicProofPath := flag.String("proof", "", "submit a previously exported public proof JSON; no key or journal needed")
	exportPath := flag.String("export-proof", "", "exclusive output path for the public proof JSON")
	send := flag.Bool("send", false, "explicitly broadcast the initial main payment")
	flag.Parse()
	if !*execute {
		log.Print("Use --execute --send for phase one, then --execute --resume --operation-id ID; or --execute --proof FILE")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	provider := must(ethclient.DialContext(ctx, *chainRPC))
	defer provider.Close()
	rpc := must(sdk.NewProtocolRpcClient(*protocolRPC))
	pool := must(sdk.ProtocolFromRPC(ctx, rpc, provider))
	if *publicProofPath != "" {
		source := must(shared.SourceClient(pool))
		var packet provenance.ProvenanceProof
		decoder := json.NewDecoder(strings.NewReader(string(must(os.ReadFile(*publicProofPath)))))
		decoder.DisallowUnknownFields()
		check(decoder.Decode(&packet))
		accepted := must(sdk.SubmitProvenanceProof(ctx, source, packet))
		log.Printf("Source submission: %+v", accepted)
		tracking := must(provenance.TrackingFromResult(accepted, ""))
		log.Printf("Source recognition: %+v", must(source.WaitTracking(ctx, tracking.ID, 2*time.Second, 2*time.Minute)))
		return
	}
	publicWallet := must(sdk.WalletFromAddress(*address))
	backend := must(native.NewNativeProver(native.NativeOptions{BBBinary: os.Getenv("BB_BINARY"), NoirBinary: os.Getenv("NOIR_EXECUTOR")}))
	source := must(shared.SourceCapability(pool))
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: rpc, Prover: backend, Provenance: source})
	database := must(sql.Open("sqlite", "examples/wallet.sqlite"))
	defer database.Close()
	storageKey := must(hex.DecodeString(strings.TrimPrefix(os.Getenv("INVISI_EXAMPLE_STORAGE_KEY"), "0x")))
	defer clear(storageKey)
	codec := must(walletsession.NewEncryptedWalletCodec(storageKey))
	store := must(walletsqlite.NewWalletStore(ctx, database, codec))
	var signer *bind.TransactOpts
	if !*resume {
		key := must(crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("OWNER_PRIVATE_KEY"), "0x")))
		if crypto.PubkeyToAddress(key.PublicKey) != publicWallet.OwnerAddress() {
			log.Fatal("OWNER_PRIVATE_KEY must match the Note owner")
		}
		signer = must(bind.NewKeyedTransactorWithChainID(key, new(big.Int).SetUint64(pool.Context().ChainID)))
	}
	managed := must(walletsession.Connect(ctx, walletsession.Options{Invisi: invisi, Wallet: publicWallet, Protocol: pool, Store: store, Namespace: "source-recovery-example", Signer: signer}))
	if !*resume {
		pending := must(managed.Refresh(ctx, walletsession.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock}))
		_ = must(pending.Decrypt(ctx, shared.AppDeliveryDecryptor{}))
		token := common.HexToAddress(*tokenText)
		balance := must(managed.Balance(ctx, token))
		if balance.AvailableKnown == nil || balance.AvailableKnown.Cmp(big.NewInt(2)) < 0 {
			log.Fatal("fund Notes first")
		}
		output := must(sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: new(big.Int).SetUint64(pool.Context().ChainID), ProtocolAddress: pool.Context().ProtocolAddress, TokenAddress: token, Amount: new(big.Int).Div(balance.AvailableKnown, big.NewInt(2)), Recipient: sdk.InvisiRecipient{Address: *recipient}}))
		prepared := must(managed.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}}))
		authorization := must(prepared.ExportWalletAuthorization(ctx))
		signature := must(shared.Sign(authorization, publicWallet.OwnerAddress(), "OWNER_PRIVATE_KEY"))
		prepared = must(prepared.WithWalletSignature(ctx, signature))
		ready := must(prepared.Prove(ctx))
		log.Printf("Saved operation ID: %s", ready.ID())
		if !*send {
			log.Print("Ready; no payment sent without --send. Retain this encrypted journal.")
			return
		}
		tx := must(managed.SendTransaction(ctx, ready))
		log.Printf("Main transaction: %s", tx.Hash())
		receipt := must(bind.WaitMined(ctx, provider, tx))
		if receipt.Status != 1 {
			log.Fatal("main transaction reverted; retain the original operation record")
		}
		log.Printf("Restart with --resume --operation-id %s", ready.ID())
		return
	}
	if *operationID == "" {
		log.Fatal("set the original --operation-id")
	}
	restored := must(managed.Resume(ctx, *operationID))
	if *transactionHash != "" {
		restored = must(restored.TrackTransaction(ctx, common.HexToHash(*transactionHash)))
	}
	confirmed := must(restored.Reconcile(ctx))
	if confirmed.Status() != "confirmed" {
		log.Fatal("original payment is not canonically confirmed; resume it later")
	}
	ready := must(confirmed.ProveProvenance(ctx, sdk.ProvenanceProveOptions{Reprove: *reprove}))
	packet := must(ready.ExportProvenanceProof(ctx))
	if *exportPath != "" {
		file := must(os.OpenFile(*exportPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600))
		check(json.NewEncoder(file).Encode(packet))
		check(file.Sync())
		check(file.Close())
		log.Printf("Saved public proof: %s", *exportPath)
	}
	accepted := must(ready.SubmitProvenance(ctx))
	tracking := must(provenance.TrackingFromResult(accepted, ""))
	if tracking.OperationDigest != packet.OperationDigest {
		log.Fatal("Source tracking digest mismatch")
	}
	log.Printf("Source recognition: %+v", must(source.Client.WaitTracking(ctx, tracking.ID, 2*time.Second, 2*time.Minute)))
}
