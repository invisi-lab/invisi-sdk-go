// Restore the same EVM owner and Invisi receiving address from one account key.
package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	// This public test key must never hold real funds. Applications own mnemonic/account derivation.
	privateKey := "0000000000000000000000000000000000000000000000000000000000000001"
	account, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return err
	}
	evmAddress := crypto.PubkeyToAddress(account.PublicKey)

	wallet, err := sdk.WalletFromPrivateKey("0x" + privateKey)
	if err != nil {
		return err
	}
	address, err := wallet.GetAddress()
	if err != nil {
		return err
	}
	fmt.Println("1. EVM owner:", evmAddress.Hex())
	fmt.Println("2. Invisi receiving address:", address, len(address), "characters")
	fmt.Println("3. Short address:", wallet.GetShortAddress())

	publicWallet, err := sdk.WalletFromAddress(address)
	if err != nil {
		return err
	}
	fmt.Println("4. Address-only wallet owner:", publicWallet.OwnerAddress())
	viewer, err := wallet.GetViewer()
	if err != nil {
		return err
	}
	detector, err := wallet.GetDetector()
	if err != nil {
		return err
	}
	fmt.Printf("5. Public viewing key: 0x%x\n", viewer.GetPublicKey())
	fmt.Printf("6. Public clue key: 0x%x\n", detector.GetClueKey())

	// Only explicitly exported copies leave the capability object; never print them.
	viewingPrivateKey := viewer.ExportKey()
	detectionKey := detector.ExportKey()
	clear(viewingPrivateKey)
	clear(detectionKey)
	return nil
}
