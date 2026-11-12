package transaction

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

// GetConfirmation observes inclusion/finality, independently of intended operation effects.
func (p *PreparedTransaction) GetConfirmation(ctx context.Context, hash common.Hash) (protocolrpc.TransactionConfirmation, error) {
	if err := p.operation.Runtime.Protocol.ValidateDeployment(ctx, p.operation.Runtime.Options.Provider); err != nil {
		return protocolrpc.TransactionConfirmation{}, err
	}
	return protocolrpc.ReadTransactionConfirmation(ctx, p.operation.Runtime.Options.Provider, hash), nil
}
