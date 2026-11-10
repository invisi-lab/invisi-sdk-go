package primitives

import "testing"

func TestDecodeWalletPlaintextRejectsMalformed(t *testing.T) {
	if _, e := DecodeNotePlaintext([]byte{1}, NoteDecryptionContext{}); e == nil {
		t.Fatal("accepted malformed plaintext")
	}
}
