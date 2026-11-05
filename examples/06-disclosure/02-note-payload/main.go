// 02 note payload.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	ctx := context.Background()

	// 1. Load local demonstration accounts.
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if len(accounts) < 2 {
		log.Fatal("this example requires at least 2 initialized accounts")
	}
	bob := accounts[1]

	// 2. Connect to the chain and pin the Protocol RPC deployment.
	provider, err := ethclient.DialContext(ctx, shared.Settings.ChainRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	chain, err := provider.ChainID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if chain.Uint64() != shared.Settings.ChainID {
		log.Fatal("chain does not match Settings.ChainID")
	}
	client, err := sdk.NewProtocolRpcClient(shared.Settings.ProtocolRPCURL)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, client, provider)
	if err != nil {
		log.Fatal(err)
	}
	token := common.HexToAddress(shared.Settings.TokenAddress)
	verifier, err := sdk.NewDisclosureVerifier(sdk.DisclosureVerifierOptions{Protocol: pool, Provider: provider, TokenAddress: token, Confirmations: 1})
	if err != nil {
		log.Fatal(err)
	}
	var disclosure *sdk.Disclosure
	var reference sdk.DisclosureReference

	// 3. Discover authenticated notes and choose this transaction's inputs.
	candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, bob.PublicWallet, bob.ViewingKey, token)
	if err != nil {
		log.Fatal(err)
	}
	if len(candidates) == 0 {
		log.Fatal("no authenticated unspent note to disclose")
	}
	selected := candidates[0]
	viewer, e := sdk.NewViewingCapability(bob.PublicWallet.OwnerAddress(), bob.ViewingKey, bob.PublicWallet.ViewingPublicKey())
	if e != nil {
		log.Fatal(e)
	}
	disclosure, err = viewer.DiscloseNote(selected.Output.Ciphertext, core.NoteDecryptionContext{
		NoteEncryptionContext: core.NoteEncryptionContext{
			ProtocolContext:  selected.Input.Note.ProtocolContext,
			OutputCommitment: core.BigIntToElement(selected.Input.Commitment.Big()),
		},
		TokenAddress: selected.Input.Note.TokenAddress,
		OwnerAddress: selected.Input.Note.OwnerAddress,
	})
	if err != nil {
		log.Fatal(err)
	}
	reference = sdk.DisclosureReference{
		TransactionHash: selected.Output.TransactionHash,
		TreeID:          selected.Input.TreeID,
		LeafIndex:       selected.Input.LeafIndex,
		Role:            "funds",
	}
	data, e := core.FundsV1.Decode(selected.Input.Note.Payload)
	if e != nil {
		log.Fatal("selected payload uses a different application layout:", e)
	}
	fmt.Println("Decoded application memo bytes:", len(data.Memo))

	// 4. Export this one disclosure locally; never print its sensitive contents.
	exported, err := json.Marshal(disclosure.Export())
	if err != nil {
		log.Fatal(err)
	}
	defer clear(exported)
	imported, err := sdk.DisclosureFromExport(exported)
	if err != nil {
		log.Fatal(err)
	}
	verified, err := verifier.Verify(ctx, imported, reference)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Verified canonical disclosure block:", verified.Block.Number, "Confirmations:", verified.Confirmations)
	fmt.Println("Redacted inspection:", disclosure, "Scope:", disclosure.GetScope())
}
