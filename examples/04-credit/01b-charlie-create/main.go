// Charlie creates an independent 400-unit finite Credit grant for Alice.
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
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
	spenderAddress := flag.String("spender", "", "spender public long receiving address")
	available := flag.Int64("available-credit", 400, "initial finite credit in token base units")
	unlimited := flag.Bool("unlimited", false, "unlimited mode requires available-credit zero")
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
	coreContext := primitives.ProtocolContext{
		ChainID:         primitives.ElementFromUint64(pool.Context().ChainID),
		ProtocolAddress: primitives.BigIntToElement(pool.Context().ProtocolAddress.Big()),
	}
	c := credit.Context{ProtocolContext: coreContext, TokenAddress: token, Timestamp: head.Time}
	deadline := new(big.Int).SetUint64(head.Time + 300)
	var prepared *credit.Prepared
	spenderWallet, err := sdk.WalletFromAddress(*spenderAddress)
	if err != nil {
		log.Fatal(err)
	}
	initial := big.NewInt(*available)
	if initial.Sign() < 0 {
		log.Fatal("available credit must be nonnegative")
	}

	contract, e := contracts.NewInvisiProtocolCaller(pool.Context().ProtocolAddress, provider)
	if e != nil {
		log.Fatal(e)
	}
	allocation, e := contract.CreditAllocation(&bind.CallOpts{Context: ctx, BlockNumber: head.Number}, publicWallet.OwnerAddress(), token)
	if e != nil {
		log.Fatal(e)
	}
	nonce, e := primitives.GenerateNonce()
	if e != nil {
		log.Fatal(e)
	}
	key, e := credit.GenerateSharedKey()
	if e != nil {
		log.Fatal(e)
	}

	// A grant is postpaid authority: finite available credit is quota, not deposited token principal.
	state := credit.State{
		AvailableCredit:       primitives.BigIntToElement(initial),
		Unlimited:             *unlimited,
		ProtocolContext:       coreContext,
		TokenAddress:          primitives.BigIntToElement(token.Big()),
		OwnerWalletIdentity:   primitives.BigIntToElement(publicWallet.OwnerAddress().Big()),
		SpenderWalletIdentity: primitives.BigIntToElement(spenderWallet.OwnerAddress().Big()),
		RevocationShard:       primitives.ElementFromUint64(allocation.Shard),
		RevocationSlot:        primitives.ElementFromUint64(uint64(allocation.Slot)),
		Nonce:                 nonce,
		Expiry:                primitives.ElementFromUint64(head.Time + 7*86400),
		SharedKey:             key,
		Policy: credit.Policy{
			SingleLimit:  primitives.ElementFromUint64(100),
			DailyLimit:   primitives.ElementFromUint64(1000),
			MonthlyLimit: primitives.ElementFromUint64(10000),
		},
	}
	snapshot := credit.DeliveryState{
		State:                state,
		OwnerReceivingKeys:   primitives.ReceivingKeys{ViewingPublicKey: publicWallet.ViewingPublicKey(), ClueKey: publicWallet.ClueKey()},
		SpenderReceivingKeys: primitives.ReceivingKeys{ViewingPublicKey: spenderWallet.ViewingPublicKey(), ClueKey: spenderWallet.ClueKey()},
	}
	ds, e := shared.CreditDeliveries(snapshot)
	if e != nil {
		log.Fatal(e)
	}
	commitment, e := state.Commitment()
	if e != nil {
		log.Fatal(e)
	}
	prepared, err = credit.PrepareCreate(c, credit.CreateParams{
		Token:            token,
		Owner:            publicWallet.OwnerAddress(),
		RevocationShard:  allocation.Shard,
		RevocationSlot:   allocation.Slot,
		CreditCommitment: common.Hash(commitment.Bytes()),
		Deadline:         deadline,
	}, ds, state)
	log.Printf("Retain original revocation reference: owner=%s token=%s shard=%d slot=%d", publicWallet.OwnerAddress(), token, allocation.Shard, allocation.Slot)

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
