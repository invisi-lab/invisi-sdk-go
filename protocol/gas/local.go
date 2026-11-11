package gas

import "strconv"

type EstimateResult struct {
	Source    string  `json:"source"`
	Available bool    `json:"available"`
	GasUnits  *uint64 `json:"gasUnits,omitempty"`
	Reason    string  `json:"reason,omitempty"`
	Reference *Result `json:"reference,omitempty"`
}

func EstimateGas(shape Shape) EstimateResult {
	row, ok := reference(shape)
	if !ok {
		return EstimateResult{Source: "benchmark", Reason: "shape-not-covered"}
	}
	n, err := strconv.ParseUint(row.GasUnits, 10, 64)
	if err != nil || n == 0 {
		return EstimateResult{Source: "benchmark", Reason: "invalid benchmark measurement"}
	}
	return EstimateResult{Source: "benchmark", Available: true, GasUnits: &n, Reference: &row}
}
