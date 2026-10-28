package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
	"math/big"
)

// Transfer demonstrates managed input selection, reservations and transaction tracking.
// Create a session with Store and geth Signer as shown in main.go.
func Transfer(ctx context.Context, session *sdk.WalletSession, provider *ethclient.Client, recipient string) error {
	tokenAddress := common.HexToAddress("0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6")
	prepared, err := session.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: tokenAddress, Outputs: []sdk.PaymentOutput{{To: recipient, Amount: big.NewInt(30)}}})
	if err != nil {
		return err
	}
	fmt.Println("Review:", prepared.Preview())
	reference, err := prepared.EstimateGas()
	if err != nil {
		return err
	}
	fmt.Println("Before proving: reference gas units", reference.GasUnits)
	authorized, err := prepared.Authorize(ctx)
	if err != nil {
		return err
	}
	ready, err := authorized.Prove(ctx, prover.Options{OnProgress: func(event prover.Progress) { fmt.Println("Proof phase:", event.Phase) }})
	if err != nil {
		return err
	}
	// The thin session delegate uses geth's estimation, nonce selection and sending.
	transaction, err := session.SendTransaction(ctx, ready)
	if err != nil {
		return err
	}
	fmt.Println("Transaction hash:", transaction.Hash())
	if _, err = bind.WaitMined(ctx, provider, transaction); err != nil {
		return err
	}
	tracked, err := session.Resume(ctx, ready.ID())
	if err != nil {
		return err
	}
	result, err := tracked.Reconcile(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Canonical operation status:", result.Status())
	return nil
}
