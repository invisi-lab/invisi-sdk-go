package credit

import (
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func TestOptionalUpdatePreservesAndClearsPayload(t *testing.T) {
	current, _ := fixture(t)
	current.Payload = []byte("keep")
	patch := UpdatePatch{}
	change := patch.Change(current)
	next, err := current.Update(change, 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	if next.State.Policy != current.Policy || next.State.Expiry != current.Expiry || string(next.State.Payload) != "keep" {
		t.Fatal("omitted fields changed")
	}
	empty := []byte{}
	zero := primitives.ElementFromUint64(0)
	change = (UpdatePatch{Payload: empty, SingleLimit: &zero}).Change(current)
	next, err = current.Update(change, 2000000000)
	if err != nil || len(next.State.Payload) != 0 || !next.State.SingleLimit.IsZero() {
		t.Fatal("explicit zero/empty not applied", err)
	}
}
