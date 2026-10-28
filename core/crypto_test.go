package core

import "testing"

func TestOrdinaryNullifierBindsReceivingAddress(t *testing.T) {
	e := ElementFromUint64
	first, err := ComputeNoteNullifier(e(31337), e(20), e(30), e(40), e(50), e(60))
	if err != nil {
		t.Fatal(err)
	}
	second, err := ComputeNoteNullifier(e(31337), e(20), e(31), e(40), e(50), e(60))
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("ordinary nullifier did not bind receiving address")
	}
}
