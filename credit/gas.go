package credit

import (
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/gas"
)

func (p *Prepared) EstimateGas() (protocol.GasEstimate, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	positive := func(n *big.Int) bool { return n != nil && n.Sign() > 0 }
	shape := gas.Shape{}
	creditDeliveries := 0
	switch {
	case p.intent.Revoke != nil:
		// 调用方在导出 calldata 时才确定；本地预估不能假定 owner 直接发送。
		shape = gas.Shape{"kind": "revoke-credit", "authorization": "unknown", "feeEnabled": positive(p.intent.Revoke.FeeAmount)}
	case p.intent.Create != nil:
		creditDeliveries = 2
		shape = gas.Shape{"kind": "create-credit", "feeEnabled": positive(p.intent.Create.FeeAmount)}
	case p.intent.Update != nil:
		creditDeliveries = 2
		shape = gas.Shape{"kind": "update-credit", "feeEnabled": positive(p.intent.Update.FeeAmount)}
	case p.intent.Draw != nil:
		d := p.intent.Draw
		capacity, err := DrawCapacity(len(d.Nullifiers))
		if err != nil {
			return protocol.GasEstimate{}, err
		}
		creditDeliveries = 2 * len(d.Nullifiers)
		shape = gas.Shape{"kind": "draw-credit", "capacity": capacity, "creditInputs": len(d.Nullifiers), "distinctPayers": len(d.Payers), "sourceTrees": len(d.Sources), "feeEnabled": positive(d.FeeAmount)}
	default:
		return gas.EstimateGasChecked(nil)
	}
	extra := 0
	for i, delivery := range p.intent.Deliveries {
		baseline := 160
		if i < creditDeliveries {
			baseline = 485
		}
		extra += max(0, len(delivery.Ciphertext)-baseline) + max(0, len(delivery.Clue)-68)
	}
	shape["extraDeliveryBytes"] = extra
	return gas.EstimateGasChecked(shape)
}
