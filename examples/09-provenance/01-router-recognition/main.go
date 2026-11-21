package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	provenanceexample "github.com/invisi-labs/invisi-sdk-go/examples/09-provenance"
	"github.com/invisi-labs/invisi-sdk-go/examples/shared"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
)

func main() {
	execute := flag.Bool("execute", false, "query real Source recognition")
	chainRPC := flag.String("chain-rpc", "", "trusted canonical Chain RPC")
	protocolRPC := flag.String("protocol-rpc", "", "Protocol RPC for the same deployment")
	tx := flag.String("transaction-hash", "", "actual Router transaction hash")
	commitment := flag.String("commitment", "", "actual Note commitment from that transaction")
	flag.Parse()
	if !*execute {
		log.Print("Review deployment and Source trust, then pass --execute with RPC URLs, transaction hash and commitment")
		return
	}
	for _, value := range []string{*tx, *commitment} {
		decoded, err := hexutil.Decode(value)
		if err != nil || len(decoded) != 32 {
			log.Fatal("transaction hash and commitment must be 32-byte hex values")
		}
	}
	if *chainRPC == "" || *protocolRPC == "" {
		log.Fatal("both RPC URLs are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	provider, err := ethclient.DialContext(ctx, *chainRPC)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	rpc, err := sdk.NewProtocolRpcClient(*protocolRPC)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, rpc, provider)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = shared.WaitNativeFinality(ctx, provider, common.HexToHash(*tx), time.Minute); err != nil {
		log.Fatal(err)
	}
	source, err := shared.SourceClient(pool)
	if err != nil {
		log.Fatal(err)
	}
	transaction, note, err := provenanceexample.Recognition(ctx, source, common.HexToHash(*tx), common.HexToHash(*commitment))
	if err != nil {
		log.Fatal(err)
	}
	// Preserve pending and reason fields. This does not prove commitment membership
	// in the given transaction; use canonical output evidence for that association.
	if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"transaction": transaction, "note": note}); err != nil {
		log.Fatal(err)
	}
}
