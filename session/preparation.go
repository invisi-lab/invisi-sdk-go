package session

import (
	"context"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"

	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func (s *WalletSession) PrepareTransfer(ctx context.Context, p sdk.TransferPreparation) (*Operation, error) {
	return s.prepareWithSnapshot(ctx, func(r *transaction.Runtime) (*transaction.Execution, error) { return r.PrepareTransfer(ctx, p) })
}
func (s *WalletSession) PrepareShield(ctx context.Context, p sdk.ShieldPreparation) (*Operation, error) {
	return s.prepareWithSnapshot(ctx, func(r *transaction.Runtime) (*transaction.Execution, error) { return r.PrepareShield(ctx, p) })
}
