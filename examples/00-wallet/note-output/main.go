// Build three output intents offline. This example sends no transaction.
package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// This ephemeral wallet is only for demonstration; do not fund it.
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
	options := sdk.NoteOutputOptions{
		ChainID:         big.NewInt(31337),
		ProtocolAddress: common.HexToAddress("0x1111111111111111111111111111111111111111"),
		TokenAddress:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
		Amount:          big.NewInt(10),
	}

	// 1. A complete 0i address encrypts immediately, with no registry lookup.
	// The nonce is optional; omit it to use a cryptographically random field element.
	options.Recipient = sdk.InvisiRecipient{Address: address, Nonce: big.NewInt(42)}
	encrypted, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encrypted output commitment:", encrypted.Commitment())

	// 2. An EVM recipient defaults to an envelope and Protocol RPC lookup.
	// Creation stays offline. PrepareShield/PrepareTransfer resolves registration
	// and fails if it is absent. AddressLookup: "chain" selects direct chain reads.
	options.Recipient = sdk.EVMRecipient{Address: wallet.OwnerAddress()}
	pending, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EVM delivery pending:", pending.DeliveryPending())

	// 3. A merchant can supply a per-note identity and expected commitment.
	// The sender does not need the owner's address or nonce in this mode.
	options.Recipient = sdk.IdentityRecipient{Identity: encrypted.Identity()}
	opaque, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	if err := opaque.AssertCommitment(encrypted.Commitment()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Opaque output has a private opening:", opaque.Note() != nil)

	// An EVM output can also explicitly omit delivery and skip registration.
	// The recipient must retain its complete note to recover and spend it.
	envelope := false
	options.Recipient = sdk.EVMRecipient{Address: wallet.OwnerAddress(), Envelope: &envelope}
	withoutEnvelope, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Explicit no-envelope output:", len(withoutEnvelope.Delivery().Ciphertext) == 0)
}
