package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

// Recover reopens an operation after a crash. A timeout never releases its inputs.
func Recover(ctx context.Context, session *sdk.WalletSession, operationID string, transactionHash common.Hash) error {
	operation, err := session.Resume(ctx, operationID)
	if err != nil {
		return err
	}
	// An application that broadcast outside the session supplies its persisted hash.
	if transactionHash != (common.Hash{}) {
		operation, err = operation.TrackTransaction(ctx, transactionHash)
		if err != nil {
			return err
		}
	}
	result, err := operation.Reconcile(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Recovered status:", result.Status())
	return nil
}
