package broadcaster

import (
	"encoding/json"
	"testing"
)

func TestOptionalConfirmationStrictWire(t *testing.T) {
	block := `{"number":"10","hash":"0x1111111111111111111111111111111111111111111111111111111111111111","timestamp":10}`
	for _, raw := range []string{`{"level":"pending","status":"available","block":null}`, `{"level":"included","status":"unsupported","block":null}`, `{"level":"finalized","status":"available","block":` + block + `}`} {
		var c Confirmation
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{`{"level":"finalized","status":"available","block":null}`, `{"level":"pending","status":"allowed","block":null}`, `{"level":"pending","status":"available"}`, `{"level":"safe","status":"available","block":{"number":"01"}}`} {
		var c Confirmation
		if err := json.Unmarshal([]byte(raw), &c); err == nil {
			t.Fatal("invalid relay claim accepted", raw)
		}
	}
}
