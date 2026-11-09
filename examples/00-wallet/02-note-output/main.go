// Compare full-address, registered EVM, opaque-identity and no-envelope Note outputs offline.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func main() {
	// These fixture values are public material only.
	var request wallet.Request
	var response wallet.Response
	if err := json.Unmarshal([]byte(`{
  "version": 1,
  "id": "wallet.public",
  "ownerAddress": "0x7e5f4552091a69125d5dfcb7b8c2659029395bdf",
  "method": "wallet.public",
  "params": {}
}`), &request); err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{
  "version": 1,
  "id": "wallet.public",
  "ownerAddress": "0x7e5f4552091a69125d5dfcb7b8c2659029395bdf",
  "method": "wallet.public",
  "result": {
    "viewingPublicKey": "0x0396f42bb298c8a0e74539294780d709799f9838c0b69a134f38c6d968f2921397",
    "clueKey": "0x1292f7299a65930b0476394e55b6475425fb3ee7445176537042966a4653bf06",
    "spendingPublicKey": "0x0479be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"
  }
}`), &response); err != nil {
		log.Fatal(err)
	}
	publicWallet, err := sdk.WalletFromResponse(request, response)
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

	// 2. A full address encrypts immediately and accepts a caller-chosen nonce.
	options.Recipient = sdk.InvisiRecipient{Address: address, Nonce: big.NewInt(42)}
	encrypted, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encrypted output commitment:", encrypted.Commitment())

	// 3. An EVM recipient resolves registered receiving keys during preparation.
	options.Recipient = sdk.EVMRecipient{Address: publicWallet.OwnerAddress()}
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
	options.Recipient = sdk.EVMRecipient{Address: publicWallet.OwnerAddress(), Envelope: &envelope}
	withoutEnvelope, err := sdk.NewNoteOutput(options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Explicit no-envelope output:", len(withoutEnvelope.Delivery().Ciphertext) == 0)
}
