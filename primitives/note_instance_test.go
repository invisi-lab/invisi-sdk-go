package primitives

import "testing"

func TestNoteHashBindsProtocolAddress(t *testing.T) {
	chain, protocol, owner, token := ElementFromUint64(31337), ElementFromUint64(0x1234), ElementFromUint64(1), ElementFromUint64(0x5678)
	amount, nonce := ElementFromUint64(70), ElementFromUint64(123)
	identity, err := ComputeNoteIdentity(owner, nonce)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := ComputeNoteCommitment(chain, protocol, identity, token, amount)
	if err != nil {
		t.Fatal(err)
	}
	want, err := Poseidon2Hash(ElementFromUint64(DNote), chain, protocol, identity, token, amount)
	if err != nil {
		t.Fatal(err)
	}
	if leaf != want {
		t.Fatal("Note commitment must bind Protocol address")
	}
	nullifier, err := ComputeNoteNullifier(chain, protocol, nonce, leaf)
	if err != nil {
		t.Fatal(err)
	}
	wantNullifier, err := Poseidon2Hash(ElementFromUint64(DNull), chain, protocol, nonce, leaf)
	if err != nil {
		t.Fatal(err)
	}
	if nullifier != wantNullifier {
		t.Fatal("Note nullifier must explicitly bind Protocol address")
	}
}
