// Resume a Credit journal without reconstructing its grant transition or spending reserved inputs twice.
package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"

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
	// Configure public actors and deployment; the native host retains account access.
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
	encoder := json.NewEncoder(os.Stdout)
	decoder := json.NewDecoder(os.Stdin)

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

	// Send the complete request array to the native host, then read its explicit response array.
	if err = encoder.Encode(pending.WalletRequests); err != nil {
		log.Fatal(err)
	}

	var refreshResponses []wallet.Response
	if err = decoder.Decode(&refreshResponses); err != nil {
		log.Fatal(err)
	}

	// Apply only after every reply passes binding and financial validation; failures leave the store unchanged.
	if _, err = pending.ApplyWalletResponses(ctx, refreshResponses); err != nil {
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
		if err = operation.Reconcile(ctx, managed.CreditReconciler(1)); err != nil {
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
			request, err := operation.ExportWalletRequest(ctx, "recover-credit-1")
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
			ordered, err := wallet.ValidateResponses(requests, responses)
			if err != nil {
				log.Fatal(err)
			}
			if err = operation.WithWalletResponse(ctx, request, ordered[0]); err != nil {
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
	if err = encoder.Encode(call); err != nil {
		log.Fatal(err)
	}
}
