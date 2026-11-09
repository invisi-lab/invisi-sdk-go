package core

import (
	"bytes"
	"testing"
)

func TestApplicationCodecsAreExplicitAndMemoOnly(t *testing.T) {
	for _, codec := range []payloadCodecV1{FundsV1} {
		for _, memo := range [][]byte{nil, []byte("invoice"), {0xff, 0, 0xfe}, bytes.Repeat([]byte{1}, 256)} {
			payload, err := codec.Encode(PayloadDataV1{Memo: memo})
			if err != nil {
				t.Fatal(err)
			}
			if len(payload) != 1+len(memo) || payload[0] != 1 {
				t.Fatal("invalid layout")
			}
			result, err := codec.Decode(payload)
			if err != nil || !bytes.Equal(result.Memo, memo) {
				t.Fatal("memo mismatch", err)
			}
			if len(result.Memo) > 0 {
				result.Memo[0] ^= 0xff
				if !bytes.Equal(payload[1:], memo) {
					t.Fatal("decode aliases input")
				}
			}
		}
		for _, payload := range [][]byte{nil, {0}, {99, 1}, append([]byte{1}, make([]byte, 257)...)} {
			if _, err := codec.Decode(payload); err == nil {
				t.Fatal("invalid app payload accepted")
			}
		}
		if _, err := codec.Encode(PayloadDataV1{Memo: make([]byte, 257)}); err == nil {
			t.Fatal("oversize memo accepted")
		}
	}
}
func TestUnknownApplicationPayloadIsOpaque(t *testing.T) {
	note, private, public, context := memoNote(t)
	note.Payload = append([]byte{99, 0xff}, bytes.Repeat([]byte{0xa5}, 1000)...)
	ciphertext, err := EncryptPrivateTransferNote(public, note, context.NoteEncryptionContext)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := DecryptPrivateTransferNote(private, ciphertext, context)
	if err != nil || !bytes.Equal(opened.Payload, note.Payload) {
		t.Fatal("opaque payload changed", err)
	}
	if _, err = FundsV1.Decode(opened.Payload); err == nil {
		t.Fatal("unknown app version accepted")
	}
	for _, header := range []byte{0, 2, 5, 0x80, 0x82} {
		bad := bytes.Clone(ciphertext)
		bad[0] = header
		if _, err = DecryptPrivateTransferNote(private, bad, context); err == nil {
			t.Fatal("retired/unknown envelope accepted")
		}
	}
}
