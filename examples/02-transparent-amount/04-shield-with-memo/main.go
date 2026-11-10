// Alice funds a 10-unit Note for Bob with an authenticated application memo.
package main

import (
	"context"
	"flag"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func main() {
	// Configure public actors and deployment.
	ctx := context.Background()
	chainRPC := flag.String("chain-rpc", "http://127.0.0.1:18565", "Chain RPC URL")
	protocolRPC := flag.String("protocol-rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	address := flag.String("wallet-address", "", "public long receiving address")
	recipient := flag.String("recipient", "", "recipient public long receiving address")
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
	memo, err := core.FundsV1.Encode(core.PayloadDataV1{Memo: []byte("Invoice 42")})
	if err != nil {
		log.Fatal(err)
	}
	output, err := sdk.NewNoteOutput(sdk.PaymentRequest{
		ChainID:         new(big.Int).SetUint64(pool.Context().ChainID),
		ProtocolAddress: pool.Context().ProtocolAddress,
		TokenAddress:    token,
		Amount:          big.NewInt(10),
		Payload:         memo,
		Recipient:       sdk.InvisiRecipient{Address: *recipient},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Freeze recipients and amounts before authorization.
	prepared, err := invisi.PrepareShield(ctx, sdk.ExplicitShieldPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: publicWallet, Protocol: pool},
		ShieldPreparation: sdk.ShieldPreparation{
			TokenAddress: token,
			Outputs:      []sdk.NoteOutput{output},
			Funding:      sdk.FundingSource{Payer: publicWallet.OwnerAddress(), Amount: big.NewInt(10), Authorization: "caller"},
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
	log.Printf("%+v", call)
}
