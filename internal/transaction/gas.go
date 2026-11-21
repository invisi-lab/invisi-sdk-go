package transaction

import (
	"github.com/invisi-labs/invisi-sdk-go/rpc/broadcaster"
	"math/big"

	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/protocol/gas"
)

func (p *PreparedTransaction) EstimateGas() (protocol.GasEstimate, error) {
	return EstimatePlanGasChecked(p.operation.Plan)
}

func EstimatePlanGas(p Plan) protocol.GasEstimate {
	result, err := EstimatePlanGasChecked(p)
	if err != nil {
		return gas.EstimateGas(nil)
	}
	return result
}

func EstimatePlanGasChecked(p Plan) (protocol.GasEstimate, error) {
	positive := func(n *big.Int) bool { return n != nil && n.Sign() > 0 }
	if p.Kind != "shield" && p.Kind != "transact" {
		return gas.EstimateGasChecked(nil)
	}
	if p.Kind == "transact" && (p.Target != "" && p.Target != "evm" || p.Broadcaster != nil && p.Broadcaster.Signed.Quote.Execution == "aggregate") {
		return gas.EstimateGasChecked(gas.Shape{"kind": "aggregate"})
	}
	shape := gas.Shape{"kind": "shield", "noteOutputs": len(p.Outputs), "feeEnabled": positive(p.Shield.FeeAmount)}
	if p.Kind == "transact" {
		shape = gas.Shape{"kind": "transact", "capacity": p.Capacity, "noteInputs": len(p.Inputs), "noteOutputs": len(p.Outputs), "sourceTrees": len(p.Params.Sources), "deposit": positive(p.Funding.Amount), "withdrawal": positive(p.Params.WithdrawalAmount), "feeEnabled": positive(p.Params.FeeAmount)}
	}
	extra := 0
	for _, output := range p.Outputs {
		extra += max(0, len(output.Delivery.Ciphertext)-160) + max(0, len(output.Delivery.Clue)-68)
	}
	shape["extraDeliveryBytes"] = extra
	shape["permit"] = positive(p.Funding.Permit.Deadline)
	shape["fundingSignature"] = p.FundingMode == "signature"
	return gas.EstimateGasChecked(shape)
}

// EstimateRouterGas adds the local Router model to this prepared Core preview.
// Caller-supplied CoreGasUnits is replaced by the prepared transaction estimate.
func (p *PreparedTransaction) EstimateRouterGas(input broadcaster.RouterGasInput) (broadcaster.RouterGasResult, error) {
	core, e := p.EstimateGas()
	if e != nil {
		return broadcaster.RouterGasResult{}, e
	}
	if !core.Available || core.GasUnits == nil {
		return broadcaster.RouterGasResult{Source: "model", ModelID: broadcaster.RouterGasModelID, Reason: "core-gas-unavailable"}, nil
	}
	input.CoreGasUnits = *core.GasUnits
	return broadcaster.EstimateRouterGas(input), nil
}
