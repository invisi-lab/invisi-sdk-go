package credit

import (
	"encoding/json"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"testing"
)

func TestStatePersistenceIsCanonicalAndRejectsLegacyCredit(t *testing.T) {
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
	if err = json.Unmarshal([]byte(`{"chainId":"31337","amount":"100","spenderPayout":"1"}`), &restored); err == nil {
		t.Fatal("accepted prefunded Credit as current state")
	}
	var encoded map[string]any
	if err = json.Unmarshal(raw, &encoded); err != nil {
		t.Fatal(err)
	}
	encoded["fields"].(map[string]any)["creditSharedKey"] = core.ElementToBigInt(core.ElementFromUint64(66)).String() + "0"
	bad, _ := json.Marshal(encoded)
	if err = json.Unmarshal(bad, &restored); err != nil {
		t.Fatal(err)
	}
	changed, _ := restored.Commitment()
	if changed == before {
		t.Fatal("serialized grant key not committed")
	}
}
