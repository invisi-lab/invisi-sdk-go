package transaction

import (
	"context"

	native "github.com/invisi-lab/invisi-sdk-go/internal/testsupport"
)

func WalletFromPrivateKey(key string) (*Wallet, error) {
	e, err := native.NewExecutor(key)
	if err != nil {
		return nil, err
	}
	return ConnectWallet(context.Background(), e.OwnerAddress(), e)
}
