// Bob discovers an active owner grant and extends expiry while preserving policy and usage.
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
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

	operationID := flag.String("operation-id", "credit-example-1", "unique durable operation ID")

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

	token := common.HexToAddress(*tokenText)
	head, err := provider.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	inventory, err := store.ReadSnapshot(ctx, managed.Scope())
	if err != nil {
		log.Fatal(err)
	}
	var selected common.Hash
	for _, g := range inventory.PostpaidCredit.Grants {
		if g.Role == credit.Owner && !g.Consumed && !g.Revoked && g.ReservedBy == "" && common.BigToAddress(primitives.ElementToBigInt(g.Delivery.State.TokenAddress)) == token {
			selected = g.Commitment
			break
		}
	}
	if selected == (common.Hash{}) {
		log.Fatal("no active owner grant")
	}
	expiry := primitives.ElementFromUint64(head.Time + 14*86400)
	creditSession, err := managed.Credit()
	if err != nil {
		log.Fatal(err)
	}
	operation, err := creditSession.PrepareUpdate(ctx, *operationID, credit.UpdatePreparation{TokenAddress: token, Credit: selected, UpdatePatch: credit.UpdatePatch{Expiry: &expiry}})
	if err != nil {
		log.Fatal(err)
	}

	// Export the owner/spender request only after the managed Credit session validates current chain state.
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

	// Generate and independently verify the Credit proof for the frozen grant transition.
	if err = operation.Prove(ctx, backend, prover.Options{}); err != nil {
		log.Fatal(err)
	}

	// Export the unsigned call while retaining the managed grant reservation for reconciliation.
	call, err := operation.Calldata(ctx, publicWallet.OwnerAddress())
	if err != nil {
		log.Fatal(err)
	}
	// Mark the call as externally visible before export so an uncertain submission keeps its reservation.
	if err = operation.Externalize(ctx, common.Hash{}); err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", call)
}
