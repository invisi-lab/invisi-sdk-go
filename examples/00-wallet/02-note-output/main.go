// Create complete note outputs without preparing or sending a transaction.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/examples/shared"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// 1. Create ephemeral receiving keys for this offline walkthrough.
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	wallet, err := shared.LocalWallet("0x" + hex.EncodeToString(crypto.FromECDSA(key)))
	if err != nil {
		log.Fatal(err)
	}
	address, err := wallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}
	options := sdk.PaymentRequest{
		ChainID:         big.NewInt(31337),
		ProtocolAddress: common.HexToAddress("0x1111111111111111111111111111111111111111"),
		TokenAddress:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
		Amount:          big.NewInt(10),
	}

	// 2. A full address encrypts immediately and accepts a caller-chosen nonce.
	options.Recipient = sdk.InvisiRecipient{Address: address, Nonce: big.NewInt(42)}
	encrypted, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encrypted output commitment:", encrypted.Commitment())

	// 3. An EVM recipient resolves registered receiving keys during preparation.
	options.Recipient = sdk.EVMRecipient{Address: wallet.OwnerAddress()}
	pending, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EVM delivery pending:", pending.DeliveryPending())

	// 4. A precomputed identity needs no envelope or private note opening.
	options.Recipient = sdk.IdentityRecipient{Identity: encrypted.Identity()}
	opaque, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	if err := opaque.AssertCommitment(encrypted.Commitment()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Opaque output has a private opening:", opaque.Note() != nil)

	// 5. An explicit no-envelope EVM output does not require registration.
	envelope := false
	options.Recipient = sdk.EVMRecipient{Address: wallet.OwnerAddress(), Envelope: &envelope}
	withoutEnvelope, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Explicit no-envelope output:", len(withoutEnvelope.Delivery().Ciphertext) == 0)
}
