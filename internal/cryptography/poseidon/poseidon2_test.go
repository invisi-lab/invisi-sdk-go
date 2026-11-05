package poseidon

import (
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/internal/cryptography/field"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Test vectors from Noir circuits (authoritative). These must match
// Barretenberg (bb.js) output bit-for-bit.
func TestPoseidon2Hash_NoirVectors(t *testing.T) {
	tests := []struct {
		name   string
		inputs []uint64
		want   string
	}{
		{
			name:   "Hash(1,2)",
			inputs: []uint64{1, 2},
			want:   "0x038682aa1cb5ae4e0a3f13da432a95c77c5c111f6f030faf9cad641ce1ed7383",
		},
		{
			name:   "Hash(1,2,3,4)",
			inputs: []uint64{1, 2, 3, 4},
			want:   "0x130bf204a32cac1f0ace56c78b731aa3809f06df2731ebcf6b3464a15788b1b9",
		},
		{
			name:   "Hash(0,0)",
			inputs: []uint64{0, 0},
			want:   "0x0b63a53787021a4a962a452c2921b3663aff1ffd8d5510540f8e659e782956f1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			elems := make([]fr.Element, len(tc.inputs))
			for i, v := range tc.inputs {
				elems[i].SetUint64(v)
			}
			got, err := Poseidon2Hash(elems...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotHex := field.ElementToHex(got)
			if gotHex != tc.want {
				t.Errorf("Poseidon2Hash(%v)\n  got  %s\n  want %s", tc.inputs, gotHex, tc.want)
			}
		})
	}
}

// Test vector from bb.js snapshot: hash(4, 8).
func TestPoseidon2Hash_BBjsSnapshot(t *testing.T) {
	var a, b fr.Element
	a.SetUint64(4)
	b.SetUint64(8)

	got, err := Poseidon2Hash(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gotHex := field.ElementToHex(got)
	want := "0x2bcaeb6d58bb38baf753d58c3c96618fea82163345295eb40e88344eeb0ce2a1"
	if gotHex != want {
		t.Errorf("Poseidon2Hash(4, 8)\n  got  %s\n  want %s", gotHex, want)
	}
}

// Poseidon2Hash with empty input must return an error.
func TestPoseidon2Hash_EmptyInput(t *testing.T) {
	_, err := Poseidon2Hash()
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

// Determinism: same inputs always produce the same output.
func TestPoseidon2Hash_Deterministic(t *testing.T) {
	var a, b fr.Element
	a.SetUint64(42)
	b.SetUint64(99)

	h1, err := Poseidon2Hash(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := Poseidon2Hash(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 != h2 {
		t.Errorf("non-deterministic: %s != %s", field.ElementToHex(h1), field.ElementToHex(h2))
	}
}

// Different inputs must produce different outputs.
func TestPoseidon2Hash_DifferentInputs(t *testing.T) {
	var a, b fr.Element
	a.SetUint64(1)
	b.SetUint64(2)

	h1, err := Poseidon2Hash(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := Poseidon2Hash(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 == h2 {
		t.Error("different inputs produced same hash")
	}
}

// Single element input (rate=3 means it absorbs [x, 0, 0] then squeezes).
func TestPoseidon2Hash_SingleElement(t *testing.T) {
	var a fr.Element
	a.SetUint64(1)

	_, err := Poseidon2Hash(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Multi-rate boundary: exactly 3 elements (one full absorb, no leftover).
func TestPoseidon2Hash_ExactRate(t *testing.T) {
	var a, b, c fr.Element
	a.SetUint64(1)
	b.SetUint64(2)
	c.SetUint64(3)

	_, err := Poseidon2Hash(a, b, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Inputs spanning multiple rate blocks.
func TestPoseidon2Hash_MultiBlock(t *testing.T) {
	elems := make([]fr.Element, 7) // 3 + 3 + 1
	for i := range elems {
		elems[i].SetUint64(uint64(i + 1))
	}
	_, err := Poseidon2Hash(elems...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
