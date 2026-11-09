// Resume a durable operation; a saved plan without a proof obtains fresh wallet authorization.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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
	recipient := flag.String("recipient", "", "recipient public long receiving address")
	operationID := flag.String("operation-id", "", "durable operation to resume")
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
	_ = recipient
	_ = token

	// Recovery reuses the durable plan/proof, but never a persisted private wallet authorization witness.
	prepared, err := managed.Resume(ctx, *operationID)
	if err != nil {
		log.Fatal(err)
	}
	if prepared.Status() == "prepared" || prepared.Status() == "authorized" {

		// The host approves only this frozen spending statement; no wallet key is passed to the SDK.
		request, err := prepared.ExportWalletRequest(ctx, "recover-authorize-1")
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
