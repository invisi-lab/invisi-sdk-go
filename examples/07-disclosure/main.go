// Offline disclosure releases only explicitly selected output data.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	wallet, err := sdk.WalletFromPrivateKey("0x0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		return err
	}
	nonce, err := core.GenerateNonce()
	if err != nil {
		return err
	}
	// These numeric deployment addresses are an offline illustration, never a live deployment.
	context := core.ProtocolContext{ChainID: core.ElementFromUint64(31337), ProtocolAddress: core.ElementFromUint64(1)}
	note := core.Note{ProtocolContext: context, WalletIdentity: core.BigIntToElement(wallet.WalletIdentity().Big()), TokenAddress: core.ElementFromUint64(2), Amount: core.ElementFromUint64(70), Nonce: nonce}
	fmt.Println("1. Construct a scoped commitment opening")
	disclosure, err := sdk.DisclosureFromNote(note)
	if err != nil {
		return err
	}
	// This explicit export is sensitive. The application selects its recipient and transport.
	exported, err := json.Marshal(disclosure.Export())
	if err != nil {
		return err
	}
	imported, err := sdk.DisclosureFromExport(exported)
	clear(exported)
	if err != nil {
		return err
	}
	commitment := disclosure.Export().Context.OutputCommitment
	output := sdk.DisclosureOutput{ProtocolContext: context, TokenAddress: note.TokenAddress, OutputCommitment: commitment, Role: "funds"}
	result, err := imported.Verify(output)
	if err != nil {
		return err
	}
	fmt.Println("2. Recomputed disclosed amount:", result.Note.Amount.String())
	fmt.Println("3. Default inspection redacts private material:", disclosure)
	fmt.Println("4. Scope:", disclosure.GetScope())
	// Use sdk.DisclosureVerifier with a trusted Protocol/provider for independent onchain inclusion.
	return nil
}
