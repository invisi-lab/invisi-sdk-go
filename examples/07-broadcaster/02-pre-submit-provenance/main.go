package main

import (
	"context"
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
)

func must[T any](value T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return value
}
func main() {
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient receiving address")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
	execute := flag.Bool("execute", false, "discover, prepare and prove")
	submit := flag.Bool("submit", false, "explicitly send the payment")
	flag.Parse()
	if !*execute {
		log.Print("Pass --execute; add --submit to send payment after proving")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	provider := must(ethclient.DialContext(ctx, *chainRPC))
	defer provider.Close()
	rpc := must(sdk.NewProtocolRpcClient(*protocolRPC))
	pool := must(sdk.ProtocolFromRPC(ctx, rpc, provider))
	publicWallet := must(sdk.WalletFromAddress(*address))
	backend := must(native.NewNativeProver(native.NativeOptions{BBBinary: os.Getenv("BB_BINARY"), NoirBinary: os.Getenv("NOIR_EXECUTOR")}))
	source := must(shared.SourceCapability(pool))
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: rpc, Prover: backend, Provenance: source})
	key := must(crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("OWNER_PRIVATE_KEY"), "0x")))
	if crypto.PubkeyToAddress(key.PublicKey) != publicWallet.OwnerAddress() {
		log.Fatal("OWNER_PRIVATE_KEY must match the Note owner")
	}
	signer := must(bind.NewKeyedTransactorWithChainID(key, new(big.Int).SetUint64(pool.Context().ChainID)))
	managed := must(walletsession.Connect(ctx, walletsession.Options{Invisi: invisi, Wallet: publicWallet, Protocol: pool, Store: walletsession.NewMemoryWalletStore(), Signer: signer}))
	pending := must(managed.Refresh(ctx, walletsession.RefreshOptions{FromBlock: pool.Deployment().Protocol.FromBlock}))
	_ = must(pending.Decrypt(ctx, shared.AppDeliveryDecryptor{}))
	token := common.HexToAddress(*tokenText)
	balance := must(managed.Balance(ctx, token))

	if balance.AvailableKnown == nil || balance.AvailableKnown.Cmp(big.NewInt(2)) <= 0 {
		log.Fatal("fund and recognize Notes first; leave funds for fees")
	}
	output := must(sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: new(big.Int).SetUint64(pool.Context().ChainID), ProtocolAddress: pool.Context().ProtocolAddress, TokenAddress: token, Amount: new(big.Int).Div(balance.AvailableKnown, big.NewInt(2)), Recipient: sdk.InvisiRecipient{Address: *recipient}}))
	prepared := must(managed.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: []sdk.NoteOutput{output}}))
	authorization := must(prepared.ExportWalletAuthorization(ctx))
	signature := must(shared.Sign(authorization, publicWallet.OwnerAddress(), "OWNER_PRIVATE_KEY"))
	prepared = must(prepared.WithWalletSignature(ctx, signature))
	ready := must(prepared.Prove(ctx))
	ready = must(ready.ProveProvenance(ctx))
	packet := must(ready.ExportProvenanceProof(ctx))
	log.Printf("Public Source proof: operation=%s sidecars=%d", packet.OperationDigest, len(packet.Sidecars))
	accepted := must(ready.SubmitProvenance(ctx))
	tracking := must(provenance.TrackingFromResult(accepted, ""))
	if string(accepted["broadcastDecision"]) != `"allow"` || tracking.OperationDigest != packet.OperationDigest {
		log.Fatal("Source did not accept the exact frozen operation")
	}
	log.Printf("Source tracking: %s", tracking.ID)
	if !*submit {
		log.Print("Source accepted; no payment sent without --submit")
		return
	}
	tx := must(managed.SendTransaction(ctx, ready))
	log.Printf("Main transaction: %s", tx.Hash())
	receipt := must(bind.WaitMined(ctx, provider, tx))
	if receipt.Status != 1 {
		log.Fatal("main transaction reverted; retain the original operation record")
	}
	log.Printf("Source recognition: %+v", must(source.Client.WaitTracking(ctx, tracking.ID, 2*time.Second, 2*time.Minute)))
}
