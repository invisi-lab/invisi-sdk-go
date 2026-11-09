// Compare full-address, registered EVM, opaque-identity and no-envelope Note outputs offline.
package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// A complete public receiving address is enough to encrypt; no native exchange is needed.
	publicWallet, err := sdk.WalletFromAddress(
		"0i0e0525SFRf53yh2aljmm3sn9JQ5nJk7LqWt0G2ajnRY2pe698y550QXhp9uELXpcCZMf5Y608rRDJ68jJgfEWy5j7U5e5eVNPVz8vW2W2kmyw4p9lVLwW3z3WeFHQs5KdfR980CX",
	)
	if err != nil {
		log.Fatal(err)
	}
	address, err := publicWallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}

	options := sdk.PaymentRequest{
		ChainID:         big.NewInt(31337),
		ProtocolAddress: common.HexToAddress("0x1111111111111111111111111111111111111111"),
		TokenAddress:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
		Amount:          big.NewInt(10),
	}

	// 1. A full address encrypts immediately and accepts a caller-chosen nonce.
	options.Recipient = sdk.InvisiRecipient{Address: address, Nonce: big.NewInt(42)}
	encrypted, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encrypted output commitment:", encrypted.Commitment())

	// 2. An EVM recipient resolves registered receiving keys during preparation.
	options.Recipient = sdk.EVMRecipient{Address: publicWallet.OwnerAddress()}
	pending, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EVM delivery pending:", pending.DeliveryPending())

	// 3. A precomputed identity needs no envelope or private note opening.
	options.Recipient = sdk.IdentityRecipient{Identity: encrypted.Identity()}
	opaque, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	if err := opaque.AssertCommitment(encrypted.Commitment()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Opaque output has a private opening:", opaque.Note() != nil)

	// 4. An explicit no-envelope EVM output does not require registration.
	envelope := false
	options.Recipient = sdk.EVMRecipient{Address: publicWallet.OwnerAddress(), Envelope: &envelope}
	withoutEnvelope, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Explicit no-envelope output:", len(withoutEnvelope.Delivery().Ciphertext) == 0)
}
