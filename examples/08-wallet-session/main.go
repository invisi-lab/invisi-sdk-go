// Connect the two RPC services and restore owned notes into encrypted SQLite storage.
package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	walletsqlite "github.com/invisi-lab/invisi-sdk-go/sdk/sqlite"
)

func required(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", errors.New("set " + name)
	}
	return value, nil
}
func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
func run(ctx context.Context) error {
	chainURL, err := required("CHAIN_RPC_URL")
	if err != nil {
		return err
	}
	protocolURL, err := required("PROTOCOL_RPC_URL")
	if err != nil {
		return err
	}
	accountKey, err := required("WALLET_ACCOUNT_PRIVATE_KEY")
	if err != nil {
		return err
	}
	storageKeyText, err := required("WALLET_STORAGE_KEY")
	if err != nil {
		return err
	}
	storageKey, err := hex.DecodeString(strings.TrimPrefix(storageKeyText, "0x"))
	if err != nil {
		return err
	}
	defer clear(storageKey)

	provider, err := ethclient.DialContext(ctx, chainURL)
	if err != nil {
		return err
	}
	defer provider.Close()
	protocolRPC, err := sdk.NewProtocolRpcClient(protocolURL)
	if err != nil {
		return err
	}
	protocol, err := sdk.ProtocolFromRPC(ctx, protocolRPC, provider)
	if err != nil {
		return err
	}
	wallet, err := sdk.WalletFromPrivateKey(accountKey)
	if err != nil {
		return err
	}

	codec, err := sdk.NewEncryptedWalletCodec(storageKey)
	if err != nil {
		return err
	}
	database, err := sql.Open("sqlite", "./wallet.sqlite")
	if err != nil {
		return err
	}
	defer database.Close()
	store, err := walletsqlite.NewWalletStore(ctx, database, codec)
	if err != nil {
		return err
	}

	// The official noir-execute and BB binaries run only when an operation proves.
	localProver, err := native.NewNativeProver(native.NativeOptions{})
	if err != nil {
		return err
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(accountKey, "0x"))
	if err != nil {
		return err
	}
	chainID, err := provider.ChainID(ctx)
	if err != nil {
		return err
	}
	signer, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return err
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{
		ProtocolRPC: protocolRPC,
		Provider:    provider,
		Store:       store,
		Signer:      signer,
		Prover:      localProver,
	})
	session, err := invisi.Connect(ctx, sdk.ConnectionOptions{Wallet: wallet, Protocol: protocol, Namespace: "examples"})
	if err != nil {
		return err
	}
	fmt.Println("1. Connected Protocol:", protocol.Context().ProtocolAddress)
	result, err := session.Refresh(ctx, sdk.RefreshOptions{FromBlock: protocol.Deployment().Protocol.FromBlock})
	if err != nil {
		return err
	}
	fmt.Println("2. Refreshed owned inventory through:", result.Block.Number)

	// The seeded local token; edit this constant for another deployment.
	tokenAddress := common.HexToAddress("0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6")
	balance, err := session.Balance(ctx, tokenAddress)
	if err != nil {
		return err
	}
	fmt.Println("3. Available known note funds:", balance.AvailableKnown)
	return nil
}
