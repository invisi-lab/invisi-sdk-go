// Connect a public-only wallet to an explicit local cryptographic executor.
package main

import (
	"context"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"log"
)

func main() {
	executor, err := native.NewExecutor("0x0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		log.Fatal(err)
	}
	wallet, err := sdk.ConnectWallet(context.Background(), executor.OwnerAddress(), executor)
	if err != nil {
		log.Fatal(err)
	}
	address, err := wallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EVM owner:", wallet.OwnerAddress())
	fmt.Println("Receiving address:", address)
	fmt.Printf("Public viewing key: 0x%x\nPublic clue key: 0x%x\n", wallet.ViewingPublicKey(), wallet.ClueKey())
}
