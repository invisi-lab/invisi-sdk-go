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

func main() {
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient receiving address")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
	finalityTimeout := flag.Duration("finality-timeout", 2*time.Minute, "independent native finality deadline")
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
	if *publicProofPath != "" {
		source, err := shared.SourceClient(pool)
		if err != nil {
			log.Fatal(err)
		}
		var packet provenance.ProvenanceProof
		proofBytes, err := os.ReadFile(*publicProofPath)
		if err != nil {
			log.Fatal(err)
		}
		decoder := json.NewDecoder(strings.NewReader(string(proofBytes)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&packet); err != nil {
			log.Fatal(err)
		}
		accepted, err := sdk.SubmitProvenanceProof(ctx, source, packet)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Source submission: %+v", accepted)
		tracking, err := provenance.TrackingFromResult(accepted, "")
		if err != nil {
			log.Fatal(err)
		}
		recognition, err := source.WaitTracking(context.Background(), tracking.ID, 2*time.Second, 2*time.Minute)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Source recognition: %+v", recognition)
		return
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
	var signer *bind.TransactOpts
	if !*resume {
		key, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("OWNER_PRIVATE_KEY"), "0x"))
		if err != nil {
			log.Fatal(err)
		}
		if crypto.PubkeyToAddress(key.PublicKey) != publicWallet.OwnerAddress() {
			log.Fatal("OWNER_PRIVATE_KEY must match the Note owner")
		}
		signer, err = bind.NewKeyedTransactorWithChainID(key, new(big.Int).SetUint64(pool.Context().ChainID))
		if err != nil {
			log.Fatal(err)
		}
	}
	managed, err := walletsession.Connect(ctx, walletsession.Options{
		Invisi: invisi, Wallet: publicWallet, Protocol: pool, Store: store,
		Namespace: "source-recovery-example", Signer: signer,
	})
	if err != nil {
		log.Fatal(err)
	}
	if !*resume {
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
		if balance.AvailableKnown == nil || balance.AvailableKnown.Cmp(big.NewInt(2)) < 0 {
			log.Fatal("fund Notes first")
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
		log.Printf("Saved operation ID: %s", ready.ID())
		if !*send {
			log.Print("Ready; no payment sent without --send. Retain this encrypted journal.")
			return
		}
		tx, err := managed.SendTransaction(ctx, ready)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Main transaction: %s", tx.Hash())
		finality, err := shared.WaitNativeFinality(context.Background(), provider, tx.Hash(), *finalityTimeout)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Native finality: %+v", finality)
		log.Printf("Restart with --resume --operation-id %s", ready.ID())
		return
	}
	if *operationID == "" {
		log.Fatal("set the original --operation-id")
	}
	restored, err := managed.Resume(ctx, *operationID)
	if err != nil {
		log.Fatal(err)
	}
	if *transactionHash != "" {
		restored, err = restored.TrackTransaction(ctx, common.HexToHash(*transactionHash))
		if err != nil {
			log.Fatal(err)
		}
	}
	confirmed, err := restored.Reconcile(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if confirmed.Status() != "confirmed" {
		log.Fatal("original payment is not canonically confirmed; resume it later")
	}
	ready, err := confirmed.ProveProvenance(ctx, sdk.ProvenanceProveOptions{Reprove: *reprove})
	if err != nil {
		log.Fatal(err)
	}
	packet, err := ready.ExportProvenanceProof(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if *exportPath != "" {
		file, err := os.OpenFile(*exportPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			log.Fatal(err)
		}
		if err := json.NewEncoder(file).Encode(packet); err != nil {
			log.Fatal(err)
		}
		if err := file.Sync(); err != nil {
			log.Fatal(err)
		}
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("Saved public proof: %s", *exportPath)
	}
	accepted, err := ready.SubmitProvenance(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tracking, err := provenance.TrackingFromResult(accepted, "")
	if err != nil {
		log.Fatal(err)
	}
	if tracking.OperationDigest != packet.OperationDigest {
		log.Fatal("Source tracking digest mismatch")
	}
	recognition, err := source.Client.WaitTracking(context.Background(), tracking.ID, 2*time.Second, 2*time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Source recognition: %+v", recognition)
}
