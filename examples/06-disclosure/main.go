package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	mode := flag.String("mode", "note-opening", "note-opening, note-payload or credit-opening")
	actor := flag.Int("actor", 0, "account index whose output is disclosed")
	flag.Parse()
	ctx := context.Background()
	accounts, err := shared.GetAccounts(shared.Settings.AccountsFile)
	if err != nil {
		log.Fatal(err)
	}
	if *actor < 0 || *actor >= len(accounts) {
		log.Fatal("actor index out of range")
	}
	alice := accounts[*actor]
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
	if *mode == "credit-opening" {
		candidates, err := shared.DiscoverCredits(ctx, provider, client, pool, alice.PublicWallet, alice.ViewingKey, token, "owner")
		if err != nil {
			log.Fatal(err)
		}
		if len(candidates) == 0 {
			log.Fatal("no authenticated owned Credit to disclose")
		}
		selected := candidates[0]
		disclosure, err = sdk.DisclosureFromCredit(selected.Input.State.Note)
		if err != nil {
			log.Fatal(err)
		}
		reference = sdk.DisclosureReference{TransactionHash: selected.Output.TransactionHash, TreeID: selected.Input.TreeID, LeafIndex: selected.Input.LeafIndex, Role: "owner"}
	} else {
		candidates, err := shared.DiscoverNotes(ctx, provider, client, pool, alice.PublicWallet, alice.ViewingKey, token)
		if err != nil {
			log.Fatal(err)
		}
		if len(candidates) == 0 {
			log.Fatal("no authenticated unspent note to disclose")
		}
		selected := candidates[0]
		switch *mode {
		case "note-opening":
			disclosure, err = sdk.DisclosureFromNote(selected.Input.Note)
		case "note-payload":
			viewer, e := sdk.NewViewingCapability(alice.PublicWallet.OwnerAddress(), alice.ViewingKey, alice.PublicWallet.ViewingPublicKey())
			if e != nil {
				log.Fatal(e)
			}
			disclosure, err = viewer.DiscloseNote(selected.Output.Ciphertext, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: selected.Input.Note.ProtocolContext, OutputCommitment: core.BigIntToElement(selected.Input.Commitment.Big())}, TokenAddress: selected.Input.Note.TokenAddress, WalletIdentity: selected.Input.Note.WalletIdentity})
		default:
			log.Fatal("unknown disclosure mode")
		}
		if err != nil {
			log.Fatal(err)
		}
		reference = sdk.DisclosureReference{TransactionHash: selected.Output.TransactionHash, TreeID: selected.Input.TreeID, LeafIndex: selected.Input.LeafIndex, Role: "funds"}
		if *mode == "note-payload" {
			data, e := core.FundsV1.Decode(selected.Input.Note.Payload)
			if e != nil {
				log.Fatal("selected payload uses a different application layout:", e)
			}
			fmt.Println("Decoded application memo bytes:", len(data.Memo))
		}
	}
	// Export is sensitive. This local round trip models an explicitly selected recipient transport.
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
