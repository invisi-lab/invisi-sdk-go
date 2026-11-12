// Resume a Credit journal without reconstructing its grant transition or spending reserved inputs twice.
package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/invisi-lab/invisi-sdk-go/examples/shared"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsession "github.com/invisi-lab/invisi-sdk-go/session"
	walletsqlite "github.com/invisi-lab/invisi-sdk-go/session/sqlite"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func main() {
	// Configure public actors and deployment.
	ctx := context.Background()
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC URL")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	address := flag.String("wallet-address", "", "public long receiving address")
	operationID := flag.String("operation-id", "", "durable operation to resume")
	txHash := flag.String("transaction-hash", "", "external transaction hash to reconcile instead of exporting")
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

	// The storage encryption key protects inventory and journals; it is independent of wallet keys.
	database, err := sql.Open("sqlite", "examples/wallet.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	storageKeyText := os.Getenv("INVISI_EXAMPLE_STORAGE_KEY")
	storageKey, err := wallet.Bytes(storageKeyText, 32)
	if err != nil {
		log.Fatal("application storage key: ", err)
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
		Invisi:    invisi,
		Wallet:    publicWallet,
		Protocol:  pool,
		Store:     store,
		Namespace: "examples",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Stage candidate discovery without changing the persisted inventory or checkpoints.
	pending, err := managed.Refresh(ctx, walletsession.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock})
	if err != nil {
		log.Fatal(err)
	}

	if _, err = pending.Decrypt(ctx, shared.AppDeliveryDecryptor{}); err != nil {
		log.Fatal(err)
	}

	// Credit has its own journal; managed.Resume restores Note operations instead.
	creditSession, err := managed.Credit()
	if err != nil {
		log.Fatal(err)
	}
	history, err := creditSession.GetHistory(ctx)
	if err != nil {
		log.Fatal(err)
	}
	status := ""
	externalized := false
	for _, saved := range history.Operations {
		if saved.ID == *operationID {
			status, externalized = saved.Status, saved.Externalized
		}
	}
	operation, err := creditSession.Resume(ctx, *operationID, backend)
	if err != nil {
		log.Fatal(err)
	}

	// Canonical reconciliation releases reservations only after checking the exact call and receipt.
	if *txHash != "" {
		hashBytes, decodeErr := hex.DecodeString(strings.TrimPrefix(*txHash, "0x"))
		if decodeErr != nil || len(hashBytes) != 32 {
			log.Fatal("transaction-hash must be a 32-byte hex hash")
		}
		if err = operation.Externalize(ctx, common.HexToHash(*txHash)); err != nil {
			log.Fatal(err)
		}
		if err = operation.Reconcile(ctx, managed.CreditReconciler("latest")); err != nil {
			log.Fatal(err)
		}
		return
	}
	if externalized {
		log.Fatal("operation already externalized; supply transaction-hash for reconciliation")
	}

	// Authorization witnesses are intentionally absent from storage, so unproved transitions need host approval again.
	if status == "prepared" || status == "authorized" {
		preview, err := operation.Preview(ctx)
		if err != nil {
			log.Fatal(err)
		}
		// Direct-owner revocation has no wallet-spending authorization or proof.
		if preview.Kind != "revoke-credit" {
			authorization, err := operation.ExportSigningRequest(ctx)
			if err != nil {
				log.Fatal(err)
			}
			signature, err := shared.Sign(authorization, publicWallet.OwnerAddress(), "OWNER_PRIVATE_KEY")
			if err != nil {
				log.Fatal(err)
			}
			if err = operation.WithWalletSignature(ctx, signature); err != nil {
				log.Fatal(err)
			}
			if err = operation.Prove(ctx, backend, prover.Options{}); err != nil {
				log.Fatal(err)
			}
		}
	}

	// The frozen transition retains its reservation when calldata leaves this process.
	call, err := operation.Calldata(ctx, publicWallet.OwnerAddress())
	if err != nil {
		log.Fatal(err)
	}
	if err = operation.Externalize(ctx, common.Hash{}); err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", call)
}
