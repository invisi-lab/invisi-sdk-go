package session

import (
	"context"

	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

func (s *WalletSession) PrepareTransfer(ctx context.Context, p sdk.TransferPreparation) (*Operation, error) {
	execution, e := s.runtimeSnapshot().PrepareTransfer(ctx, p)
	if e != nil {
		return nil, e
	}
	return operationFromExecution(s, execution), nil
}
func (s *WalletSession) PrepareShield(ctx context.Context, p sdk.ShieldPreparation) (*Operation, error) {
	execution, e := s.runtimeSnapshot().PrepareShield(ctx, p)
	if e != nil {
		return nil, e
	}
	return operationFromExecution(s, execution), nil
}
