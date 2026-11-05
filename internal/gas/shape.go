package gas

import (
	"errors"
	"math/big"
	"strings"
)

func SameGasCircuit(a, b *GasBenchmarkCircuit) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.CircuitName == b.CircuitName && a.Target == b.Target &&
		strings.EqualFold(a.ArtifactSHA256, b.ArtifactSHA256) && strings.EqualFold(a.VKSHA256, b.VKSHA256)
}

func SameGasDeployment(a, b GasBenchmarkDeployment) bool {
	return a.ChainID == b.ChainID && strings.EqualFold(a.ProtocolAddress, b.ProtocolAddress) &&
		strings.EqualFold(a.TokenAddress, b.TokenAddress) && strings.EqualFold(a.RuntimeCodeHash, b.RuntimeCodeHash)
}

func MeasuredGas(m GasBenchmarkMeasurement) (GasReference, error) {
	total := new(big.Int)
	for _, sample := range m.Samples {
		n, _ := new(big.Int).SetString(sample.GasUsed, 10)
		total.Add(total, n)
	}
	count := big.NewInt(int64(len(m.Samples)))
	total.Add(total, new(big.Int).Sub(count, big.NewInt(1)))
	return GasResult(total.Quo(total, count))
}

func GasResult(units *big.Int) (GasReference, error) {
	if !units.IsUint64() || units.Sign() <= 0 {
		return GasReference{}, errors.New("gas reference exceeds uint64")
	}
	return GasReference{GasUnits: units.Uint64()}, nil
}

func GasDimensions(s GasBenchmarkShape, capacity bool) []int {
	values := []int{s.NoteInputs, s.NoteOutputs, s.SourceTrees, s.CreditInputs, s.CreditOutputs, s.CreditRemainders, s.DeliveryBytes}
	if capacity {
		values = append(values, s.Capacity)
	}
	return values
}

// EstimateGasShape mirrors the TS upward-only workload envelope. It is intentionally
// coarse for unmeasured combinations and is not a calibrated bound on actual gas.
func EstimateGasShape(request GasEstimateRequest, profiles []GasBenchmarkProfile) (GasReference, error) {
	if request.Shape.Target == "noir-recursive" {
		return GasReference{}, errors.New("estimate the complete aggregate, not an individual recursive member")
	}
	var matching, family []GasBenchmarkMeasurement
	var exact GasReference
	for _, p := range profiles {
		for _, m := range p.Measurements {
			if m.Shape.Kind != request.Shape.Kind || m.Shape.Target != request.Shape.Target {
				continue
			}
			family = append(family, m)
			if SameGasCircuit(m.Circuit, request.Circuit) {
				matching = append(matching, m)
				if m.Shape == request.Shape {
					candidate, err := MeasuredGas(m)
					if err != nil {
						return GasReference{}, err
					}
					if candidate.GasUnits > exact.GasUnits {
						exact = candidate
					}
				}
			}
		}
	}
	if exact.GasUnits > 0 {
		return exact, nil
	}
	candidates := matching
	if len(candidates) == 0 && strings.Contains(request.Shape.Kind, "credit") {
		return GasReference{}, errors.New("no gas reference for this Credit circuit identity")
	}
	if len(candidates) == 0 {
		candidates = family
	}
	if len(candidates) == 0 {
		return GasReference{}, errors.New("no gas reference for this operation kind and target")
	}
	// Policy variants share a public workload. Combine their observed samples
	// before scaling so a cheaper variant cannot lower a modeled reference.
	var baselines []GasBenchmarkMeasurement
	for _, candidate := range candidates {
		found := false
		for i := range baselines {
			if baselines[i].Shape == candidate.Shape && SameGasCircuit(baselines[i].Circuit, candidate.Circuit) {
				baselines[i].Samples = append(baselines[i].Samples, candidate.Samples...)
				found = true
				break
			}
		}
		if !found {
			candidate.Samples = append([]GasBenchmarkSample{}, candidate.Samples...)
			baselines = append(baselines, candidate)
		}
	}
	var minimum *big.Int
	for _, m := range baselines {
		numerator, denominator := big.NewInt(1), big.NewInt(1)
		current, baseline := GasDimensions(request.Shape, len(matching) == 0), GasDimensions(m.Shape, len(matching) == 0)
		for i := range current {
			n := new(big.Int).Add(big.NewInt(int64(current[i])), big.NewInt(1))
			d := new(big.Int).Add(big.NewInt(int64(baseline[i])), big.NewInt(1))
			if new(big.Int).Mul(n, denominator).Cmp(new(big.Int).Mul(numerator, d)) > 0 {
				numerator, denominator = n, d
			}
		}
		units := new(big.Int)
		for _, sample := range m.Samples {
			n, _ := new(big.Int).SetString(sample.GasUsed, 10)
			if n.Cmp(units) > 0 {
				units.Set(n)
			}
		}
		units.Mul(units, numerator)
		units.Add(units, new(big.Int).Sub(denominator, big.NewInt(1)))
		units.Quo(units, denominator)
		if request.Shape.Funding == "signature" && m.Shape.Funding != "signature" {
			units.Mul(units, big.NewInt(2))
		}
		if request.Shape.Permit && !m.Shape.Permit {
			units.Mul(units, big.NewInt(2))
		}
		if minimum == nil || units.Cmp(minimum) < 0 {
			minimum = units
		}
	}
	return GasResult(minimum)
}
