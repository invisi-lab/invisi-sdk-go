package sdk

import "github.com/invisi-lab/invisi-sdk-go/gas"

type GasReference = gas.Result
type GasReferenceShape = gas.Shape

func GetGasReference(shape gas.Shape) (gas.Result, bool) { return gas.Reference(shape) }
