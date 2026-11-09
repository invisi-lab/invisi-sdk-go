package sdk

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/wallet"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

func NewViewingCapability(identity common.Address, key []byte, public []byte) (*ViewingCapability, error) {
	return transaction.NewViewingCapability(identity, key, public)
}

type NullifierCapability = transaction.NullifierCapability

type ViewingCapability = transaction.ViewingCapability

type Wallet = transaction.Wallet

type WalletReference = transaction.WalletReference

func ConnectWallet(ctx context.Context, owner common.Address, executor wallet.Executor) (*Wallet, error) {
	return transaction.ConnectWallet(ctx, owner, executor)
}

func WalletFromAddress(address string) (*Wallet, error) {
	return transaction.WalletFromAddress(address)
}

func WalletFromResponse(request wallet.Request, response wallet.Response) (*Wallet, error) {
	return transaction.WalletFromResponse(request, response)
}
