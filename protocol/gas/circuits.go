package gas

import (
	_ "embed"
	"encoding/json"
)

//go:embed data/circuits.json
var circuitJSON []byte
var Circuits = func() map[string]GasBenchmarkCircuit {
	var out map[string]GasBenchmarkCircuit
	if err := json.Unmarshal(circuitJSON, &out); err != nil {
		panic(err)
	}
	return out
}()
