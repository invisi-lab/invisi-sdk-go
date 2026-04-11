package receiver

import (
	"math/big"
	"testing"
)

func TestStringIdToBigInt_BasicConversion(t *testing.T) {
	result := StringIdToBigInt("hello")
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.Sign() <= 0 {
		t.Error("result should be positive")
	}
}

func TestStringIdToBigInt_DifferentStringsProduceDifferentValues(t *testing.T) {
	a := StringIdToBigInt("invoice-001")
	b := StringIdToBigInt("invoice-002")
	if a.Cmp(b) == 0 {
		t.Error("different strings should produce different values")
	}
}

func TestStringIdToBigInt_SameStringProducesSameValue(t *testing.T) {
	a := StringIdToBigInt("refund-abc")
	b := StringIdToBigInt("refund-abc")
	if a.Cmp(b) != 0 {
		t.Errorf("same string should produce same value: %s vs %s", a, b)
	}
}

func TestStringIdToBigInt_ResultWithinFieldModulus(t *testing.T) {
	// BN254 scalar field modulus
	fieldModulus, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	// Test with a very long string to ensure modular reduction
	longStr := ""
	for i := 0; i < 100; i++ {
		longStr += "abcdefghij"
	}
	result := StringIdToBigInt(longStr)
	if result.Cmp(fieldModulus) >= 0 {
		t.Error("result should be less than field modulus")
	}
}

func TestStringIdToBigInt_EmptyString(t *testing.T) {
	result := StringIdToBigInt("")
	if result == nil {
		t.Fatal("result is nil")
	}
	// Empty string -> 0
	if result.Sign() != 0 {
		t.Errorf("empty string should produce 0, got %s", result)
	}
}

func TestStringIdToBigInt_MatchesTSSDK(t *testing.T) {
	// Verify the algorithm matches TS SDK:
	// for each char, result = (result << 8) | charCode
	// then result % fieldModulus
	//
	// For "A" (charCode 65):
	// result = 65
	result := StringIdToBigInt("A")
	if result.Cmp(big.NewInt(65)) != 0 {
		t.Errorf("StringIdToBigInt('A') = %s, want 65", result)
	}

	// For "AB" (charCodes 65, 66):
	// result = (65 << 8) | 66 = 16706
	result = StringIdToBigInt("AB")
	if result.Cmp(big.NewInt(16706)) != 0 {
		t.Errorf("StringIdToBigInt('AB') = %s, want 16706", result)
	}
}
