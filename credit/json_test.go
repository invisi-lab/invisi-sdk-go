package credit

import (
	"encoding/json"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func TestStatePersistenceIsCanonicalAndBindsGrantKey(t *testing.T) {
	s, _ := fixture(t)
	s.Payload = []byte{1, 2, 3}
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var restored State
	if err = json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	before, _ := s.Commitment()
	after, _ := restored.Commitment()
	if before != after || string(s.Payload) != string(restored.Payload) {
		t.Fatal("persistence changed grant")
	}
	var encoded map[string]any
	if err = json.Unmarshal(raw, &encoded); err != nil {
		t.Fatal(err)
	}
	encoded["fields"].(map[string]any)["creditSharedKey"] = primitives.ElementToBigInt(primitives.ElementFromUint64(66)).String() + "0"
	bad, _ := json.Marshal(encoded)
	if err = json.Unmarshal(bad, &restored); err != nil {
		t.Fatal(err)
	}
	changed, _ := restored.Commitment()
	if changed == before {
		t.Fatal("serialized grant key not committed")
	}
}

func TestStateJSONRejectsInvalidSchemaAndFields(t *testing.T) {
	state, _ := fixture(t)
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []string{"format", "missing", "decimal", "boolean", "missingBoolean", "unlimitedBalance"} {
		t.Run(change, func(t *testing.T) {
			var encoded map[string]any
			if err := json.Unmarshal(raw, &encoded); err != nil {
				t.Fatal(err)
			}
			fields := encoded["fields"].(map[string]any)
			switch change {
			case "format":
				encoded["version"] = 999
			case "missing":
				delete(fields, "creditSharedKey")
			case "boolean":
				encoded["unlimited"] = 1
			case "missingBoolean":
				delete(encoded, "unlimited")
			case "unlimitedBalance":
				encoded["unlimited"] = true
			case "decimal":
				fields["creditSharedKey"] = "01"
			}
			invalid, err := json.Marshal(encoded)
			if err != nil {
				t.Fatal(err)
			}
			var decoded State
			if err := json.Unmarshal(invalid, &decoded); err == nil {
				t.Fatal("accepted invalid state JSON")
			}
		})
	}
}
