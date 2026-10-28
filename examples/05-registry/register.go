package registry

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

// Register selects the registry from the same trusted deployment used by the Protocol.
func Register(ctx context.Context, pool *sdk.Protocol, provider *ethclient.Client, wallet *sdk.Wallet, backend sdk.Prover, signer *bind.TransactOpts) error {
	registry, err := sdk.NewAddressRegistry(pool, provider)
	if err != nil {
		return err
	}
	fmt.Println("1. Prepare receiving-key registration using the registry's own nonce")
	prepared, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: wallet, Prover: backend})
	if err != nil {
		return err
	}
	fmt.Println("2. Review registration locally:", prepared.Preview())
	authorized, err := prepared.Authorize(ctx)
	if err != nil {
		return err
	}
	proved, err := authorized.Prove(ctx, prover.Options{OnProgress: func(event prover.Progress) { fmt.Println("Proof phase:", event.Phase) }})
	if err != nil {
		return err
	}
	call, err := proved.ExportTransaction(ctx)
	if err != nil {
		return err
	}
	options := *signer
	options.Context = ctx
	contract := bind.NewBoundContract(call.To, abi.ABI{}, provider, provider, provider)
	transaction, err := contract.RawTransact(&options, call.Data)
	if err != nil {
		return err
	}
	tracked, err := proved.TrackTransaction(ctx, transaction.Hash())
	if err != nil {
		return err
	}
	if _, err = bind.WaitMined(ctx, provider, transaction); err != nil {
		return err
	}
	confirmed, err := tracked.Reconcile(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Canonical registration status:", confirmed.Status())
	address, err := registry.Resolve(ctx, wallet.OwnerAddress().Hex())
	if err != nil {
		return err
	}
	fmt.Println("4. Resolved long receiving address:", address)
	return nil
}
