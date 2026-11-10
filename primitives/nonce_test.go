package primitives

import (
	"math/big"
	"testing"
)

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatal(err)
	}
	b := ElementToBigInt(nonce)
	if b.Sign() < 0 {
		t.Error("nonce is negative")
	}
	if b.Cmp(FieldModulus()) >= 0 {
		t.Error("nonce >= FIELD_MODULUS")
	}
}

func TestGenerateNonceUnique(t *testing.T) {
	n1, _ := GenerateNonce()
	n2, _ := GenerateNonce()
	if n1.Equal(&n2) {
		t.Error("two consecutive nonces should not be equal")
	}
}

func TestFieldModulusValue(t *testing.T) {
	expected, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	if FieldModulus().Cmp(expected) != 0 {
		t.Errorf("FieldModulus mismatch")
	}
}

func TestFieldModulusReturnsIndependentCopy(t *testing.T) {
	value := FieldModulus()
	value.SetUint64(1)
	if FieldModulus().BitLen() != 254 {
		t.Fatal("modulus mutated through exported value")
	}
}
