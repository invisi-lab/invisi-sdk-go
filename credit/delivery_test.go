package credit

import (
	hexcodec "encoding/hex"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"strings"
	"testing"
)

func bytesHex(t *testing.T, s string) []byte {
	t.Helper()
	b, e := hexcodec.DecodeString(strings.TrimPrefix(s, "0x"))
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestCanonicalDeferredDelivery(t *testing.T) {
	s, f := fixture(t)
	d := f["delivery"].(map[string]any)
	commitment, e := s.Commitment()
	if e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"owner", "spender"} {
		role := Owner
		if name == "spender" {
			role = Spender
		}
		owner := s.OwnerWalletIdentity
		if role == Spender {
			owner = s.SpenderWalletIdentity
		}
		context := DeliveryContext{ProtocolContext: s.ProtocolContext, TokenAddress: s.TokenAddress, Commitment: commitment, RecipientIdentity: owner}
		cipher := bytesHex(t, d[name+"Ciphertext"].(string))
		private := bytesHex(t, d[name+"PrivateKey"].(string))
		if len(cipher) != 484 || cipher[0] != 0x82 {
			t.Fatal("wrong canonical size")
		}
		got, e := Decrypt(private, cipher, context, role)
		if e != nil {
			t.Fatal(e)
		}
		if c, e := got.State.Commitment(); e != nil || c != commitment {
			t.Fatal("state mismatch", e)
		}
		roundtrip, e := Encrypt(got, context, role)
		if e != nil {
			t.Fatal(e)
		}
		if _, e := Decrypt(private, roundtrip, context, role); e != nil {
			t.Fatal(e)
		}
		cipher[0] = 0x81
		if _, e := Decrypt(private, cipher, context, role); e == nil {
			t.Fatal("accepted retired envelope")
		}
		context.RecipientIdentity = core.ElementFromUint64(999)
		if _, e := Decrypt(private, roundtrip, context, role); e == nil {
			t.Fatal("accepted wrong identity")
		}
	}
}
