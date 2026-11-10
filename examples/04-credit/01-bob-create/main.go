// Bob creates a 400-unit finite Credit grant for Alice without depositing Note principal.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
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
	spenderAddress := flag.String("spender", "", "spender public long receiving address")
	available := flag.Int64("available-credit", 400, "initial finite credit in token base units")
	unlimited := flag.Bool("unlimited", false, "unlimited mode requires available-credit zero")
	relayURL := flag.String("broadcaster", "", "optional HTTPS registered Broadcaster endpoint")
	relaySigner := flag.String("quote-signer", "", "registered quote signer")
	relayRecipient := flag.String("fee-recipient", "", "registered fee recipient")
	relayPolicy := flag.String("provenance-policy", "", "optional authenticated required policy ID")
	relayMaxFee := flag.String("max-fee", "1", "maximum fee in token base units")
	execute := flag.Bool("execute", false, "run the optional relay workflow")
	submit := flag.Bool("submit", false, "explicitly submit the quoted request")
	flag.Parse()
	if *relayURL != "" && !*execute {
		log.Print("add --execute to prepare/quote/prove; --submit broadcasts")
		return
	}
	if *submit && *relayURL == "" {
		log.Fatal("--submit requires --broadcaster")
	}

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
	nonce, e := core.GenerateNonce()
	if e != nil {
		log.Fatal(e)
	}
	key, e := credit.GenerateSharedKey()
	if e != nil {
		log.Fatal(e)
	}

	// A grant is postpaid authority: finite available credit is quota, not deposited token principal.
	state := credit.State{
		AvailableCredit:       core.BigIntToElement(initial),
		Unlimited:             *unlimited,
		ProtocolContext:       coreContext,
		TokenAddress:          core.BigIntToElement(token.Big()),
		OwnerWalletIdentity:   core.BigIntToElement(publicWallet.OwnerAddress().Big()),
		SpenderWalletIdentity: core.BigIntToElement(spenderWallet.OwnerAddress().Big()),
		RevocationShard:       core.ElementFromUint64(allocation.Shard),
		RevocationSlot:        core.ElementFromUint64(uint64(allocation.Slot)),
		Nonce:                 nonce,
		Expiry:                core.ElementFromUint64(head.Time + 7*86400),
		SharedKey:             key,
		Policy: credit.Policy{
			SingleLimit:  core.ElementFromUint64(100),
			DailyLimit:   core.ElementFromUint64(1000),
			MonthlyLimit: core.ElementFromUint64(10000),
		},
	}
	snapshot := credit.DeliveryState{
		State:                state,
		OwnerReceivingKeys:   core.ReceivingKeys{ViewingPublicKey: publicWallet.ViewingPublicKey(), ClueKey: publicWallet.ClueKey()},
		SpenderReceivingKeys: core.ReceivingKeys{ViewingPublicKey: spenderWallet.ViewingPublicKey(), ClueKey: spenderWallet.ClueKey()},
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

	var relay *broadcaster.Client
	if *relayURL != "" {
		relay, err = broadcaster.NewClient(broadcaster.Configuration{Endpoint: *relayURL, QuoteSigner: common.HexToAddress(*relaySigner), FeeRecipient: common.HexToAddress(*relayRecipient)})
		if err != nil {
			log.Fatal(err)
		}
		if *relayPolicy != "" {
			relay, err = relay.WithProvenancePolicy(common.HexToHash(*relayPolicy))
			if err != nil {
				log.Fatal(err)
			}
		}
		maxFee, ok := new(big.Int).SetString(*relayMaxFee, 10)
		if !ok {
			log.Fatal("invalid maximum fee")
		}
		if err = operation.WithBroadcaster(ctx, relay, credit.WithBroadcasterOptions{MaxFeeAmount: maxFee}); err != nil {
			log.Fatal(err)
		}
		preview, err := operation.Preview(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Final quoted Credit preview: %+v", preview)
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

	if relay != nil {
		if *submit {
			status, err := operation.SubmitBroadcaster(ctx, relay)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Broadcaster operation=%s state=%s tx=%s", *operationID, status.State, status.TransactionHash)
		} else {
			log.Printf("Quoted/proved operation %s retained; --submit was not supplied", *operationID)
		}
		return
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
