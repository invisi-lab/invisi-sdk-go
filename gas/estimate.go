package gas

import shared "github.com/invisi-lab/invisi-sdk-go/protocol/gas"

type Shape = shared.Shape
type Result = shared.EstimateResult

func EstimateGas(shape Shape) Result { return shared.EstimateGas(shape) }

func EstimateGasChecked(shape Shape) (Result, error) { return shared.EstimateGasChecked(shape) }
