// Alice funds nine 10-unit Notes, one for each recipient, from 90 public token units.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// Configure public actors and deployment; the native host retains account access.
	ctx := context.Background()
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC URL")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipientsText := flag.String("recipients", "", "nine comma-separated public long receiving addresses")
	tokenText := flag.String("token", "0x2222222222222222222222222222222222222222", "token contract")
	flag.Parse()

	publicWallet, err := sdk.WalletFromAddress(*address)
	if err != nil {
		log.Fatal(err)
	}

	// Authenticate the Protocol RPC deployment against Chain RPC before preparing financial intent.
	provider, err := ethclient.DialContext(ctx, *chainRPC)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()
	client, err := sdk.NewProtocolRpcClient(*protocolRPC)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := sdk.ProtocolFromRPC(ctx, client, provider)
	if err != nil {
		log.Fatal(err)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: client})
	token := common.HexToAddress(*tokenText)
	encoder := json.NewEncoder(os.Stdout)
	recipients := strings.Split(*recipientsText, ",")
	if len(recipients) != 9 {
		log.Fatal("exactly nine recipients required")
	}
	outputs := make([]sdk.NoteOutput, 0, 9)
	for _, recipient := range recipients {
		output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
			ChainID:         new(big.Int).SetUint64(pool.Context().ChainID),
			ProtocolAddress: pool.Context().ProtocolAddress,
			TokenAddress:    token,
			Amount:          big.NewInt(10),
			Recipient:       sdk.InvisiRecipient{Address: recipient},
		})
		if err != nil {
			log.Fatal(err)
		}
		outputs = append(outputs, output)
	}

	// Freeze recipients and amounts before authorization so the native response approves this exact intent.
	prepared, err := invisi.PrepareShield(ctx, sdk.ExplicitShieldPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: publicWallet, Protocol: pool},
		ShieldPreparation: sdk.ShieldPreparation{
			TokenAddress: token,
			Outputs:      outputs,
			Funding:      sdk.FundingSource{Payer: publicWallet.OwnerAddress(), Amount: big.NewInt(90), Authorization: "caller"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	prepared, err = prepared.Ready(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Export unsigned calldata for the external signer; submission and canonical reconciliation remain explicit.
	call, err := prepared.ExportTransaction(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if err = encoder.Encode(call); err != nil {
		log.Fatal(err)
	}
}
