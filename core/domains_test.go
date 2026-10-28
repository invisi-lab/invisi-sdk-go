package core

import "testing"

func TestCurrentDomainConstants(t *testing.T) {
	got := []uint64{DNoteSecret, DIdent, DNote, DNull, DCredit, DCNull, DCreditKeyHash, DCreditRecovery, DCreditRemainderNonce}
	want := []uint64{1, 2, 3, 4, 5, 6, 7, 9, 10}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tag %d: got %d want %d", i, got[i], want[i])
		}
	}
	if TreeDepth != 16 {
		t.Fatalf("TreeDepth=%d", TreeDepth)
	}
}
