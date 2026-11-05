package transaction

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func (o *Execution) EncodeCredit(proof []byte, caller *common.Address) (*protocol.Calldata, error) {
	switch o.Plan.Kind {
	case "create-credit":
		return protocol.BuildCreateCreditCalldata(protocol.BuildCreateCreditCalldataParams{Context: o.Plan.Context, Proof: proof, Params: *o.Plan.CreateCredit, Deliveries: o.Deliveries(), Funding: o.Plan.Funding, Caller: caller})
	case "credit-draw":
		return protocol.BuildCreditDrawCalldata(protocol.BuildCreditDrawCalldataParams{Context: o.Plan.Context, Proof: proof, Params: *o.Plan.DrawCredit, Deliveries: o.Deliveries()})
	case "cancel-credit":
		return protocol.BuildCancelCreditCalldata(protocol.BuildCancelCreditCalldataParams{Context: o.Plan.Context, Proof: proof, Params: *o.Plan.CancelCredit})
	case "top-up-credit":
		return protocol.BuildTopUpCreditCalldata(protocol.BuildTopUpCreditCalldataParams{Context: o.Plan.Context, Proof: proof, Params: *o.Plan.TopUpCredit, Deliveries: o.Deliveries(), Funding: o.Plan.Funding, Caller: caller})
	}
	return nil, errors.New("unsupported Credit operation")
}
