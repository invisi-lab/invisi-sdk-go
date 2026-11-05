package sdk

import (
	"github.com/ethereum/go-ethereum/common"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type Detector = transaction.Detector

func NewViewingCapability(identity common.Address, key []byte, public []byte) (*ViewingCapability, error) {
	return transaction.NewViewingCapability(identity, key, public)
}

type NullifierCapability = transaction.NullifierCapability

type PublicWallet = transaction.PublicWallet

func PublicWalletFromAddress(address string) (*PublicWallet, error) {
	return transaction.PublicWalletFromAddress(address)
}

type ViewingCapability = transaction.ViewingCapability

type Wallet = transaction.Wallet

type WalletAuthorization = transaction.WalletAuthorization

type WalletAuthorizationProvider = transaction.WalletAuthorizationProvider

func WalletFromPrivateKey(privateKey string) (*Wallet, error) {
	return transaction.WalletFromPrivateKey(privateKey)
}

type WalletReference = transaction.WalletReference
