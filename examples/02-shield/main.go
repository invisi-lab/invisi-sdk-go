// Caller-funded shield uses ordinary geth approval and sending, without a Store or proof.
package shield

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
	"strings"
)

func Run(ctx context.Context, provider *ethclient.Client, protocolRPC *sdk.ProtocolRpcClient, pool *sdk.Protocol, wallet *sdk.Wallet, signer *bind.TransactOpts, recipient string) error {
	if signer == nil {
		return errors.New("geth signer required")
	}
	tokenAddress := common.HexToAddress("0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6")
	// Standard ERC20 approval belongs to geth, not an Invisi SDK abstraction.
	tokenABI, err := abi.JSON(strings.NewReader(`[{"type":"function","name":"approve","stateMutability":"nonpayable","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"type":"bool"}]}]`))
	if err != nil {
		return err
	}
	token := bind.NewBoundContract(tokenAddress, tokenABI, provider, provider, provider)
	options := *signer
	options.Context = ctx
	approval, err := token.Transact(&options, "approve", pool.Context().ProtocolAddress, big.NewInt(100))
	if err != nil {
		return err
	}
	receipt, err := bind.WaitMined(ctx, provider, approval)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return errors.New("ERC20 approval failed")
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: provider, ProtocolRPC: protocolRPC, Signer: signer})
	prepared, err := invisi.PrepareShield(ctx, sdk.ExplicitShieldPreparation{
		TransactionIdentity: sdk.TransactionIdentity{Wallet: wallet, Protocol: pool},
		ShieldPreparation: sdk.ShieldPreparation{TokenAddress: tokenAddress, Outputs: []sdk.PaymentOutput{{To: recipient, Amount: big.NewInt(50)}, {To: recipient, Amount: big.NewInt(50)}},
			Funding: sdk.FundingSource{Payer: signer.From, Amount: big.NewInt(100), Authorization: "caller"}}})
	if err != nil {
		return err
	}
	fmt.Println("Review:", prepared.Preview())
	ready, err := prepared.Ready(ctx)
	if err != nil {
		return err
	}
	call, err := ready.ExportTransaction(ctx)
	if err != nil {
		return err
	}
	contract := bind.NewBoundContract(call.To, abi.ABI{}, provider, provider, provider)
	transaction, err := contract.RawTransact(&options, call.Data)
	if err != nil {
		return err
	}
	receipt, err = bind.WaitMined(ctx, provider, transaction)
	if err != nil {
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return errors.New("shield transaction failed")
	}
	fmt.Println("Shield confirmed:", transaction.Hash())
	return nil
}
