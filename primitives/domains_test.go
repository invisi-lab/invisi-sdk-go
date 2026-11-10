package primitives

import "testing"

func TestCurrentDomainConstants(t *testing.T) {
	got := []uint64{DIdent, DNote, DNull}
	want := []uint64{2, 3, 4}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tag %d: got %d want %d", i, got[i], want[i])
		}
	}
	if TreeDepth != 16 {
		t.Fatalf("TreeDepth=%d", TreeDepth)
	}
}
