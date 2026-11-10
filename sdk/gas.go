package sdk

import (
	"github.com/invisi-lab/invisi-sdk-go/gas"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

type GasReference = gas.Result
type GasReferenceShape = gas.Shape

func GetGasReference(shape gas.Shape) (gas.Result, bool) { return gas.Reference(shape) }

type GasEstimate = protocol.GasEstimate
type GasEstimator = protocol.GasEstimator
