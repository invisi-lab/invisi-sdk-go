// Bob discovers an active owner grant and extends expiry while preserving policy and usage.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
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
	token := common.HexToAddress(*tokenText)
	head, err := provider.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	coreContext := core.ProtocolContext{
		ChainID:         core.ElementFromUint64(pool.Context().ChainID),
		ProtocolAddress: core.BigIntToElement(pool.Context().ProtocolAddress.Big()),
	}
	c := credit.Context{ProtocolContext: coreContext, TokenAddress: token, Timestamp: head.Time}
	deadline := new(big.Int).SetUint64(head.Time + 300)
	var prepared *credit.Prepared
	var changedMode *bool
	increaseAmount, decreaseAmount := new(big.Int), new(big.Int)

	// Reuse the encrypted inventory populated above; a second ephemeral discovery would repeat host prompts.
	inventory, e := store.ReadSnapshot(ctx, managed.Scope())
	if e != nil {
		log.Fatal(e)
	}
	candidates, e := shared.CreditCandidates(ctx, provider, client, pool, inventory, token, credit.Owner)
	if e != nil {
		log.Fatal(e)
	}
	if len(candidates) == 0 {
		log.Fatal("no active owner grant")
	}
	selected := candidates[0]
	old := selected.Input.State
	policy := old.Policy
	expiry := core.ElementFromUint64(head.Time + 14*86400)

	// Only policy/expiry/quota change; accumulated usage and the original grant identity survive the update.
	change := credit.Change{
		Policy:                  policy,
		Expiry:                  expiry,
		Unlimited:               changedMode,
		IncreaseAvailableCredit: core.BigIntToElement(increaseAmount),
		DecreaseAvailableCredit: core.BigIntToElement(decreaseAmount),
	}
	next, e := old.Update(change, deadline.Uint64())
	if e != nil {
		log.Fatal(e)
	}
	snapshot := selected.Delivery
	snapshot.State = next.State
	ds, e := shared.CreditDeliveries(snapshot)
	if e != nil {
		log.Fatal(e)
	}
	prepared, err = credit.PrepareUpdate(c, credit.UpdateParams{
		Token:            token,
		Owner:            publicWallet.OwnerAddress(),
		Source:           protocol.InputSource{TreeId: new(big.Int).SetUint64(selected.Input.TreeID), MerkleRoot: selected.Membership.Root},
		Nullifier:        common.Hash(next.Nullifier.Bytes()),
		CreditCommitment: common.Hash(next.StateCommitment.Bytes()),
		RevocationShard:  core.ElementToBigInt(old.RevocationShard).Uint64(),
		RevocationSlot:   uint16(core.ElementToBigInt(old.RevocationSlot).Uint64()),
		Deadline:         deadline,
	}, ds, selected.Input, change)

	if err != nil {
		log.Fatal(err)
	}

	// Keep Credit inventory and cross-role grant reservations separate from spendable Notes.
	creditSession, err := managed.Credit()
	if err != nil {
		log.Fatal(err)
	}
	operation, err := creditSession.Prepare(ctx, *operationID, prepared)
	if err != nil {
		log.Fatal(err)
	}

	// Export the owner/spender request only after the managed Credit session validates current chain state.
	request, err := operation.ExportWalletRequest(ctx, "credit-1")
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
	if err = operation.WithWalletResponse(ctx, request, ordered[0]); err != nil {
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
	if err = encoder.Encode(call); err != nil {
		log.Fatal(err)
	}
}
