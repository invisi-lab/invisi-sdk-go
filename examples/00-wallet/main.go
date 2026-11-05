// Derive an offline wallet and expose only its public identity and viewing capabilities.
package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// Generate a fresh ephemeral key for this offline demonstration; no local account file is read.
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	wallet, err := sdk.WalletFromPrivateKey("0x" + hex.EncodeToString(crypto.FromECDSA(key)))
	if err != nil {
		log.Fatal(err)
	}
	address, err := wallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}
	publicWallet, err := sdk.PublicWalletFromAddress(address)
	if err != nil {
		log.Fatal(err)
	}
	viewer, err := wallet.GetViewer()
	if err != nil {
		log.Fatal(err)
	}
	detector, err := wallet.GetDetector()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EVM owner:", publicWallet.OwnerAddress())
	fmt.Println("Long receiving address:", address)
	fmt.Println("EVM owner address length:", len(publicWallet.OwnerAddress().Hex()))
	fmt.Printf("Public viewing key: 0x%x\nPublic clue key: 0x%x\n", viewer.GetPublicKey(), detector.GetClueKey())
	fmt.Println("Private capabilities remain local:", viewer, detector)
}
