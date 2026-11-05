package transaction

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	gasmodel "github.com/invisi-lab/invisi-sdk-go/internal/gas"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

// EstimateGas returns only GasUnits, without proving, store access or network calls.
func (o *Execution) EstimateGas() (gasmodel.GasReference, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	request, err := o.GasRequest()
	if err != nil {
		return gasmodel.GasReference{}, err
	}
	profiles := o.Runtime.Options.GasBenchmarks
	if err := gasmodel.ValidateGasProfiles(profiles); err != nil {
		return gasmodel.GasReference{}, err
	}
	var exact gasmodel.GasReference
	for _, p := range profiles {
		if !gasmodel.SameGasDeployment(p.Deployment, request.Deployment) {
			continue
		}
		for _, m := range p.Measurements {
			if m.Shape == request.Shape && gasmodel.SameGasCircuit(m.Circuit, request.Circuit) {
				candidate, err := gasmodel.MeasuredGas(m)
				if err != nil {
					return gasmodel.GasReference{}, err
				}
				if candidate.GasUnits > exact.GasUnits {
					exact = candidate
				}
			}
		}
	}
	if exact.GasUnits > 0 {
		return exact, nil
	}
	return gasmodel.EstimateGasShape(request, gasmodel.DefaultGasProfiles)
}

// EstimateGas is available as soon as preparation completes.
func (p *PreparedTransaction) EstimateGas() (gasmodel.GasReference, error) {
	return p.operation.EstimateGas()
}

func (o *Execution) GasRequest() (gasmodel.GasEstimateRequest, error) {
	p := o.Plan
	shape := gasmodel.GasBenchmarkShape{Kind: p.Kind, Target: p.Target, Capacity: p.Capacity,
		NoteInputs: len(p.Inputs), NoteOutputs: len(p.Outputs), CreditInputs: len(p.CreditInputs), Funding: p.FundingMode}
	if shape.Target == "" {
		shape.Target = "evm"
	}
	if p.Kind == "transact" {
		shape.Kind = "transfer"
		if p.Params.WithdrawalAmount != nil && p.Params.WithdrawalAmount.Sign() > 0 {
			shape.Kind = "withdraw"
			if !p.NeedsAuthorization {
				shape.Kind = "sweep"
			}
		}
		shape.SourceTrees = len(p.Params.Sources)
	}
	if p.CreateCredit != nil {
		shape.SourceTrees = len(p.CreateCredit.Sources)
	}
	if p.TopUpCredit != nil {
		shape.SourceTrees = len(p.TopUpCredit.Creation.Sources) + 1
	}
	if p.CancelCredit != nil {
		shape.SourceTrees = 1
	}
	if p.CreditOutput != nil {
		shape.CreditOutputs = 1
	}
	if p.DrawCredit != nil {
		shape.Kind = "draw-credit"
		shape.SourceTrees = len(p.DrawCredit.Sources)
		for _, hash := range p.DrawCredit.RemainderCommitments {
			if hash != ([32]byte{}) {
				shape.CreditRemainders++
			}
		}
		shape.CreditOutputs = shape.CreditRemainders
	}
	shape.Permit = p.Funding.Permit.Deadline != nil && p.Funding.Permit.Deadline.Sign() != 0
	for _, d := range o.Deliveries() {
		shape.DeliveryBytes += len(d.Ciphertext) + len(d.Clue)
	}
	request := gasmodel.GasEstimateRequest{Shape: shape, Deployment: gasmodel.GasBenchmarkDeployment{
		ChainID: p.Context.ChainID.String(), ProtocolAddress: p.Context.ProtocolAddress.Hex(),
		TokenAddress: p.Context.TokenAddress.Hex(), RuntimeCodeHash: o.Runtime.Protocol.Deployment().Protocol.RuntimeCodeHash.Hex(),
	}}
	if p.Kind != "shield" {
		words, err := o.Words()
		if err != nil {
			return request, err
		}
		inputs := make([]common.Hash, len(words))
		for i := range words {
			inputs[i] = words[i]
		}
		statement, err := prover.Statement(o.CircuitName(), shape.Target, inputs)
		if err != nil {
			return request, fmt.Errorf("gas reference circuit: %w", err)
		}
		request.Circuit = &gasmodel.GasBenchmarkCircuit{CircuitName: statement.CircuitName, Target: statement.Target, ArtifactSHA256: statement.ArtifactSHA256, VKSHA256: statement.VKSHA256}
	}
	return request, nil
}
