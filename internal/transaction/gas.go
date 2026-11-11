package transaction

import (
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/gas"
	"math/big"
)

func (p *PreparedTransaction) EstimateGas() (protocol.GasEstimate, error) {
	return EstimatePlanGas(p.operation.Plan), nil
}
func EstimatePlanGas(p Plan) protocol.GasEstimate {
	positive := func(n *big.Int) bool { return n != nil && n.Sign() > 0 }
	funding, withdrawal, fee := positive(p.Funding.Amount), positive(p.Params.WithdrawalAmount), positive(p.Params.FeeAmount)
	scenario := "mixed"
	if len(p.Inputs) == 0 {
		scenario = "public-funding"
	} else if len(p.Outputs) == 0 && !funding && !fee {
		scenario = "withdrawal"
	} else if !funding && !withdrawal && !fee {
		scenario = "private"
	}
	shape := gas.Shape{"kind": "transact", "capacity": p.Capacity, "noteInputs": len(p.Inputs), "noteOutputs": len(p.Outputs), "sourceTrees": len(p.Params.Sources), "noteCiphertextBytes": 160, "clueBytes": 68, "scenario": scenario}
	for _, output := range p.Outputs {
		d := output.Delivery
		if len(d.Ciphertext) != 160 || len(d.Clue) != 68 {
			return gas.EstimateGas(nil)
		}
	}
	if p.Kind != "transact" || p.Target != "evm" {
		return gas.EstimateGas(nil)
	}
	return gas.EstimateGas(shape)
}
