// Bootstrap public wallet material through an external native request/response exchange.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func main() {
	owner := flag.String("owner", "0x7e5f4552091a69125d5dfcb7b8c2659029395bdf", "public owner address")
	flag.Parse()

	request, err := wallet.Prepare("public-1", common.HexToAddress(*owner), wallet.Public, struct{}{})
	if err != nil {
		log.Fatal(err)
	}
	requests := []wallet.Request{request}
	if err = json.NewEncoder(os.Stdout).Encode(requests); err != nil {
		log.Fatal(err)
	}

	var responses []wallet.Response
	if err = json.NewDecoder(os.Stdin).Decode(&responses); err != nil {
		log.Fatal(err)
	}

	// Match by ID rather than arrival order; attachment independently verifies the signature and statement.
	ordered, err := wallet.ValidateResponses(requests, responses)
	if err != nil {
		log.Fatal(err)
	}
	publicWallet, err := sdk.WalletFromResponse(request, ordered[0])
	if err != nil {
		log.Fatal(err)
	}
	address, err := publicWallet.GetAddress()
	if err != nil {
		log.Fatal(err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(address); err != nil {
		log.Fatal(err)
	}
}
