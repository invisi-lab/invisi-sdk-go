package session

import "github.com/invisi-lab/invisi-sdk-go/protocol"

func (o *Operation) EstimateGas() (protocol.GasEstimate, error) { return o.prepared.EstimateGas() }
