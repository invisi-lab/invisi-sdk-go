// Alice discovers spender grants and draws seven gross units from each of up to two grants.
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"math/big"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
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
	recipientAddress := flag.String("recipient", "", "ordinary Note recipient public long receiving address")
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
	recipientWallet, err := sdk.WalletFromAddress(*recipientAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Reuse the encrypted inventory populated above; a second ephemeral discovery would repeat host prompts.
	inventory, e := store.ReadSnapshot(ctx, managed.Scope())
	if e != nil {
		log.Fatal(e)
	}
	candidates, e := shared.CreditCandidates(ctx, provider, client, pool, inventory, token, credit.Spender)
	if e != nil {
		log.Fatal(e)
	}
	if len(candidates) == 0 {
		log.Fatal("no active grants; run owner creation examples and approve the Protocol for each payer first")
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i].Input.State, candidates[j].Input.State
		if x := a.OwnerWalletIdentity.Cmp(&b.OwnerWalletIdentity); x != 0 {
			return x < 0
		}
		if x := a.RevocationShard.Cmp(&b.RevocationShard); x != 0 {
			return x < 0
		}
		return a.RevocationSlot.Cmp(&b.RevocationSlot) < 0
	})
	if len(candidates) > 2 {
		candidates = candidates[:2]
	}
	p := credit.DrawParams{Token: token, TotalAmount: new(big.Int), Deadline: deadline}
	inputs := []credit.LocatedState{}
	ds := []protocol.Delivery{}
	sources := map[uint64]common.Hash{}

	// Gross debits drive grant quota and usage; the ordinary output Note receives the combined net amount.
	debit := primitives.ElementFromUint64(7)
	for _, candidate := range candidates {
		s := candidate.Input.State
		if primitives.ElementToBigInt(s.Expiry).Cmp(deadline) < 0 {
			log.Fatal("selected grant expires before request deadline")
		}
		owner := common.BigToAddress(primitives.ElementToBigInt(s.OwnerWalletIdentity))
		pi := len(p.Payers) - 1
		if pi < 0 || p.Payers[pi].Payer != owner {
			p.Payers = append(p.Payers, credit.Payer{Payer: owner, Amount: new(big.Int)})
			pi++
		}
		p.Payers[pi].Amount.Add(p.Payers[pi].Amount, big.NewInt(7))
		next, e := s.Draw(debit, head.Time)
		if e != nil {
			log.Fatal(e)
		}
		p.GrantRefs = append(p.GrantRefs, credit.GrantRef{
			PayerIndex: big.NewInt(int64(pi)),
			Shard:      primitives.ElementToBigInt(s.RevocationShard).Uint64(),
			Slot:       uint16(primitives.ElementToBigInt(s.RevocationSlot).Uint64()),
		})
		p.Nullifiers = append(p.Nullifiers, common.Hash(next.Nullifier.Bytes()))
		p.CreditCommitments = append(p.CreditCommitments, common.Hash(next.StateCommitment.Bytes()))
		p.RecoveryAmounts = append(p.RecoveryAmounts, common.Hash(next.RecoveryAmount.Bytes()))
		p.TotalAmount.Add(p.TotalAmount, big.NewInt(7))
		candidate.Input.Debit = debit
		inputs = append(inputs, candidate.Input)
		sources[candidate.Input.TreeID] = candidate.Membership.Root
		snapshot := candidate.Delivery
		snapshot.State = next.State
		deliveries, e := shared.CreditDeliveries(snapshot)
		if e != nil {
			log.Fatal(e)
		}
		ds = append(ds, deliveries...)
	}
	ids := []uint64{}
	for id := range sources {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		p.Sources = append(p.Sources, protocol.InputSource{TreeId: new(big.Int).SetUint64(id), MerkleRoot: sources[id]})
	}
	nonce, e := primitives.GenerateNonce()
	if e != nil {
		log.Fatal(e)
	}
	note := primitives.Note{
		ProtocolContext: coreContext,
		OwnerAddress:    primitives.BigIntToElement(recipientWallet.OwnerAddress().Big()),
		TokenAddress:    primitives.BigIntToElement(token.Big()),
		Amount:          primitives.BigIntToElement(p.TotalAmount),
		Nonce:           nonce,
	}
	prepared, err = credit.PrepareDrawToNote(c, p, ds, inputs, note, &primitives.ReceivingKeys{ViewingPublicKey: recipientWallet.ViewingPublicKey(), ClueKey: recipientWallet.ClueKey()})
	log.Printf("Gross payer debit=%s; ordinary Note recipient=%s", p.TotalAmount, recipientWallet.OwnerAddress())

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
