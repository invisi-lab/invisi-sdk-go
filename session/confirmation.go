package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

// GetConfirmation observes the primary chain without claiming a worker or changing reservations.
func (o *Operation) GetConfirmation(ctx context.Context) (protocolrpc.TransactionConfirmation, error) {
	if err := o.session.protocol.ValidateDeployment(ctx, o.session.options.Provider); err != nil {
		return protocolrpc.TransactionConfirmation{}, err
	}
	state, err := o.session.options.Store.ReadSnapshot(ctx, o.session.scope)
	if err != nil {
		return protocolrpc.TransactionConfirmation{}, err
	}
	record, ok := state.Operations[o.plan.ID]
	if !ok {
		return protocolrpc.TransactionConfirmation{}, errors.New("operation is unavailable")
	}
	if record.TransactionHash == (common.Hash{}) {
		return protocolrpc.TransactionConfirmation{Level: "pending", Status: protocolrpc.FinalityAvailable}, nil
	}
	return protocolrpc.ReadTransactionConfirmation(ctx, o.session.options.Provider, record.TransactionHash), nil
}
