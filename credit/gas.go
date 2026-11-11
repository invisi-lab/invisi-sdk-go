package credit

import (
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/gas"
)

func (p *Prepared) EstimateGas() (protocol.GasEstimate, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.intent.Revoke != nil {
		return gas.EstimateGas(nil), nil
	}
	if p.intent.Create != nil || p.intent.Update != nil {
		if len(p.intent.Deliveries) != 2 {
			return gas.EstimateGas(nil), nil
		}
		for _, d := range p.intent.Deliveries {
			if len(d.Ciphertext) != 485 || len(d.Clue) != 68 {
				return gas.EstimateGas(nil), nil
			}
		}
		kind, scenario := "create-credit", "first-allocation-empty-tree"
		fee := p.intent.Create != nil && p.intent.Create.FeeAmount != nil && p.intent.Create.FeeAmount.Sign() > 0
		if p.intent.Update != nil {
			kind, scenario = "update-credit", "after-draw-preserve-usage"
			fee = p.intent.Update.FeeAmount != nil && p.intent.Update.FeeAmount.Sign() > 0
		}
		mechanism := "none"
		if fee {
			mechanism = "owner-pull"
		}
		return gas.EstimateGas(gas.Shape{"kind": kind, "scenario": scenario, "feeEnabled": fee, "feeMechanism": mechanism}), nil
	}
	d := p.intent.Draw
	if d == nil {
		return gas.EstimateGas(nil), nil
	}
	capacity, err := DrawCapacity(len(d.Nullifiers))
	if err != nil {
		return protocol.GasEstimate{}, err
	}
	shape := gas.Shape{"kind": "draw-credit", "capacity": capacity, "creditInputs": len(d.Nullifiers), "distinctPayers": len(d.Payers), "sourceTrees": len(d.Sources), "noteOutputs": 1, "creditOutputs": len(d.CreditCommitments), "creditCiphertextBytes": 485, "noteCiphertextBytes": 160, "clueBytes": 68, "scenario": "distinct-payers", "feeEnabled": false, "feeMechanism": "none"}
	if d.FeeAmount != nil && d.FeeAmount.Sign() > 0 {
		shape["feeEnabled"] = true
		shape["feeMechanism"] = "draw-deduction"
	}
	if len(d.Payers) != len(d.Nullifiers) {
		return gas.EstimateGas(nil), nil
	}
	for _, delivery := range p.intent.Deliveries {
		if len(delivery.Ciphertext) != 485 && len(delivery.Ciphertext) != 160 || len(delivery.Clue) != 68 {
			return gas.EstimateGas(nil), nil
		}
	}
	return gas.EstimateGas(shape), nil
}
