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
func (s *WalletSession) PrepareCreateCredit(ctx context.Context, p sdk.CreateCreditPreparation) (*Operation, error) {
	execution, e := s.runtimeSnapshot().PrepareCreateCredit(ctx, p)
	if e != nil {
		return nil, e
	}
	return operationFromExecution(s, execution), nil
}
func (s *WalletSession) PrepareDrawCredit(ctx context.Context, p sdk.DrawCreditPreparation) (*Operation, error) {
	execution, e := s.runtimeSnapshot().PrepareDrawCredit(ctx, p)
	if e != nil {
		return nil, e
	}
	return operationFromExecution(s, execution), nil
}
func (s *WalletSession) PrepareTopUpCredit(ctx context.Context, p sdk.TopUpCreditPreparation) (*Operation, error) {
	execution, e := s.runtimeSnapshot().PrepareTopUpCredit(ctx, p)
	if e != nil {
		return nil, e
	}
	return operationFromExecution(s, execution), nil
}
func (s *WalletSession) PrepareCancelCredit(ctx context.Context, p sdk.CancelCreditPreparation) (*Operation, error) {
	execution, e := s.runtimeSnapshot().PrepareCancelCredit(ctx, p)
	if e != nil {
		return nil, e
	}
	return operationFromExecution(s, execution), nil
}
