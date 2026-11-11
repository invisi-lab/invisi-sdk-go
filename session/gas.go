package session

import (
	"github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func (o *Operation) EstimateGas() (protocol.GasEstimate, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return transaction.EstimatePlanGas(o.plan), nil
}
