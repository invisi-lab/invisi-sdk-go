package transaction

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/wallet"

	native "github.com/invisi-lab/invisi-sdk-go/internal/testsupport"
)

func WalletFromPrivateKey(key string) (*Wallet, error) {
	e, err := native.NewExecutor(key)
	if err != nil {
		return nil, err
	}
	return connectTestWallet(context.Background(), e.OwnerAddress(), e)
}

func connectTestWallet(ctx context.Context, owner common.Address, client wallet.Client) (*Wallet, error) {
	q, err := wallet.Prepare("test-public", owner, wallet.Public, struct{}{})
	if err != nil {
		return nil, err
	}
	responses, err := client.Execute(ctx, []wallet.Request{q})
	if err != nil {
		return nil, err
	}
	ordered, err := wallet.ValidateResponses([]wallet.Request{q}, responses)
	if err != nil {
		return nil, err
	}
	return WalletFromResponse(q, ordered[0])
}
