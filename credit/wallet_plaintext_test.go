package credit

import "testing"

func TestDecodeWalletPlaintextRejectsMalformed(t *testing.T) {
	if _, e := DecodeDeliveryPlaintext([]byte{1}, DeliveryContext{}, Owner, nil); e == nil {
		t.Fatal("accepted malformed plaintext")
	}
}
