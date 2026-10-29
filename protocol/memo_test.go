package protocol

import "testing"

func TestVariableMemoEnvelopeValidation(t *testing.T) {
	for _, credit := range []bool{false, true} {
		minimum, version := 160, byte(2)
		if credit {
			minimum, version = 402, 4
		}
		clue := make([]byte, 68)
		clue[64] = 24
		for _, length := range []int{minimum, minimum + 1, minimum + 256} {
			ciphertext := make([]byte, length)
			ciphertext[0] = version
			if err := ValidateDelivery(Delivery{Ciphertext: ciphertext, Clue: clue}, credit, false); err != nil {
				t.Fatal(length, err)
			}
		}
		for _, length := range []int{0, minimum - 1, minimum + 257} {
			ciphertext := make([]byte, length)
			if length > 0 {
				ciphertext[0] = version
			}
			if err := ValidateDelivery(Delivery{Ciphertext: ciphertext, Clue: clue}, credit, false); err == nil {
				t.Fatal("invalid length accepted", length)
			}
		}
		ciphertext := make([]byte, minimum)
		ciphertext[0] = version - 1
		if err := ValidateDelivery(Delivery{Ciphertext: ciphertext, Clue: clue}, credit, false); err == nil {
			t.Fatal("retired version accepted")
		}
	}
}
