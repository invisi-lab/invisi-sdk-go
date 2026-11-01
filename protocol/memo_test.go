package protocol

import "testing"

func TestOpaqueDeliveryValidation(t *testing.T) {
	clue := make([]byte, 68)
	clue[64] = 24
	for _, credit := range []bool{false, true} {
		for _, length := range []int{1, 160, 498, 4096} {
			for _, header := range []byte{0, 1, 2, 129, 255} {
				ciphertext := make([]byte, length)
				ciphertext[0] = header
				if err := ValidateDelivery(Delivery{Ciphertext: ciphertext, Clue: clue}, credit, false); err != nil {
					t.Fatalf("opaque delivery %d/%d: %v", length, header, err)
				}
			}
		}
		for _, length := range []int{0, 4097} {
			if ValidateDelivery(Delivery{Ciphertext: make([]byte, length), Clue: clue}, credit, false) == nil {
				t.Fatal("invalid length accepted", length)
			}
		}
	}
}
