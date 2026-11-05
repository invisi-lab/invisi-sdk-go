package sdk

import gasmodel "github.com/invisi-lab/invisi-sdk-go/internal/gas"

func DefineGasBenchmarks(data []byte) ([]GasBenchmarkProfile, error) {
	return gasmodel.DefineGasBenchmarks(data)
}

type GasBenchmarkCircuit = gasmodel.GasBenchmarkCircuit

type GasBenchmarkDeployment = gasmodel.GasBenchmarkDeployment

type GasBenchmarkFingerprints = gasmodel.GasBenchmarkFingerprints

type GasBenchmarkMeasurement = gasmodel.GasBenchmarkMeasurement

type GasBenchmarkProfile = gasmodel.GasBenchmarkProfile

type GasBenchmarkRuntimeContract = gasmodel.GasBenchmarkRuntimeContract

type GasBenchmarkSample = gasmodel.GasBenchmarkSample

type GasBenchmarkShape = gasmodel.GasBenchmarkShape

type GasBenchmarkSource = gasmodel.GasBenchmarkSource

type GasBenchmarkToolchain = gasmodel.GasBenchmarkToolchain

type GasReference = gasmodel.GasReference
